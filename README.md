# Archon Base

**The self-hosted backend for AI agents.** Vectors, functions, and storage on your own infrastructure.

Archon Base is an open-source platform that does for AI workloads what Appwrite and Supabase do for classic apps. One install gives you managed vector databases, serverless agent functions, and S3-compatible storage for models and datasets, behind one API, one CLI, and one auth model.

> Status: foundation and preview stage. The Project/DatabaseInstance API foundations, PostgreSQL startup wiring, migration runner, student K3s tooling, and local dashboard preview exist; authentication, reconciliation, and most platform features are not implemented yet. Start with the [New User Guide](docs/NEW_USER_GUIDE.md).

## Why Archon Base

| Need | Typical answer today | Archon Base |
|---|---|---|
| Vector database | Run Qdrant, Weaviate or Milvus yourself | Provisioned declaratively, with backups and upgrades |
| Agent runtime | Sandbox service or DIY containers | Sandboxed functions with bindings to your own data |
| Model and file storage | Separate S3 setup | Built in, with a model registry and events |
| Glue (auth, secrets, tracing) | Written per project | One control plane |
| Lock-in | Hyperscaler services | Runs anywhere Kubernetes runs |

## Core components

- **Vector Database Hosting**: Qdrant (default), Chroma (dev and small workloads), Weaviate, Milvus, pgvector, and optionally ScyllaDB, packaged as Engine Packages.
- **Agent Functions**: serverless agents on gVisor by default, Firecracker or Kata microVMs for untrusted code.
- **Storage**: S3-compatible object storage (SeaweedFS by default) with a model registry and dataset versioning.
- **Model Gateway**: OpenAI-compatible endpoint in front of vLLM, Ollama and optional external providers.
- **Planned**: knowledge (RAG) pipelines, tracing and evals, guardrails and approvals, MCP hosting, budgets.

## Repository layout

```
archon-base/
  cmd/archon-api/          control plane API (skeleton)
  cmd/archon/              CLI (skeleton)
  internal/server/         HTTP handlers
  internal/api/            v1 resource types
  api/v1/examples/         example manifests
  engines/                 Engine Packages (qdrant, chroma, pgvector)
  sdk/python/              Python SDK (stubs)
  examples/support-agent/  sample agent function
  deploy/                  Dockerfile, docker-compose, Helm chart
  docs/                    Technical Design Document
```

## Quick start (dev preview)

Requirements: Go 1.22 or newer, Docker with Compose.

```bash
# run the API locally (set ARCHON_DB_URL to enable resource persistence)
make run
curl localhost:8080/healthz
curl localhost:8080/livez
curl localhost:8080/readyz

# or run the dev stack (API, Postgres, NATS, Qdrant)
make up
```

The API binds to `127.0.0.1:8080` by default. Set `ARCHON_API_ADDR` explicitly when exposing it beyond the local host. `/livez` checks the process; `/readyz` also requires the configured resource store to be reachable.

Continuous integration runs Go tests, vet, formatting, builds, frontend syntax checks, Python checks, YAML/JSON validation, Compose validation, and Helm validation through [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

Build binaries:

```bash
make build
./bin/archon version
```

## Open-source building blocks

Archon Base integrates existing projects behind swappable adapters: Qdrant, Chroma, Weaviate, Milvus and pgvector for vectors; vLLM, Ollama and Hugging Face TEI for inference; Temporal for durable workflows; OpenTelemetry and Langfuse for tracing; promptfoo and Ragas for evals; gVisor and Firecracker for sandboxing; SeaweedFS for storage. The full table with roles and license notes is in section 6.1 of the design document.

## Example manifests

Create a vector database:

```yaml
apiVersion: archon.dev/v1
kind: DatabaseInstance
metadata:
  name: docs-index
  project: support-bot
spec:
  engine: qdrant
  plan: standard-2
  replicas: 3
```

Deploy an agent function:

```yaml
apiVersion: archon.dev/v1
kind: AgentFunction
metadata:
  name: support-agent
  project: support-bot
spec:
  runtime: python3.12
  entrypoint: app:handle
  isolation: sandbox
  bindings:
    vector: [ { instance: docs-index, access: read-write } ]
    models: [ { alias: chat-llm } ]
```

Full examples are in `api/v1/examples/` and `examples/support-agent/`.

## Roadmap

| Phase | Focus |
|---|---|
| 0. Foundations | CI, release pipeline, control plane skeleton, single-node installer |
| 1. MVP | Qdrant and pgvector packages, storage, HTTP functions on gVisor, CLI, dev emulator |
| 2. Agent platform | Model Gateway, knowledge pipelines, tracing and evals, durable workflows, MCP hosting, Weaviate |
| 3. Production hardening | HA install, RBAC, guardrails, approvals, budgets, Milvus, model registry |
| 4. Ecosystem | Plugin SDK, marketplace, multi-cluster, enterprise features, air-gapped bundles |

## Self-hosting requirements

| Profile | CPU | RAM | Disk |
|---|---|---|---|
| Enthusiast / K3s practice | 2 CPU cores | 8 GB | 20 GB free space |
| Single node (evaluation) | 4 vCPU | 16 GB | 100 GB SSD |
| Team (3 nodes) | 8 vCPU each | 32 GB each | 500 GB NVMe each |
| Production | Workload dependent | Vector RAM at about 1.5 to 2x raw index size | NVMe for databases |

The enthusiast profile is intended for students and local K3s practice. It also assumes **1 GB graphics memory** for lightweight graphics/GPU experiments; model inference, large vector indexes, and concurrent workloads need more capacity.

See [Self-hosting Requirements](docs/self-hosting-requirements.md) for the student setup and profile guidance. Dependencies for production: Linux kernel 5.10 or newer, containerd, Kubernetes 1.28 or newer, a CSI driver, a load balancer, cert-manager, and optionally the NVIDIA GPU Operator.

## Documentation

- [Technical Design Document](docs/technical-design-document.md): architecture, engine comparison, security, deployment, roadmap.
- [Architecture Diagram](docs/ARCHITECTURE.md): control plane, data plane, provisioning flow, identity foundation, and security boundaries.
- [Core v1 Resource Schema](docs/core-v1-resource-schema.md): the shared Project and DatabaseInstance contract.
- [Phase 1 Local Execution Specification](docs/phase-1-local-execution-spec.md): the local vertical slice without Kubernetes.
- [Self-hosting Requirements](docs/self-hosting-requirements.md): enthusiast/K3s practice profile and larger deployment starting points.
- [Student K3s Setup](docs/student-k3s-setup.md): prerequisite checks, single-node installation, and local monitoring.
- [Product Control Plane TODO](PRODUCT_CONTROL_PLANE_TODO.md): UI, dashboard, admin panel, security, operations, and phased product delivery backlog.
- [Control Plane UI](control-plane-web/README.md): local dashboard, resources, operations, admin, audit, and settings preview.
- [New User Guide](docs/NEW_USER_GUIDE.md): a simple explanation of the architecture, current capabilities, and first real product slice.
- [Contributing](CONTRIBUTING.md)

## License

Archon Base is licensed under the [Apache License 2.0](LICENSE). Third-party integrations may have separate licenses; review their terms before bundling or redistributing them.

## Naming note

Check trademark and registry availability for "Archon Base" before public launch.
