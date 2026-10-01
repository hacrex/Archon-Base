# Archon Base: Self-Hostable AI Cloud Platform
### Technical Design Document, v1.0

## 1. Summary

Archon Base is an open-source, self-hostable backend platform for AI applications. It does for AI workloads what Appwrite and Supabase do for classic apps. One install gives a team the building blocks every AI app needs:

1. **Vector and data stores**, provisioned on demand (Qdrant, Chroma, Weaviate, Milvus, pgvector, ScyllaDB).
2. **Agent Functions**, serverless and sandboxed runtimes for agents, tools and pipelines.
3. **Object storage**, for models, datasets, embeddings snapshots and artifacts.
4. **AI platform services**, including a model gateway, knowledge pipelines, tool registry, memory, guardrails and evals.

One control plane exposes a single API, CLI, console and auth model across all of it. Everything runs on infrastructure the user owns, from a single VM to a multi-node Kubernetes cluster.

### Goals
- One-command install on a laptop or VM; Helm-based HA install on Kubernetes.
- No hard dependency on any hyperscaler service.
- Strong isolation for untrusted and LLM-generated code.
- Pluggable engines, runtimes and providers so the platform grows without core changes.
- Agent-first developer experience: bindings, tracing, evals and approvals built in.

### Non-goals (v1)
- Training foundation models. Fine-tuning orchestration comes later.
- Replacing full MLOps suites. Archon Base integrates with them.

### Positioning
Supabase is the closest general backend, but it has no agent-specific tooling. Sandbox projects (E2B, Daytona) and vector databases (Qdrant, Weaviate, Milvus) each cover one slice. Archon Base's value is the integration: one install, one auth model, and bindings that connect agents to vectors, storage and models.

---

## 2. Architecture Overview

```
          +-------------------------------------------------------+
          |  Console (web)   CLI (archon)   SDKs (Py/TS/Go)   REST |
          +---------------------------+---------------------------+
                                      |
                             API Gateway (Envoy)
                        authN/Z, rate limit, routing
                                      |
   +----------------------------------+----------------------------------+
   |                          CONTROL PLANE                              |
   | Project/Tenant | Provisioner | Function Ctl | Storage Svc | Secrets   |
   | Model Gateway  | Tool Registry | Knowledge Svc | Eval Svc | Audit    |
   |        Metadata: PostgreSQL        Events: NATS JetStream           |
   +--------+-------------+--------------+-------------+-----------------+
            |             |              |             |
     +------v-----+ +-----v------+ +-----v-----+ +-----v------+
     | DB Operators| | Agent      | | Object    | | Inference  |
     | (vector DBs)| | Runtime    | | Storage   | | (vLLM etc) |
     +------+------+ +-----+------+ +-----+-----+ +-----+------+
            |              |              |             |
   ============== DATA PLANE (Kubernetes) ===================================
   Cilium CNI | CSI storage | GPU operator | OpenTelemetry | OpenBao/Vault
```

### Plane separation
- **Control plane**: stateless Go services plus Postgres and NATS. It stores desired state and reconciles it.
- **Data plane**: the actual databases, sandboxes and storage nodes. It can share a cluster with the control plane or run remotely, which enables multi-cluster and edge setups later.

### Design principles
- **Declarative first.** Every resource (database, agent, bucket, pipeline) is a versioned manifest. API, CLI and console produce the same manifests, so GitOps works naturally.
- **Operators over scripts.** Install, upgrade, backup and scale logic lives in Kubernetes operators.
- **Multi-tenant by default.** Organization, Project, Environment is enforced at the gateway, database and network layers.
- **Bindings, not credentials.** Agents request access to a resource; the platform injects short-lived credentials.

---

## 3. Core Components

### 3.1 Vector Database Hosting

#### Engine comparison

