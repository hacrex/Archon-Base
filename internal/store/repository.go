// Package store contains persistence adapters for the Archon control plane.
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hacrex/Archon-Base/internal/api"
)

var ErrNotFound = errors.New("resource not found")
var ErrProjectNotEmpty = errors.New("project has active database instances")

type Repository struct {
	db             *sql.DB
	organizationID string
}

func NewRepository(db *sql.DB, organizationID string) (*Repository, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	if strings.TrimSpace(organizationID) == "" {
		return nil, errors.New("organization ID is required")
	}
	return &Repository{db: db, organizationID: organizationID}, nil
}

func (r *Repository) CreateProject(ctx context.Context, project *api.Project) error {
	if project == nil {
		return errors.New("project is required")
	}
	project.Normalize()
	if err := project.Validate(); err != nil {
		return err
	}
	labels, err := json.Marshal(project.Metadata.Labels)
	if err != nil {
		return fmt.Errorf("marshal project labels: %w", err)
	}
	annotations, err := json.Marshal(project.Metadata.Annotations)
	if err != nil {
		return fmt.Errorf("marshal project annotations: %w", err)
	}
	spec, err := json.Marshal(project.Spec)
	if err != nil {
		return fmt.Errorf("marshal project spec: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin project create: %w", err)
	}
	defer tx.Rollback()
	const query = `
		INSERT INTO projects (organization_id, name, display_name, description, environment, region, spec, labels, annotations, phase, generation)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9::jsonb, $10, $11)
		RETURNING id::text, generation, created_at, updated_at`
	var createdAt, updatedAt time.Time
	if err := tx.QueryRowContext(ctx, query, r.organizationID, project.Metadata.Name, project.Spec.DisplayName, project.Spec.Description, project.Spec.Environment, project.Spec.Region, spec, labels, annotations, project.Status.Phase, project.Metadata.Generation).
		Scan(&project.Metadata.UID, &project.Metadata.Generation, &createdAt, &updatedAt); err != nil {
		return fmt.Errorf("insert project: %w", err)
	}
	project.Status.LastReconciledAt = &updatedAt
	if err := insertOutbox(ctx, tx, "Project", project.Metadata.UID, "ProjectCreated", project); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit project create: %w", err)
	}
	return nil
}

func (r *Repository) GetProject(ctx context.Context, name string) (*api.Project, error) {
	const query = `
		SELECT id::text, name, display_name, description, environment, region, spec, labels, annotations,
		       phase, message, generation, observed_generation, created_at, updated_at
		FROM projects WHERE organization_id = $1 AND name = $2 AND deleted_at IS NULL`
	row := r.db.QueryRowContext(ctx, query, r.organizationID, name)
	return scanProject(row)
}

func (r *Repository) UpdateProject(ctx context.Context, name string, project *api.Project) error {
	if project == nil {
		return errors.New("project is required")
	}
	project.Normalize()
	project.Metadata.Name = name
	if err := project.Validate(); err != nil {
		return err
	}
	spec, err := json.Marshal(project.Spec)
	if err != nil {
		return fmt.Errorf("marshal project spec: %w", err)
	}
	labels, err := json.Marshal(project.Metadata.Labels)
	if err != nil {
		return fmt.Errorf("marshal project labels: %w", err)
	}
	annotations, err := json.Marshal(project.Metadata.Annotations)
	if err != nil {
		return fmt.Errorf("marshal project annotations: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin project update: %w", err)
	}
	defer tx.Rollback()
	var id string
	if err := tx.QueryRowContext(ctx, `SELECT id::text FROM projects WHERE organization_id = $1 AND name = $2 AND deleted_at IS NULL FOR UPDATE`, r.organizationID, name).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("find project: %w", err)
	}
	var generation int64
	var updatedAt time.Time
	if err := tx.QueryRowContext(ctx, `
		UPDATE projects SET display_name = $1, description = $2, environment = $3, region = $4,
			spec = $5::jsonb, labels = $6::jsonb, annotations = $7::jsonb, generation = generation + 1
		WHERE id = $8
		RETURNING generation, updated_at`, project.Spec.DisplayName, project.Spec.Description, project.Spec.Environment, project.Spec.Region, spec, labels, annotations, id).
		Scan(&generation, &updatedAt); err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	project.Metadata.UID = id
	project.Metadata.Generation = generation
	project.Status.LastReconciledAt = &updatedAt
	if err := insertOutbox(ctx, tx, "Project", id, "ProjectUpdated", project); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit project update: %w", err)
	}
	return nil
}

