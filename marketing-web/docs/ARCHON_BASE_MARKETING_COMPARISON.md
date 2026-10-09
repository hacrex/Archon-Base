# Archon Base --- Marketing Website Competitive Comparison

> **Purpose:** Marketing and positioning reference for the Archon Base
> website.
>
> **Primary positioning:** Archon Base is an open-source self-hosted AI
> Backend for agents, models, data, tools, and compute.

------------------------------------------------------------------------

## 1. Executive Positioning

### Archon Base

**Archon Base is an open-source self-hosted AI Backend for agents, models,
data, tools, and compute.**

It provides a unified backend for building and operating AI applications:

-   Agent runtimes
-   Model gateways
-   Vector databases
-   Object storage
-   Tool and MCP infrastructure
-   RAG and knowledge pipelines
-   Sandboxed execution
-   GPU workloads
-   Identity, policies, bindings, and secrets
-   Observability, evaluations, usage, and governance
-   Kubernetes-native deployment and lifecycle management

The goal is not to replace every AI application framework or generic
application backend.

The goal is to provide the AI-specific backend services they need, with
self-hosting, portability, and developer control built in.

------------------------------------------------------------------------

# 2. The Market Landscape

The AI developer stack is increasingly divided into several layers.

``` text
┌─────────────────────────────────────────────────────────────┐
│                    AI APPLICATIONS                          │
│ Chat apps • SaaS AI • Coding agents • Research apps       │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│              AI APPLICATION / WORKFLOW LAYER               │
│ Dify • Langflow • Open WebUI • Custom AI applications      │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│                 AGENT EXECUTION LAYER                      │
│ E2B • Sandboxes • Agent runtimes • Tool execution           │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│                    SELF-HOSTED AI BACKEND                  │
│                         ARCHON BASE                         │
│ Auth • Projects • Vectors • Storage • Functions • Models   │
│ Tools • RAG • Policies • Evals • Observability              │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│                  COMPUTE / DATA PLANE                       │
│ Kubernetes • GPUs • Qdrant • S3 • vLLM • NATS • Cilium     │
└─────────────────────────────────────────────────────────────┘
```

This product positioning is important.

Archon Base should not primarily market itself as another chat UI,
visual AI builder, or generic non-AI backend-as-a-service platform.

------------------------------------------------------------------------

# 3. Competitive Landscape

  ----------------------------------------------------------------------------
  Platform          Primary Category       Core Strength     Archon Base
                                                             Relationship
  ----------------- ---------------------- ----------------- -----------------
  **Supabase**      Backend-as-a-Service   Postgres, Auth,   Complementary
                                           Storage,          
                                           Realtime,         
                                           Functions         

  **Appwrite**      Open-source Backend /  Auth, Databases,  Complementary /
                    Agent Backend          Storage,          adjacent
                                           Functions,        
                                           Hosting, MCP      

  **E2B**           Agent Execution        Secure microVM    Runtime partner /
                    Infrastructure         sandboxes         adapter

  **Open WebUI**    Self-hosted AI         Chat, models,     Application layer
                    Platform / Interface   RAG, tools,       
                                           agents            

  **Dify**          AI Application         RAG, agents,      Application layer
                    Platform               workflows, LLMOps 

  **Langflow**      Visual AI Builder      Agent and RAG     Application layer
                                           workflows, APIs,  
                                           MCP               

  **Archon Base**   AI Infrastructure      Agents, models,   Infrastructure
                    Control Plane          compute, data,    layer
                                           tools, governance 
  ----------------------------------------------------------------------------

The objective is not to claim that these platforms are interchangeable.

They solve different problems.

------------------------------------------------------------------------

# 4. Archon Base vs Supabase

## What Supabase is

Supabase is a developer platform centered around Postgres, with
integrated authentication, storage, realtime functionality, edge
functions, and vector search capabilities.

Supabase also supports self-hosting.

## What Archon Base is

Archon Base is designed around AI infrastructure rather than general
application backend infrastructure.

Its primary resources include:

-   AI agents
-   Agent runtimes
-   Model endpoints
-   GPU workloads
-   Vector databases
-   Object storage for AI artifacts
-   Tools and MCP servers
-   Knowledge pipelines
-   Evaluations
-   Sandboxed execution
-   AI-specific policies and governance

## Positioning

