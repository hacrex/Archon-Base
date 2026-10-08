# Archon Base Codebase Audit and Completion Plan

**Audit date:** 2026-10-08  
**Branch:** `main`  
**Audited commit:** `99cf247 Wire PostgreSQL startup and modularize control plane`

## 1. Executive summary

Archon Base has a useful foundation, but it is not yet the Supabase/Appwrite-style product described by the long-term documentation.

The current shipped implementation is:

```text
Go resource models and validation
        ↓
HTTP CRUD handlers
        ↓
PostgreSQL repository + migration runner
        ↓
Local preview dashboard
        ↓
Student K3s installer + resource monitor
```

The following major layers are still missing:

- Authentication and authorization
- A real CLI workflow
- Qdrant provisioning/reconciliation
- A real persistent web console connected to the API
- Agent execution
- Model gateway
- Kubernetes controllers
- Audit/event consumers
- Production deployment hardening

The correct next goal is not to implement every feature in the architecture document. The correct next goal is a small, demonstrable vertical slice:

```text
Start local dependencies → create Project → create DatabaseInstance
→ reconcile one Qdrant collection → show real status in the dashboard
```

## 2. Audit status

### Repository state

- Working tree clean.
- Local `main` is synchronized with `origin/main`.
- Current commit is `99cf247`.
- Approximately 10,000 lines are present, including design documents, marketing material, code, tests, and deployment files.

### Checks run

Passed:

- `go test ./... -count=1`
- `go vet ./...`
- `go build ./...`
- Go formatting check
- JavaScript syntax checks
- JSON route manifest validation
- `bash -n scripts/install-k3s-student.sh`
- Python bytecode compilation for the SDK and example
- `git diff --check`

Not run because the tools are unavailable in this environment:

- Docker Compose validation: Docker CLI unavailable.
- Helm lint/template validation: Helm CLI unavailable.
- Kubernetes deployment test: Kubernetes CLI/cluster unavailable.

Passing tests prove compilation and unit behavior. They do **not** prove a complete end-to-end product workflow.

## 3. What is implemented today?

| Area | Status | Notes |
|---|---|---|
| Go module and license | Available | Module path and Apache-2.0 license exist. |
| Project model | Available | Validation, normalization, JSON decoding, PostgreSQL persistence. |
| DatabaseInstance model | Available | Validation, normalization, PostgreSQL persistence. |
| HTTP CRUD handlers | Available | Project and DatabaseInstance handlers exist. |
| PostgreSQL migration runner | Available | Ordered forward migrations and `schema_migrations`. |
| Transactional outbox writes | Available | Resource mutations write outbox rows. No publisher yet. |
| API startup wiring | Available with configuration | Set `ARCHON_DB_URL`; otherwise resource routes remain unavailable. |
| Student K3s installer | Available | Prerequisite checks, dry-run, explicit confirmation. |
| Resource monitor | Available | Local JSON, health, and Prometheus endpoints. |
| Control-plane UI | Preview | Modular shell, but still uses local hardcoded data. |
| CLI | Stub | `version` works; most commands return not implemented. |
| Qdrant adapter | Missing | No health client, collection provisioning, or reconciliation. |
| Agent runtime | Missing | Python SDK is a shape/stub, not an executable runtime. |
| Authentication/RBAC | Missing | No identity, sessions, permissions, or project isolation. |
| Audit log | Missing | UI boundary exists; durable authenticated audit is not implemented. |
| Kubernetes controllers | Missing | Helm chart is only an early deployment skeleton. |

## 4. Important findings

### P0: the first end-to-end product workflow is incomplete

The repository documents this target flow:

```bash
archon dev
archon project create support-bot --display-name "Support Bot"
archon db create docs-index --project support-bot --engine qdrant --plan standard-2 --storage 20Gi
archon agent deploy ...
archon agent invoke ...
```

But the CLI commands are still stubs, Qdrant reconciliation is absent, and the agent runtime is absent. The documented Phase 1 definition of done is therefore not met.

### P0: the dashboard is not connected to the API