| Engine | Strengths | Weaknesses | Ops weight | License note |
|---|---|---|---|---|
| **Qdrant** | Fast filtered search, simple binary, good quantization | Smaller ecosystem than Milvus | Low | Apache 2.0 |
| **Weaviate** | Hybrid (BM25 + vector), modules, multi-tenancy | Memory hungry at scale, larger config surface | Medium | BSD-3 |
| **Milvus** | Highest scale ceiling, many index types, GPU indexes | Many moving parts (etcd, object store, message queue) | High | Apache 2.0 |
| **pgvector** | No new infra, transactional joins with relational data | Slower at very large scale, index build cost | Low | PostgreSQL license |
| **ScyllaDB** | Extreme throughput, predictable latency, strong write rates | Vector search is newer, fewer ANN options, license changed to source-available (verify terms) | Medium | Verify current license |
| **Chroma** | Great developer experience, embeddable and server modes, simple API | Best suited to dev and small to mid workloads; verify distributed mode maturity before offering it as a scale tier | Low | Apache 2.0 |

#### Recommended tiering
- **Default: Qdrant.** Best balance of speed and operational simplicity.
- **Hybrid search: Weaviate.**
- **Scale tier: Milvus**, with dependencies provisioned by the operator.
- **Relational plus vector: pgvector** through managed Postgres (CloudNativePG).
- **High throughput: ScyllaDB**, optional once features and license fit.
- **Developer and small-workload tier: Chroma.** It also powers the embedded store in the `archon dev` local emulator, so the same client code runs on a laptop and in production.

#### Engine Packages
Each engine ships as an **Engine Package**: Helm chart or operator, config JSON schema, health probes, backup strategy, upgrade policy and conformance tests. Users create instances declaratively:

```yaml
apiVersion: archon.dev/v1
kind: DatabaseInstance
metadata:
  name: docs-index
  project: support-bot
spec:
  engine: qdrant
  version: "1.x"
  plan: standard-2          # maps to CPU/RAM/disk
  replicas: 3
  shards: 2
  storage: { class: fast-ssd, size: 200Gi }
  backup: { schedule: "0 2 * * *", retention: 14d, target: bucket://backups }
  network: { exposure: private }   # private | project | public
```

#### Managed capabilities
- Install, configure and size from plans.
- Rolling upgrades with pre-flight checks and automatic rollback.
- Scheduled and on-demand snapshots; point-in-time restore where the engine supports it.
- Horizontal scale (replicas, shards) and vertical scale (plan change).
- Metrics, slow-query logs and index-health dashboards.
- Connection broker that issues short-lived credentials.
- Optional **unified vector API** (`upsert`, `query`, `delete`, `filter`) for portability. Native endpoints stay available so engine-specific features are never blocked.

---

### 3.2 Agent Services (Serverless Agent Functions)

#### Definition
A deployable unit of code (Python, TypeScript, Go or any container) with declared triggers, resource limits, permissions and bindings to platform services.

#### Isolation options

| Technology | Cold start | Isolation | Use case |
|---|---|---|---|
| Container (runc) | Fast | Low | Trusted internal code only |
| **gVisor** | Fast | Medium-high | Default for most tenants |
| **Firecracker / Kata microVM** | Moderate | High | Untrusted code, LLM-generated code, code interpreter |
| WASM (Wasmtime) | Very fast | High, limited languages | Lightweight tools and plugins |

Policy: gVisor by default; microVMs for functions marked `untrusted: true` and for the code interpreter tool.

#### Execution modes
- **Request/response**: HTTP invocation with scale to zero.
- **Event-driven**: NATS subjects, storage events, database change hooks, cron.
- **Durable workflows**: long-running agents with checkpointing, retries and resume.
- **Streaming**: SSE and WebSocket for token streaming.
- **Queued jobs**: batch and async work with priorities and dead-letter handling.

#### Scaling
KEDA-based autoscaling on concurrency, queue depth or custom metrics. Warm pools for latency-sensitive functions. Separate GPU pools with time-slicing or MIG for fractional GPUs.

#### Agent manifest

