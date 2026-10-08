package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	archondb "github.com/hacrex/Archon-Base/internal/db"
	"github.com/hacrex/Archon-Base/internal/server"
	"github.com/hacrex/Archon-Base/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const localOrganizationID = "00000000-0000-0000-0000-000000000001"

func main() {
	addr := envOr("ARCHON_API_ADDR", "127.0.0.1:8080")
	srv := server.New()
	var database *sql.DB
	if databaseURL := os.Getenv("ARCHON_DB_URL"); databaseURL != "" {
		var err error
		database, err = sql.Open("pgx", databaseURL)
		if err != nil {
			log.Fatalf("open postgres: %v", err)
		}
		defer database.Close()
		configurePool(database)
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

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	stop, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal()
	go func() {
		<-stop.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownContext); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}
	}()

	log.Printf("archon-api listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func configurePool(database *sql.DB) {
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)
	database.SetConnMaxIdleTime(5 * time.Minute)
	database.SetConnMaxLifetime(30 * time.Minute)
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
