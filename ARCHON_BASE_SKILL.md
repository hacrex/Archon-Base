# Archon Base — Project Skill / Engineering Specification

> **Project:** Archon Base  
> **Purpose:** Self-hostable AI cloud platform / AI infrastructure control plane  
> **Status:** Architecture baseline v1.1  
> **Source:** Archon Base Technical Design v1.0 plus architecture review findings  
> **Audience:** Human engineers, AI coding agents, platform engineers, security engineers, contributors

---

## 0. Purpose of This Skill File

This file is the governing engineering context for implementing Archon Base.

When working on this repository, an AI coding agent or engineer MUST:

1. Preserve the architectural principles defined here.
2. Prefer existing open-source components behind stable Archon interfaces rather than rebuilding infrastructure.
3. Treat security, tenant isolation, reconciliation, idempotency, and failure recovery as first-class requirements.
4. Keep the Archon core narrow and extensible.
5. Avoid adding infrastructure dependencies without an explicit architectural reason.
6. Never silently weaken isolation, authorization, credential handling, or auditability.
7. Prefer declarative resources and reconciliation over imperative scripts.
8. Keep all APIs, CRDs, events, SDKs, CLI commands, and console operations aligned around the same resource model.
9. Treat this file as implementation guidance, not permission to invent unsupported platform behavior.

If a proposed implementation conflicts with this file, stop and identify the conflict before proceeding.

---

# 1. Product Definition

Archon Base is an open-source, self-hostable backend and infrastructure platform for AI applications.

It provides a unified control plane for:

- AI agents
- serverless/sandboxed agent functions
- vector databases
- object storage
- model inference
- model gateways
- knowledge/RAG pipelines
- tools and MCP servers
- agent memory
- evaluation
- observability
- governance
- secrets
- quotas and usage

The platform is intended to run on infrastructure owned by the user, from a single-node development installation to a multi-node Kubernetes deployment.

The product should feel conceptually similar to a combination of:

- Supabase/Appwrite-style developer platform
- Kubernetes control plane
- serverless agent runtime
- self-hosted AI model gateway
- AI infrastructure management layer

Archon Base is NOT intended to become a new implementation of every underlying infrastructure component.

---

# 2. Core Architectural Principle

## Narrow Core, Broad Adapters

Archon owns:

- resource model
- control plane
- reconciliation
- identity
- authorization
- bindings
- credential brokering
- policy
- agent lifecycle
- unified APIs
- usage/accounting
- audit
- developer experience

Specialized systems should perform specialized work.

Examples:

| Capability | Preferred implementation |
|---|---|
| Kubernetes | Kubernetes / k3s |
| Metadata | PostgreSQL |
| Events | NATS JetStream |
| Vector search | Qdrant initially |
| Object storage | S3-compatible backend / SeaweedFS initially |
| Secrets | OpenBao initially |
| Sandbox | gVisor initially |
| Strong isolation | Firecracker/Kata later |
| Inference | vLLM initially |
| GPU orchestration | Kubernetes GPU ecosystem |
| Networking | Cilium |
| Observability | OpenTelemetry |
| Workflows | Temporal or another established engine |
| Auth | OIDC provider / Keycloak or Ory |
| Policy | OPA or Cedar |

Every integration MUST sit behind an Archon abstraction where practical.

---

# 3. Architecture

## 3.1 Logical Architecture

```text
                         ARCHON BASE
                              |
             +----------------+----------------+
             |                                 |
       Developer Plane                    Admin Plane
             |                                 |
      CLI / SDK / REST                    Console / API
             |                                 |
             +---------------+-----------------+
                             |
                        API Gateway
                             |
             +---------------v----------------+
             |          CONTROL PLANE         |
             |                                |
             | IAM / Policy                   |
             | Organization / Project         |
             | Resource API                   |
             | Provisioner                    |
             | Agent Controller               |
             | Model Controller               |
             | Storage Controller             |
             | Credential Broker              |
             | Event Controller               |
             | Audit                          |
             | Usage / Billing                |
             +---------------+----------------+
                             |
                   +---------+---------+
                   |                   |
              PostgreSQL          NATS JetStream
                   |                   |
                   +---------+---------+
                             |
                       Reconciliation
                             |
             +---------------v----------------+
             |           DATA PLANE           |
             |                                |
             | Agent Runtime                  |
             | Vector DB                      |
             | Object Storage                 |
             | Model Gateway                  |
             | Inference                     |
             | Knowledge Pipelines            |
             | Workflow Engine                |
             +---------------+----------------+
                             |
             +---------------+----------------+
             |               |                |
          Cilium          Sandbox            GPU
                         gVisor/microVM     Operator
```