The control-plane UI presents Projects, Resources, Operations, Admin, Audit, Settings, and allocation views, but the data is local preview state. Buttons do not persist changes and the UI does not fetch the Go API.

The next UI work must connect only the Project and DatabaseInstance routes first. Do not expand the dashboard surface until those two resources are real.

### P1: API readiness and lifecycle behavior are incomplete

The API currently has `/healthz` and `/v1/version`, but the planned `/livez` and dependency-aware `/readyz` endpoints are missing.

The HTTP server also lacks:

- Read, write, idle, and header timeouts.
- Graceful shutdown on SIGTERM/SIGINT.
- Request IDs and structured request logs.
- Dependency health reporting.
- A consistent API error envelope with field-level details.

### P1: API binds broadly by default

`ARCHON_API_ADDR` defaults to `:8080`, which binds to all interfaces on many systems. The local execution specification says loopback should be the default.

Recommended default:

```text
127.0.0.1:8080
```

Use an explicit warning when an operator chooses `0.0.0.0`.

### P1: the API does not list Projects

There is `POST /v1/projects` and individual Project operations, but no `GET /v1/projects` collection endpoint. A real dashboard cannot load the organization project list without this endpoint.

Add:

```text
GET /v1/projects
```

with pagination, filtering, and stable ordering before connecting the Projects view.

### P1: migration runner needs multi-instance hardening

The migration runner checks whether a migration exists and then inserts the version. Two API instances starting simultaneously can race.

Before multi-replica deployment:

- Use a PostgreSQL advisory lock around migration execution, or
- Use a locking row/transaction strategy, and
- Test concurrent startup.

### P1: Compose is not yet a reliable local runtime

The Compose file has the main services, but:

- Images are not consistently pinned to immutable versions.
- There are no service health checks.
- `depends_on` does not wait for service readiness.
- Qdrant is exposed directly on the host.
- The monitor is not included in the stack.
- NATS is configured but no application consumer/publisher uses it.
- There is no automatic smoke test for startup, migration, CRUD, and teardown.

### P1: Helm deployment is only a skeleton

The Helm Deployment currently has a readiness probe on `/healthz`, but the product needs dependency-aware readiness. It does not show a migration Job, database configuration Secret, PostgreSQL dependency, NATS wiring, resource limits, security context, NetworkPolicy, service account, or controller deployment.

Do not call the Helm chart production-ready until those pieces are implemented and tested with a real cluster or a reproducible local Kubernetes environment.

### P2: examples and SDK are ahead of the implementation

The example AgentFunction manifest contains:

- Source
- Triggers
- Resources
- Bindings
- Secrets
- Egress

But the Go `AgentFunctionSpec` currently validates only:

- Runtime
- Entrypoint
- Isolation

The Python SDK exposes vector and LLM client shapes but raises `NotImplementedError` at runtime.

Either label the full example as target design, or reduce it to the fields actually implemented. Do not imply that the current example can deploy.

### P2: frontend modularization is structural but not yet easy to maintain

The frontend now has modules, which is an improvement, but `src/routes.js` still contains very large one-line template renderers. New contributors will still struggle to edit individual pages.

Next refactor:

```text
src/routes/overview.js
src/routes/projects.js
src/routes/resources.js
src/routes/operations.js
src/routes/admin.js
src/routes/audit.js
src/routes/settings.js
src/components/status-badge.js
src/components/data-table.js
src/components/empty-state.js
```

### P2: duplicated and stale roadmap sources

There are two large TODO systems:

- `todo.md` — Phase 1 engineering implementation.
- `PRODUCT_CONTROL_PLANE_TODO.md` — product/control-plane roadmap.

Some items are stale or duplicated. For example:

- Repository integration tests are checked in one section but still unchecked in another.
- Migration runner work is complete, but the “current next action” still describes it as future work.
- The UI plan still describes the old monolithic `app.js` responsibility.

Choose one source of truth for active work and keep the other as a product roadmap.

## 5. TODO reconciliation

### Root `todo.md`

#### Completed

