# Archon Base Implementation TODO

**Goal:** Ship a reliable Phase 1 local vertical slice before expanding into Kubernetes, multiple engines, or enterprise features.

**Definition of done for Phase 1:** A developer can start Archon locally, create a Project, create a Qdrant DatabaseInstance, deploy the example Python agent, invoke it, inspect status/logs, and tear everything down without Kubernetes.

## 0. Repository and engineering foundation

- [x] Correct the Go module path and internal imports.
- [x] Add Apache-2.0 license.
- [x] Add core v1 resource schema document.
- [x] Add PostgreSQL migrations for Projects, DatabaseInstances, and outbox events.
- [x] Add Phase 1 local execution specification.
- [x] Add baseline security, threat-modeling, monitoring, and incident-response guidance.
- [ ] Add GitHub Actions CI.
  - [ ] `go test ./...`
  - [ ] `go vet ./...`
  - [ ] `make fmt-check`
  - [ ] Helm lint/template validation
  - [ ] Python package checks
  - [ ] YAML and JSON schema validation
- [ ] Add a migration runner and `schema_migrations` table.
- [ ] Add dependency pinning and vulnerability scanning.
- [ ] Add a `NOTICE`/third-party attribution process.

## 1. Core v1 API contract

- [x] Expand Go API types to match `docs/core-v1-resource-schema.md`.
  - [x] Common metadata: UID, generation, labels, annotations.
  - [x] Common status: phase, conditions, observed generation, message.
  - [x] Project spec and quota types.
  - [x] DatabaseInstance storage, backup, and network types.
- [x] Add strict JSON decoding and unknown-field rejection.
- [x] Add validation for names, environments, engines, plans, quantities, replicas, and shards.
- [ ] Define typed API errors with stable codes and field paths.
- [ ] Add request IDs and structured JSON logging.
- [ ] Add HTTP server timeouts and graceful shutdown.
- [ ] Add `/livez` and dependency-aware `/readyz`.
- [x] Implement Project endpoints:
  - [x] `POST /v1/projects`
  - [x] `GET /v1/projects/{project}`
  - [x] `PATCH /v1/projects/{project}`
  - [x] `DELETE /v1/projects/{project}`
- [x] Implement DatabaseInstance endpoints:
  - [x] `POST /v1/projects/{project}/databases`
  - [x] `GET /v1/projects/{project}/databases/{name}`
  - [x] `LIST /v1/projects/{project}/databases`
  - [x] `PATCH /v1/projects/{project}/databases/{name}`
  - [x] `DELETE /v1/projects/{project}/databases/{name}`
- [ ] Support idempotency keys for mutating requests.
- [x] Add API unit tests.
- [x] Add API handler and routing tests.
- [x] Add PostgreSQL repository integration tests using testcontainers-go.

## 2. PostgreSQL control plane

- [ ] Add a database package with bounded connection pooling.
- [x] Implement Project repository methods.
- [x] Implement DatabaseInstance repository methods.
- [x] Enforce parent/child ownership and soft-deletion rules.
- [x] Persist spec snapshots and status transitions.
- [x] Increment generation on desired-state changes.
- [ ] Add optimistic concurrency/resource-version checks.
- [x] Implement transactional outbox writes for every resource mutation.
- [ ] Add outbox publisher interface with retry/backoff behavior.
- [ ] Add database integration tests using disposable PostgreSQL.
- [ ] Document migration and rollback procedures.

## 3. Local development runtime

- [ ] Update Compose with pinned versions and health checks.
- [ ] Add a local PostgreSQL readiness check.
- [ ] Add a local Qdrant readiness check.
- [ ] Implement `archon dev` startup orchestration.
  - [ ] Create `.archon/` with mode `0700`.
  - [ ] Start required Compose services.
  - [ ] Apply migrations.
  - [ ] Start the API.
  - [ ] Poll `/livez` and `/readyz` with a bounded timeout.
  - [ ] Write PID and context metadata.
- [ ] Implement `archon dev down` cleanup behavior.
- [ ] Ensure startup failure cleans up child processes.
- [ ] Keep local services loopback-only by default.
- [ ] Add an end-to-end smoke test for startup and teardown.

## 4. Qdrant DatabaseInstance adapter

