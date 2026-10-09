# Archon Base Product Control Plane TODO

**Purpose:** Build a self-hosted AI infrastructure control plane with the usability of products such as Supabase and Appwrite while keeping Archon Base focused on the infrastructure underneath AI applications.

**Product rule:** Do not present planned or preview capabilities as available. Every feature should be labeled **Available**, **Preview**, or **Planned** in the UI, API documentation, and marketing content.

**Primary positioning:**

> Archon Base gives teams one control plane to provision, operate, secure, and observe self-hosted AI infrastructure.

## Current baseline

- [x] Core v1 `Project` and `DatabaseInstance` resource models.
- [x] PostgreSQL persistence and transactional outbox writes.
- [x] Project and DatabaseInstance CRUD API handlers.
- [x] Student K3s installation script.
- [x] Local resource monitoring daemon.
- [ ] Authentication and authorization.
- [ ] Persistent web console; the current modular console remains a local preview.
- [ ] Production-ready controllers and reconciliation.
- [ ] Agent, model, storage, tool, and governance services.

## Definition of product-control-plane done

A user can create an account or sign in, create an organization and project, connect a self-hosted cluster, provision supported resources, inspect health and usage, manage team access, review audit events, and safely delete resources from a responsive web console. Every action is backed by an authenticated API, produces an auditable event, and clearly communicates whether the capability is available, preview, or planned.

---

## 0. Product architecture and contracts

### 0.1 Product boundaries

- [ ] Define the control-plane boundary versus the Kubernetes/data-plane boundary.
- [ ] Define which resources are owned by Archon Base and which are adapter-backed.
- [x] Define the organization, project, environment, cluster, and resource hierarchy.
- [ ] Define the difference between desired state, observed state, and provider state.
- [ ] Define product capability statuses: `available`, `preview`, `planned`, `unsupported`, `degraded`.
- [x] Define the initial single-organization local mode without pretending it is multi-tenant production.
- [ ] Define which features are available in local, single-node, team, and production deployment profiles.

### 0.2 Stable contracts

- [ ] Version the REST API under `/v1`.
- [ ] Publish OpenAPI for all control-plane endpoints.
- [x] Define stable error codes and field paths.
- [x] Define request IDs and correlation IDs.
- [ ] Define pagination, filtering, sorting, and search conventions.
- [ ] Define optimistic concurrency/resource-version behavior.
- [ ] Define idempotency-key behavior for mutating requests.
- [ ] Define asynchronous operation and long-running task responses.
- [ ] Define webhook/event contract for resource lifecycle changes.
- [ ] Define SDK compatibility and deprecation policy.

### 0.3 Product information architecture

- [ ] Create a shared navigation model for web console, CLI, API docs, and SDKs.
- [ ] Define global navigation: Overview, Projects, Infrastructure, Models, Data, Agents, Security, Observability, Settings.
- [ ] Define project-level navigation: Overview, Resources, Deployments, Activity, Secrets, Members, Settings.
- [ ] Define organization-level navigation: Overview, Projects, Clusters, Members, Roles, Audit, Usage, Settings.
- [ ] Define admin-only navigation separately from end-user navigation.
- [ ] Define empty, loading, error, degraded, and unavailable states for every page.
- [ ] Define responsive behavior for desktop, tablet, and mobile widths.

---

## 1. Web console foundation

### 1.1 Application shell

- [x] Choose and document the frontend runtime and component strategy: dependency-free local web app for the first slice.
- [x] Add accessible design tokens for colors, typography, spacing, motion, borders, and status states.
- [x] Build the local control-plane shell with sidebar, top bar, breadcrumbs, and user/context controls; authentication remains a follow-up.
- [x] Add organization and project context switcher surfaces.
- [x] Add environment and cluster context surfaces.
- [x] Add global search and notification entry points with explicit not-connected messaging.
- [ ] Add keyboard shortcuts and a visible shortcut help panel.
- [ ] Add command palette actions for common operations.
- [x] Add notification center entry point for asynchronous operation results and system warnings.
- [x] Add global status banner for degraded dependencies and maintenance windows.
- [x] Add route-level fallback rendering and recoverable planned-feature states.
- [ ] Add session-expiry handling with safe reauthentication.
- [x] Add keyboard focus, semantic headings, responsive layout, and reduced-motion behavior.

