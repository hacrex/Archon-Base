# Archon Base Marketing Website

Minimal Neo-Brutalist marketing website scaffold for **Archon Base**, positioned as:

> **The open-source infrastructure layer for AI.**

## Positioning

Archon Base is an open-source AI infrastructure control plane for self-hosted agents, models, data, tools, and compute.

The site should position Archon Base as infrastructure underneath AI applications such as Open WebUI, Dify, Langflow and custom applications rather than as another AI chat UI or workflow builder.

## Design

- Minimal
- Neo-brutalist
- Developer-first
- Infrastructure-oriented
- Off-white + black + lime
- Square corners
- Hard borders
- Offset shadows
- Large editorial typography
- Technical monospace labels
- Restrained animation

## Structure

```text
archon-base-marketing-website/
├── README.md
├── docs/
│   ├── UIUX.md
│   └── MARKETING-CONTENT.md
├── public/
├── src/
│   ├── components/
│   ├── sections/
│   ├── styles/
│   └── data/
└── package.json
```

## Page Map

```text
/
├── Home
├── /platform
├── /architecture
├── /integrations
├── /compare
├── /developers
├── /deploy
└── /community
```

## Homepage Flow

```text
Header
  ↓
Hero
  ↓
Technology Strip
  ↓
Problem
  ↓
Control Plane
  ↓
Capabilities
  ↓
How It Works
  ↓
Ecosystem
  ↓
Developer Experience
  ↓
IaC
  ↓
Security
  ↓
Deployment
  ↓
Open Source
  ↓
Final CTA
  ↓
Footer
```

## Suggested Stack

If implementation has not already been chosen:

- Next.js
- TypeScript
- Tailwind CSS
- Framer Motion or Motion
- Lucide icons
- MDX for documentation/content
- Geist or Space Grotesk
- JetBrains Mono

Avoid unnecessary UI libraries if they fight the visual language.

## Important Product Rule

Do not present roadmap features as shipped features.

Use:
- `Available`
- `Preview`
- `Planned`

when appropriate.

## Competitive Positioning

The website should communicate:

```text
AI APPLICATIONS
Open WebUI / Dify / Langflow / Custom Apps
                ↓
         ARCHON BASE
                ↓
Kubernetes / GPU / Storage / Models / Data
```

Message:

> **Archon Base doesn't replace the AI ecosystem. It operates the infrastructure underneath it.**

## Naming Note

The name "Archon" is used by multiple current projects and products, so the marketing site should consistently use the full brand **Archon Base** and establish its domain/GitHub identity clearly. This is a branding/discoverability consideration, not a product-positioning change.
