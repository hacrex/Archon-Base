package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MigrationRunner applies forward SQL migrations once per filename version.
type MigrationRunner struct{ Directory string }

func (r MigrationRunner) Run(ctx context.Context, database *sql.DB) error {
	if database == nil {
		return errors.New("database is required")
	}
	if strings.TrimSpace(r.Directory) == "" {
		return errors.New("migration directory is required")
	}
	if _, err := database.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	migrations, err := load(r.Directory)
	if err != nil {
		return err
	}
	for _, m := range migrations {
		var applied bool
		if err := database.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, m.version).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", m.version, err)
		}
		if applied {
			continue
		}
		script, err := os.ReadFile(m.path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", m.version, err)
		}
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", m.version, err)
		}
		if _, err = tx.ExecContext(ctx, string(script)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", m.version, err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, m.version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", m.version, err)
		}
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", m.version, err)
		}
	}
	return nil
}

type migration struct{ version, path string }

func load(directory string) ([]migration, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read migration directory %q: %w", directory, err)
	}
	result := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") || strings.HasSuffix(entry.Name(), "_down.sql") {
			continue
		}
		result = append(result, migration{version: strings.TrimSuffix(entry.Name(), ".sql"), path: filepath.Join(directory, entry.Name())})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].version < result[j].version })
	return result, nil
}