```yaml
apiVersion: archon.dev/v1
kind: AgentFunction
metadata:
  name: support-agent
  project: support-bot
spec:
  runtime: python3.12
  entrypoint: app.main:handle
  source: { git: "https://git.example.com/org/support-agent", ref: main }
  triggers:
    - http: { path: /chat, auth: project-key }
    - event: { subject: "storage.support-docs.created" }
  resources: { cpu: "1", memory: 1Gi, timeout: 300s, concurrency: 10, gpu: 0 }
  isolation: sandbox            # sandbox | microvm
  bindings:
    vector:  [ { instance: docs-index, access: read-write } ]
    storage: [ { bucket: support-docs, access: read } ]
    models:  [ { alias: chat-llm } ]
    tools:   [ { name: crm-lookup, approval: required } ]
  secrets: [ CRM_API_TOKEN ]
  egress: { allow: ["api.crm.example.com"] }   # default deny
```

#### SDK example (Python)

```python
from archon import agent, vector, llm

@agent.handler
async def handle(req):
    query = req.json["message"]
    hits = await vector("docs-index").query(text=query, top_k=5)
    context = "\n".join(h.payload["text"] for h in hits)
    return await llm("chat-llm").chat(
        [{"role": "system", "content": context},
         {"role": "user", "content": query}], stream=True)
```

---

### 3.3 File and Model Storage

#### Requirements
S3-compatible API, large objects (model weights of tens of GB), multipart and resumable uploads, versioning, lifecycle rules, event notifications and per-project quotas.

| Backend | Fit | Note |
|---|---|---|
| **SeaweedFS** | Handles small and large files, simple to run | Recommended default |
| **Garage** | Lightweight, geo-distributed small clusters | Good for edge and small installs |
| **Ceph (RGW)** | Large enterprise deployments | Higher operational cost |
| **MinIO** | Mature and popular | Review current license and edition split (AGPL) first |

A storage adapter interface makes backends swappable by configuration.

#### AI-specific storage features
- **Model Registry**: name, version, framework, checksum, lineage, signed manifests.
- **Dataset versioning**: immutable snapshots with content hashing and Parquet preview.
- **Artifact cache**: node-local cache with lazy loading, so large weights do not re-download on each cold start.
- **Event hooks**: new objects can trigger functions or knowledge pipelines.
- **Signed URLs** and per-project bucket policies.

---

### 3.4 Model Gateway

An OpenAI-compatible endpoint in front of self-hosted inference (vLLM, Ollama, TGI) and optional external providers.
- Per-project keys, rate limits, quotas and token accounting.
- Routing and fallback rules (for example, local model first, external on failure).
- Response and embedding cache.
- Request and response logging with PII redaction options.
- Model aliases (`chat-llm`) so agents never hard-code a model name.

---

## 4. Additional Features (Recommended)

These go beyond the original scope. They are ordered by impact on adoption.

### 4.1 Knowledge and RAG Pipelines (high value)
- Declarative **ingestion pipelines**: source, parse, chunk, embed, index.
- **Connectors**: S3 and bucket events, Git, Google Drive, Confluence, Notion, websites, SQL databases.
- Parsers for PDF, DOCX, HTML, images (OCR) and tables.
- Incremental sync and re-embedding when the embedding model changes.
- Hybrid retrieval, reranking and metadata filtering as platform primitives.

### 4.2 Agent Governance and Safety (high value)
- **Guardrails** at the gateway: prompt-injection detection, PII redaction, output policy checks.
- **Human-in-the-loop approvals** for high-risk tools, with Slack, email or console prompts.
- **Agent identity**: each agent gets its own identity and scoped, delegated credentials, not shared keys.
- Tool permission scopes and per-call audit.
- Kill switch and rate caps per agent to stop runaway loops and cost spikes.

### 4.3 Evals and Observability for Agents (high value)
- Trace viewer for every LLM call, tool call and retrieval (OpenTelemetry-based).
- **Eval suites**: datasets, scorers (rule-based and LLM-judge), regression gates in CI.
- **Replay**: re-run recorded traces against a new agent version before promotion.
- Online feedback collection (thumbs, labels) feeding eval datasets.
- Cost and latency breakdown per agent, per model, per project.

