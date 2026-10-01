// Package api defines the declarative resource types (v1).
// These will later be generated into Kubernetes CRDs.
package api

const APIVersion = "archon.dev/v1"

type Metadata struct {
	Name    string `json:"name" yaml:"name"`
	Project string `json:"project" yaml:"project"`
}

type DatabaseInstance struct {
	APIVersion string       `json:"apiVersion" yaml:"apiVersion"`
	Kind       string       `json:"kind" yaml:"kind"`
	Metadata   Metadata     `json:"metadata" yaml:"metadata"`
	Spec       DatabaseSpec `json:"spec" yaml:"spec"`
}

type DatabaseSpec struct {
	Engine   string `json:"engine" yaml:"engine"` // qdrant | chroma | weaviate | milvus | pgvector | scylladb
	Version  string `json:"version" yaml:"version"`
	Plan     string `json:"plan" yaml:"plan"`
	Replicas int    `json:"replicas" yaml:"replicas"`
	Shards   int    `json:"shards" yaml:"shards"`
}

type AgentFunction struct {
	APIVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Kind       string            `json:"kind" yaml:"kind"`
	Metadata   Metadata          `json:"metadata" yaml:"metadata"`
	Spec       AgentFunctionSpec `json:"spec" yaml:"spec"`
}

type AgentFunctionSpec struct {
	Runtime    string `json:"runtime" yaml:"runtime"`
	Entrypoint string `json:"entrypoint" yaml:"entrypoint"`
	Isolation  string `json:"isolation" yaml:"isolation"` // sandbox | microvm
}