``` text
Supabase
Application Backend
        ↓
Postgres • Auth • Storage • Realtime • Functions

Archon Base
AI Infrastructure
        ↓
Agents • Models • GPU • Vector DB • Storage
Tools • RAG • Sandboxes • Evals • Governance
```

### Marketing message

> **Supabase gives your application a backend. Archon Base gives your AI
> workloads an infrastructure platform.**

### Potential integration

``` text
Application
    │
    ├── Supabase
    │     ├── Application DB
    │     ├── Auth
    │     └── Realtime
    │
    └── Archon Base
          ├── Agents
          ├── Models
          ├── GPU
          ├── RAG
          ├── Vector DB
          └── Sandboxed execution
```

**Do not position Archon Base as a replacement for Supabase's Postgres
developer experience.**

------------------------------------------------------------------------

# 5. Archon Base vs Appwrite

## What Appwrite is

Appwrite provides an open-source backend platform covering
authentication, databases, storage, functions, messaging, realtime
services, hosting, and newer agent/MCP capabilities.

Its current positioning increasingly targets agents and developers.

## What makes Archon Base different

Archon Base focuses on the infrastructure required to operate AI
workloads themselves.

The distinction is:

``` text
Appwrite
Build and deploy applications.

Archon Base
Build and operate AI backends for applications and agents.
```

### Architectural difference

  Area                                Appwrite           Archon Base
  ----------------------------------- ------------------ --------------------------
  Authentication                      Core               Core
  Application database                Core               Metadata/control plane
  Object storage                      Core               AI/data/artifact focused
  Functions                           Core               Agent execution
  AI models                           Integration        Core infrastructure
  GPU lifecycle                       Not primary        Core concern
  Vector databases                    Not primary        Core
  Sandboxed AI execution              Not primary        Core
  AI model gateway                    Not primary        Core
  Agent governance                    Growing            Core
  Kubernetes infrastructure control   Limited/indirect   Core
  AI workload scheduling              Not primary        Core

### Marketing message

> **Appwrite is a general backend for applications. Archon Base is an
> AI Backend for applications and agents.**

------------------------------------------------------------------------

# 6. Archon Base vs E2B

## What E2B is

E2B specializes in running AI agents inside isolated environments. Its
platform uses microVM-based sandboxes and provides execution
capabilities such as shells, filesystems, browsers, networking, storage,
observability, and lifecycle management.

E2B supports cloud, BYOC, private, and self-hosted deployment models.

## Where Archon Base differs

E2B is highly focused on the execution boundary.

Archon Base is broader:

``` text
E2B
Agent
  ↓
Secure execution environment
  ↓
MicroVM
```

``` text
Archon Base
Project
  ↓
Identity / Policy
  ↓
Agent
  ↓
Runtime
  ├── gVisor
  ├── Firecracker
  ├── Kata
  ├── Containers
  └── WASM
  ↓
Models / Tools / Data / Storage / Network
```

## Strategic approach

Archon Base should avoid unnecessarily rebuilding specialized sandbox
infrastructure.

A runtime abstraction can support multiple execution backends:

``` text
RuntimeProvider
├── ContainerRuntime
├── GVisorRuntime
├── FirecrackerRuntime
├── KataRuntime
├── WASMRuntime
└── E2BRuntime
```

### Marketing message

> **E2B specializes in the sandbox. Archon Base manages the AI
> infrastructure around the sandbox.**

------------------------------------------------------------------------

# 7. Archon Base vs Open WebUI

## What Open WebUI is

Open WebUI is a self-hosted AI platform and interface supporting local
and cloud models, RAG, tools, agents, plugins, web search, and other AI
interaction capabilities.

It can connect to providers such as Ollama, OpenAI-compatible APIs,
Anthropic, and vLLM.

## Architectural distinction

``` text
Open WebUI
        ↓
User Interface
        ↓
Models / Agents / Knowledge / Tools
```

``` text
Open WebUI
        ↓
      Archon Base
        ↓
Model Gateway • Agents • RAG • Vector DB
GPU • Storage • Sandboxes • Governance
```

### Marketing message

> **Open WebUI is where users interact with AI. Archon Base provides the
> backend services that power AI agents, data, tools, and workflows.**

Archon Base can therefore be positioned as infrastructure that Open
WebUI and similar interfaces can consume.

------------------------------------------------------------------------

# 8. Archon Base vs Dify

## What Dify is