### 4.4 Prompt and Config Management (medium)
- Versioned **prompt registry** with environments, diffs and rollback.
- Feature flags and A/B or canary rollouts for agent versions.
- Config and secrets promotion between dev, staging and prod.

### 4.5 Protocol and Ecosystem Support (medium)
- **MCP server hosting**: deploy and manage MCP servers as first-class functions with auth.
- **A2A support** so agents on the platform can discover and call each other.
- Tool marketplace and **agent templates** (support bot, research agent, document Q&A, data analyst).
- OpenAPI import to auto-generate tools from existing APIs.

### 4.6 Data Platform Extras (medium)
- **Vector index snapshots and branching**: clone an index to test a new embedding model or chunking strategy, then promote.
- **Engine migration tool**: move a collection between Qdrant, Weaviate, Milvus and pgvector.
- **Semantic cache** to reduce repeat LLM spend.
- Agent memory service: short-term (Redis-compatible) and long-term (vector-backed), with retention policies.
- Built-in **queues, pub/sub and realtime** channels for agent-to-UI updates.

### 4.7 Cost, Capacity and Scheduling (medium)
- Budgets and alerts per project, with hard caps on tokens and GPU hours.
- Chargeback and showback reports.
- GPU-aware scheduler with priorities, preemption and bin-packing.
- Spot and idle node policies for cost control.

### 4.8 Enterprise and Compliance (later)
- SSO/SAML, SCIM provisioning, bring-your-own-key (BYOK) encryption.
- Data residency controls and per-project encryption keys.
- Audit export to SIEM, retention policies, compliance report templates.
- Air-gapped installer bundle and offline model and package mirrors.

### 4.9 Developer Experience (high value for adoption)
- **Local dev emulator**: `archon dev` runs a lightweight stack on a laptop with the same APIs, using embedded Chroma for vectors.
- Terraform provider and Kubernetes CRDs for infrastructure as code.
- Project templates and scaffolds (`archon init --template rag-chat`).
- Preview environments per pull request.
- Public status of each resource with clear conditions and error messages.

### Feature priority summary

| Feature | Impact | Effort | Suggested phase |
|---|---|---|---|
| Knowledge and RAG pipelines | High | Medium | 2 |
| Agent tracing and evals | High | Medium | 2 |
| Guardrails and approvals | High | Medium | 2 to 3 |
| Local dev emulator | High | Medium | 1 to 2 |
| MCP and A2A hosting | Medium | Low | 2 |
| Index branching and migration | Medium | High | 3 |
| Semantic cache | Medium | Low | 3 |
| Cost controls and budgets | Medium | Medium | 3 |
| Marketplace and templates | Medium | Medium | 4 |
| Enterprise and compliance | Medium | High | 3 to 4 |

---

## 5. Cross-Cutting Concerns

### 5.1 Security

| Layer | Controls |
|---|---|
| **Identity** | OIDC (Keycloak or Ory), SSO/SAML, service accounts, short-lived tokens, agent identities |
| **Authorization** | RBAC at org, project and resource level; policy engine (OPA or Cedar) for fine-grained rules |
| **Secrets** | OpenBao or Vault, or Kubernetes secrets with KMS; injected at runtime only |
| **Network** | Default-deny policies (Cilium), per-project namespaces, mTLS, per-function egress allowlists |
| **Execution** | gVisor or microVM sandboxes, read-only root filesystems, seccomp, no host mounts, CPU/memory/time limits |
| **Data** | TLS everywhere, encryption at rest, optional per-project keys, encrypted backups |
| **Supply chain** | Image signing (cosign), SBOMs, vulnerability scans on every build and Engine Package |
| **Audit** | Immutable log of API calls, tool calls and secret access, exportable to SIEM |
| **AI-specific** | Prompt-injection guardrails, PII redaction, tool scopes, approval steps, spend caps |