---

# 4. Control Plane Contract

The control plane is the heart of Archon.

## 4.1 Source of Truth

For Archon resources:

**PostgreSQL is the control-plane source of truth.**

Kubernetes resources are derived desired state for data-plane execution.

The system MUST NOT depend on an ephemeral Kubernetes object alone to represent Archon ownership, tenancy, lifecycle, or billing state.

## 4.2 Reconciliation Flow

```text
API Request
   |
   v
Validate
   |
   v
Authorize
   |
   v
PostgreSQL
   |
   +--> Outbox Event
             |
             v
        NATS JetStream
             |
             v
        Controller
             |
             v
      Kubernetes API / CRD
             |
             v
        Operator/Runtime
             |
             v
        Actual State
             |
             v
       Status Update
             |
             v
        PostgreSQL/API
```

## 4.3 Transactional Outbox

Any operation that changes PostgreSQL state and produces an event MUST use a transactional outbox.

Example:

```text
BEGIN

INSERT resource
INSERT outbox_event

COMMIT
```

A separate publisher delivers outbox events to NATS.

Never implement:

```text
write database
publish event
```

as two unrelated operations when correctness depends on both succeeding.

---

# 5. Resource Lifecycle

All platform resources MUST expose a consistent lifecycle.

Minimum state model:

```text
Pending
   |
Provisioning
   |
Ready
```

Failure/degradation paths:

```text
Provisioning
   +--> Failed
   +--> Degraded
   +--> Ready
```

Deletion:

```text
Ready
  |
Deleting
  |
Deleted
```

## 5.1 Status Requirements

Resources SHOULD expose:

```yaml
status:
  phase: Ready
  observedGeneration: 12
  conditions:
    - type: Ready
      status: "True"
      reason: Provisioned
      message: Resource is operational
```

Controllers MUST be idempotent.

Repeated reconciliation MUST produce the same desired state.

---

# 6. Declarative Resource Model

Every major Archon resource MUST have:

- apiVersion
- kind
- metadata
- tenant/project/environment ownership
- spec
- status
- generation
- observedGeneration
- conditions

Core resources include:

```text
Organization
Project
Environment

DatabaseInstance
Bucket
AgentFunction
Model
ModelDeployment
Tool
Pipeline
Workflow
Secret
Policy
AgentIdentity
CredentialBinding
EvalSuite
EvalRun
EventSubscription
Backup
Operation
```

---

# 7. Tenancy Model

Hierarchy:

```text
Organization
    |
    +-- Project
          |
          +-- Environment
                |
                +-- Resource
```

Environment examples:

```text
dev
staging
prod
```

Cross-project access MUST be denied by default.

Cross-environment access MUST be denied by default.

A resource MUST always have a clearly resolvable ownership scope.

---

# 8. Identity Model

Archon MUST distinguish:

```text
Human User
Service Account
Agent Identity
API Key
Workload Identity
```

An agent MUST NOT use a shared project-wide credential as its primary identity.

Preferred flow:

```text
Agent
 |
 v
Agent Identity
 |
 v
Policy Evaluation
 |
 v
Binding
 |
 v
Credential Broker
 |
 v
Short-lived Credential
 |
 v
Target Resource
```

---

# 9. Authorization

Authorization MUST be evaluated at multiple layers where relevant:

1. API gateway
2. Control plane
3. Resource controller
4. Kubernetes/network policy
5. Data/resource endpoint

Example permissions:

```text
database.read
database.write
database.admin

bucket.read
bucket.write
bucket.admin

model.invoke
model.admin

agent.invoke
agent.deploy
agent.admin

secret.read
secret.write

tool.invoke
tool.admin
```

Use RBAC for common access control.

Use OPA or Cedar for fine-grained policy where required.

Never rely exclusively on UI restrictions.