Dify is an open-source LLM application platform focused on building AI
applications using workflows, agents, RAG, model integrations, prompt
tooling, APIs, and LLMOps capabilities.

## The key distinction

Dify focuses on **building AI applications**.

Archon Base focuses on **operating AI infrastructure**.

``` text
Dify
Build the AI application.

Archon Base
Provision and operate the infrastructure.
```

### Example

A company could use:

``` text
Dify
  ↓
Application workflow
  ↓
Archon Base API / MCP
  ↓
Archon Model Gateway
  ↓
vLLM / external provider
```

And:

``` text
Dify
  ↓
RAG workflow
  ↓
Archon Base
  ├── Qdrant
  ├── Object Storage
  ├── Knowledge Pipeline
  └── Model Gateway
```

### Marketing message

> **Dify helps you build AI applications. Archon Base provides the
> self-hosted backend services those applications connect to.**

------------------------------------------------------------------------

# 9. Archon Base vs Langflow

## What Langflow is

Langflow is a low-code AI builder for agentic and RAG applications. It
provides visual flows, reusable components, APIs, integrations, agent
capabilities, and MCP support.

## Architectural distinction

``` text
Langflow
Visual AI composition
        ↓
Agent / RAG / workflow
        ↓
Archon Base
```

Archon Base can provide:

-   Model endpoints
-   Agent runtimes
-   Vector databases
-   Object storage
-   Secrets
-   Tool infrastructure
-   Sandboxed execution
-   Usage and governance
-   Kubernetes resources

### Marketing message

> **Langflow helps you compose AI workflows. Archon Base provides the
> backend resources and functions those workflows use.**

------------------------------------------------------------------------

# 10. The Most Important Positioning

The website should avoid:

> "Archon Base is an alternative to Supabase, Appwrite, Dify, Open
> WebUI, Langflow and E2B."

That creates an unnecessarily broad competitive claim.

Instead:

> **Archon Base is the self-hosted AI Backend for applications and
> agents.**

And:

> **Build with the AI tools you already love. Run them on infrastructure
> you control.**

------------------------------------------------------------------------

# 11. Recommended Website Category

### Primary category

**Open-Source AI Infrastructure**

### Secondary category

**Self-Hostable AI Cloud Platform**

### Technical category

**Kubernetes-Native AI Infrastructure Control Plane**

### Short category

**AI Infrastructure Control Plane**

------------------------------------------------------------------------

# 12. Recommended Homepage Hero

## Option A --- Infrastructure-first

### Headline

**The Open-Source Infrastructure Layer for AI**

### Subheadline

Build, run, and operate AI agents, models, vector databases, tools, and
GPU workloads on infrastructure you control.

### CTA

**Get Started**

**View GitHub**

------------------------------------------------------------------------

## Option B --- Self-hosting-first

### Headline

**Your AI Cloud. Your Infrastructure. Your Control.**

### Subheadline

A self-hostable AI infrastructure platform for agents, models, data,
tools, and compute --- built on Kubernetes and open source.

### CTA

**Deploy Archon Base**

**Explore the Architecture**

------------------------------------------------------------------------

## Option C --- Platform-first

### Headline

**One Control Plane for Your AI Infrastructure**

### Subheadline

Provision agents, models, vector databases, storage, tools, sandboxes,
and GPU workloads through one API, CLI, and control plane.

### CTA

**Start Building**

**View GitHub**

------------------------------------------------------------------------

# 13. Recommended One-Line Description

> **Archon Base is an open-source, self-hostable AI infrastructure
> control plane for agents, models, data, tools, and compute.**

------------------------------------------------------------------------

# 14. Recommended Short Description

> Archon Base brings the infrastructure behind modern AI applications
> into one self-hostable control plane. Provision agents, model
> endpoints, vector databases, object storage, tools, sandboxes, and GPU
> workloads while keeping your data and infrastructure under your
> control.

------------------------------------------------------------------------

# 15. The "Think of It As" Message

### Simple version

> **Think of Archon Base as a Kubernetes-native backend platform for AI
> infrastructure.**

### Familiar comparison

> **Supabase and Appwrite provide backend infrastructure for
> applications. Archon Base provides infrastructure for AI workloads.**

### Technical comparison

> **Kubernetes gives you primitives. Archon Base gives you an AI-focused
> control plane on top of those primitives.**

------------------------------------------------------------------------

