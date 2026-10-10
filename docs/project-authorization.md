# Project-scoped authorization

## Design decision

Archon Base has two authorization scopes:

1. **Organization scope** — the authenticated session carries the user's active organization and organization role.
2. **Project scope** — access to a project is granted by `project_memberships`, unless the organization role is `owner` or `admin`.

Organization `operator`, `developer`, and `viewer` roles do **not** implicitly receive access to every project. They need an explicit project membership. This keeps a future multi-project organization from accidentally exposing resources across projects.

## PostgreSQL schema

Migration `004_project_authorization.sql` creates:

```text
project_memberships
├── project_id  → projects.id, ON DELETE CASCADE
├── user_id     → users.id, ON DELETE CASCADE
├── role        ∈ owner | admin | operator | developer | viewer
├── created_at
└── updated_at

PRIMARY KEY (project_id, user_id)
```

The `(user_id, project_id)` index supports listing a user's accessible projects. The `(project_id, role)` index supports project administration and future member-management screens. Membership writes are idempotent with `ON CONFLICT` updates.

The migration assumes `projects`, `users`, `organization_memberships`, and `set_updated_at()` were created by migrations 001 and 002. The down script removes only the project-scoped table and trigger.

## Effective access rules

| Request | Minimum effective role |
|---|---:|
| List accessible projects | viewer |
| Read a project or database instance | viewer |
| Create a project | organization developer |
| Create or update a database instance | project developer |
| Update project metadata | project developer |
| Delete a database instance | project operator |
| Delete a project | project admin |

`owner` and `admin` organization memberships satisfy any project requirement. All other users are evaluated against their explicit project membership. The creator's effective organization role is inserted as an explicit membership immediately after a successful project creation.

## Middleware flow

```text
request
  │
  ├─ request ID / CORS
  ├─ bearer authentication
  │    └─ session → Principal{user, organization, org role}
  ├─ project authorization
  │    ├─ map URL + HTTP method to minimum role
  │    ├─ query project + organization membership + project membership
  │    └─ continue or return 403
  └─ resource handler
```

Authorization failures return the stable `forbidden` error code. If project authorization has not been configured, the API returns `503 authorization_unavailable` rather than silently falling back to organization-wide access.

Project collection `GET /v1/projects` uses a principal-scoped query and never returns projects outside the caller's organization or explicit project grants. `POST /v1/projects` is organization-scoped and requires developer access before the project is created.

## PostgreSQL configuration

For local development:

```bash
cp .env.example .env
# edit POSTGRES_PASSWORD for any shared environment
docker compose --env-file .env -f deploy/docker-compose.yml up --build
```

The API reads:

- `ARCHON_DB_URL` — pgx connection URL
- `ARCHON_MIGRATIONS_DIR` — migration directory, default `migrations`
- `ARCHON_ORGANIZATION_ID` — organization used by the local single-organization repository

Startup creates `schema_migrations`, applies forward migrations in filename order, and stops before serving if PostgreSQL or a migration is unavailable. Compose now waits for the PostgreSQL health check before starting the API.

The checked-in default credentials are for disposable local development only. Use environment substitution or a secret manager for shared or production deployments.