### 1.2 Shared UI components

- [ ] Data table with server-side pagination, filters, sorting, column visibility, and bulk actions.
- [ ] Resource status badge with text and icon, never color alone.
- [ ] Condition/timeline component for reconciliation state.
- [ ] YAML/JSON editor with schema validation and read-only mode.
- [ ] Log viewer with search, timestamps, levels, copy, and download.
- [ ] Metrics chart component with time range and refresh controls.
- [ ] Confirmation dialog for destructive actions with resource name and impact.
- [ ] Asynchronous operation drawer with progress, events, retry, and cancellation state.
- [ ] Secret input component that never displays secret values after save.
- [ ] Usage meter with clear units and profile limits.
- [ ] Empty-state component with documentation and next action.
- [ ] Feature-status component for Available, Preview, and Planned capabilities.
- [ ] Toasts that do not replace persistent error or audit information.

### 1.3 Design system and product voice

- [x] Keep the product console technical, clear, and infrastructure-oriented.
- [x] Use plain language for destructive and security-sensitive actions.
- [x] Avoid generic AI gradients, decorative dashboards, and fake activity data.
- [x] Show timestamps, sources, freshness, and unavailable states for preview surfaces.
- [x] Distinguish product health from infrastructure health.
- [x] Document copy standards through the local UI plan and route states.
- [ ] Add visual regression tests for core layouts and responsive breakpoints.

---

## 2. Authentication and account onboarding

### 2.1 Authentication

- [x] Add local development authentication that is clearly labeled development-only.
- [x] Add email/password authentication foundation with bcrypt credentials and bearer sessions; production hardening remains.
- [ ] Add OIDC/OAuth provider support.
- [ ] Add optional passkeys/WebAuthn.
- [ ] Add MFA enrollment and recovery flows.
- [x] Add server-side session revocation; device/session management UI remains.
- [ ] Add secure cookie settings for HTTPS embedded previews and production deployments.
- [ ] Add CSRF protection where cookie authentication is used.
- [ ] Add login rate limiting and abuse protection.
- [ ] Add account deletion and data export workflows.
- [ ] Add clear recovery behavior when the identity provider is unavailable.

### 2.2 First-run onboarding

- [ ] Create first organization flow.
- [ ] Create first project flow.
- [ ] Choose deployment profile: Local, Enthusiast/K3s, Single Node, Team, or Production.
- [ ] Show profile requirements before installation or connection.
- [ ] Run host prerequisite checks before cluster installation.
- [ ] Connect an existing K3s/Kubernetes cluster.
- [ ] Install Archon Base components into a selected namespace.
- [ ] Verify API, database, data plane, and monitor health.
- [ ] Show a setup checklist that can be resumed later.
- [ ] Seed only clearly marked demo data, never fake production activity.

---

## 3. User dashboard

### 3.1 Organization overview

- [ ] Show organization name, profile, connected clusters, projects, and members.
- [ ] Show control-plane health and data-plane health separately.
- [ ] Show recent operations and warnings.
- [ ] Show resource counts by type and lifecycle state.
- [ ] Show current capacity: CPU, memory, disk, GPU, and cluster nodes.
- [ ] Show usage trends only when backed by real measurements.
- [ ] Show recommended next actions based on actual incomplete setup.
- [ ] Show a clear student/K3s profile card with current versus recommended capacity.

### 3.2 Project overview

- [ ] Show project metadata, environment, region, owner, and lifecycle state.
- [ ] Show resources grouped by type and status.
- [x] Show project capacity and quota consumption in the local control-plane preview.
- [ ] Show recent project operations and audit activity.
- [ ] Show project members and effective permissions.
- [ ] Show project-level documentation links and examples.
- [ ] Support project archive and deletion with dependency checks.

### 3.3 Resource experience

- [ ] Create Project form with schema validation and defaults.
- [ ] Project detail page with spec, status, conditions, events, and metadata.
- [ ] Project edit flow with generation/concurrency handling.
- [ ] DatabaseInstance create form with engine, version, plan, storage, backup, and network settings.
- [ ] DatabaseInstance detail page with endpoint status, capacity, conditions, events, and recent operations.
- [ ] DatabaseInstance list with engine, plan, phase, health, storage, and last update.
- [ ] Resource YAML view and copy/download actions.
- [ ] Resource activity timeline linked to outbox/audit events.
- [ ] Resource deletion flow that explains child-resource and data-loss consequences.
- [ ] Preview UI for future resources without implying backend availability.