### 5.2 Observability
OpenTelemetry for traces, metrics and logs. Default stack: Prometheus, Loki, Tempo, Grafana (all replaceable). Platform dashboards cover database health, function latency, cold starts, GPU use and token cost.

### 5.3 Multi-tenancy and quotas
Organization, Project, Environment (dev, staging, prod). Quotas for CPU, memory, GPU, storage, vector count and tokens. Resource limits and priority classes protect against noisy neighbors.

### 5.4 Extensibility
- Engine Packages (stores), Runtime Packages (languages and sandboxes), Provider plugins (models), Connector plugins (knowledge sources).
- Webhooks and an event bus for external integrations.
- Console extensions through a plugin manifest.
- Plugins are versioned, signed and loaded through a stable gRPC or OCI interface.

---

## 6. Technology Choices

| Concern | Choice | Reason |
|---|---|---|
| Control plane language | Go | Kubernetes ecosystem, operators, static binaries |
| Hot-path services (gateway, invoker) | Go, Rust where latency demands | Performance and safety |
| Orchestration | Kubernetes (k3s for small installs) | Operators and autoscaling ecosystem |
| Metadata DB | PostgreSQL (CloudNativePG) | Reliable, supports pgvector internally |
| Messaging | NATS JetStream | Light footprint, streams plus request/reply |
| Gateway | Envoy or a Gateway API implementation | AuthN, rate limiting, observability |
| Autoscaling | KEDA plus an activator for scale to zero | Event-driven scaling |
| Sandboxing | gVisor, Firecracker or Kata | Layered isolation |
| CNI and policy | Cilium | Network policy, mTLS, visibility |
| Object storage | SeaweedFS (default) | See section 3.3 |
| Inference | vLLM, Ollama, TGI behind the Model Gateway | Open ecosystem |
| Secrets | OpenBao or Vault | Dynamic credentials |
| Console | TypeScript and React | Large contributor pool |
| Packaging | Helm, Kustomize, Terraform modules | Fits DevOps practice |

### 6.1 Integrated open-source tools

Archon Base integrates existing open-source projects instead of rebuilding them. Licenses change over time, so verify each one before bundling or redistributing.

| Area | Tools | Role in Archon Base | License (verify) |
|---|---|---|---|
| Vector engines | Qdrant, Chroma, Weaviate, Milvus, pgvector | Engine Packages | Apache 2.0, BSD-3, PostgreSQL |
| Inference | vLLM, Ollama, llama.cpp, Hugging Face TEI | Backends behind the Model Gateway; TEI for embeddings | Apache 2.0, MIT |
| Gateway routing | LiteLLM (or custom) | Provider routing, fallbacks, key management | MIT |
| Agent frameworks | LangChain and LangGraph, LlamaIndex, CrewAI, Haystack | First-class runtimes and templates, no lock-in to one framework | MIT, Apache 2.0 |
| Document parsing | Docling, Unstructured, Apache Tika | Parsers in knowledge pipelines | MIT, Apache 2.0 |
| Workflows | Temporal (or Argo Workflows) | Durable agent workflows instead of a custom engine | MIT, Apache 2.0 |
| LLM observability | OpenTelemetry, OpenLLMetry, Langfuse | Trace capture and trace UI options | Apache 2.0, MIT core |
| Evals | promptfoo, Ragas, DeepEval | Pluggable scorers in eval suites | MIT, Apache 2.0 |
| Guardrails and PII | NeMo Guardrails, Microsoft Presidio | Policy checks and redaction at the gateway | Apache 2.0, MIT |
| Sandboxing | gVisor, Firecracker, Kata Containers, Wasmtime | Function isolation tiers | Apache 2.0 |
| Cache and queues | Valkey, NATS JetStream | Semantic cache, agent memory, events | BSD-3, Apache 2.0 |
| Identity | Keycloak, Ory | OIDC, SSO | Apache 2.0 |
| Storage | SeaweedFS, Garage, Ceph | Object storage backends | Apache 2.0, AGPL (Garage), LGPL (Ceph) |
| Secrets | OpenBao | Dynamic credentials | MPL 2.0 |
| Platform | Kubernetes, KEDA, Cilium, CloudNativePG, cert-manager | Orchestration, scaling, networking, Postgres | Apache 2.0 |