- Go module foundation.
- License and core schema documentation.
- PostgreSQL schema and migration runner.
- Project and DatabaseInstance models.
- Strict decoding and validation.
- CRUD handlers.
- Repository and outbox writes.
- API tests.
- Testcontainers integration test.
- Student K3s installer.
- Local monitor.

#### Still required for Phase 1

1. CI and automated artifact validation.
2. Typed API errors and field paths.
3. Request IDs and structured logs.
4. HTTP timeouts and graceful shutdown.
5. `/livez` and `/readyz`.
6. Bounded database pool configuration.
7. Optimistic concurrency/resource versions.
8. Idempotency keys.
9. Migration/rollback operational documentation.
10. Compose health checks and pinned images.
11. `archon dev` and `archon dev down`.
12. Qdrant adapter and reconciliation.
13. Local trusted agent subprocess runtime.
14. Real CLI project/database/agent commands.
15. Authentication and authorization.
16. Audit and metrics.
17. End-to-end smoke tests.

### `PRODUCT_CONTROL_PLANE_TODO.md`

#### UI foundation mostly present

- Basic shell.
- Responsive layout.
- Preview state labels.
- Project capacity preview.
- Resource allocation preview.
- Admin/audit/settings information architecture.

#### Product-control-plane work still required

1. Stable API contracts and OpenAPI.
2. Authentication and sessions.
3. Organization/project/member hierarchy.
4. RBAC and project-scoped authorization.
5. Real dashboard API integration.
6. Resource detail and condition timelines.
7. Cluster connection and installation flows.
8. Reconciliation state machine.
9. Capacity validation before provisioning.
10. Audit and privileged support tools.
11. Model, agent, storage, and governance modules.
12. Production Kubernetes installation and upgrades.

## 6. Recommended completion order

Do these phases in order. Do not jump to models, MCP, marketplace, or enterprise features before Phase 4 is complete.

### Phase 0 — Freeze the product scope

**Goal:** Make the project easy to understand.

Tasks:

- Declare the first product milestone: `Project → DatabaseInstance → Qdrant`.
- Mark all other modules as Planned.
- Make `todo.md` the engineering source of truth.
- Make `PRODUCT_CONTROL_PLANE_TODO.md` the product roadmap.
- Remove stale “current next action” text.
- Update README examples so they only show executable behavior.
- Add a capability matrix with Available, Preview, Planned, and Unsupported.

Exit criteria:

- A new contributor can understand the current product in 10 minutes.
- Every README command either works or is explicitly marked planned.

### Phase 1 — Make the API production-shaped

**Goal:** Make the control-plane API reliable before adding more features.

Tasks:

1. Change default bind address to `127.0.0.1:8080`.
2. Add `GET /livez`.
3. Add dependency-aware `GET /readyz`.
4. Add HTTP server timeouts.
5. Add signal handling and graceful shutdown.
6. Add request ID middleware.
7. Add structured JSON request logs.
8. Add stable error codes and field paths.
9. Add `GET /v1/projects`.
10. Add pagination/filter/sort conventions.
11. Configure bounded PostgreSQL pool size, idle time, and lifetime.
12. Add optimistic concurrency using generation/resource version.
13. Add idempotency keys to POST/PATCH mutations.
14. Add API contract tests.
15. Publish an initial OpenAPI document.

Exit criteria:

- API starts, becomes ready only when PostgreSQL is available, and shuts down cleanly.
- A client can list and mutate Projects and DatabaseInstances safely.
- Concurrent requests cannot silently overwrite each other.

### Phase 2 — Make local development one command

**Goal:** A student can start and stop the product without understanding infrastructure internals.

Tasks:

1. Pin PostgreSQL, Qdrant, NATS, and API image versions.
2. Add health checks to every Compose service.
3. Keep services loopback-only by default.
4. Implement `archon dev`.
5. Create `.archon/` with mode `0700`.
6. Start required Compose services.
7. Apply migrations.
8. Start the API with generated local configuration.
9. Poll `/livez` and `/readyz` with a finite timeout.
10. Write PID/context metadata.
11. Print exact URLs and next commands.
12. Implement `archon dev down`.
13. Clean up child processes on startup failure.
14. Add one smoke test for startup and teardown.

