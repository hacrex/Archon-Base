# Archon Base Control Plane UI

This is the first local UI slice for the Archon Base product control plane. It is a dependency-free, route-aware dashboard preview that establishes the application shell and information architecture before authentication and API wiring.

## Run locally

From this directory:

```bash
python3 -m http.server 4174 --bind 0.0.0.0
```

Open `http://127.0.0.1:4174/#/overview`.

## Routes

- `#/overview` — organization and project dashboard
- `#/projects` — project inventory
- `#/resources` — resource inventory and lifecycle status
- `#/resources/allocation` — CPU, memory, storage, and GPU request/limit allocation view
- `#/operations` — reconciliation and operation history
- `#/admin` — admin panel information architecture
- `#/audit` — audit event boundary
- `#/settings` — organization/project settings

## Frontend structure

The browser entrypoint is `app.js`, which bootstraps the modules under `src/`:

```text
src/state.js       local preview state
src/components.js  shared headings, statuses, meters, and modals
src/routes.js      page renderers and route-specific views
src/app.js         hash routing and interaction handlers
```

## Product boundary

The preview uses clearly labeled local preview data. It does not submit mutations, authenticate users, or imply that planned agent, model, GPU, MCP, or cluster capabilities are shipped. The Projects view now shows project capacity and quota context, while Resource Allocation provides a read-only request/limit management surface. The next implementation slice should connect the existing Project and DatabaseInstance API to these routes, then add authentication and authorization before enabling admin mutations.

The structure follows the comparison notes: Archon Base is the infrastructure control plane below application builders such as Supabase-backed applications, Appwrite applications, Open WebUI, Dify, and Langflow. It is not a replacement chat UI or general application backend.