---

## 4. Admin panel

### 4.1 Admin access model

- [x] Define organization owner, admin, operator, developer, viewer, and billing/audit roles.
- [ ] Define platform administrator separately from organization administrator.
- [ ] Require explicit authorization for platform-wide actions.
- [ ] Add privileged-action confirmation and reason capture.
- [ ] Add break-glass access workflow with time limit and audit event.
- [ ] Add admin session timeout and reauthentication for sensitive actions.

### 4.2 Organization administration

- [ ] Manage organization profile and defaults.
- [ ] Invite, suspend, remove, and restore members.
- [x] Assign organization roles and resolve the authenticated organization membership; project-level permissions remain.
- [ ] View pending invitations and expiration.
- [ ] Configure allowed identity providers.
- [ ] Configure organization API keys and service accounts.
- [ ] Rotate and revoke keys with affected-resource preview.
- [ ] Configure retention and audit policies.

### 4.3 Platform administration

- [ ] Show connected control-plane instances and cluster inventory.
- [ ] Show node health, K3s/Kubernetes version, and capacity.
- [ ] Show operator/controller health and reconciliation backlog.
- [ ] Show migration status and schema version.
- [ ] Show outbox backlog, publish failures, and retry state.
- [ ] Show engine package versions and compatibility.
- [ ] Enable/disable adapters only through explicit configuration.
- [ ] Manage feature flags with scope, owner, expiry, and audit history.
- [ ] Add maintenance mode with visible user messaging.
- [ ] Add safe configuration validation before applying changes.

### 4.4 Audit and support tools

- [ ] Search audit events by actor, organization, project, resource, action, and time.
- [ ] Display before/after metadata without exposing secrets.
- [ ] Export audit events in a documented format.
- [ ] Add request ID and trace links for support investigation.
- [ ] Add read-only diagnostics bundle generation.
- [ ] Redact tokens, credentials, prompts, signed URLs, and secret values.
- [ ] Log all privileged support access.

---

## 5. Infrastructure and deployment management

### 5.1 K3s and Kubernetes

- [ ] Add cluster connection wizard.
- [ ] Validate Kubernetes API reachability and permissions.
- [ ] Validate supported Kubernetes/K3s versions.
- [ ] Validate storage class, ingress/load balancer, and DNS prerequisites.
- [ ] Validate optional NVIDIA GPU Operator and device plugin.
- [ ] Add namespace and ownership labels.
- [ ] Install CRDs and controllers idempotently.
- [ ] Add upgrade plan with compatibility checks and rollback guidance.
- [ ] Add cluster disconnect flow that preserves or explicitly removes resources.
- [ ] Add single-node and multi-node deployment profiles.
- [ ] Add air-gapped installation documentation and artifact checks.

### 5.2 Resource lifecycle

- [ ] Define reconcile state machine for every resource.
- [ ] Add create, update, pause, resume, delete, and retry operation states.
- [ ] Add bounded retries and exponential backoff.
- [ ] Add dependency graph and deletion protection.
- [ ] Add observed-generation tracking.
- [ ] Add provider error normalization.
- [ ] Add drift detection and remediation policy.
- [ ] Add dry-run/preview changes before apply.
- [ ] Add rollback where provider capabilities allow it.
- [ ] Add import/adopt flow for existing supported resources.

### 5.3 Capacity and scheduling

- [x] Show CPU, memory, disk, and GPU capacity in the local control-plane preview.
- [x] Show allocatable versus requested versus used capacity for the selected project.
- [x] Add a read-only resource request/limit allocation management surface; persist edits after the resource API and authorization layer are connected.
- [ ] Add GPU profile selection and compatibility warnings.
- [ ] Add placement and node-selector configuration.
- [ ] Add capacity warnings before provisioning.
- [ ] Add student-profile warnings when the host is below 8 GB RAM, 2 cores, or 20 GB disk.
- [ ] Add no-GPU guidance for CPU-only practice.

---

## 6. AI platform modules

### 6.1 Models and model gateway

