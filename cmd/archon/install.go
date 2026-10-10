package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hacrex/Archon-Base/internal/auth"
	archondb "github.com/hacrex/Archon-Base/internal/db"
	"github.com/hacrex/Archon-Base/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var databaseNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)

func install(args []string) error {
	flagSet, flags := newInstallFlagSet()
	if err := flagSet.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*flags.email) == "" {
		return errors.New("--admin-email is required")
	}
	if strings.TrimSpace(*flags.displayName) == "" {
		*flags.displayName = "Archon Admin"
	}
	if !databaseNamePattern.MatchString(*flags.databaseName) {
		return errors.New("--database-name must contain only letters, numbers, and underscores and start with a letter or underscore")
	}
	password, err := installPassword(*flags.password, *flags.passwordStdin)
	if err != nil {
		return err
	}
	role := auth.RoleOwner
	adminURL := firstNonEmpty(*flags.adminDBURL, os.Getenv("ARCHON_ADMIN_DB_URL"))
	targetURL := firstNonEmpty(*flags.databaseURL, os.Getenv("ARCHON_DB_URL"))
	if targetURL == "" && adminURL != "" {
		targetURL, err = databaseURLForName(adminURL, *flags.databaseName)
		if err != nil {
			return fmt.Errorf("derive product database URL: %w", err)
		}
	}
	if targetURL == "" {
		return errors.New("a target database URL is required; use --database-url or ARCHON_DB_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if adminURL != "" {
		if err := createDatabase(ctx, adminURL, *flags.databaseName); err != nil {
			return err
		}
	}
	database, err := sql.Open("pgx", targetURL)
	if err != nil {
		return fmt.Errorf("open product database: %w", err)
	}
	defer database.Close()
	database.SetMaxOpenConns(5)
	database.SetMaxIdleConns(2)
	if err := database.PingContext(ctx); err != nil {
		return fmt.Errorf("product database is not ready: %w", err)
	}
	if err := (archondb.MigrationRunner{Directory: *flags.migrationsDir}).Run(ctx, database); err != nil {
		return fmt.Errorf("apply product database migrations: %w", err)
	}
	organizationID := firstNonEmpty(*flags.organizationID, envOr("ARCHON_ORGANIZATION_ID", localOrganizationID))
	repository, err := store.NewRepository(database, organizationID)
	if err != nil {
		return err
	}
	user, membership, err := repository.BootstrapUser(ctx, *flags.email, *flags.displayName, password, role)
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return fmt.Errorf("an admin with that email already exists; refusing to overwrite it")
		}
		return fmt.Errorf("create initial admin: %w", err)
	}
	fmt.Printf("Archon Base installed\n")
	fmt.Printf("Product database: %s\n", *flags.databaseName)
	fmt.Printf("Admin email: %s\n", user.Email)
	fmt.Printf("Admin user ID: %s\n", user.ID)
	fmt.Printf("Organization ID: %s\n", membership.Organization)
	fmt.Printf("Role: %s\n", membership.Role)
	fmt.Printf("Start the API with ARCHON_DB_URL=%s\n", targetURL)
	return nil
}

type installFlags struct {
	email, displayName, password, databaseURL, adminDBURL, databaseName, organizationID, migrationsDir *string
	passwordStdin                                                                                      *bool
}

func newInstallFlagSet() (*flag.FlagSet, *installFlags) {
	flagSet := flag.NewFlagSet("install", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	result := &installFlags{}
	result.email = flagSet.String("admin-email", "", "initial administrator email")
	result.displayName = flagSet.String("admin-name", "Archon Admin", "initial administrator display name")
	result.password = flagSet.String("admin-password", "", "initial administrator password; prefer --password-stdin")
	result.passwordStdin = flagSet.Bool("password-stdin", false, "read the administrator password from stdin")
	result.databaseURL = flagSet.String("database-url", "", "product database URL; defaults to ARCHON_DB_URL")
	result.adminDBURL = flagSet.String("admin-db-url", "", "PostgreSQL administrative URL used to create the product database")
	result.databaseName = flagSet.String("database-name", envOr("ARCHON_DATABASE_NAME", "archon"), "product database name")
	result.organizationID = flagSet.String("organization-id", "", "organization UUID")
	result.migrationsDir = flagSet.String("migrations-dir", envOr("ARCHON_MIGRATIONS_DIR", "migrations"), "migration directory")
	return flagSet, result
}

func installPassword(password string, stdin bool) (string, error) {
	if stdin && password != "" {
		return "", errors.New("use only one of --admin-password or --password-stdin")
	}
	if stdin {
		data, err := io.ReadAll(io.LimitReader(os.Stdin, 4097))
		if err != nil {
			return "", fmt.Errorf("read administrator password: %w", err)
		}
		password = strings.TrimRight(string(data), "\r\n")
	}
	if password == "" {
		return "", errors.New("an administrator password is required; use --password-stdin or --admin-password")
	}
	return password, nil
}

func createDatabase(ctx context.Context, adminURL, databaseName string) error {
	database, err := sql.Open("pgx", adminURL)
	if err != nil {
		return fmt.Errorf("open PostgreSQL administrative database: %w", err)
	}
	defer database.Close()
	if err := database.PingContext(ctx); err != nil {
		return fmt.Errorf("PostgreSQL administrative database is not ready: %w", err)
	}
	if _, err := database.ExecContext(ctx, `CREATE DATABASE `+quoteIdentifier(databaseName)); err != nil && !strings.Contains(strings.ToLower(err.Error()), "already exists") {
		return fmt.Errorf("create product database %q: %w", databaseName, err)
	}
	return nil
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func databaseURLForName(raw, databaseName string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parsed.Path = "/" + databaseName
	return parsed.String(), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
