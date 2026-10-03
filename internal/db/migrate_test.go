package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMigrationsOrdersUpFilesAndSkipsDownFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"002_second.sql", "001_first.sql", "001_first_down.sql", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("-- test\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	got, err := load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].version != "001_first" || got[1].version != "002_second" {
		t.Fatalf("unexpected migrations: %#v", got)
	}
}