- [ ] Define Model and ModelEndpoint v1 resources.
- [ ] Add local OpenAI-compatible endpoint registration.
- [ ] Add Ollama/vLLM provider adapters.
- [ ] Add external provider configuration without storing secrets in plain text.
- [ ] Add model routing and fallback policies.
- [ ] Add health checks and latency/error reporting.
- [ ] Add token usage accounting when reliable provider data is available.
- [ ] Add model access policies and project bindings.
- [ ] Clearly label unsupported providers and preview adapters.

### 6.2 Agents and runtimes

- [ ] Finalize AgentFunction manifest schema.
- [ ] Add agent deployment UI and YAML editor.
- [ ] Add runtime, entrypoint, resource, trigger, and binding validation.
- [ ] Add trusted local subprocess mode with explicit warning.
- [ ] Add production runtime abstraction.
- [ ] Add gVisor integration before claiming sandbox isolation.
- [ ] Add microVM runtime integration only after threat-model and lifecycle review.
- [ ] Add invocation history, logs, timeout, crash, and retry views.
- [ ] Add tool and data-binding permission review.

### 6.3 Data and knowledge

- [ ] Add supported vector engine catalog.
- [x] Add Qdrant provisioning and health integration for the local Early Access slice.
- [ ] Add pgvector option where supported.
- [ ] Add object-storage resource and bucket policy model.
- [ ] Add dataset/artifact metadata and retention.
- [ ] Add RAG pipeline model only after storage and model contracts stabilize.
- [ ] Add backup and restore workflows.
- [ ] Add data deletion and retention confirmation.

### 6.4 Tools and MCP

- [ ] Define Tool and MCPServer resources.
- [ ] Add connection test without exposing credentials.
- [ ] Add allowlisted tool bindings.
- [ ] Add egress and network policy integration.
- [ ] Add tool invocation audit events.
- [ ] Add credential rotation and revocation.
- [ ] Add explicit preview status until runtime isolation is available.

---

## 7. Security and governance

- [ ] Add project-scoped authorization checks to every API handler.
- [ ] Add object-level authorization tests.
- [ ] Add secret encryption at rest with configurable key provider.
- [ ] Add secret versioning, rotation, and revocation.
- [ ] Add network policy abstraction and default-deny guidance.
- [ ] Add egress controls for agents and tools.
- [ ] Add workload identity and service-account bindings.
- [ ] Add image provenance, signature, SBOM, and vulnerability status.
- [ ] Add policy checks before deployment.
- [ ] Add approval workflow for privileged or production actions.
- [ ] Add audit events for auth, resource changes, secrets, bindings, and admin access.
- [ ] Add privacy controls for prompts, documents, traces, and logs.
- [ ] Add data export and deletion workflows.
- [ ] Run threat modeling for every new runtime or provider adapter.

---

## 8. Observability and operations

- [ ] Add API request metrics, latency, status, and error metrics.
- [ ] Add reconciliation duration, success, failure, and backlog metrics.
- [ ] Add cluster/node CPU, memory, disk, and GPU metrics.
- [ ] Add logs with request IDs and structured fields.
- [ ] Add traces across API, controller, provider adapter, and runtime.
- [ ] Add service health and dependency health pages.
- [ ] Add alert definitions for capacity, failed reconciliation, disk pressure, and auth anomalies.
- [ ] Add maintenance and incident status pages inside the console.
- [ ] Add runbook links beside operational failures.
- [ ] Add support bundle export with secrets and sensitive payloads redacted.
- [ ] Add retention configuration for logs, metrics, traces, and audit events.

---

## 9. Usage, quotas, and billing-ready foundations

- [ ] Define usage event schema independently from billing provider choice.
- [ ] Track resource-hours, storage, vector points, tokens, GPU time, and agent invocations where measurable.
- [ ] Add organization/project quotas.
- [ ] Add quota enforcement before provisioning.
- [ ] Add usage dashboard with source and freshness indicators.
- [ ] Add exportable usage reports.
- [ ] Add billing provider integration only after usage semantics stabilize.
- [ ] Keep self-hosted installations functional without a hosted billing dependency.

---

## 10. Developer experience

