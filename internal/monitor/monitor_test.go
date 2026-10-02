package monitor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEvaluateStudentProfile(t *testing.T) {
	base := Snapshot{
		CPU:    CPUStatus{Cores: 2},
		Memory: MemoryStats{TotalBytes: 8 * 1024 * 1024 * 1024},
		Disk:   DiskStats{TotalBytes: 20 * 1024 * 1024 * 1024},
	}
	result := Evaluate(base)
	if !result.Ready {
		t.Fatalf("expected minimum student profile to be ready: %#v", result)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("expected optional GPU warning, got %#v", result.Warnings)
	}

	base.CPU.Cores = 1
	base.Memory.TotalBytes = 4 * 1024 * 1024 * 1024
	result = Evaluate(base)
	if result.Ready {
		t.Fatal("undersized profile must not be ready")
	}
}

func TestWriteJSONAtomicSnapshot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "status.json")
	snapshot := Snapshot{CollectedAt: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC), Profile: "enthusiast-k3s"}
	if err := WriteJSON(path, snapshot); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var decoded Snapshot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if decoded.Profile != snapshot.Profile {
		t.Fatalf("unexpected snapshot: %#v", decoded)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("temporary snapshot file was not removed")
	}
}