# 16. Feature Comparison for Marketing

  -----------------------------------------------------------------------------------------------------------------------------------
  Capability             Archon     Supabase           Appwrite                 E2B       Open WebUI        Dify             Langflow
                           Base                                                                                  
  ------------------- --------- ------------ ------------------ ------------------- ---------------- ----------- --------------------
  Self-hosting                ✓            ✓                  ✓                   ✓                ✓           ✓                    ✓

  AI infrastructure           ✓          ---                ---             Partial              ---         ---                  ---
  control plane                                                                                                  

  Agent runtime               ✓    Functions          Functions                   ✓            Agent           ✓                    ✓
                                                                                         integration             

  Sandboxed execution         ✓          ---                ---                   ✓   Execution/tool           ✓            Dependent
                                                                                               layer             

  Model gateway               ✓      Limited            Limited                 ---                ✓           ✓                    ✓

  Vector                      ✓     pgvector   Database-focused                 ---     Integrations           ✓         Integrations
  infrastructure                                                                                                 

  Object storage              ✓            ✓                  ✓                   ✓     Data storage           ✓         Integrations

  GPU workload                ✓          ---                ---                   ✓              ---         ---                  ---
  management                                                                                                     

  RAG infrastructure          ✓     Building            Limited                 ---                ✓           ✓                    ✓
                                      blocks                                                                     

  Tool/MCP                    ✓      Limited                  ✓                   ✓                ✓           ✓                    ✓
  infrastructure                                                                                                 

  AI evaluations              ✓          ---                ---       Agent-focused          Limited           ✓   Workflow-dependent

  AI governance               ✓      General            General   Execution-focused                ✓           ✓                    ✓

  Kubernetes-native           ✓   Deployment  Deployment option                   ✓        Supported   Supported            Supported
  infrastructure                      option                                                                     

  Multi-runtime               ✓          ---          Functions             MicroVM    Plugins/tools     Sandbox           Components
  execution                                                                                                      

  Declarative                 ✓      Partial            Partial                   ✓          Partial     Partial              Partial
  infrastructure                                                                                                 
  -----------------------------------------------------------------------------------------------------------------------------------

**Important:** This table is a positioning aid, not a claim that the
other platforms lack every capability listed as "---". Capabilities and
product boundaries evolve rapidly.

------------------------------------------------------------------------

# 17. The Core Differentiator

Archon Base should own this concept:

## Infrastructure Abstraction for AI

Instead of forcing developers to assemble:

``` text
Kubernetes
+
PostgreSQL
+
NATS
+
Qdrant
+
S3
+
vLLM
+
GPU Operator
+
Cilium
+
OpenBao
+
gVisor
+
Observability
+
Agent runtime
+
Model gateway
+
RAG pipeline
+
Policy
```

Archon Base provides a unified control plane over these components.

``` text
                     ARCHON BASE
                         │
       ┌─────────────────┼─────────────────┐
       │                 │                 │
     Agents           Models             Data
       │                 │                 │
   Runtime           Gateway        Vector / Object
       │                 │                 │
       └─────────────────┼─────────────────┘
                         │
                    Governance
                         │
                 Kubernetes Data Plane
```

------------------------------------------------------------------------

# 18. Why Kubernetes Alone Is Not Enough

Kubernetes provides primitives for:

-   Containers
-   Networking
-   Scheduling
-   Storage
-   Secrets
-   Services
-   Workloads

But an AI platform still needs higher-level concepts:

``` text
AI Agent
Model
Model Endpoint
Vector Database
Knowledge Base
Tool
MCP Server
Inference Pool
GPU Pool
Evaluation
Prompt
Binding
Credential
AI Usage
AI Policy
```

Archon Base turns infrastructure primitives into AI-native platform
resources.

### Website message

> **Kubernetes manages workloads. Archon Base manages AI workloads.**

------------------------------------------------------------------------

# 19. The Open-Source Story

Archon Base should emphasize:

### Self-hostable

Run it on your own infrastructure.

### Open architecture

Use the components that fit your environment.

### Provider-neutral

Connect local models, open-weight models, cloud models, and different
infrastructure providers.

### Kubernetes-native

Use Kubernetes as the underlying data-plane foundation.

### Adapter-first

Integrate mature open-source projects rather than rebuilding every
component.

### Portable

Move workloads between bare metal, VMs, private clouds, and public
clouds.

------------------------------------------------------------------------