- [ ] Define the engine adapter interface.
- [ ] Implement the Qdrant health client.
- [ ] Implement deterministic project/resource namespace derivation.
- [ ] Create the local collection/namespace during reconciliation.
- [ ] Make reconciliation idempotent.
- [ ] Record `Pending`, `Ready`, `Degraded`, and `Failed` conditions.
- [ ] Add bounded retries and exponential backoff.
- [ ] Add query/upsert/delete binding operations.
- [ ] Add Qdrant adapter conformance tests.
- [ ] Keep unsupported engines rejected in Phase 1 with typed errors.

## 5. Python agent runtime

- [ ] Finalize the agent manifest schema.
- [ ] Validate runtime, entrypoint, resources, triggers, and bindings.
- [ ] Implement local trusted-subprocess execution.
  - [ ] Temporary working directory.
  - [ ] Environment-variable allowlist.
  - [ ] Configurable timeout.
  - [ ] Captured stdout/stderr.
  - [ ] Request ID propagation.
  - [ ] Explicit trusted-code-only warning.
- [ ] Implement Python SDK request context.
- [ ] Implement vector binding client against managed Qdrant.
- [ ] Implement OpenAI-compatible model client.
- [ ] Implement streaming response handling.
- [ ] Add agent crash, timeout, and malformed-response handling.
- [ ] Add SDK tests and an example invocation test.
- [ ] Do not claim sandbox isolation until gVisor or microVM execution exists.

## 6. CLI and developer experience

- [ ] Replace command stubs with a command framework.
- [ ] Implement `archon init`.
- [ ] Implement `archon project create/get/delete`.
- [ ] Implement `archon db create/get/list/delete`.
- [ ] Implement `archon agent deploy/invoke/logs`.
- [ ] Implement `archon dev` and `archon dev down`.
- [ ] Add `--output json` for scripting.
- [ ] Add clear errors and non-zero exit codes.
- [ ] Add local context selection and API endpoint configuration.
- [ ] Add shell completion later, after command shapes stabilize.

## 7. Security baseline before public use

- [ ] Add authentication to the API.
- [ ] Add project-scoped authorization.
- [ ] Add API key/service-account lifecycle.
- [ ] Add input size limits and rate limits.
- [ ] Keep PostgreSQL, Qdrant, and internal runtimes private.
- [ ] Redact secrets, prompts, signed URLs, and request bodies from logs.
- [ ] Add audit events for authentication, resource mutations, bindings, and secret access.
- [ ] Add egress policy abstraction.
- [ ] Define the production sandbox boundary.
- [ ] Add image signing, SBOM generation, and image scanning.
- [ ] Run a cross-project authorization test suite.

## 8. Observability and operations

- [ ] Add metrics for API requests, latency, errors, and readiness.
- [ ] Add resource reconciliation metrics.
- [ ] Add outbox backlog and publish-failure metrics.
- [ ] Add Qdrant health and query metrics.
- [ ] Add agent invocation, timeout, crash, and dependency metrics.
- [ ] Add trace/request correlation.
- [ ] Add backup and restore runbooks.
- [ ] Add failure-injection tests for PostgreSQL, Qdrant, and agent processes.
- [ ] Add operational dashboards after metric names stabilize.

## 9. Kubernetes and production expansion

Start only after the Phase 1 local definition of done is met.

- [ ] Define CRD projection from the core v1 resource model.
- [ ] Implement controller/reconciler interfaces shared with local mode.
- [ ] Implement transactional outbox publisher to NATS JetStream.
- [ ] Add Qdrant Kubernetes deployment/operator integration.
- [ ] Add resource requests, limits, security contexts, and NetworkPolicies.
- [ ] Add liveness/readiness/startup probes.
- [ ] Add Secrets/ConfigMaps and external dependency configuration.
- [ ] Add PodDisruptionBudgets and migration Jobs.
- [ ] Add backup/restore controllers.
- [ ] Add gVisor runtime integration.
- [ ] Add HA installation documentation and tested Helm releases.

## 10. Later platform capabilities

- [ ] Object storage and signed URLs.
- [ ] Model gateway and token accounting.
- [ ] Knowledge/RAG pipelines.
- [ ] Tracing and evaluation suites.
- [ ] Guardrails and human approvals.
- [ ] MCP server hosting.
- [ ] Budgets and quotas.
- [ ] Additional vector engines.
- [ ] Multi-cluster data planes.
- [ ] SSO/SCIM, BYOK, compliance exports, and air-gapped bundles.

## Current next action

Implement the first unchecked items in Sections 1–3:

1. Add typed v1 Go resource models and validation.
2. Add the PostgreSQL repository and migration runner.
3. Add API tests for Project and DatabaseInstance lifecycle.
4. Add CI so every subsequent implementation is checked automatically.
