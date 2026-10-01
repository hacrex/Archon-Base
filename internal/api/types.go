// Package api defines the declarative Archon Base resource types (v1).
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

const APIVersion = "archon.dev/v1"

var (
	namePattern     = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	quantityPattern = regexp.MustCompile(`^[0-9]+(m|Ki|Mi|Gi|Ti|Pi|Ei)?$`)
)

type Metadata struct {
	UID         string            `json:"uid,omitempty" yaml:"uid,omitempty"`
	Name        string            `json:"name" yaml:"name"`
	Project     string            `json:"project,omitempty" yaml:"project,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Generation  int64             `json:"generation,omitempty" yaml:"generation,omitempty"`
}

type Condition struct {
	Type               string    `json:"type" yaml:"type"`
	Status             string    `json:"status" yaml:"status"`
	Reason             string    `json:"reason,omitempty" yaml:"reason,omitempty"`
	Message            string    `json:"message,omitempty" yaml:"message,omitempty"`
	ObservedGeneration int64     `json:"observedGeneration,omitempty" yaml:"observedGeneration,omitempty"`
	LastTransitionTime time.Time `json:"lastTransitionTime,omitempty" yaml:"lastTransitionTime,omitempty"`
}

type ResourceStatus struct {
	Phase              string      `json:"phase" yaml:"phase"`
	ObservedGeneration int64       `json:"observedGeneration" yaml:"observedGeneration"`
	Conditions         []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Message            string      `json:"message,omitempty" yaml:"message,omitempty"`
	LastReconciledAt   *time.Time  `json:"lastReconciledAt,omitempty" yaml:"lastReconciledAt,omitempty"`
}

type Project struct {
	APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string         `json:"kind" yaml:"kind"`
	Metadata   Metadata       `json:"metadata" yaml:"metadata"`
	Spec       ProjectSpec    `json:"spec" yaml:"spec"`
	Status     ResourceStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type ProjectSpec struct {
	DisplayName string  `json:"displayName" yaml:"displayName"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Environment string  `json:"environment,omitempty" yaml:"environment,omitempty"`
	Region      string  `json:"region,omitempty" yaml:"region,omitempty"`
	Quotas      *Quotas `json:"quotas,omitempty" yaml:"quotas,omitempty"`
}

type Quotas struct {
	CPU            string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory         string `json:"memory,omitempty" yaml:"memory,omitempty"`
	Storage        string `json:"storage,omitempty" yaml:"storage,omitempty"`
	VectorPoints   int64  `json:"vectorPoints,omitempty" yaml:"vectorPoints,omitempty"`
	TokensPerMonth int64  `json:"tokensPerMonth,omitempty" yaml:"tokensPerMonth,omitempty"`
}

