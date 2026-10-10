package main

import (
	"strings"
	"testing"
)

func TestDatabaseURLForName(t *testing.T) {
	got, err := databaseURLForName("postgres://user:pass@localhost:5432/postgres?sslmode=disable", "archon")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "/archon?") || !strings.Contains(got, "sslmode=disable") {
		t.Fatalf("database URL = %q", got)
	}
}

func TestInstallPasswordGuardsConflictingInputs(t *testing.T) {
	if _, err := installPassword("one", true); err == nil {
		t.Fatal("expected conflicting password input to fail")
	}
	if _, err := installPassword("", false); err == nil {
		t.Fatal("expected missing password to fail")
	}
}