---

# 10. Binding Model

Agents receive access through bindings rather than manually configured credentials.

Example:

```yaml
bindings:
  vector:
    - instance: docs-index
      access: read

  storage:
    - bucket: support-docs
      access: read

  models:
    - alias: chat-llm

  tools:
    - name: crm-lookup
      approval: required
```

A binding MUST resolve to:

```text
identity
+
authorization
+
resource
+
permission
+
credential mechanism
```

Bindings must be revocable.

---

# 11. Credential Broker

The credential broker is a core platform service.

Responsibilities:

- issue short-lived credentials
- scope credentials
- rotate credentials
- revoke credentials
- enforce audience/resource
- inject credentials at runtime
- prevent credential leakage
- audit credential access

Credentials SHOULD be short-lived whenever the target system supports it.

Secrets MUST NOT be embedded into container images.

Secrets MUST NOT be committed to Git.

Secrets MUST NOT appear in logs or traces.

---

# 12. Agent Runtime

An AgentFunction is:

- executable code
- immutable version
- runtime
- entrypoint
- resource limits
- triggers
- bindings
- secrets
- network policy
- sandbox policy
- identity
- deployment configuration

Example:

```yaml
apiVersion: archon.dev/v1
kind: AgentFunction

metadata:
  name: support-agent
  project: support-bot

spec:
  runtime: python3.12
  entrypoint: app.main:handle

  resources:
    cpu: "1"
    memory: 1Gi
    timeout: 300s
    concurrency: 10
    gpu: 0

  isolation:
    mode: sandbox

  bindings:
    vector:
      - instance: docs-index
        access: read

  egress:
    allow:
      - api.crm.example.com
```

---

# 13. Agent Lifecycle

Agents SHOULD follow:

```text
Draft
  |
Build
  |
Scan
  |
Sign
  |
Publish
  |
Deploy
  |
Canary
  |
Active
  |
Deprecated
  |
Retired
```

Each deployment MUST reference an immutable artifact/version.

Do not deploy mutable Git branches directly into production without producing a versioned artifact.

---

# 14. Sandbox Security

Default sandbox:

**gVisor**

Higher-risk workloads:

**Firecracker/Kata microVM**

Potential lightweight plugins:

**WASM**

Do not treat a sandbox as automatically secure.

Threat model MUST include:

- filesystem escape
- container escape
- kernel exploitation
- Kubernetes API access
- credential theft
- metadata-service access
- SSRF
- DNS exfiltration
- network scanning
- crypto mining
- fork bombs
- resource exhaustion
- malicious dependencies
- supply-chain compromise

Agents MUST NOT access:

```text
Kubernetes API
host filesystem
host PID namespace
host network
cloud metadata endpoints
privileged devices
```

unless explicitly required and policy-approved.

---

# 15. Network Security

Use Cilium where Kubernetes networking is available.

Default posture:

**deny by default**

Required controls:

- namespace isolation
- project isolation
- environment isolation
- egress allowlists
- ingress controls
- DNS controls
- mTLS where appropriate
- network auditability

Example:

```yaml
egress:
  mode: allowlist
  allow:
    - host: api.crm.example.com
      port: 443
```

Network policy MUST be generated from the Archon resource/policy model rather than manually maintained independently.

---

# 16. Object Storage

The platform needs S3-compatible storage for:

- user files
- documents
- datasets
- model weights
- artifacts
- backups
- snapshots

Initial backend:

**SeaweedFS**

Other backends are adapters.

Storage SHOULD support classes such as:

```text
hot
warm
archive
model-cache
backup
```

Never assume model artifacts, user documents and backups have identical retention requirements.

---

# 17. Backup and Disaster Recovery

Backups MUST cover the platform state required to reconstruct an installation.

Potential backup domains:

```text
PostgreSQL
NATS durable streams
Kubernetes/Archon resource state
Secrets metadata
Object storage
Vector databases
Model registry
Agent artifacts
Configuration
```

Backups MUST have:

- retention
- encryption
- integrity validation
- restore procedure
- restore verification
- documented RPO
- documented RTO

A backup is not considered successful until restore validation is possible.

Use snapshot/generation identifiers to correlate:

```text
metadata
database
object storage
vector state
```

