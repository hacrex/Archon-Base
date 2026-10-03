package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	archondb "github.com/hacrex/Archon-Base/internal/db"
	"github.com/hacrex/Archon-Base/internal/server"
	"github.com/hacrex/Archon-Base/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const localOrganizationID = "00000000-0000-0000-0000-000000000001"

func main() {
	addr := envOr("ARCHON_API_ADDR", ":8080")
	srv := server.New()
	if databaseURL := os.Getenv("ARCHON_DB_URL"); databaseURL != "" {
		database, err := sql.Open("pgx", databaseURL)
		if err != nil {
			log.Fatalf("open postgres: %v", err)
		}
		defer database.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := database.PingContext(ctx); err != nil {
			log.Fatalf("postgres is not ready: %v", err)
		}
		migrationsDir := envOr("ARCHON_MIGRATIONS_DIR", "migrations")
		if err := (archondb.MigrationRunner{Directory: migrationsDir}).Run(ctx, database); err != nil {
			log.Fatalf("run migrations: %v", err)
		}
		organizationID := envOr("ARCHON_ORGANIZATION_ID", localOrganizationID)
		repository, err := store.NewRepository(database, organizationID)
		if err != nil {
			log.Fatalf("configure repository: %v", err)
		}
		srv = server.New(repository)
		log.Printf("postgres connected; migrations applied from %s", migrationsDir)
	} else {
		log.Printf("ARCHON_DB_URL is not set; resource routes remain unavailable")
	}
	log.Printf("archon-api listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