# 20. The Ecosystem Story

Archon Base should be presented as an ecosystem rather than an isolated
application.

``` text
                 AI APPLICATIONS
                       │
       ┌───────────────┼────────────────┐
       │               │                │
   Open WebUI        Dify          Langflow
       │               │                │
       └───────────────┼────────────────┘
                       │
                   API / MCP
                       │
                 ARCHON BASE
                       │
     ┌─────────────────┼─────────────────┐
     │                 │                 │
   Models            Agents             Data
     │                 │                 │
   vLLM             Runtime          Qdrant/S3
   Ollama           Sandbox          PostgreSQL
   Providers        Tools            Storage
                       │
                 Kubernetes
                       │
              Compute / GPU / Network
```

This creates a more powerful narrative:

> **Archon Base does not replace the AI ecosystem. It connects and
> operates it.**

------------------------------------------------------------------------

# 21. Website Feature Pillars

The marketing site should organize the product around 7--8 major pillars
rather than listing dozens of technologies.

## 1. AI Compute

Run inference and AI workloads on CPU, GPU, or dedicated accelerator
pools.

## 2. Agent Runtime

Deploy isolated, scalable AI agents and tool-execution workloads.

## 3. Model Gateway

One API for local, self-hosted, and external models.

## 4. AI Data

Provision vector databases, object storage, datasets, embeddings, and
knowledge stores.

## 5. Tools & MCP

Register, govern, expose, and consume AI tools and MCP servers.

## 6. Security & Governance

Identity, policies, secrets, isolation, egress controls, approvals, and
audit.

## 7. Observability & Evals

Trace agent execution, monitor infrastructure, measure cost and latency,
and evaluate model behavior.

## 8. Infrastructure Control

Manage AI resources declaratively through API, CLI, Terraform, and
Kubernetes-native resources.

------------------------------------------------------------------------

# 22. Who Archon Base Is For

## AI Engineers

Deploy models, agents, RAG systems, and tool-enabled applications
without assembling the entire infrastructure stack manually.

## Platform Engineers

Provide an internal AI platform on top of Kubernetes.

## DevOps / SRE Teams

Operate AI infrastructure using familiar infrastructure-as-code,
observability, networking, security, and lifecycle patterns.

## Startups

Build AI products without immediately depending on a
hyperscaler-specific AI platform.

## Enterprises

Run AI workloads inside controlled infrastructure, including private,
regulated, and potentially air-gapped environments.

## Open-Source Builders

Self-host and customize the complete AI infrastructure stack.

------------------------------------------------------------------------

# 23. Key Marketing Messages

### Message 1

> **Build AI applications without rebuilding AI infrastructure.**

### Message 2

> **Your models. Your data. Your GPUs. Your infrastructure.**

### Message 3

> **One control plane for agents, models, data, tools, and compute.**

### Message 4

> **Open-source AI infrastructure that runs where you need it.**

### Message 5

> **From one server to Kubernetes clusters.**

### Message 6

> **Use the AI frameworks you like. Archon Base runs underneath them.**

------------------------------------------------------------------------

# 24. Recommended Comparison Section on the Website

## "Where Archon Base Fits"

Instead of a traditional competitor attack page, use a neutral ecosystem
diagram.

``` text
                     BUILD WITH
       ┌────────────┬────────────┬────────────┐
       │            │            │            │
   Open WebUI      Dify      Langflow      Custom Apps
       │            │            │            │
       └────────────┴──────┬─────┴────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │  ARCHON BASE  │
                    │               │
                    │ AI CONTROL    │
                    │ PLANE         │
                    └───────┬───────┘
                            │
       ┌────────────────────┼────────────────────┐
       │                    │                    │
     Models               Agents                Data
       │                    │                    │
   vLLM/Ollama         Sandboxes             Qdrant
   Cloud APIs          Tools/MCP             S3
   Model Gateway       Workflows             PostgreSQL
       │                    │                    │
       └────────────────────┼────────────────────┘
                            ▼
                     Kubernetes / GPU
```

### Section headline

**The infrastructure layer for the AI ecosystem.**

### Section copy

> Use Archon Base with the AI applications, frameworks, models, and
> infrastructure you already use. Archon Base provides the control plane
> that connects them.

------------------------------------------------------------------------

# 25. Competitive Messaging Rules

## Do

