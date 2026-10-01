# Archon Base Core v1 Resource Schema

**Status:** draft for Phase 1

This document defines the smallest durable resource contract shared by the REST API, CLI, local emulator, SDKs, and future Kubernetes controllers. The API uses `apiVersion: archon.dev/v1`.

## Common envelope

Every resource has this envelope:

```yaml
apiVersion: archon.dev/v1
kind: Project | DatabaseInstance
metadata:
  name: string              # DNS-compatible, immutable after creation
  project: string           # required for namespaced resources
  labels: {}
  annotations: {}
spec: {}
status:
  phase: Pending | Ready | Degraded | Failed | Deleting
  observedGeneration: integer
  conditions: []
  message: string
  lastReconciledAt: RFC3339 timestamp
```

`metadata.name` is the user-facing stable name. The server assigns an opaque UUID and stores it as `metadata.uid` in API responses. `spec` is desired state; `status` is controller-observed state and is never accepted as authoritative input.

## Project

```yaml
apiVersion: archon.dev/v1
kind: Project
metadata:
  name: support-bot
  labels: {}
  annotations: {}
spec:
  displayName: Support Bot
  description: Customer support resources
  environment: dev
  region: local
  quotas:
    cpu: "4"
    memory: 8Gi
    storage: 100Gi
    vectorPoints: 1000000
    tokensPerMonth: 0
status:
  phase: Ready
  observedGeneration: 1
```

Rules:

- `metadata.name` is unique within an organization/tenant; Phase 1 uses one implicit local organization.
- `spec.displayName` is required and may be changed.
- `spec.environment` defaults to `dev`; allowed values are `dev`, `staging`, and `prod`.
- Quota values are optional. Zero means unlimited only when explicitly configured by the server; otherwise the server applies installation defaults.
- Project deletion is soft-deletion first and is blocked while child resources exist unless `force` is explicitly supported later.

## DatabaseInstance

```yaml
apiVersion: archon.dev/v1
kind: DatabaseInstance
metadata:
  name: docs-index
  project: support-bot
  labels: {}
  annotations: {}
spec:
  engine: qdrant
  version: "1.12"
  plan: standard-2
  replicas: 1
  shards: 1
  storage:
    class: local
    size: 20Gi
  backup:
    schedule: ""
    retentionDays: 7
    target: ""
  network:
    exposure: private
status:
  phase: Pending
  observedGeneration: 1
```

Rules:

- `metadata.project` must reference an existing Project.
- `spec.engine` is required. Phase 1 supports only `qdrant`; other engines are rejected with a typed validation error.
- `spec.plan` is required and maps to local resource defaults. `standard-2` is the initial development plan.
- `replicas` and `shards` must be positive integers; Phase 1 accepts `1` only and reserves higher values for a controller implementation.
- `storage.size` is required and must be a Kubernetes-style quantity such as `20Gi`.
- `network.exposure` defaults to `private`; Phase 1 does not expose instances publicly.
- `status.phase` is controller-owned. Failed reconciliation must preserve the last error in `status.message` and a condition.

## Condition shape

```yaml
conditions:
  - type: Ready | Reconciled | Error
    status: "True" | "False" | "Unknown"
    reason: string
    message: string
    observedGeneration: integer
    lastTransitionTime: RFC3339 timestamp
```

## API behavior

- `POST /v1/projects` creates a Project and returns `201 Created`.
- `GET /v1/projects/{project}` returns a Project.
- `PATCH /v1/projects/{project}` updates mutable spec fields and increments `metadata.generation`.
- `DELETE /v1/projects/{project}` requests deletion and returns `202 Accepted` while children are being removed.
- `POST /v1/projects/{project}/databases` creates a DatabaseInstance and returns `202 Accepted` because provisioning is asynchronous.
- `GET /v1/projects/{project}/databases/{name}` returns desired and observed state.
- `DELETE /v1/projects/{project}/databases/{name}` requests teardown and returns `202 Accepted`.

All mutating requests must support an idempotency key. Resource writes must be authorized before persistence and must emit an outbox event in the same PostgreSQL transaction.

## Compatibility and versioning

- Unknown `spec` fields are rejected in strict API mode so typos do not silently change behavior.
- New optional fields may be added within v1.
- Removing or changing field meaning requires `archon.dev/v2`.
- Database rows store the submitted spec as JSONB so the API can preserve forward-compatible fields while typed validation remains the gatekeeper.
