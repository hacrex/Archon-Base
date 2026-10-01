package store

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hacrex/Archon-Base/internal/api"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRepositoryPostgres(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil && os.Getenv("DOCKER_HOST") == "" {
		t.Skip("Docker is not available; run this test in a Docker-enabled environment")
	}
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("archon_test"),
		postgres.WithUsername("archon"),
		postgres.WithPassword("archon"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	connectionString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "001_core_resources.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	repository, err := NewRepository(db, "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}
	project := &api.Project{Metadata: api.Metadata{Name: "support-bot"}, Spec: api.ProjectSpec{DisplayName: "Support Bot"}}
	if err := repository.CreateProject(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if project.Metadata.UID == "" {
		t.Fatal("create project did not return a UID")
	}

	loadedProject, err := repository.GetProject(ctx, "support-bot")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if loadedProject.Spec.DisplayName != "Support Bot" || loadedProject.Spec.Environment != "dev" {
		t.Fatalf("unexpected project: %#v", loadedProject)
	}

	instance := &api.DatabaseInstance{Metadata: api.Metadata{Name: "docs-index", Project: "support-bot"}, Spec: api.DatabaseSpec{Engine: "qdrant", Plan: "standard-2", Storage: api.StorageSpec{Size: "20Gi"}}}
	if err := repository.CreateDatabaseInstance(ctx, instance); err != nil {
		t.Fatalf("create database instance: %v", err)
	}
	if instance.Metadata.UID == "" {
		t.Fatal("create database instance did not return a UID")
	}

	loadedInstance, err := repository.GetDatabaseInstance(ctx, "support-bot", "docs-index")
	if err != nil {
		t.Fatalf("get database instance: %v", err)
	}
	if loadedInstance.Spec.Engine != "qdrant" || loadedInstance.Spec.Replicas != 1 {
		t.Fatalf("unexpected database instance: %#v", loadedInstance)
	}

	instances, err := repository.ListDatabaseInstances(ctx, "support-bot")
	if err != nil {
		t.Fatalf("list database instances: %v", err)
	}
	if len(instances) != 1 || instances[0].Metadata.Name != "docs-index" {
		t.Fatalf("unexpected list: %#v", instances)
	}

	project.Spec.DisplayName = "Support Platform"
	if err := repository.UpdateProject(ctx, "support-bot", project); err != nil {
		t.Fatalf("update project: %v", err)
	}
	loadedProject, err = repository.GetProject(ctx, "support-bot")
	if err != nil {
		t.Fatalf("get updated project: %v", err)
	}
	if loadedProject.Spec.DisplayName != "Support Platform" || loadedProject.Metadata.Generation != 2 {
		t.Fatalf("project update was not persisted: %#v", loadedProject)
	}

	instance.Spec.Plan = "standard-4"
	if err := repository.UpdateDatabaseInstance(ctx, "support-bot", "docs-index", instance); err != nil {
		t.Fatalf("update database instance: %v", err)
	}
	if err := repository.UpdateDatabaseInstanceStatus(ctx, instance.Metadata.UID, "Ready", "reconciled", instance.Metadata.Generation); err != nil {
		t.Fatalf("update database status: %v", err)
	}
	loadedInstance, err = repository.GetDatabaseInstance(ctx, "support-bot", "docs-index")
	if err != nil {
		t.Fatalf("get updated database instance: %v", err)
	}
	if loadedInstance.Spec.Plan != "standard-4" || loadedInstance.Status.Phase != "Ready" {
		t.Fatalf("database update was not persisted: %#v", loadedInstance)
	}

	var eventCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM outbox_events`).Scan(&eventCount); err != nil {
		t.Fatalf("count outbox events: %v", err)
	}
	if eventCount < 4 {
		t.Fatalf("expected create/update outbox events, got %d", eventCount)
	}

	if err := repository.DeleteDatabaseInstance(ctx, "support-bot", "docs-index"); err != nil {
		t.Fatalf("delete database instance: %v", err)
	}
	if _, err := repository.GetDatabaseInstance(ctx, "support-bot", "docs-index"); err != ErrNotFound {
		t.Fatalf("expected deleted database to be hidden, got %v", err)
	}
	if err := repository.DeleteProject(ctx, "support-bot"); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if _, err := repository.GetProject(ctx, "support-bot"); err != ErrNotFound {
		t.Fatalf("expected deleted project to be hidden, got %v", err)
	}
}
