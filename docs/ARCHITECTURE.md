# Archon Base Architecture

**Architecture status:** Early Access foundation  
**Current release boundary:** local control plane, PostgreSQL resource state, optional Qdrant provisioning, student/K3s tooling, and identity foundation  
**Not yet production-ready:** authentication, authorization, Kubernetes controllers, agent isolation, multi-tenancy, and high availability

## 1. Product architecture at a glance

Archon Base is designed as a control plane for self-hosted AI infrastructure. The control plane stores desired resources and coordinates provisioning; the data plane runs the actual databases, model servers, agents, and workloads.

```mermaid
flowchart LR
    User[Developer / Operator]
    CLI[Archon CLI\nplanned workflows]
    Web[Control Plane Web UI\nlocal preview]
    API[Archon API\nGo HTTP server]
    PG[(PostgreSQL\nresource state + migrations)]
    Outbox[(Transactional Outbox\nresource events)]
    QP[Qdrant Provisioner\nHTTP adapter]
    Q[(Qdrant\nvector data plane)]
    Monitor[Archon Monitor\nlocal host metrics]
    K3s[K3s / Kubernetes\nfuture data plane]
    Agent[Agent Runtime\ntrusted local mode planned]
    Models[Model Gateway\nplanned]
    Storage[Object Storage\nplanned]

    User --> CLI
    User --> Web
    CLI --> API
    Web -. preview today .-> API
    API --> PG
    API --> Outbox
    API --> QP
    QP --> Q
    Monitor -. local status .-> Web
    API -. future controllers .-> K3s
    K3s --> Agent
    K3s --> Models
    K3s --> Storage
    Agent --> Q
    Agent --> Models
```

### Current implementation boundary

The solid path is the current Early Access foundation:

```text
HTTP client/API request
        ↓
Go API validation and handlers
        ↓
PostgreSQL repository
        ↓
Project/DatabaseInstance state + outbox event
        ↓
Optional Qdrant HTTP provisioner
        ↓
Qdrant collection
```

The dashed paths are planned or preview-only and must not be presented as shipped functionality.

## 2. Runtime deployment topology

### Local development and student profile

```mermaid
flowchart TB
    subgraph Host[Developer laptop or student Linux host]
        Browser[Browser]
        CLI[archon CLI]
        API[archon-api\n127.0.0.1:8080]
        Monitor[archon-monitor\n127.0.0.1:9105]

        subgraph Compose[Docker Compose local stack]
            PG[(PostgreSQL)]
            Q[(Qdrant)]
            NATS[(NATS\nconfigured, publisher planned)]
        end

        Browser --> UI[control-plane-web\nstatic preview]
        UI --> API
        CLI --> API
        API --> PG
        API --> Q
        API --> NATS
        Monitor --> HostMetrics[CPU / memory / disk / optional GPU]
        Browser --> Monitor
    end

    Student[K3s student installer] -. installs .-> K3s[K3s single-node\nstudent practice]
```

### Intended self-hosted cluster topology

```mermaid
flowchart TB
    Client[Web UI / CLI / API clients]
    Gateway[Ingress / TLS gateway\nplanned]
    Control[Archon Control Plane]
    Auth[Identity + authorization\nplanned]
    DB[(PostgreSQL\ncontrol-plane state)]
    Bus[(NATS JetStream\noutbox publisher planned)]
    Controller[Archon Controllers\nplanned]
    Qdrant[Qdrant cluster]
    Models[Model servers\nOllama / vLLM planned]
    Runtime[Agent runtime\ngVisor or microVM planned]
    Object[Object storage\nplanned]
    Nodes[Kubernetes nodes]

    Client --> Gateway --> Control
    Control --> Auth
    Control --> DB
    Control --> Bus
    Bus --> Controller
    Controller --> Qdrant
    Controller --> Models
    Controller --> Runtime
    Controller --> Object
    Qdrant --> Nodes
    Models --> Nodes
    Runtime --> Nodes
    Object --> Nodes
```

## 3. Control-plane request flow

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant A as Archon API
    participant P as PostgreSQL
    participant O as Outbox
    participant QP as Qdrant Provisioner
    participant Q as Qdrant

    C->>A: POST /v1/projects/{project}/databases
    A->>A: Decode strictly, normalize, validate
    A->>P: Insert DatabaseInstance
    P-->>A: UID, generation, Pending state
    A->>O: Insert DatabaseInstanceCreated event
    O-->>A: Transaction committed
    A->>QP: Ensure collection when configured
    QP->>Q: GET /readyz
    Q-->>QP: Ready
    QP->>Q: PUT /collections/archon_project_instance
    Q-->>QP: Collection created or already present
    QP-->>A: Provisioning result
    A->>P: Update phase and observed generation
    A-->>C: 202 Accepted + resource status
    C->>A: GET /v1/projects/{project}/databases/{name}
    A->>P: Read desired and observed state
    P-->>A: Resource
    A-->>C: Current resource status
```

### Resource state model

```mermaid
stateDiagram-v2
    [*] --> Pending: Resource created
    Pending --> Ready: Provider health + provisioning succeed
    Pending --> Degraded: Provider unavailable or partial result
    Pending --> Failed: Validation/provider failure
    Degraded --> Ready: Retry succeeds
    Degraded --> Failed: Retry budget exhausted
    Ready --> Pending: Desired generation changes
    Ready --> Deleting: Delete requested
    Degraded --> Deleting: Delete requested
    Failed --> Deleting: Delete requested
    Deleting --> [*]: Soft deletion recorded