Integration rules:
- **Adapter first.** Every integrated tool sits behind an Archon interface (Engine Package, provider plugin, scorer plugin), so it can be swapped.
- **Prefer permissive licenses.** Flag copyleft or source-available components (AGPL, SSPL, Elastic, BSL) and keep them optional.
- **Pin and scan.** Pin versions, sign images, and scan every bundled component in CI.
- **Contribute upstream** where possible instead of forking.

---

## 7. Deployment Strategy

### 7.1 Install tiers

| Tier | Target | Topology | Method |
|---|---|---|---|
| **Dev / single node** | Laptop or small VM | All services on one host, one vector engine, embedded k3s | Installer script or Docker Compose |
| **Team** | 3 nodes | k3s or kubeadm, HA control plane, replicated Postgres | Helm chart |
| **Production** | 5+ nodes, optional GPU | Dedicated pools (control, database, function, GPU, storage), multi-AZ | Helm plus Terraform modules (bare metal, VMware, AWS, GCP, Azure) |
| **Air-gapped** | Regulated environments | Offline bundle of images, models and Engine Packages | OCI bundle plus local registry |

### 7.2 Minimum resources

| Profile | CPU | RAM | Disk | Notes |
|---|---|---|---|---|
| Single node (evaluation) | 4 vCPU | 16 GB | 100 GB SSD | One small vector instance, a few functions |
| Team (3 nodes) | 8 vCPU each | 32 GB each | 500 GB NVMe each | HA control plane, replicated databases |
| Production | Workload dependent | Plan vector RAM at about 1.5 to 2x raw index size | NVMe for databases, SSD or HDD tier for objects | Add GPU nodes for inference |

### 7.3 Dependencies
Linux kernel 5.10 or newer (KVM if using microVMs), containerd, Kubernetes 1.28 or newer, a CSI driver with a fast storage class, a load balancer (MetalLB on bare metal), cert-manager, and optionally the NVIDIA GPU Operator.

### 7.4 Lifecycle operations
- **Upgrades**: versioned releases; `archon upgrade` runs pre-flight checks, upgrades the control plane first, then engines per instance.
- **Backup and DR**: platform state and tenant data back up to a user-defined target; restore runbooks and a tested restore command.
- **Configuration**: one `archon.yaml` values file with sensible defaults.
- **Multi-cluster**: later phase, one control plane managing several data planes.

---

## 8. Developer Experience

### CLI

```bash
archon init --template rag-chat       # scaffold a project
archon dev                            # local emulator
archon db create docs-index --engine qdrant --plan standard-2
archon bucket create support-docs
archon pipeline create docs-ingest --source bucket://support-docs --index docs-index
archon deploy ./support-agent
archon logs support-agent --follow
archon trace open <trace-id>
archon eval run support-agent --dataset golden-set
archon secrets set CRM_API_TOKEN
archon backup run docs-index
```

### API surface (REST and gRPC)
- `/v1/projects/{p}/databases`: CRUD, scale, backup, restore, credentials
- `/v1/projects/{p}/functions`: deploy, invoke, versions, rollback, logs
- `/v1/projects/{p}/buckets` and `/objects`: S3-compatible plus management API
- `/v1/projects/{p}/models`: registry and gateway config
- `/v1/projects/{p}/pipelines`: knowledge ingestion
- `/v1/projects/{p}/tools`: tool registry and MCP servers
- `/v1/projects/{p}/evals`: datasets, runs, results
- `/v1/projects/{p}/events`: subscriptions and webhooks
- `/v1/iam`: users, roles, service accounts, agent identities, API keys

