# Phase 1 Local Execution Specification

**Goal:** run one complete Archon vertical slice on a developer laptop without Kubernetes.

The slice is intentionally narrow: one local API process, PostgreSQL metadata, one Qdrant instance, one Python agent, and a CLI. NATS, object storage, model gateway, and sandbox isolation are interfaces or optional services in this phase; they must not be falsely presented as production capabilities.

## 1. Supported topology

```text
archon CLI
    |
    v
archon-api (:8080) ---- PostgreSQL (:5432)
    |
    +---- Qdrant (:6333) [one managed local instance]
    |
    +---- local agent runner [subprocess, development-only]
```

`archon dev` starts PostgreSQL and Qdrant with Docker Compose, applies migrations, starts the API, and writes a local context file under `.archon/`. The API owns resource state. The local provisioner reconciles DatabaseInstance resources to a named Qdrant collection/endpoint and records status.

## 2. Explicit non-goals

Phase 1 does not provide:

- Kubernetes reconciliation or HA
- public multi-tenant hosting
- untrusted-code execution
- gVisor/Firecracker isolation
- production secrets management
- external model-provider credentials stored by Archon
- durable event delivery through NATS
- backups that are safe to treat as disaster recovery
- public exposure of Qdrant or the local agent runner

The local runner must display a warning that it is a trusted developer subprocess.

## 3. Prerequisites

- Go 1.22+
- Python 3.12+
- Docker Engine and Compose v2
- `curl`
- a local OpenAI-compatible endpoint or test model for agent invocation

Verify:

```bash
go version
python3 --version
docker compose version
```

## 4. Local configuration

`archon dev` creates `.archon/config.yaml` with disposable development values:

```yaml
api: http://127.0.0.1:8080
postgres:
  url: postgres://archon:archon@127.0.0.1:5432/archon?sslmode=disable
qdrant:
  url: http://127.0.0.1:6333
runtime:
  mode: local-subprocess
model:
  baseUrl: http://127.0.0.1:11434/v1
  apiKeyEnv: ARCHON_LOCAL_MODEL_API_KEY
```

The API must bind to loopback by default. Binding to `0.0.0.0` requires an explicit `--listen` option and a warning.

## 5. Startup sequence

1. Check prerequisites and fail with an actionable message.
2. Create `.archon/` with mode `0700`.
3. Start only PostgreSQL and Qdrant from the development Compose profile.
4. Wait for PostgreSQL readiness using `pg_isready` or a bounded SQL connection loop.
5. Wait for Qdrant readiness using its health endpoint.
6. Apply migrations transactionally. Record the migration version in a `schema_migrations` table.
7. Start `archon-api` with the local config.
8. Poll `/livez` and `/readyz`; do not report success until both are healthy.
9. Write the API PID and child process metadata to `.archon/run/`.
10. Print the API URL, project context, and teardown command.

Startup must have a finite timeout, recommended 60 seconds, and must clean up newly created processes if a required step fails.

## 6. Resource lifecycle

### 6.1 Create a project

```bash
archon project create support-bot --display-name "Support Bot"
```

The CLI sends a strict `Project` manifest. The API validates it, inserts the Project row, and returns `201`. Creation is synchronous because no data-plane dependency is required.

### 6.2 Create a Qdrant database instance

```bash
archon db create docs-index --project support-bot \
  --engine qdrant --plan standard-2 --storage 20Gi
```

The API transaction must:

1. validate the project and DatabaseInstance spec;
2. insert the desired resource with `phase = Pending`;
3. insert an outbox event named `DatabaseInstanceCreated`;
4. commit before returning `202 Accepted`.

The local provisioner then:

1. derives a deterministic Qdrant collection/namespace from project and resource IDs;
2. verifies Qdrant connectivity;
3. creates the collection if absent;
4. updates `status.phase` to `Ready` or `Failed`;
5. records `observed_generation` and a condition.

The operation must be idempotent. Re-running the command with the same resource name updates or returns the existing desired state rather than creating duplicates.

### 6.3 Deploy the example agent

For Phase 1, deployment is a local runner registration rather than a sandboxed deployment:

```bash
archon agent deploy examples/support-agent \
  --project support-bot \
  --manifest examples/support-agent/agent.yaml
```

The API validates the manifest against the same typed schema used by the CLI. The runner loads the Python package in a subprocess with:

- a temporary working directory;
- a restricted environment allowlist;
- no host filesystem mounts other than the source directory;
- a configurable timeout;
- loopback-only network access by default.

This is not a security boundary. The command must state that explicitly.

### 6.4 Invoke and inspect

```bash
archon agent invoke support-agent --data '{"message":"How do I reset my password?"}'
archon db get docs-index --project support-bot
archon agent logs support-agent --follow
```

The runner injects a request context and binding clients. The vector client targets the managed local Qdrant instance. The model client targets the configured OpenAI-compatible local endpoint. Errors must preserve a request ID and be visible in CLI output.

## 7. Failure behavior

- PostgreSQL unavailable: `/readyz` is unhealthy; mutations fail with `503` and no partial row is written.
- Qdrant unavailable: DatabaseInstance remains `Pending` or becomes `Degraded`; the API remains live.
- Agent timeout: terminate the subprocess, record a failed invocation, and return `504`.
- Agent crash: capture stderr, return a typed runtime error, and leave the resource deployed.
- Migration failure: stop startup and do not serve the API.
- Ctrl-C: stop the API, terminate child runners, and leave Compose services running unless `archon dev down` is requested.

## 8. Verification checklist

A Phase 1 implementation is complete when a clean machine can:

```bash
archon dev
archon project create support-bot --display-name "Support Bot"
archon db create docs-index --project support-bot --engine qdrant --plan standard-2 --storage 20Gi
until archon db get docs-index --project support-bot | grep -q Ready; do sleep 1; done
archon agent deploy examples/support-agent --project support-bot --manifest examples/support-agent/agent.yaml
archon agent invoke support-agent --data '{"message":"hello"}'
archon dev down
```

Automated tests must cover:

- manifest validation and strict unknown-field handling;
- project/database uniqueness and foreign keys;
- idempotent create;
- outbox write atomicity;
- Qdrant reconciliation retry;
- agent timeout and crash handling;
- readiness behavior;
- cleanup after failed startup.

## 9. Exit criteria for Kubernetes work

Do not begin the Kubernetes controller until this local path has:

- stable v1 resource types;
- migrations and rollback documentation;
- repeatable local startup;
- a working Qdrant adapter;
- status conditions and generation handling;
- idempotent mutation tests;
- structured logs and request IDs;
- a documented security boundary for the local runner.