func (r *Repository) DeleteProject(ctx context.Context, name string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin project delete: %w", err)
	}
	defer tx.Rollback()
	var projectID string
	if err := tx.QueryRowContext(ctx, `SELECT id::text FROM projects WHERE organization_id = $1 AND name = $2 AND deleted_at IS NULL FOR UPDATE`, r.organizationID, name).Scan(&projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("find project: %w", err)
	}
	var active int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM database_instances WHERE project_id = $1 AND deleted_at IS NULL`, projectID).Scan(&active); err != nil {
		return fmt.Errorf("check project children: %w", err)
	}
	if active > 0 {
		return ErrProjectNotEmpty
	}
	var phase string
	if err := tx.QueryRowContext(ctx, `UPDATE projects SET phase = 'Deleting', deleted_at = now(), generation = generation + 1 WHERE id = $1 RETURNING phase`, projectID).Scan(&phase); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	payload := map[string]string{"name": name, "phase": phase}
	if err := insertOutbox(ctx, tx, "Project", projectID, "ProjectDeleted", payload); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit project delete: %w", err)
	}
	return nil
}

func (r *Repository) CreateDatabaseInstance(ctx context.Context, instance *api.DatabaseInstance) error {
	if instance == nil {
		return errors.New("database instance is required")
	}
	instance.Normalize()
	if err := instance.Validate(); err != nil {
		return err
	}
	storage, err := json.Marshal(instance.Spec.Storage)
	if err != nil {
		return fmt.Errorf("marshal storage: %w", err)
	}
	backup, err := json.Marshal(instance.Spec.Backup)
	if err != nil {
		return fmt.Errorf("marshal backup: %w", err)
	}
	network, err := json.Marshal(instance.Spec.Network)
	if err != nil {
		return fmt.Errorf("marshal network: %w", err)
	}
	spec, err := json.Marshal(instance.Spec)
	if err != nil {
		return fmt.Errorf("marshal database spec: %w", err)
	}
	labels, err := json.Marshal(instance.Metadata.Labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}
	annotations, err := json.Marshal(instance.Metadata.Annotations)
	if err != nil {
		return fmt.Errorf("marshal annotations: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin database create: %w", err)
	}
	defer tx.Rollback()
	var projectID string
	if err := tx.QueryRowContext(ctx, `SELECT id::text FROM projects WHERE organization_id = $1 AND name = $2 AND deleted_at IS NULL`, r.organizationID, instance.Metadata.Project).Scan(&projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("project %q: %w", instance.Metadata.Project, ErrNotFound)
		}
		return fmt.Errorf("find project: %w", err)
	}
	const query = `
		INSERT INTO database_instances (project_id, name, engine, version, plan, replicas, shards, storage, backup, network, spec, labels, annotations, phase, generation)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9::jsonb, $10::jsonb, $11::jsonb, $12::jsonb, $13::jsonb, $14, $15)
		RETURNING id::text, generation, created_at, updated_at`
	var createdAt, updatedAt time.Time
	if err := tx.QueryRowContext(ctx, query, projectID, instance.Metadata.Name, instance.Spec.Engine, instance.Spec.Version, instance.Spec.Plan, instance.Spec.Replicas, instance.Spec.Shards, storage, backup, network, spec, labels, annotations, instance.Status.Phase, instance.Metadata.Generation).
		Scan(&instance.Metadata.UID, &instance.Metadata.Generation, &createdAt, &updatedAt); err != nil {
		return fmt.Errorf("insert database instance: %w", err)
	}
	instance.Status.LastReconciledAt = &updatedAt
	if err := insertOutbox(ctx, tx, "DatabaseInstance", instance.Metadata.UID, "DatabaseInstanceCreated", instance); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit database create: %w", err)
	}
	return nil
}

func (r *Repository) GetDatabaseInstance(ctx context.Context, projectName, name string) (*api.DatabaseInstance, error) {
	const query = `
		SELECT d.id::text, p.name, d.name, d.engine, d.version, d.plan, d.replicas, d.shards,
		       d.storage, d.backup, d.network, d.phase, d.message, d.generation, d.observed_generation,
		       d.labels, d.annotations
		FROM database_instances d JOIN projects p ON p.id = d.project_id
		WHERE p.organization_id = $1 AND p.name = $2 AND d.name = $3 AND p.deleted_at IS NULL AND d.deleted_at IS NULL`
	row := r.db.QueryRowContext(ctx, query, r.organizationID, projectName, name)
	return scanDatabaseInstance(row)
}

func (r *Repository) UpdateDatabaseInstance(ctx context.Context, projectName, name string, instance *api.DatabaseInstance) error {
	if instance == nil {
		return errors.New("database instance is required")
	}
	instance.Normalize()
	instance.Metadata.Project = projectName
	instance.Metadata.Name = name
	if err := instance.Validate(); err != nil {
		return err
	}
	storage, err := json.Marshal(instance.Spec.Storage)
	if err != nil {
		return fmt.Errorf("marshal storage: %w", err)
	}
	backup, err := json.Marshal(instance.Spec.Backup)
	if err != nil {
		return fmt.Errorf("marshal backup: %w", err)
	}
	network, err := json.Marshal(instance.Spec.Network)
	if err != nil {
		return fmt.Errorf("marshal network: %w", err)
	}
	spec, err := json.Marshal(instance.Spec)
	if err != nil {
		return fmt.Errorf("marshal database spec: %w", err)
	}
	labels, err := json.Marshal(instance.Metadata.Labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}
	annotations, err := json.Marshal(instance.Metadata.Annotations)
	if err != nil {
		return fmt.Errorf("marshal annotations: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin database update: %w", err)
	}
	defer tx.Rollback()
	var id string
	if err := tx.QueryRowContext(ctx, `
		SELECT d.id::text FROM database_instances d JOIN projects p ON p.id = d.project_id
		WHERE p.organization_id = $1 AND p.name = $2 AND d.name = $3 AND d.deleted_at IS NULL FOR UPDATE`, r.organizationID, projectName, name).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("find database instance: %w", err)
	}
	var generation int64
	var updatedAt time.Time
	if err := tx.QueryRowContext(ctx, `
		UPDATE database_instances SET engine = $1, version = $2, plan = $3, replicas = $4, shards = $5,
			storage = $6::jsonb, backup = $7::jsonb, network = $8::jsonb, spec = $9::jsonb,
			labels = $10::jsonb, annotations = $11::jsonb, generation = generation + 1
		WHERE id = $12
		RETURNING generation, updated_at`, instance.Spec.Engine, instance.Spec.Version, instance.Spec.Plan, instance.Spec.Replicas, instance.Spec.Shards, storage, backup, network, spec, labels, annotations, id).
		Scan(&generation, &updatedAt); err != nil {
		return fmt.Errorf("update database instance: %w", err)
	}
	instance.Metadata.UID = id
	instance.Metadata.Generation = generation
	instance.Status.LastReconciledAt = &updatedAt
	if err := insertOutbox(ctx, tx, "DatabaseInstance", id, "DatabaseInstanceUpdated", instance); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit database update: %w", err)
	}
	return nil
}

func (r *Repository) ListDatabaseInstances(ctx context.Context, projectName string) ([]api.DatabaseInstance, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id::text, p.name, d.name, d.engine, d.version, d.plan, d.replicas, d.shards,
		       d.storage, d.backup, d.network, d.phase, d.message, d.generation, d.observed_generation,
		       d.labels, d.annotations
		FROM database_instances d JOIN projects p ON p.id = d.project_id
		WHERE p.organization_id = $1 AND p.name = $2 AND p.deleted_at IS NULL AND d.deleted_at IS NULL
		ORDER BY d.name`, r.organizationID, projectName)
	if err != nil {
		return nil, fmt.Errorf("list database instances: %w", err)
	}
	defer rows.Close()
	instances := []api.DatabaseInstance{}
	for rows.Next() {
		instance, err := scanDatabaseInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, *instance)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate database instances: %w", err)
	}
	return instances, nil
}

