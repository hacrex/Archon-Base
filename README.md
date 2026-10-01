# Archon Base

**The self-hosted backend for AI agents.** Vectors, functions, and storage on your own infrastructure.

Archon Base is an open-source platform that does for AI workloads what Appwrite and Supabase do for classic apps. One install gives you managed vector databases, serverless agent functions, and S3-compatible storage for models and datasets, behind one API, one CLI, and one auth model.

> Status: design and scaffold stage. The control plane skeleton runs, but most features are not implemented yet. See the roadmap below.

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

## Quick start (dev)

Requirements: Go 1.22 or newer, Docker with Compose.

```bash
# run the API locally
make run
curl localhost:8080/healthz

# or run the dev stack (API, Postgres, NATS, Qdrant)
make up
```

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
  entrypoint: app.main:handle
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
| Single node (evaluation) | 4 vCPU | 16 GB | 100 GB SSD |
| Team (3 nodes) | 8 vCPU each | 32 GB each | 500 GB NVMe each |
| Production | Workload dependent | Vector RAM at about 1.5 to 2x raw index size | NVMe for databases |

Dependencies for production: Linux kernel 5.10 or newer, containerd, Kubernetes 1.28 or newer, a CSI driver, a load balancer, cert-manager, and optionally the NVIDIA GPU Operator.

## Documentation

- [Technical Design Document](docs/technical-design-document.md): architecture, engine comparison, security, deployment, roadmap.
- [Contributing](CONTRIBUTING.md)

## License

To be decided. Update this section and add a `LICENSE` file before the first public release.

## Naming note

Check trademark and registry availability for "Archon Base" before public launch.
