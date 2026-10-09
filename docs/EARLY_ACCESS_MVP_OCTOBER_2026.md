# Archon Base Early Access MVP — October 2026

**Target release:** October 31, 2026  
**Planning window:** October 8–31, 2026  
**Product label:** Early Access / self-hosted developer preview

## 1. Reality-based goal

Archon Base should not attempt to become the complete AI Backend equivalent of Supabase/Appwrite before November. That would require complete authentication, multi-tenant authorization, storage, model routing, agent isolation, audit, upgrades, and operational hardening.

The achievable October goal is a **small but real control plane**:

> A developer can start Archon Base locally, bootstrap an account, create an AI project, provision a Qdrant vector backend, see real resource status in the dashboard, and run a trusted local support-agent example.

This is a credible Early Access product because the core workflow is real rather than a collection of disconnected preview screens.

## 2. MVP promise

### What users can do

1. Start the local AI Backend stack with one command.
2. Bootstrap an owner account and sign in.
3. Check prerequisites and receive actionable errors.
4. Create and list AI projects.
5. Create, list, inspect, update, and delete vector backends.
6. Provision one local Qdrant collection from a vector backend resource.
7. See Pending, Ready, Degraded, and Failed status.
8. Open the dashboard and see real API-backed data.
9. Run the example trusted local Python agent.
10. Query the managed vector binding from the example agent.
11. Stop the local stack safely.

### What the MVP does not promise

- Public multi-tenant hosting.
- Production-grade agent sandboxing.
- gVisor or microVM isolation.
- High availability.
- Multi-cluster management.
- Billing or usage metering.
- General-purpose storage platform.
- External provider secret management.
- Full enterprise RBAC/SSO.
- MCP hosting.
- Model marketplace.
- Disaster recovery guarantees.

These remain roadmap items and must remain visibly labeled Planned.

## 3. Scope freeze

### Must ship

- API startup with PostgreSQL.
- Idempotent migrations.
- `/livez` and `/readyz`.
- Project collection endpoint.
- Database pool configuration.
- Typed API errors.
- Request IDs and structured logs.
- `archon dev` and `archon dev down`.
- `archon project` commands.
- `archon db` commands.
- Qdrant adapter and reconciliation loop.
- Real dashboard data loading.
- Project and DatabaseInstance create forms.
- Status and error states in the dashboard.
- Trusted local agent invocation.
- Automated CI.
- One clean-machine smoke test.
- Early Access installation and limitations documentation.

### Defer until after October

- Full authentication provider integrations.
- Organization invitations and enterprise RBAC.
- Kubernetes controllers.
- Multi-node production installation.
- Agent sandbox isolation.
- Object storage and model registry.
- Model routing and token accounting.
- RAG pipelines, evals, guardrails, and MCP.
- Billing, quotas, marketplace, and multi-cluster.

## 4. User experience

### First-run flow

```text
Install prerequisites
        ↓
archon dev
        ↓
Open dashboard
        ↓
Create Project: support-bot
        ↓
Create DatabaseInstance: docs-index
        ↓
Qdrant reconciliation
        ↓
Status becomes Ready
        ↓
Deploy and invoke trusted local support-agent
```

### Required UI states

Every MVP page must support:

- Loading
- Empty
- Ready
- Pending
- Degraded
- Failed
- API unavailable
- Planned/unavailable feature

The dashboard must never display hardcoded resource counts when the API is connected.

## 5. Schedule

### October 8–10 — Scope and reliability foundation

- Freeze MVP scope.
- Add GitHub Actions CI.
- Add `/livez` and `/readyz`.
- Add HTTP timeouts and graceful shutdown.
- Change local bind default to `127.0.0.1:8080`.
- Add request IDs and structured logs.
- Add typed errors with field paths.
- Add bounded PostgreSQL pool configuration.

**Exit check:** API behavior is deterministic and CI is green.

### October 11–14 — CLI and local runtime

- Implement `archon dev`.
- Implement `archon dev down`.
- Add prerequisite checks.
- Add Compose health checks.
- Pin local image versions.
- Add startup timeout and cleanup.
- Add `archon project create/get/list/delete`.
- Add `archon db create/get/list/delete`.
- Add JSON output mode.

