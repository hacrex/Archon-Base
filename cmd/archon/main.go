// Command archon is the Archon Base CLI.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/hacrex/Archon-Base/internal/auth"
	archondb "github.com/hacrex/Archon-Base/internal/db"
	"github.com/hacrex/Archon-Base/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const localOrganizationID = "00000000-0000-0000-0000-000000000001"

const usage = `archon: CLI for Archon Base

Usage:
  archon <command> [args]

Commands:
  version         print CLI version
  bootstrap-user  create the first active user and organization owner
  init            scaffold a new project        (planned)
  dev             run the local emulator        (planned)
  db              manage vector databases       (planned)
  bucket          manage buckets                (planned)
  deploy          deploy an agent function      (planned)
  logs            stream function logs          (planned)

Bootstrap example:
  printf '%s\n' 'replace-with-a-password' | archon bootstrap-user \\
    --email owner@example.com --display-name 'Local Owner' --password-stdin

The command requires ARCHON_DB_URL. It applies migrations before creating the
user, stores only a bcrypt password hash, and refuses duplicate email addresses.
`

func main() {
	if len(os.Args) < 2 {
		_, _ = os.Stdout.Write([]byte(usage))
		return
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("archon 0.0.1-dev")
	case "bootstrap-user":
		if err := bootstrapUser(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "archon bootstrap-user: %v\n", err)
			os.Exit(1)
		}
	case "init", "dev", "db", "bucket", "deploy", "logs":
		fmt.Printf("archon %s: not implemented yet\n", os.Args[1])
		os.Exit(2)
	default:
		fmt.Printf("unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}

func bootstrapUser(args []string) error {
	flags := flag.NewFlagSet("bootstrap-user", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	email := flags.String("email", "", "user email address")
	displayName := flags.String("display-name", "", "user display name")
	password := flags.String("password", "", "password; prefer --password-stdin")
	passwordStdin := flags.Bool("password-stdin", false, "read the password from stdin")
	organizationID := flags.String("organization-id", envOr("ARCHON_ORGANIZATION_ID", localOrganizationID), "organization UUID")
	roleValue := flags.String("role", "owner", "organization role")
	migrationsDir := flags.String("migrations-dir", envOr("ARCHON_MIGRATIONS_DIR", "migrations"), "migration directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*email) == "" || strings.TrimSpace(*displayName) == "" {
		return errors.New("--email and --display-name are required")
	}
	if *passwordStdin && *password != "" {
		return errors.New("use only one of --password or --password-stdin")
	}
	if *passwordStdin {
		data, err := io.ReadAll(io.LimitReader(os.Stdin, 4097))
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		*password = strings.TrimRight(string(data), "\r\n")
	}
	if *password == "" {
		return errors.New("a password is required; use --password-stdin or --password")
	}
	role, err := auth.ParseRole(*roleValue)
	if err != nil {
		return err
	}
	databaseURL := os.Getenv("ARCHON_DB_URL")
	if databaseURL == "" {
		return errors.New("ARCHON_DB_URL is required")
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer database.Close()
	database.SetMaxOpenConns(5)
	database.SetMaxIdleConns(2)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres is not ready: %w", err)
	}
	if err := (archondb.MigrationRunner{Directory: *migrationsDir}).Run(ctx, database); err != nil {
		return err
	}
	repository, err := store.NewRepository(database, *organizationID)
	if err != nil {
		return err
	}
	user, membership, err := repository.BootstrapUser(ctx, *email, *displayName, *password, role)
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return fmt.Errorf("a user with that email already exists; refusing to overwrite it")
		}
		return err
	}
	fmt.Printf("created user %s (%s) as %s in organization %s\n", user.Email, user.ID, membership.Role, membership.Organization)
	return nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