type DatabaseInstance struct {
	APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string         `json:"kind" yaml:"kind"`
	Metadata   Metadata       `json:"metadata" yaml:"metadata"`
	Spec       DatabaseSpec   `json:"spec" yaml:"spec"`
	Status     ResourceStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type DatabaseSpec struct {
	Engine   string      `json:"engine" yaml:"engine"`
	Version  string      `json:"version,omitempty" yaml:"version,omitempty"`
	Plan     string      `json:"plan" yaml:"plan"`
	Replicas int         `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	Shards   int         `json:"shards,omitempty" yaml:"shards,omitempty"`
	Storage  StorageSpec `json:"storage" yaml:"storage"`
	Backup   BackupSpec  `json:"backup,omitempty" yaml:"backup,omitempty"`
	Network  NetworkSpec `json:"network,omitempty" yaml:"network,omitempty"`
}

type StorageSpec struct {
	Class string `json:"class,omitempty" yaml:"class,omitempty"`
	Size  string `json:"size" yaml:"size"`
}

type BackupSpec struct {
	Schedule      string `json:"schedule,omitempty" yaml:"schedule,omitempty"`
	RetentionDays int    `json:"retentionDays,omitempty" yaml:"retentionDays,omitempty"`
	Target        string `json:"target,omitempty" yaml:"target,omitempty"`
}

type NetworkSpec struct {
	Exposure string `json:"exposure,omitempty" yaml:"exposure,omitempty"`
}

// AgentFunction remains part of the v1 public model while its runtime is being built.
type AgentFunction struct {
	APIVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Kind       string            `json:"kind" yaml:"kind"`
	Metadata   Metadata          `json:"metadata" yaml:"metadata"`
	Spec       AgentFunctionSpec `json:"spec" yaml:"spec"`
}

type AgentFunctionSpec struct {
	Runtime    string `json:"runtime" yaml:"runtime"`
	Entrypoint string `json:"entrypoint" yaml:"entrypoint"`
	Isolation  string `json:"isolation" yaml:"isolation"`
}

type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string { return e.Field + ": " + e.Message }

type ValidationErrors []FieldError

func (e ValidationErrors) Error() string {
	parts := make([]string, 0, len(e))
	for _, field := range e {
		parts = append(parts, field.Error())
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

func (e ValidationErrors) Add(field, message string) ValidationErrors {
	return append(e, FieldError{Field: field, Message: message})
}

func (p *Project) Normalize() {
	if p.APIVersion == "" {
		p.APIVersion = APIVersion
	}
	if p.Kind == "" {
		p.Kind = "Project"
	}
	if p.Spec.Environment == "" {
		p.Spec.Environment = "dev"
	}
	if p.Spec.Region == "" {
		p.Spec.Region = "local"
	}
	if p.Metadata.Generation == 0 {
		p.Metadata.Generation = 1
	}
	if p.Status.Phase == "" {
		p.Status.Phase = "Pending"
	}
}

func (d *DatabaseInstance) Normalize() {
	if d.APIVersion == "" {
		d.APIVersion = APIVersion
	}
	if d.Kind == "" {
		d.Kind = "DatabaseInstance"
	}
	if d.Spec.Replicas == 0 {
		d.Spec.Replicas = 1
	}
	if d.Spec.Shards == 0 {
		d.Spec.Shards = 1
	}
	if d.Spec.Network.Exposure == "" {
		d.Spec.Network.Exposure = "private"
	}
	if d.Metadata.Generation == 0 {
		d.Metadata.Generation = 1
	}
	if d.Status.Phase == "" {
		d.Status.Phase = "Pending"
	}
}

func (p Project) Validate() error {
	var errs ValidationErrors
	if p.APIVersion != APIVersion {
		errs = errs.Add("apiVersion", "must be "+APIVersion)
	}
	if p.Kind != "Project" {
		errs = errs.Add("kind", "must be Project")
	}
	if !namePattern.MatchString(p.Metadata.Name) {
		errs = errs.Add("metadata.name", "must be a lowercase DNS-compatible name")
	}
	if strings.TrimSpace(p.Spec.DisplayName) == "" {
		errs = errs.Add("spec.displayName", "is required")
	}
	if p.Spec.Environment != "dev" && p.Spec.Environment != "staging" && p.Spec.Environment != "prod" {
		errs = errs.Add("spec.environment", "must be dev, staging, or prod")
	}
	if strings.TrimSpace(p.Spec.Region) == "" {
		errs = errs.Add("spec.region", "is required")
	}
	if p.Metadata.Generation < 1 {
		errs = errs.Add("metadata.generation", "must be positive")
	}
	if p.Spec.Quotas != nil {
		errs = validateQuantity(errs, "spec.quotas.cpu", p.Spec.Quotas.CPU)
		errs = validateQuantity(errs, "spec.quotas.memory", p.Spec.Quotas.Memory)
		errs = validateQuantity(errs, "spec.quotas.storage", p.Spec.Quotas.Storage)
		if p.Spec.Quotas.VectorPoints < 0 {
			errs = errs.Add("spec.quotas.vectorPoints", "must not be negative")
		}
		if p.Spec.Quotas.TokensPerMonth < 0 {
			errs = errs.Add("spec.quotas.tokensPerMonth", "must not be negative")
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (d DatabaseInstance) Validate() error {
	var errs ValidationErrors
	if d.APIVersion != APIVersion {
		errs = errs.Add("apiVersion", "must be "+APIVersion)
	}
	if d.Kind != "DatabaseInstance" {
		errs = errs.Add("kind", "must be DatabaseInstance")
	}
	if !namePattern.MatchString(d.Metadata.Name) {
		errs = errs.Add("metadata.name", "must be a lowercase DNS-compatible name")
	}
	if !namePattern.MatchString(d.Metadata.Project) {
		errs = errs.Add("metadata.project", "must be a lowercase DNS-compatible project name")
	}
	if !validEngine(d.Spec.Engine) {
		errs = errs.Add("spec.engine", "must be qdrant, chroma, weaviate, milvus, pgvector, or scylladb")
	}
	if strings.TrimSpace(d.Spec.Plan) == "" {
		errs = errs.Add("spec.plan", "is required")
	}
	if d.Spec.Replicas < 1 {
		errs = errs.Add("spec.replicas", "must be positive")
	}
	if d.Spec.Shards < 1 {
		errs = errs.Add("spec.shards", "must be positive")
	}
	if !quantityPattern.MatchString(d.Spec.Storage.Size) {
		errs = errs.Add("spec.storage.size", "must be a quantity such as 20Gi")
	}
	if d.Spec.Backup.RetentionDays < 0 {
		errs = errs.Add("spec.backup.retentionDays", "must not be negative")
	}
	if d.Spec.Network.Exposure != "private" && d.Spec.Network.Exposure != "project" && d.Spec.Network.Exposure != "public" {
		errs = errs.Add("spec.network.exposure", "must be private, project, or public")
	}
	if d.Metadata.Generation < 1 {
		errs = errs.Add("metadata.generation", "must be positive")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (f AgentFunctionSpec) Validate() error {
	var errs ValidationErrors
	if strings.TrimSpace(f.Runtime) == "" {
		errs = errs.Add("spec.runtime", "is required")
	}
	if strings.TrimSpace(f.Entrypoint) == "" {
		errs = errs.Add("spec.entrypoint", "is required")
	}
	if f.Isolation != "sandbox" && f.Isolation != "microvm" {
		errs = errs.Add("spec.isolation", "must be sandbox or microvm")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func DecodeStrict(data []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("decode resource: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("decode resource: trailing data")
	}
	return nil
}

func validateQuantity(errs ValidationErrors, field, value string) ValidationErrors {
	if value != "" && !quantityPattern.MatchString(value) {
		return errs.Add(field, "must be a quantity such as 1 or 8Gi")
	}
	return errs
}

func validEngine(value string) bool {
	switch value {
	case "qdrant", "chroma", "weaviate", "milvus", "pgvector", "scylladb":
		return true
	}
	return false
}

func EncodeSpec(value any) ([]byte, error)    { return json.Marshal(value) }
func DecodeSpec(data []byte, value any) error { return json.Unmarshal(data, value) }