-   Explain the architectural layer.
-   Show integrations.
-   Emphasize self-hosting and infrastructure ownership.
-   Explain Kubernetes-native design.
-   Highlight AI-specific infrastructure abstractions.
-   Show how Archon Base works with existing AI applications.
-   Use concrete technical capabilities.
-   Present other platforms accurately.

## Do not

-   Claim Archon Base replaces every platform.
-   Claim competitors lack features they actually provide.
-   Market every feature as a unique invention.
-   Use vague "all-in-one AI platform" messaging without explaining the
    infrastructure layer.
-   Turn the website into a long competitor feature checklist.
-   Overpromise enterprise capabilities before they are implemented.
-   Present roadmap features as currently available.

------------------------------------------------------------------------

# 26. Recommended Website Navigation

``` text
Home
Platform
  ├── Agents
  ├── Models
  ├── AI Data
  ├── Tools & MCP
  ├── Compute & GPU
  ├── Security
  └── Observability

Architecture
Integrations
Developers
  ├── Docs
  ├── CLI
  ├── API
  ├── SDKs
  └── Terraform

Deploy
  ├── Local
  ├── Single Node
  ├── Kubernetes
  └── Air-Gapped

Community
GitHub
Roadmap
```

------------------------------------------------------------------------

# 27. Recommended SEO Positioning

## Primary keywords

-   open source AI infrastructure
-   self hosted AI platform
-   AI infrastructure platform
-   self hosted AI infrastructure
-   AI infrastructure control plane
-   Kubernetes AI platform
-   open source AI cloud
-   self hosted AI cloud
-   AI agent infrastructure
-   Kubernetes AI infrastructure

## Secondary keywords

-   self hosted LLM infrastructure
-   GPU AI platform
-   AI agent runtime
-   self hosted AI agents
-   AI model gateway
-   self hosted RAG infrastructure
-   vector database platform
-   AI Kubernetes platform
-   open source agent infrastructure
-   AI platform for Kubernetes

------------------------------------------------------------------------

# 28. Recommended Brand Statement

> **Archon Base is the open-source control plane for AI
> infrastructure.**
>
> Run agents, models, vector databases, storage, tools, and GPU
> workloads on infrastructure you control.
>
> Built for Kubernetes. Designed for self-hosting. Open by architecture.

------------------------------------------------------------------------

# 29. Recommended Homepage Closing Statement

> The future of AI infrastructure should not require every team to
> assemble the same stack from scratch.
>
> Archon Base provides the control plane.
>
> You choose the models.
>
> You choose the runtimes.
>
> You choose the data stores.
>
> You choose the infrastructure.
>
> **Archon Base connects and operates them.**

------------------------------------------------------------------------

# 30. Final Positioning Summary

  -----------------------------------------------------------------------
  Question                            Answer
  ----------------------------------- -----------------------------------
  What is Archon Base?                Open-source AI infrastructure
                                      control plane

  Who is it for?                      AI engineers, platform teams,
                                      DevOps/SRE, startups, enterprises,
                                      OSS builders

  What does it manage?                Agents, models, compute, data,
                                      tools, storage, policies and AI
                                      operations

  Where does it run?                  Self-hosted infrastructure from a
                                      single node to Kubernetes

  What sits above it?                 Open WebUI, Dify, Langflow, custom
                                      AI applications and agent
                                      frameworks

  What sits below it?                 Kubernetes, GPUs, storage,
                                      networks, model servers and data
                                      services

  Is it a chat UI?                    No

  Is it a generic BaaS?               No

  Is it an AI workflow builder?       No

  Is it a sandbox product only?       No

  Is it a model provider?             No

  What is its core role?              Operate the infrastructure behind
                                      AI applications

  Core message                        **The infrastructure layer for
                                      self-hosted AI.**
  -----------------------------------------------------------------------

------------------------------------------------------------------------

# Sources Used for Current Market Positioning

The comparison was checked against current official product
documentation and websites:

-   Supabase --- official product and self-hosting documentation
-   Appwrite --- official documentation and current agent/MCP
    positioning
-   E2B --- official AI Agent Cloud and open-source documentation
-   Open WebUI --- official documentation and feature pages
-   Dify --- official project/product documentation
-   Langflow --- official product documentation and website

The competitive landscape changes quickly. Treat feature-level
comparison as a marketing snapshot and verify individual capabilities
before publishing a dated competitor comparison page.