- [ ] Add interactive `archon init` for local and K3s profiles.
- [ ] Add `archon doctor` for prerequisite, cluster, API, and monitor checks.
- [ ] Add `archon cluster connect` and `archon cluster status`.
- [ ] Add `archon project create/get/update/delete`.
- [ ] Add `archon db create/get/list/update/delete`.
- [ ] Add `archon agent deploy/invoke/logs`.
- [ ] Add `archon model list/register/test`.
- [ ] Add `archon monitor status` and `archon monitor install`.
- [ ] Add JSON output for every command.
- [ ] Add shell completion after command shapes stabilize.
- [ ] Generate SDKs and API examples from the OpenAPI contract.
- [ ] Add Terraform/provider or declarative configuration only after API lifecycle semantics stabilize.

---

## 11. Documentation and education

- [x] Document enthusiast/student hardware requirements.
- [x] Document single-node K3s installation.
- [x] Document local resource monitoring.
- [ ] Add product-control-plane overview.
- [ ] Add dashboard user guide.
- [ ] Add admin panel user guide.
- [ ] Add cluster connection guide.
- [ ] Add resource lifecycle guide.
- [ ] Add authentication and authorization guide.
- [ ] Add security hardening guide.
- [ ] Add backup and restore guide.
- [ ] Add migration and upgrade guide; the initial runner and version ledger are now implemented.
- [ ] Add troubleshooting guide organized by error code and request ID.
- [ ] Add architecture decision records for major control-plane choices.
- [ ] Add availability labels to all public documentation.

---

## 12. Quality gates

- [ ] Add CI for Go tests, vet, formatting, frontend checks, YAML, JSON, and Markdown links.
- [ ] Add API contract tests generated from OpenAPI.
- [ ] Add browser tests for authentication, onboarding, resource CRUD, admin access, and destructive flows.
- [ ] Add accessibility tests for core screens.
- [ ] Add responsive visual regression tests.
- [ ] Add PostgreSQL and Kubernetes integration tests.
- [ ] Add disposable K3s or Kubernetes integration environment.
- [ ] Add failure-injection tests for API, database, controller, provider, and runtime failures.
- [ ] Add load tests for resource listing, events, metrics, and dashboard queries.
- [ ] Add security tests for authorization bypass, secret leakage, SSRF, CSRF, and injection.
- [ ] Add release checklist requiring product-status labels and migration notes.

---

## Recommended delivery sequence

### Milestone 1 — Console foundation

- [x] Frontend shell and design system.
- [x] Split the local frontend into state, shared components, route renderers, and an application bootstrap.
- [ ] Local development authentication.
- [ ] Organization/project switchers.
- [ ] Project and DatabaseInstance pages backed by the existing API.
- [ ] Real loading, error, empty, and status states.

### Milestone 2 — Secure onboarding

- [ ] OIDC or selected production authentication.
- [ ] Organization membership and roles.
- [ ] K3s prerequisite and cluster connection wizard.
- [ ] Student profile capacity and monitor integration.
- [ ] Audit events for account and resource actions.

### Milestone 3 — Operational control plane

- [ ] Reconciliation status and operations.
- [ ] Cluster/node capacity dashboard.
- [ ] Admin panel.
- [ ] Logs, metrics, traces, and diagnostics.
- [ ] Backup, restore, upgrade, and rollback workflows.

### Milestone 4 — AI infrastructure modules

- [ ] Model gateway.
- [ ] Agent runtime.
- [ ] Vector/data services.
- [ ] Object storage.
- [ ] Tools and MCP.
- [ ] Security policies and bindings.

### Milestone 5 — Production readiness

- [ ] HA deployment.
- [ ] Multi-cluster support.
- [ ] Policy and approval workflows.
- [ ] Usage and quota enforcement.
- [ ] Disaster recovery and air-gapped installation.
- [ ] Public release documentation and compatibility matrix.

## Explicit non-goals for the first control-plane release

- [ ] Do not build a replacement chat interface.
- [ ] Do not claim to replace Supabase or Appwrite for general application backends.
- [ ] Do not claim production sandbox isolation before gVisor or microVM integration is tested.
- [ ] Do not add hosted billing as a hard dependency for self-hosted deployments.
- [ ] Do not expose privileged Kubernetes credentials to agents or browser clients.
- [ ] Do not ship fake metrics, fake activity, or fake health states in the console.
