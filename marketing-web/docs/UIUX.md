# Archon Base Marketing Website — UI/UX Specification

## Design Direction

**Style:** Minimal Neo-Brutalism  
**Personality:** Raw · Technical · Open · Bold · Controlled

Archon Base should look like infrastructure software that happens to have a marketing website.

Avoid:
- Generic AI gradients
- Glassmorphism
- Floating AI brains
- Excessive 3D
- Particle backgrounds
- Over-animated sections
- Glossy SaaS cards

Prefer:
- Off-white canvas
- Black typography
- Lime accent
- Hard borders
- Square corners
- Offset shadows
- Monospace metadata
- Technical diagrams
- Terminal UI
- Oversized editorial typography

## Color Tokens

```text
Background       #F4F1EA
Primary          #111111
Surface          #FFFFFF
Accent           #B7FF3C
Muted            #6B6B63
Code Background  #181818
Border           #111111
```

Use the lime accent sparingly for CTAs, active states, status indicators, architecture highlights and important interactions.

## Typography

Display:
- Space Grotesk
- Geist
- Satoshi
- Archivo

Technical:
- JetBrains Mono

Suggested scale:
- Hero: clamp(3.5rem, 9vw, 9rem)
- Section heading: clamp(2.5rem, 6vw, 6rem)
- Card heading: 1.75–3rem
- Body: 1rem–1.25rem
- Technical labels: 0.7–0.85rem

## Layout

- Max width: 1440px
- Desktop content width: 88–92%
- Mobile content width: calc(100% - 32px)
- Grid: 12 columns desktop
- Grid: 4 columns tablet/mobile where useful
- Border: 1–2px solid #111
- Border radius: 0 by default
- Card shadow: 6px 6px 0 #111 for selected interactive elements

## Header

```text
ARCHON BASE       Platform  Architecture  Docs  GitHub       [ GET STARTED → ]
```

Height: 72px  
Bottom border: 1px  
Sticky on desktop and mobile.

Mobile:
```text
ARCHON BASE                                      ☰
```

## Homepage Sections

### 01 Header

Minimal navigation and primary CTA.

### 02 Hero

Headline:

> THE OPEN-SOURCE INFRASTRUCTURE LAYER FOR AI.

Subheadline:

> Run agents, models, vector databases, tools and GPU workloads on infrastructure you control.

Actions:
- GET STARTED →
- VIEW GITHUB ↗

Right-side visual: infrastructure control-plane panel showing control plane, data plane, GPU utilization and health.

Technical annotations:
```text
OPEN SOURCE
KUBERNETES NATIVE
SELF HOSTED
AI INFRASTRUCTURE
```

### 03 Technology Strip

```text
BUILT AROUND OPEN TECHNOLOGY

KUBERNETES  QDRANT  VLLM  NATS
CILIUM      OPENBAO S3    OPENTELEMETRY
```

Monochrome logos.

### 04 Problem

Headline:

> AI INFRASTRUCTURE IS A STACK.

Show the infrastructure assembly:
```text
Kubernetes
+
GPU Operator
+
Model Server
+
Vector DB
+
Object Storage
+
Secrets
+
Networking
+
Observability
+
Sandboxing
+
AI Gateway
```

Closing statement:

> WHY ASSEMBLE THE SAME INFRASTRUCTURE EVERY TIME?

### 05 Control Plane

Headline:

> ALL OF THIS. ONE CONTROL PLANE.

Architecture diagram:
```text
                     ARCHON BASE
                          │
       ┌──────────────────┼──────────────────┐
       │                  │                  │
     AGENTS             MODELS              DATA
       │                  │                  │
    Runtime            Gateway           Vector DB
    Sandbox            Routing           Storage
    Tools              Inference         Knowledge
       │                  │                  │
       └──────────────────┼──────────────────┘
                          │
                     KUBERNETES
```

### 06 Core Capabilities

Use numbered editorial blocks.