Exit criteria:

```bash
archon dev
archon project create support-bot --display-name "Support Bot"
archon db create docs-index --project support-bot --engine qdrant --plan standard-2 --storage 20Gi
archon dev down
```

works on a clean supported machine.

### Phase 3 — Implement the Qdrant vertical slice

**Goal:** Turn a DatabaseInstance row into a real data-plane resource.

Tasks:

1. Define the engine adapter interface.
2. Implement Qdrant health checking.
3. Derive deterministic project/resource collection names.
4. Create the collection when a DatabaseInstance is created.
5. Make reconciliation idempotent.
6. Update `Pending`, `Ready`, `Degraded`, and `Failed` conditions.
7. Add bounded retry with exponential backoff.
8. Record observed generation.
9. Add query/upsert/delete binding operations.
10. Add adapter conformance tests.
11. Add failure tests for unavailable Qdrant.
12. Add outbox consumer or local reconciler loop.

Exit criteria:

- Creating `docs-index` creates one real Qdrant collection.
- Repeating reconciliation does not create duplicates.
- Qdrant failure is visible without taking down the API.
- The API returns accurate observed status.

### Phase 4 — Connect the dashboard to real state

**Goal:** Replace hardcoded preview data with the working API.

Tasks:

1. Add a small API client module.
2. Load Projects from `GET /v1/projects`.
3. Load DatabaseInstances from the project collection endpoint.
4. Add loading, empty, error, and degraded states.
5. Implement Create Project form.
6. Implement Create DatabaseInstance form.
7. Add Project detail page.
8. Add DatabaseInstance detail page.
9. Add conditions and operation timeline.
10. Add safe delete confirmations.
11. Keep unsupported routes clearly marked Preview/Planned.
12. Split the remaining large route templates into separate files.
13. Add browser smoke tests for overview, projects, resources, and error states.

Exit criteria:

- Refreshing the dashboard shows database-backed resources.
- Creating a resource through the UI changes the API and appears in the UI.
- No dashboard number is fake or hardcoded for the supported slice.

### Phase 5 — Implement the trusted local agent runtime

**Goal:** Complete the local Phase 1 AI application path without falsely claiming sandbox security.

Tasks:

1. Finalize AgentFunction manifest schema.
2. Decide which fields are supported in v1.
3. Validate runtime, entrypoint, resources, triggers, bindings, and egress.
4. Implement local subprocess execution.
5. Use a temporary working directory.
6. Apply an environment-variable allowlist.
7. Add timeout and cancellation.
8. Capture stdout/stderr.
9. Propagate request IDs.
10. Implement Python request context.
11. Implement Qdrant vector binding client.
12. Implement OpenAI-compatible model client.
13. Add invocation history and logs.
14. Add crash, timeout, and malformed-response handling.
15. Clearly state that trusted subprocess mode is not a security boundary.

Exit criteria:

```bash
archon agent deploy examples/support-agent --project support-bot
archon agent invoke support-agent --data '{"message":"hello"}'
```

works locally with documented limitations.

### Phase 6 — Authentication, authorization, and audit

**Goal:** Make the product safe for multiple users and projects.

Tasks:

1. Define organization, project, member, role, and service-account models.
2. Add local development authentication.
3. Add OIDC/OAuth support.
4. Add secure sessions and cookies.
5. Add CSRF protection where needed.
6. Add login rate limiting.
7. Define owner/admin/operator/developer/viewer roles.
8. Enforce project-scoped authorization on every resource route.
9. Add API key/service-account lifecycle.
10. Add secret redaction.
11. Add durable audit events for auth and mutations.
12. Add audit search and export.
13. Add authorization tests across organizations and projects.
14. Add reauthentication for sensitive actions.

Exit criteria:

- A user cannot read or mutate another project.
- Every mutation has an actor, request ID, timestamp, and auditable result.
- Secrets never appear in API responses or logs.

### Phase 7 — Kubernetes/K3s installation and controllers