**Exit check:** a clean developer machine can start and stop the local stack with documented commands.

### October 15–19 — Qdrant vertical slice

- Define engine adapter interface.
- Implement Qdrant health check.
- Implement deterministic collection naming.
- Add reconciliation worker.
- Add retry/backoff.
- Update resource conditions and observed generation.
- Add Qdrant failure tests.
- Add create/update/delete smoke tests.

**Exit check:** creating `docs-index` creates and maintains one real Qdrant collection.

### October 20–23 — Real dashboard integration

- Add API client module.
- Add Project list and detail loading.
- Add DatabaseInstance list and detail loading.
- Add real create forms.
- Add status polling or event refresh.
- Add error and empty states.
- Add safe delete confirmation.
- Remove hardcoded supported-resource dashboard values.
- Keep future modules visibly Planned.

**Exit check:** dashboard and CLI show the same real state.

### October 24–27 — Trusted local agent path

- Finalize the supported AgentFunction manifest subset.
- Implement trusted subprocess runner.
- Add timeout and cancellation.
- Add environment allowlist.
- Capture logs.
- Add Qdrant binding client.
- Add local OpenAI-compatible model client.
- Add invocation command and dashboard operation entry.
- Label the runner explicitly as non-isolated.

**Exit check:** the support-agent example runs locally and can query `docs-index`.

### October 28–29 — Security, docs, and release hardening

- Add API input/rate limits appropriate for local Early Access.
- Confirm private-by-default networking.
- Add secret and prompt redaction.
- Add dependency and image scanning.
- Add backup/restore limitation notes.
- Run clean-machine installation test.
- Run failure tests for PostgreSQL, Qdrant, and agent timeout.
- Update README and Early Access guide.

**Exit check:** limitations and security boundary are clear and tested.

### October 30 — Release candidate

- Freeze code.
- Build binaries and container image.
- Run all CI checks.
- Run the complete smoke test.
- Verify migration upgrade from an empty database.
- Verify clean teardown.
- Verify dashboard routes.
- Verify no unsupported feature is presented as Available.

### October 31 — Early Access release

Release only if all exit criteria pass. If one critical dependency fails, release as a dated developer preview rather than claiming a stable product.

## 6. Definition of done

The MVP is complete when this scenario passes on a clean supported Linux machine:

```bash
make build
archon dev
archon project create support-bot --display-name "Support Bot"
archon db create docs-index \
  --project support-bot \
  --engine qdrant \
  --plan standard-2 \
  --storage 20Gi

until archon db get docs-index --project support-bot | grep -q Ready; do
  sleep 1
done

archon agent deploy examples/support-agent \
  --project support-bot
archon agent invoke support-agent \
  --data '{"message":"hello"}'
archon agent logs support-agent
archon dev down
```

The dashboard must show:

- The created Project.
- The created DatabaseInstance.
- Real status conditions.
- Capacity and allocation data sourced from the API.
- Reconciliation or operation events.
- Clear limitations for unavailable features.

## 7. Release gates

Do not release if any of these are true:

- The dashboard still uses fake data for supported resources.
- A DatabaseInstance does not reconcile to Qdrant.
- The CLI commands are still stubs.
- API mutations have no request ID.
- PostgreSQL readiness is not checked.
- Qdrant failure causes silent success.
- The agent runner is described as sandboxed when it is not.
- Secrets or prompts are written into ordinary logs.
- A clean install requires undocumented manual steps.
- The UI presents planned functionality as available.

## 8. Post-November roadmap

After the Early Access MVP is stable:

1. Authentication and organization membership.
2. Project-scoped RBAC.
3. Durable audit log.
4. Kubernetes/K3s controllers.
5. Production agent isolation with a reviewed runtime boundary.
6. Object storage and model resources.
7. Model gateway and provider adapters.
8. Multi-cluster operations.
9. Backups, upgrades, and HA.
10. Enterprise SSO, policies, budgets, and compliance features.

## 9. Success metric

The October release succeeds if a new developer can go from clone to a real vector-backed local AI workflow in **under 30 minutes**, without reading the entire architecture document and without confusing preview features for shipped capabilities.