01 / AGENTS  
02 / MODELS  
03 / DATA  
04 / TOOLS  
05 / SECURITY  
06 / OBSERVABILITY

Each block should contain:
- Large number
- Short description
- Technical capabilities
- Small metadata labels
- Minimal interaction

### 07 How It Works

```text
APPLICATION
     ↓
ARCHON BASE
     ↓
KUBERNETES DATA PLANE
     ↓
COMPUTE + DATA + NETWORK
```

### 08 Ecosystem

Headline:

> USE THE TOOLS YOU ALREADY USE.

Show:
```text
Open WebUI
      │
Dify ─┼─ Archon Base ─ vLLM
      │
Langflow
      │
Custom Apps
```

Message:

> Archon Base doesn't replace the AI ecosystem. It operates the infrastructure underneath it.

### 09 Developer Experience

Terminal panel:

```text
$ archon init my-ai-app

✓ Project created
✓ Kubernetes connection established
✓ Model gateway configured
✓ Qdrant provisioned
✓ Object storage configured
✓ Agent runtime ready

$ archon deploy agent.yaml
```

### 10 Infrastructure as Code

Show an `AgentFunction` YAML example.

Message:

> DECLARE IT. ARCHON OPERATES IT.

### 11 Security

Dark section.

Headline:

> SECURITY IS NOT AN ADD-ON.

Flow:
```text
USER
 ↓
IDENTITY
 ↓
POLICY
 ↓
BINDING
 ↓
CREDENTIAL
 ↓
RESOURCE
```

Labels:
IDENTITY · POLICY · SANDBOX · NETWORK · SECRETS · EGRESS · AUDIT

### 12 Deployment

Headline:

> FROM A SINGLE SERVER TO A PRIVATE AI CLOUD.

Progression:
```text
ONE SERVER → TEAM CLUSTER → GPU CLUSTER → PRIVATE CLOUD → AIR-GAPPED
```

### 13 Open Source

Emphasize:
- Self-hostable
- Open architecture
- Provider-neutral
- Kubernetes-native
- Adapter-first
- Portable

### 14 Final CTA

```text
BUILD YOUR AI CLOUD.
OWN YOUR INFRASTRUCTURE.

[ VIEW ON GITHUB ↗ ]
[ READ THE DOCS → ]
```

### 15 Footer

Minimal:
```text
ARCHON BASE

Open-source AI infrastructure.

Platform
Architecture
Docs
GitHub
Roadmap
Community

© 2026 Archon Base
```

## Components

Recommended reusable components:

- `SiteHeader`
- `MobileNav`
- `BrutalistButton`
- `TechnicalLabel`
- `SectionHeader`
- `StatusDot`
- `InfrastructurePanel`
- `TechnologyStrip`
- `StackList`
- `ArchitectureDiagram`
- `CapabilityCard`
- `EcosystemDiagram`
- `TerminalWindow`
- `CodeWindow`
- `SecurityFlow`
- `DeploymentRail`
- `FinalCTA`
- `SiteFooter`

## Interaction

Use restrained motion:
- 150–250ms transitions
- 2–4px hover translation
- CTA shadow movement
- Architecture node highlighting
- Terminal reveal
- Section fade/slide on entry

Do not animate everything.

## Accessibility

- WCAG AA contrast target
- Keyboard-visible focus states
- Semantic headings
- Proper button/link semantics
- Reduced-motion support
- No information conveyed only by color
- Horizontal scrolling for code/architecture on small screens

## Mobile

Do not merely shrink desktop.

Hero:
```text
ARCHON BASE

THE OPEN-SOURCE
INFRASTRUCTURE
LAYER FOR AI.

[ GET STARTED ]
[ GITHUB ]
```

Architecture diagrams become horizontally scrollable.

Code blocks remain horizontally scrollable.

Cards become full width.

## Brand Rule

The website should communicate:

> You choose the infrastructure. Archon Base gives you the control plane.