when consistency matters.

---

# 18. Model Gateway

The Model Gateway is a first-class platform service.

Architecture:

```text
Client
  |
Authentication
  |
Authorization
  |
Quota
  |
Budget
  |
Policy
  |
Model Router
  |
Cache
  |
Provider
  |
Usage Accounting
  |
Tracing
```

Initial inference backend:

**vLLM**

Additional adapters:

- Ollama
- TGI
- llama.cpp
- external providers

The gateway SHOULD expose an OpenAI-compatible API.

Use model aliases:

```text
chat-llm
embedding-model
reasoning-model
vision-model
```

Agents should not hard-code provider-specific model IDs.

---

# 19. Model Metadata

Models SHOULD have metadata including:

```text
name
version
provider
artifact
checksum
license
context_length
capabilities
quantization
architecture
GPU requirements
CPU requirements
memory requirements
security status
```

Capabilities may include:

```text
chat
completion
embedding
vision
audio
tool_calling
structured_output
reasoning
```

---

# 20. GPU Architecture

GPU resources MUST be treated as more than:

```yaml
gpu: 1
```

Where practical, model scheduling should understand:

```text
accelerator type
GPU count
GPU memory
MIG profile
time slicing
architecture
CUDA/runtime compatibility
```

Example:

```yaml
resources:
  accelerator:
    type: nvidia
    memory: 24Gi
    count: 1
```

GPU scheduling must account for:

- GPU memory
- utilization
- model size
- quantization
- node locality
- model cache
- concurrency
- preemption policy

---

# 21. Model Lifecycle

Preferred lifecycle:

```text
Artifact
  |
Registry
  |
Validation
  |
Security Scan
  |
Deployment
  |
Warm/Cold Loading
  |
Endpoint
  |
Alias
```

Large model weights should use node-local caching where possible.

Avoid downloading tens of GB of model data for every invocation.

---

# 22. Event Architecture

NATS JetStream is the initial event system.

Assume:

**at-least-once delivery**

Consumers MUST be idempotent.

Events MUST be versioned.

Suggested event format:

```json
{
  "id": "evt_123",
  "type": "storage.object.created",
  "version": "1",
  "organization": "org_123",
  "project": "support-bot",
  "environment": "prod",
  "subject": "bucket/support-docs/object/a.pdf",
  "timestamp": "2026-01-01T00:00:00Z",
  "data": {}
}
```

Events MUST NOT contain secrets.

Sensitive payloads should be referenced rather than copied into the event stream.

---

# 23. Workflow Architecture

Do not build a custom durable workflow engine unless there is a compelling architectural reason.

Archon should expose a workflow abstraction while delegating durable execution to an established engine such as Temporal where appropriate.

Archon owns:

- workflow resources
- authorization
- tenancy
- bindings
- lifecycle
- audit

The workflow engine owns:

- execution
- retries
- checkpointing
- timers
- durable state

---

# 24. Knowledge / RAG Pipeline

Initial pipeline model:

```text
Source
  |
Parse
  |
Normalize
  |
Chunk
  |
Embed
  |
Index
  |
Retrieve
  |
Rerank
```

Initial connectors should remain narrow.

Preferred early connectors:

- S3-compatible storage
- Git
- HTTP/web

Build a connector interface rather than implementing every SaaS integration immediately.

---

# 25. Vector Database Strategy

Initial default:

**Qdrant**

Adapters may later support:

- pgvector
- Chroma
- Weaviate
- Milvus
- other engines

Archon MUST NOT hide engine-specific features behind an abstraction that makes advanced capabilities impossible.

Unified vector operations may include:

```text
upsert
query
delete
filter
```

Native engine APIs remain available.

---

# 26. API Design

REST and gRPC may coexist.

REST should be developer-friendly.

All mutating operations MUST define idempotency semantics.

Long-running operations should return an operation/resource reference instead of keeping HTTP connections open unnecessarily.

Example:

```http
POST /v1/projects/p/databases/docs-index/restore
```

Response:

```json
{
  "operation": "op_123",
  "status": "pending"
}
```

API requirements:

- versioning
- pagination
- filtering
- sorting
- idempotency
- consistent errors
- request IDs
- audit IDs
- async operations
- backwards compatibility