func (r *Repository) DeleteDatabaseInstance(ctx context.Context, projectName, name string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin database delete: %w", err)
	}
	defer tx.Rollback()
	var id string
	if err := tx.QueryRowContext(ctx, `
		SELECT d.id::text FROM database_instances d JOIN projects p ON p.id = d.project_id
		WHERE p.organization_id = $1 AND p.name = $2 AND d.name = $3 AND d.deleted_at IS NULL FOR UPDATE`, r.organizationID, projectName, name).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("find database instance: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE database_instances SET phase = 'Deleting', deleted_at = now(), generation = generation + 1 WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete database instance: %w", err)
	}
	if err := insertOutbox(ctx, tx, "DatabaseInstance", id, "DatabaseInstanceDeleted", map[string]string{"project": projectName, "name": name}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit database delete: %w", err)
	}
	return nil
}

func (r *Repository) UpdateDatabaseInstanceStatus(ctx context.Context, uid, phase, message string, observedGeneration int64) error {
	if !validPhase(phase) {
		return fmt.Errorf("invalid phase %q", phase)
	}
	result, err := r.db.ExecContext(ctx, `UPDATE database_instances SET phase = $1, message = $2, observed_generation = $3 WHERE id = $4 AND deleted_at IS NULL`, phase, message, observedGeneration, uid)
	if err != nil {
		return fmt.Errorf("update database instance status: %w", err)
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return ErrNotFound
	}
	return nil
}

