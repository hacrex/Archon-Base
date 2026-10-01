package api

import (
	"strings"
	"testing"
)

func TestProjectNormalizeAndValidate(t *testing.T) {
	project := Project{
		Metadata: Metadata{Name: "support-bot"},
		Spec:     ProjectSpec{DisplayName: "Support Bot"},
	}
	project.Normalize()
	if err := project.Validate(); err != nil {
		t.Fatalf("expected valid project, got %v", err)
	}
	if project.APIVersion != APIVersion || project.Kind != "Project" {
		t.Fatalf("normalization did not set envelope: %#v", project)
	}
	if project.Spec.Environment != "dev" || project.Spec.Region != "local" {
		t.Fatalf("normalization did not set defaults: %#v", project.Spec)
	}
	if project.Metadata.Generation != 1 {
		t.Fatalf("expected generation 1, got %d", project.Metadata.Generation)
	}
}

func TestDatabaseInstanceValidation(t *testing.T) {
	instance := DatabaseInstance{
		Metadata: Metadata{Name: "docs-index", Project: "support-bot"},
		Spec:     DatabaseSpec{Engine: "qdrant", Plan: "standard-2", Storage: StorageSpec{Size: "20Gi"}},
	}
	instance.Normalize()
	if err := instance.Validate(); err != nil {
		t.Fatalf("expected valid database instance, got %v", err)
	}
	if instance.Spec.Replicas != 1 || instance.Spec.Shards != 1 || instance.Spec.Network.Exposure != "private" {
		t.Fatalf("unexpected defaults: %#v", instance.Spec)
	}
}

func TestValidationReportsFields(t *testing.T) {
	project := Project{APIVersion: APIVersion, Kind: "Project", Metadata: Metadata{Name: "Bad Name"}, Spec: ProjectSpec{Environment: "qa", Region: ""}}
	err := project.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	message := err.Error()
	for _, field := range []string{"metadata.name", "spec.displayName", "spec.environment", "spec.region"} {
		if !strings.Contains(message, field) {
			t.Errorf("error %q does not mention %s", message, field)
		}
	}
}

func TestDecodeStrictRejectsUnknownFieldsAndTrailingData(t *testing.T) {
	var project Project
	if err := DecodeStrict([]byte(`{"apiVersion":"archon.dev/v1","kind":"Project","metadata":{"name":"demo"},"spec":{"displayName":"Demo","unknown":true}}`), &project); err == nil {
		t.Fatal("expected unknown field error")
	}
	if err := DecodeStrict([]byte(`{"apiVersion":"archon.dev/v1","kind":"Project","metadata":{"name":"demo"},"spec":{"displayName":"Demo"}} {}`), &project); err == nil {
		t.Fatal("expected trailing data error")
	}
}