```

## 4. Data model and ownership

```mermaid
erDiagram
    ORGANIZATION ||--o{ PROJECT : owns
    ORGANIZATION ||--o{ MEMBERSHIP : grants
    USER ||--o{ MEMBERSHIP : receives
    PROJECT ||--o{ DATABASE_INSTANCE : contains
    DATABASE_INSTANCE ||--o{ OUTBOX_EVENT : emits

    ORGANIZATION {
        uuid id PK
        text name
        text mode
    }
    USER {
        uuid id PK
        text email UK
        text display_name
        text status
    }
    MEMBERSHIP {
        uuid organization_id PK
        uuid user_id PK
        text role
    }
    PROJECT {
        uuid id PK
        uuid organization_id FK
        text name
        jsonb spec
        text phase
        bigint generation
    }
    DATABASE_INSTANCE {
        uuid id PK
        uuid project_id FK
        text name
        text engine
        jsonb spec
        text phase
        bigint generation
        bigint observed_generation
    }
    OUTBOX_EVENT {
        bigint id PK
        text aggregate_type
        uuid aggregate_id
        text event_type
        jsonb payload
        timestamptz published_at
    }
```

### Identity status

The identity migration creates `users` and `organization_memberships` with owner/admin/operator/developer/viewer roles. It is a storage foundation only.

Authentication and authorization must be implemented before member-management mutations or project-scoped access are exposed. The current single-organization local mode is not multi-tenant production security.

## 5. Security boundaries

```mermaid
flowchart LR
    subgraph Public[Untrusted client boundary]
        Browser[Browser]
        CLI[CLI]
        External[External API client]
    end

    subgraph Control[Control-plane boundary]
        API[API validation + request limits]
        Auth[Authentication / authorization\nplanned]
        Store[PostgreSQL repository]
        Audit[Audit stream\nplanned]
    end

    subgraph Data[Data-plane boundary]
        Qdrant[Qdrant]
        Runtime[Agent runtime\nnot isolated yet]
        Models[Model providers]
        Storage[Object storage]
    end

    Browser --> API
    CLI --> API
    External --> API
    API --> Auth
    API --> Store
    API --> Audit
    API --> Qdrant
    API --> Runtime
    Runtime --> Models
    Runtime --> Storage

    Warning[Current rule: trusted local subprocess is not a security sandbox]
    Runtime -.-> Warning
```

Current security rules:

- API binds to `127.0.0.1:8080` by default.
- Qdrant and PostgreSQL should remain private to the local stack or cluster network.
- Request bodies are size-limited.
- Request IDs and structured access logs are available.
- Secrets and prompt redaction still need to be completed throughout all future runtimes.
- User/membership data exists, but authentication and authorization are not complete.
- Do not expose trusted local agent execution to untrusted users.

## 6. Repository-to-architecture map

| Architecture area | Repository location | Current status |
|---|---|---|
| API entrypoint | `cmd/archon-api/main.go` | Implemented with PostgreSQL, migrations, readiness, and optional Qdrant wiring |
| HTTP API | `internal/server/` | Project and DatabaseInstance CRUD plus health/readiness endpoints |
| Resource contracts | `internal/api/` | Project, DatabaseInstance, User, Membership, validation |
| PostgreSQL repository | `internal/store/` | Project/DatabaseInstance persistence and status updates |
| Migration runner | `internal/db/` | Ordered, idempotent forward migration runner |
| Core schema | `migrations/001_core_resources.sql` | Projects, DatabaseInstances, outbox |
| Identity schema | `migrations/002_identity_foundation.sql` | Users and organization memberships; auth gate remains |
| Qdrant adapter | `internal/provision/qdrant.go` | Health, deterministic collection provisioning, deletion |
| Local monitor | `cmd/archon-monitor/`, `internal/monitor/` | Host resource status and Prometheus metrics |
| Web console | `control-plane-web/` | Modular local preview; real API integration remains next |
| K3s installer | `scripts/install-k3s-student.sh` | Student/enthusiast single-node installation |
| Kubernetes package | `deploy/helm/` | Early skeleton; not production-ready |
| Python SDK | `sdk/python/` | Public shape/stubs; runtime bindings remain planned |

## 7. Near-term architecture sequence

The safest implementation order is:

```text
1. Qdrant adapter and real resource status
2. API-backed dashboard for Projects and DatabaseInstances
3. Local CLI and one-command dev runtime
4. Authentication and session management
5. Authorization and project-scoped membership checks
6. Authenticated Admin Panel user management
7. Outbox publisher and reconciliation worker
8. Kubernetes controllers and Helm hardening
9. Agent runtime security boundary
10. Models, storage, RAG, tools, and enterprise features
```

Each future capability should follow this lifecycle:

```text
contract → migration → repository → API → reconciliation → UI
→ authentication → authorization → audit → tests → documentation
```

## 8. Architecture status legend

- **Available:** implemented and tested in the current repository.
- **Preview:** visible in the local UI or supported in a limited local mode.
- **Planned:** design target, not available for production use.
- **Degraded:** the component is present but a dependency or capability is unavailable.
- **Unsupported:** intentionally rejected in the current milestone.