func insertOutbox(ctx context.Context, tx *sql.Tx, aggregateType, aggregateID, eventType string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload) VALUES ($1, $2, $3, $4::jsonb)`, aggregateType, aggregateID, eventType, payload); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}

type scanner interface{ Scan(dest ...any) error }

func scanProject(row scanner) (*api.Project, error) {
	var p api.Project
	var spec, labels, annotations []byte
	var createdAt, updatedAt time.Time
	if err := row.Scan(&p.Metadata.UID, &p.Metadata.Name, &p.Spec.DisplayName, &p.Spec.Description, &p.Spec.Environment, &p.Spec.Region, &spec, &labels, &annotations, &p.Status.Phase, &p.Status.Message, &p.Metadata.Generation, &p.Status.ObservedGeneration, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan project: %w", err)
	}
	if err := json.Unmarshal(spec, &p.Spec); err != nil {
		return nil, fmt.Errorf("decode project spec: %w", err)
	}
	if err := json.Unmarshal(labels, &p.Metadata.Labels); err != nil {
		return nil, fmt.Errorf("decode project labels: %w", err)
	}
	if err := json.Unmarshal(annotations, &p.Metadata.Annotations); err != nil {
		return nil, fmt.Errorf("decode project annotations: %w", err)
	}
	p.APIVersion, p.Kind = api.APIVersion, "Project"
	p.Status.LastReconciledAt = &updatedAt
	return &p, nil
}

func scanDatabaseInstance(row scanner) (*api.DatabaseInstance, error) {
	var d api.DatabaseInstance
	var storage, backup, network, labels, annotations []byte
	if err := row.Scan(&d.Metadata.UID, &d.Metadata.Project, &d.Metadata.Name, &d.Spec.Engine, &d.Spec.Version, &d.Spec.Plan, &d.Spec.Replicas, &d.Spec.Shards, &storage, &backup, &network, &d.Status.Phase, &d.Status.Message, &d.Metadata.Generation, &d.Status.ObservedGeneration, &labels, &annotations); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan database instance: %w", err)
	}
	if err := json.Unmarshal(storage, &d.Spec.Storage); err != nil {
		return nil, fmt.Errorf("decode storage: %w", err)
	}
	if err := json.Unmarshal(backup, &d.Spec.Backup); err != nil {
		return nil, fmt.Errorf("decode backup: %w", err)
	}
	if err := json.Unmarshal(network, &d.Spec.Network); err != nil {
		return nil, fmt.Errorf("decode network: %w", err)
	}
	if err := json.Unmarshal(labels, &d.Metadata.Labels); err != nil {
		return nil, fmt.Errorf("decode labels: %w", err)
	}
	if err := json.Unmarshal(annotations, &d.Metadata.Annotations); err != nil {
		return nil, fmt.Errorf("decode annotations: %w", err)
	}
	d.APIVersion, d.Kind = api.APIVersion, "DatabaseInstance"
	return &d, nil
}

func validPhase(value string) bool {
	switch value {
	case "Pending", "Ready", "Degraded", "Failed", "Deleting":
		return true
	}
	return false
}
