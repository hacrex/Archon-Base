# Control Plane UI Phase 1 Plan

## Design direction

- **Movement:** operational console / quiet technical utility.
- **Principles:** infrastructure-first, evidence over decoration, explicit capability status, safe defaults.
- **Color philosophy:** deep blue-black for the control boundary, pale green for healthy/owned state, amber for capacity and dependency attention, neutral surfaces for data density.
- **Layout:** persistent operator sidebar with a wide evidence canvas; dashboard panels are subordinate to resource tables and operational state.
- **Signature elements:** diagonal Archon mark, compact status chips, and split control-plane/data-plane health cards.
- **Interaction:** direct navigation, clear context, and modal explanations for unavailable or planned actions; no fake submissions.
- **Animation:** only small state transitions; honor reduced-motion preferences.
- **Typography:** system sans for dense operational scanning, uppercase micro-labels for hierarchy, compact numeric metrics.
- **Brand essence:** infrastructure you control for AI workloads; precise, self-hosted, infrastructure-first.

## Structure

- `index.html`: accessible application shell, sidebar, top bar, notice, and app mount.
- `styles.css`: control-plane design tokens, responsive layout, tables, cards, status states, and modal styling.
- `app.js`: hash routing, local preview state, route renderers, context interactions, and planned-feature explanations.
- `manus-routes.json`: declared page route set.
- `README.md`: local run instructions and product boundary.

## Phase 1 scope

The first slice establishes the application shell and routes for Overview, Projects, Resources, Operations, Admin, Audit, and Settings. It presents separate control-plane and data-plane health, the student/K3s capacity context, resource lifecycle status, and explicit preview/planned boundaries. It intentionally does not add authentication, API mutations, or privileged admin behavior until those contracts exist.