---

# 27. Error Model

Errors SHOULD have a stable structure:

```json
{
  "error": {
    "code": "RESOURCE_NOT_READY",
    "message": "Database is still provisioning",
    "request_id": "req_123",
    "resource": "database/docs-index",
    "retryable": true
  }
}
```

Do not expose:

- stack traces
- internal secrets
- credentials
- sensitive infrastructure details

in normal API responses.

---

# 28. Usage and FinOps

Usage accounting should track:

```text
input tokens
output tokens
cached tokens
GPU seconds
CPU seconds
memory GB-seconds
storage GB
egress
API calls
agent invocations
tool calls
```

Hierarchy:

```text
Organization
  |
Project
  |
Environment
  |
Agent
  |
Model
  |
Request
```

Budgets and limits should be enforceable.

Example:

```text
usage
  |
cost
  |
budget
  |
alert
  |
enforcement
```

---

# 29. Observability

Use OpenTelemetry as the common instrumentation layer.

Initial ecosystem:

```text
Prometheus
Loki
Tempo
Grafana
```

Telemetry MUST account for AI-specific data sensitivity.

Configurable controls should include:

```text
prompt logging
response logging
trace sampling
trace retention
PII redaction
payload capture
```

Never assume prompts and responses are safe to retain.

Telemetry should expose:

- API latency
- reconciliation latency
- event latency
- function latency
- cold starts
- GPU utilization
- GPU memory
- model latency
- token usage
- tool failures
- retrieval latency
- workflow failures
- sandbox failures

---

# 30. Security Architecture

Required security layers:

```text
Identity
Authorization
Secrets
Network
Sandbox
Data
Supply Chain
Audit
AI Governance
```

Required controls:

- OIDC
- RBAC
- policy engine
- short-lived credentials
- default-deny networking
- sandboxing
- TLS
- encryption at rest
- image signing
- SBOM
- vulnerability scanning
- immutable audit logs
- prompt-injection protections
- PII handling
- tool scopes
- approval workflows
- spend limits

---

# 31. Supply Chain

Agent source should follow:

```text
Git
 |
Build
 |
Dependency Resolution
 |
SBOM
 |
Vulnerability Scan
 |
Image
 |
Sign
 |
Registry
 |
Admission Verification
 |
Deploy
```

Container images MUST be signed in production-oriented environments.

Pinned versions are preferred.

Unpinned `latest` tags MUST NOT be used for production platform components.

---

# 32. Plugin Security

Plugins are untrusted by default.

Installation flow:

```text
Plugin
  |
Signature Verification
  |
Publisher Verification
  |
Compatibility Check
  |
Permission Check
  |
Sandbox
  |
Install
```

Plugins should declare permissions.

A plugin MUST NOT automatically receive:

- cluster-admin
- host filesystem access
- unrestricted network
- all project secrets

---

# 33. Local Development

Two local modes are recommended.

## Lite

```text
archon dev
```

Fast local emulator with minimal dependencies.

## Production-like

```text
archon dev --profile production
```

Use containers for services such as:

```text
PostgreSQL
NATS
Qdrant
S3-compatible storage
OpenBao
```

The local emulator should preserve the same API and resource model as production.

---

# 34. Installation

The installer should perform:

```text
Detect environment
  |
Detect Kubernetes
  |
Detect CPU/GPU
  |
Validate storage
  |
Validate networking
  |
Generate secrets
  |
Install dependencies
  |
Install Archon
  |
Bootstrap admin
  |
Health checks
```

Required operational commands should include:

```bash
archon install
archon uninstall
archon upgrade
archon doctor
archon backup
archon restore
```

`archon doctor` should diagnose:

- Kubernetes
- storage
- networking
- DNS
- certificates
- GPU
- required ports
- permissions
- Archon components
- dependencies

---

# 35. Deployment Profiles

## Evaluation

```text
4 vCPU
16 GB RAM
100 GB SSD
```

Target: experimentation.

## Team

```text
3 nodes
8 vCPU/node
32 GB/node
```

Target: development teams.

## Production

Dedicated workload pools where justified:

```text
control
database
function
GPU
storage
```

## Air-gapped

Use:

```text
offline images
local registry
OCI bundles
offline packages
offline model artifacts
```