**Goal:** Move from local mode to supported self-hosted clusters.

Tasks:

1. Define local versus cluster deployment profiles.
2. Add cluster connection wizard.
3. Validate Kubernetes API access and permissions.
4. Validate K3s/Kubernetes version.
5. Validate storage class and ingress/load balancer.
6. Add namespace and ownership labels.
7. Install CRDs/controllers idempotently.
8. Share reconciliation interfaces with local mode.
9. Add Qdrant Kubernetes deployment integration.
10. Add resource requests/limits and security contexts.
11. Add NetworkPolicies.
12. Add migration Job or controlled migration init flow.
13. Add liveness/readiness/startup probes.
14. Add upgrade and rollback documentation.
15. Add backup/restore procedure.
16. Run Helm lint/template and cluster smoke tests in CI.

Exit criteria:

- A clean K3s cluster can install Archon Base.
- A Project and DatabaseInstance can reconcile into Kubernetes.
- Upgrade and rollback are tested.

### Phase 8 — Production hardening

**Goal:** Make the platform operable and secure.

Tasks:

- Add image signing, SBOM, and vulnerability scanning.
- Pin all images and dependencies.
- Add NetworkPolicies and least-privilege service accounts.
- Add encrypted secret storage or an external secret manager.
- Add metrics for API, DB, reconciliation, outbox, Qdrant, and agents.
- Add tracing and request correlation.
- Add backup/restore tests.
- Add failure-injection tests.
- Add rate limiting and abuse controls.
- Add resource quotas and capacity warnings.
- Add maintenance mode.
- Add incident response runbooks.
- Add supported-version and upgrade policy.
- Add release automation and signed artifacts.

Exit criteria:

- Security review completed.
- Restore drill completed.
- Failure scenarios have documented detection and recovery.
- Production deployment is reproducible from a tagged release.

### Phase 9 — Expand the AI platform

Only after the previous phases are stable:

1. Model and ModelEndpoint resources.
2. Ollama/vLLM adapters.
3. External provider configuration.
4. Object storage and signed URLs.
5. Dataset and artifact lifecycle.
6. Knowledge/RAG pipelines.
7. Tracing and evaluations.
8. Guardrails and approvals.
9. MCP hosting.
10. Budgets and token accounting.
11. Multi-cluster support.
12. SSO/SCIM and enterprise features.

Each module should follow the same pattern:

```text
resource schema → API → persistence → reconciliation → UI → auth → audit → tests → docs
```

## 7. CI checklist to add immediately

Create GitHub Actions with jobs for:

```text
go test ./...
go vet ./...
make fmt-check
go build ./...
node --check control-plane-web/app.js
node --check control-plane-web/src/*.js
python package checks
YAML validation
JSON validation
Docker Compose config
Helm lint/template
```

Add separate integration jobs for:

- PostgreSQL/testcontainers.
- Qdrant.
- Full local smoke test.
- Kubernetes/Helm smoke test.

## 8. Definition of Phase 1 complete

Do not mark Phase 1 complete until all of these work on a clean machine:

```bash
archon dev
archon project create support-bot --display-name "Support Bot"
archon db create docs-index --project support-bot \
  --engine qdrant --plan standard-2 --storage 20Gi
until archon db get docs-index --project support-bot | grep -q Ready; do sleep 1; done
archon agent deploy examples/support-agent --project support-bot
archon agent invoke support-agent --data '{"message":"hello"}'
archon agent logs support-agent
archon dev down
```

The dashboard must show the same real Project, DatabaseInstance, status, capacity, and operation data.

## 9. Immediate next five tasks

1. Add CI and make all validation reproducible.
2. Add `/livez`, `/readyz`, timeouts, graceful shutdown, request IDs, and typed errors.
3. Add `GET /v1/projects` and bounded database pool configuration.
4. Implement `archon dev`, `archon project`, and `archon db` commands.
5. Implement the Qdrant adapter and connect the dashboard to real API data.

These five tasks provide more product value and clarity than adding additional preview pages or future AI modules.