All resources are declarative and idempotent, with versions and status conditions.

---

## 9. Key Trade-offs

| Decision | Chosen | Alternative | Rationale |
|---|---|---|---|
| Kubernetes required | Yes, k3s for small installs | Nomad, plain Docker | Operator and autoscaling ecosystem outweigh the weight |
| Default vector engine | Qdrant | Milvus | Lower operational cost; Milvus offered for scale |
| Unified vector API | Optional layer | Mandatory abstraction | Keeps engine-specific features available |
| Default sandbox | gVisor | Firecracker everywhere | Better density and startup; microVMs for untrusted code |
| Object storage default | SeaweedFS | MinIO | Licensing flexibility; adapter allows either |
| Metadata store | PostgreSQL | SQLite or etcd only | HA, multi-user and rich queries |
| Build vs integrate | Integrate best-of-breed OSS | Build everything | Focus on control plane, DX and integration glue |

---

## 10. Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Operational complexity of many engines | Strict Engine Package contract, conformance tests, small supported default set |
| Sandbox escape | Layered isolation, microVMs for untrusted code, regular reviews and fuzzing |
| Upstream license changes | Adapter interfaces, license tracking in CI, permissive defaults |
| Cold start latency for large models | Warm pools, node-local model cache, dedicated inference services |
| Data loss in self-managed setups | Backups on by default, restore drills, clear DR docs |
| Runaway agent cost or loops | Budgets, rate caps, kill switch, loop detection |
| Scope creep | Phase gates, each phase ships a usable product |
| Name and trademark conflicts | Search registries and trademarks for "Archon Base" before public launch |

---

## 11. Roadmap

### Phase 0: Foundations (weeks 0 to 6)
- Repo layout, CI, signing, release pipeline.
- Control plane skeleton: gateway, auth, project model, Postgres, NATS.
- Single-node installer.

### Phase 1: MVP (months 2 to 4)
- Qdrant, Chroma and pgvector Engine Packages, backup and restore, connection broker.
- S3-compatible storage on SeaweedFS, buckets, signed URLs.
- HTTP-triggered functions on gVisor (Python and TypeScript), secrets and bindings.
- CLI, basic console, local dev emulator (alpha).

### Phase 2: Agent platform (months 5 to 8)
- Model Gateway with vLLM and Ollama backends and token accounting.
- Knowledge pipelines with bucket and Git connectors.
- Tracing UI, eval suites, replay.
- Event triggers, cron, durable workflows, MCP server hosting.
- Weaviate package, microVM isolation, KEDA autoscaling, GPU pools.

### Phase 3: Production hardening (months 9 to 12)
- HA Kubernetes install, rolling upgrades, point-in-time restore.
- Milvus package, ScyllaDB (subject to license review).
- RBAC and policy engine, guardrails, approvals, agent identities.
- Budgets, quotas, semantic cache, index branching and migration.
- Model Registry and dataset versioning, artifact cache.

### Phase 4: Ecosystem and enterprise (year 2)
- Plugin SDK, marketplace and agent templates, A2A support.
- Multi-cluster and edge data planes.
- SSO/SCIM, BYOK, SIEM export, compliance templates, air-gapped bundles.
- Fine-tuning job orchestration.

---

## 12. Success Metrics
- Time from install to first deployed agent under 15 minutes.
- Function cold start p95 under 1 second for sandboxed functions without large models.
- Vector query p95 within 10 percent of native engine performance.
- Zero cross-tenant isolation findings in an external security review.
- Restore drills passing on every release.
- Eval regression gates adopted by a majority of production agents.

---

## 13. Open Questions
1. Licensing for the platform itself (Apache 2.0 core with optional commercial add-ons?).
2. Whether to offer a hosted edition later, and where the open-source boundary sits.
3. Which GPU sharing approach to standardize on across mixed hardware.
4. How deep to go on workflow orchestration versus integrating Temporal or similar.
5. Whether the knowledge pipeline ships in core or as a separate optional module.