---

# 36. Dependency Classification

## Required

```text
Kubernetes
PostgreSQL
Archon API
```

## Core

```text
NATS
OpenBao
Qdrant
Object Storage
```

## Optional

```text
Prometheus
Grafana
Loki
Tempo
vLLM
```

## Extensions

```text
Milvus
Weaviate
Chroma
Temporal
MCP
A2A
LangGraph
additional providers
```

Avoid making every extension mandatory for installation.

---

# 37. MVP Boundary

The MVP MUST remain narrow.

Recommended initial stack:

```text
Kubernetes / k3s
PostgreSQL
NATS JetStream
Qdrant
S3-compatible storage
OpenBao
gVisor
Python Agent Runtime
vLLM
CLI
Basic Console
```

Initial capabilities:

1. Projects
2. Environments
3. IAM basics
4. Agent identities
5. Bindings
6. Credential brokering
7. Qdrant provisioning
8. Buckets
9. Sandboxed Python functions
10. Model Gateway
11. Basic observability
12. Backup/restore
13. CLI
14. Console

Do NOT attempt to ship every vector engine, every agent framework, every connector, every workflow system and every GPU strategy in the MVP.

---

# 38. What Should Be Deferred

Defer unless a real use case requires them:

```text
Milvus
ScyllaDB
large connector catalog
marketplace
A2A
index branching
vector migration
semantic cache
fine-tuning orchestration
multi-cluster
edge control plane
enterprise compliance bundles
advanced GPU scheduling
```

The architecture should support them without requiring them.

---

# 39. Technology Rules

Prefer:

- Go for control-plane services
- Rust only where latency/safety materially benefits
- PostgreSQL for metadata
- NATS JetStream for events
- Kubernetes for orchestration
- Cilium for networking
- OpenTelemetry for instrumentation
- Qdrant as initial vector engine
- SeaweedFS/S3-compatible storage initially
- OpenBao for secrets
- gVisor for normal sandboxing
- Firecracker/Kata for higher-risk workloads
- vLLM for initial inference

Do not introduce another technology merely because it is popular.

Every new infrastructure dependency should answer:

1. Why is it needed?
2. Why can't an existing component provide it?
3. What is its operational cost?
4. What is its license?
5. How does it affect installation?
6. How does it affect backups?
7. How does it affect upgrades?
8. How does it affect security?
9. How does it affect local development?

---

# 40. Licensing

The platform should prefer permissive licenses.

Before bundling or redistributing a component:

- verify current license
- record license
- record version
- record source
- check transitive dependencies
- track license changes

Source-available, copyleft or restrictive components should be isolated or optional where practical.

Never assume an old license table remains accurate.

---

# 41. Failure Semantics

Design for partial failure.

## API failure

Existing workloads should continue whenever possible.

## PostgreSQL failure

Running workloads should continue where possible; control-plane mutations may pause.

## NATS failure

Running workloads continue; event-driven operations should retry after recovery.

## Controller failure

Existing workloads continue; reconciliation resumes after recovery.

## Kubernetes API failure

Existing workloads should continue; new reconciliation may pause.

## Object storage failure

Agents must receive explicit, actionable failures rather than indefinite hangs.

## Model provider failure

Model Gateway should support timeout, retry and fallback policy where configured.

---

# 42. Reliability Rules

Controllers MUST be:

- idempotent
- retry-safe
- crash-safe
- observable

Handlers consuming events MUST be idempotent.

Long-running operations MUST survive process restarts.

Never assume an operation executes exactly once.

---

# 43. Testing Strategy

Every major component requires:

## Unit tests

Business logic and policy.

## Integration tests

Real dependencies where possible:

```text
Postgres
NATS
Kubernetes
Qdrant
Object storage
```

## End-to-end tests

Example:

```text
Create Project
  |
Create Bucket
  |
Create Vector DB
  |
Deploy Agent
  |
Bind Resources
  |
Invoke Agent
  |
Query Vector DB
  |
Call Model
  |
Inspect Trace
```

## Security tests

Must cover:

- tenant escape
- authorization bypass
- credential leakage
- sandbox escape
- SSRF
- network policy bypass
- secret exposure
- malicious image
- malicious plugin
- prompt injection
- tool abuse

