# Archon Base: New User Guide

## What is Archon Base?

Archon Base is intended to be the **control plane for self-hosted AI infrastructure**. It should eventually help you operate projects, vector databases, models, agents, storage, tools, security policies, and Kubernetes workloads from one place.

The important distinction is:

- **Supabase/Appwrite** mainly help you build application backends.
- **Dify/Langflow/Open WebUI** mainly help you build or use AI applications.
- **Archon Base** is intended to operate the infrastructure underneath those applications.

## What exists today?

The repository is currently a **foundation and preview**, not a finished Supabase/Appwrite-style product.

| Area | Current state |
|---|---|
| Project resource model | Implemented in Go |
| DatabaseInstance resource model | Implemented in Go |
| PostgreSQL repository | Implemented and tested with a disposable database test |
| PostgreSQL API wiring | Implemented when `ARCHON_DB_URL` is supplied |
| Migration runner | Implemented with `schema_migrations` version tracking |
| HTTP CRUD handlers | Implemented for Projects and DatabaseInstances |
| Qdrant provisioning | Implemented when `ARCHON_QDRANT_URL` is configured; creates one deterministic collection per DatabaseInstance |
| User/membership foundation | PostgreSQL schema and validation foundation implemented; authentication is still required before member mutations |
| Migration SQL | Initial tables exist |
| K3s student installer | Implemented with prerequisite checks and dry run |
| Local resource monitor | Implemented with JSON, health, and Prometheus endpoints |
| Web control-plane shell | Implemented as a local preview |
| Authentication | Login/logout repository and bearer middleware implemented |
| Authorization/RBAC | Role hierarchy and organization membership lookup implemented; project policy checks remain |
| First-user bootstrap | Implemented through `archon bootstrap-user`; duplicate emails are refused |
| Persistent web console | Not implemented; current UI uses local preview data |
| CLI workflows | Not implemented; most commands are stubs |
| Kubernetes reconciliation | Not implemented |
| Agent runtime | Not implemented |
| Model gateway | Not implemented |
| Production sandboxing | Not implemented |

## The simplest mental model

Think of the product as four layers:

```text
1. Web console / CLI
   The interface a person uses.

2. Control-plane API
   Validates desired resources and records their state.

3. Control-plane database
   Stores Projects, DatabaseInstances, status, and outbox events.

4. Data plane
   Kubernetes/K3s, Qdrant, model servers, storage, and runtimes.
```

Only the first three layers have partial foundations today. The data-plane controllers that turn a desired DatabaseInstance into a running Qdrant resource are still planned.

## What can I run now?

### Check the codebase

```bash
go test ./...
go vet ./...
go build ./...
```

### Start the API preview

```bash
make run
```

Before logging in for the first time, create the local owner account from a shell with access to PostgreSQL:

```bash
export ARCHON_DB_URL='postgres://archon:archon@127.0.0.1:5432/archon?sslmode=disable'
printf '%s\n' 'replace-with-a-strong-local-password' | \
  go run ./cmd/archon bootstrap-user \
    --email owner@example.com \
    --display-name 'Local Owner' \
    --password-stdin
```

The command applies pending migrations, creates an active user, stores a bcrypt password hash, assigns the `owner` role in the configured organization, and refuses to overwrite an existing email. Use `ARCHON_ORGANIZATION_ID` to select a different organization UUID.

Then check:

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/v1/version
curl http://127.0.0.1:8080/livez
curl http://127.0.0.1:8080/readyz

# after creating an active user credential, log in
curl -X POST http://127.0.0.1:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"your-password"}'

# use the returned token for the current principal and resource routes
curl http://127.0.0.1:8080/v1/auth/me -H 'Authorization: Bearer <token>'
```

`/healthz` and `/livez` prove that the HTTP process is alive. `/readyz` also checks that the configured resource store is ready. The API binds to loopback (`127.0.0.1:8080`) by default. To enable resource CRUD, provide `ARCHON_DB_URL`; the API pings PostgreSQL, applies forward migrations from `ARCHON_MIGRATIONS_DIR` (default `migrations`), and wires the repository before serving resource routes.

### Start the local dashboard preview

```bash
cd control-plane-web
python3 -m http.server 4174 --bind 127.0.0.1
```

Open:

```text
http://127.0.0.1:4174/#/overview
```

The dashboard demonstrates the planned product experience. Its values are local preview data and its buttons do not persist changes yet.

### Check the student host profile

```bash
sudo ./scripts/install-k3s-student.sh --check-only
```

The recommended enthusiast profile is 8 GB RAM, 2 CPU cores, 20 GB free disk, and optional 1 GB graphics memory. The K3s installer can be previewed without changing the host:

```bash
sudo ./scripts/install-k3s-student.sh --dry-run
```

## What is the first real product slice?

The first useful end-to-end slice should be deliberately small:

1. Start PostgreSQL and Qdrant.
2. Start `archon-api` with `ARCHON_DB_URL`.
3. Let the migration runner apply the initial schema.
4. Create a Project through the API.
5. Create a DatabaseInstance through the API.
6. Reconcile that DatabaseInstance into one local Qdrant collection.
7. Display those real resources in the Projects and Resources dashboard.
8. Add `/livez`, `/readyz`, and local development authentication before enabling administrative actions.

Do not start with every AI platform feature. A working Project → DatabaseInstance → Qdrant path will make the rest of the architecture understandable.

## Where should a new contributor look?

| Question | Start here |
|---|---|
| What is the product vision? | `docs/technical-design-document.md` |
| How is the product connected? | `docs/ARCHITECTURE.md` |
| What is the durable v1 resource contract? | `docs/core-v1-resource-schema.md` |
| How should local Phase 1 work? | `docs/phase-1-local-execution-spec.md` |
| What does the product UI look like? | `control-plane-web/README.md` |
| What is the product backlog? | `PRODUCT_CONTROL_PLANE_TODO.md` |
| How do I understand the code? | `internal/api`, `internal/server`, `internal/store` |
| How do I install student K3s? | `docs/student-k3s-setup.md` |
| How do I check local resources? | `cmd/archon-monitor` and `internal/monitor` |
| What is the authentication boundary? | `docs/authentication-authorization-foundation.md` |

## Important vocabulary

- **Project:** a logical application boundary with environment, region, and quotas.
- **DatabaseInstance:** a desired vector/data service attached to a Project.
- **Desired state:** what the user asks Archon Base to create or change.
- **Observed state:** what a controller has confirmed in the data plane.
- **Control plane:** the API, database, UI, and controllers that manage resources.
- **Data plane:** the actual K3s/Kubernetes workloads and infrastructure services.
- **Reconciliation:** repeatedly comparing desired state with observed state and correcting drift.
- **Outbox event:** a database event written in the same transaction as a resource mutation so a future publisher can process it reliably.

## Current product promise

Today, the honest promise is:

> Archon Base contains the early control-plane foundations and a local UI preview for self-hosted AI infrastructure.

The stronger promise—provisioning and operating AI agents, models, data, tools, and compute from a complete dashboard—remains the target architecture and roadmap, not the current shipped behavior.