## Restore tests

Every release should have automated or scheduled restore validation.

---

# 44. Architecture Decision Records

Important decisions MUST be recorded as ADRs.

Recommended ADRs:

```text
ADR-001 Control Plane Source of Truth
ADR-002 Kubernetes as Data Plane
ADR-003 PostgreSQL Metadata
ADR-004 NATS Event Architecture
ADR-005 Resource Reconciliation
ADR-006 Agent Identity
ADR-007 Credential Broker
ADR-008 gVisor Default Sandbox
ADR-009 MicroVM Isolation
ADR-010 Qdrant Default Vector Engine
ADR-011 Object Storage
ADR-012 Model Gateway
ADR-013 GPU Scheduling
ADR-014 Workflow Engine Boundary
ADR-015 Multi-Tenancy
ADR-016 Plugin Security
ADR-017 Backup/Restore
ADR-018 API Versioning
ADR-019 Local Development
ADR-020 MVP Scope
```

---

# 45. Repository Structure

Recommended structure:

```text
archon-base/
│
├── cmd/
│   ├── archon/
│   ├── archon-api/
│   ├── archon-controller/
│   └── archon-worker/
│
├── internal/
│   ├── iam/
│   ├── policy/
│   ├── projects/
│   ├── resources/
│   ├── reconciliation/
│   ├── bindings/
│   ├── credentials/
│   ├── events/
│   ├── audit/
│   ├── usage/
│   └── operations/
│
├── api/
│   ├── openapi/
│   └── proto/
│
├── sdk/
│   ├── python/
│   ├── typescript/
│   └── go/
│
├── operators/
│   ├── database/
│   ├── agent/
│   ├── storage/
│   └── model/
│
├── runtime/
│   ├── sandbox/
│   ├── python/
│   └── streaming/
│
├── gateway/
│
├── console/
│
├── charts/
│
├── deploy/
│
├── configs/
│
├── docs/
│   ├── architecture/
│   ├── security/
│   ├── operations/
│   ├── api/
│   └── adr/
│
├── tests/
│   ├── integration/
│   ├── e2e/
│   ├── security/
│   └── recovery/
│
└── scripts/
```

The actual repository may differ, but architectural separation should remain.

---

# 46. Coding Rules for AI Agents

When modifying code:

1. Read the relevant architecture documentation first.
2. Identify the resource/lifecycle affected.
3. Preserve tenant boundaries.
4. Preserve idempotency.
5. Add tests for changed behavior.
6. Do not bypass controllers with ad-hoc Kubernetes scripts.
7. Do not introduce global mutable state.
8. Do not store credentials in source.
9. Do not log secrets.
10. Do not weaken network policy for convenience.
11. Do not add cluster-admin permissions unless explicitly justified.
12. Do not use `latest` image tags.
13. Do not make external services mandatory without architectural approval.
14. Prefer interfaces over provider-specific implementations.
15. Keep provider-specific logic inside adapters.
16. Update documentation when changing public behavior.
17. Add an ADR when making a significant architectural decision.

---

# 47. AI Coding Agent Workflow

For every substantial task:

```text
1. Understand
   |
2. Locate architecture boundary
   |
3. Identify affected resources
   |
4. Identify security implications
   |
5. Implement smallest correct change
   |
6. Add tests
   |
7. Run validation
   |
8. Inspect logs/errors
   |
9. Update documentation
   |
10. Summarize architectural impact
```

Before implementing a new subsystem, answer:

```text
What problem does this solve?
Who owns its state?
What is the source of truth?
How is it reconciled?
How is it secured?
How does it fail?
How is it backed up?
How is it upgraded?
How is it observed?
How is it tested?
Can an existing component provide it?
```

---

# 48. Definition of Done

A feature is NOT complete merely because code compiles.

For platform features, Definition of Done includes:

- API implemented
- authorization implemented
- tenant isolation verified
- resource lifecycle defined
- reconciliation implemented
- idempotency handled
- status conditions exposed
- metrics/tracing added
- logs added
- failure behavior defined
- tests added
- security tests added where relevant
- backup implications considered
- upgrade implications considered
- documentation updated

---

# 49. Security Definition of Done

A security-sensitive feature must answer:

```text
Who can call it?
What identity is used?
What permissions are required?
What secrets are involved?
What network access is required?
What happens if credentials are compromised?
What is logged?
What is redacted?
Can another tenant reach it?
Can the workload reach Kubernetes?
Can it reach host resources?
Can it reach cloud metadata?
```

---

# 50. Operational Definition of Done

A production feature must define:

```text
Health check
Readiness
Metrics
Logs
Tracing
Timeout
Retry
Backoff
Failure state
Recovery
Backup
Restore
Upgrade
Rollback
```

---

# 51. Success Metrics

Initial targets:

- installation to first agent: under 15 minutes
- sandboxed function cold start: target p95 under 1 second excluding large model loading
- vector query overhead: target within approximately 10% of native engine where the abstraction is used
- zero cross-tenant isolation findings in security review
- restore drills pass
- agent eval regression gates can be used in production CI

Additional operational metrics:

- API availability
- controller reconciliation latency
- event processing latency
- agent deployment success rate
- backup success rate
- restore success rate
- GPU utilization
- GPU memory utilization
- inference latency
- inference error rate
- tool failure rate
- policy denial rate
- sandbox failure rate
- mean time to recovery

---

# 52. Phase Plan

## Phase 0 — Foundations

Build:

- repository
- CI
- signing
- release pipeline
- Go control plane
- PostgreSQL
- NATS
- authentication
- organizations/projects
- resource API
- transactional outbox
- reconciliation framework
- single-node installation

## Phase 1 — MVP

Build:

- Qdrant adapter
- S3-compatible storage
- OpenBao integration
- agent identity
- bindings
- credential broker
- gVisor runtime
- Python agents
- model gateway
- vLLM integration
- CLI
- basic console
- basic observability
- backup/restore
- local development

## Phase 2 — Agent Platform

Add:

- knowledge pipelines
- Git/S3/HTTP connectors
- event triggers
- cron
- durable workflow integration
- tracing UI
- evals
- replay
- MCP hosting
- KEDA
- microVM isolation
- GPU pools

## Phase 3 — Production Hardening

Add:

- HA Kubernetes
- advanced restore
- additional vector engines
- advanced policies
- guardrails
- approvals
- quotas
- budgets
- semantic cache
- model registry
- dataset versioning

## Phase 4 — Ecosystem

Add:

- plugin SDK
- marketplace
- templates
- A2A
- multi-cluster
- edge
- SSO/SCIM
- SIEM
- compliance
- air-gapped bundles
- fine-tuning orchestration

---

# 53. What Not To Do

Do NOT:

- build a custom vector database
- build a custom object store
- build a custom workflow engine without strong justification
- create a custom container runtime
- make every integration mandatory
- expose Kubernetes admin credentials to agents
- use shared credentials for agents
- bypass policy for internal services
- store prompts/responses indefinitely by default
- treat `gpu: 1` as sufficient GPU scheduling metadata
- deploy unsigned production images
- rely on UI-only authorization
- make production behavior fundamentally different from the API model
- add a technology simply because it is fashionable
- expand MVP scope without a concrete use case

---

# 54. Architectural North Star

Archon Base should ultimately provide this developer experience:

```text
archon init my-agent
cd my-agent

archon db create knowledge --engine qdrant
archon bucket create documents
archon model deploy llama
archon agent deploy .
```

Then:

```text
Developer
    |
    v
Archon API
    |
    +--> Identity
    +--> Policy
    +--> Resource Lifecycle
    +--> Credential Broker
    |
    v
Agent
    |
    +--> Vector
    +--> Storage
    +--> Model
    +--> Tools
    +--> Memory
```

The developer should not need to understand the underlying Kubernetes objects to use the platform.

Platform engineers must still be able to access the underlying infrastructure when needed.

---

# 55. Final Architectural Principle

Archon Base should not attempt to replace Kubernetes, databases, inference engines, storage engines, workflow engines, or AI frameworks.

It should become the **control plane that makes these components feel like one coherent AI platform**.

The core value is:

```text
One API
One CLI
One Console
One Identity Model
One Policy Model
One Binding Model
One Resource Model
One Observability Model
One Developer Experience

          ↓

Many underlying engines
```

That is the architectural boundary that should guide the project.
