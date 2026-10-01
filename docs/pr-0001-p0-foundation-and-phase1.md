# PR 0001: Establish v1 resource foundations and local Phase 1 contract

## Summary

This change converts the repository from an unversioned design scaffold into a reviewable foundation for the first local vertical slice. It fixes the placeholder Go module identity, makes formatting reproducible, selects Apache-2.0, defines the Project and DatabaseInstance v1 contract, adds PostgreSQL tables and a transactional outbox, and documents the Kubernetes-free execution path.

## Changes

- Replace `github.com/your-org/archon-base` with the repository module path.
- Add `make fmt-check` and make `make fmt` operate on explicit Go files.
- Add `LICENSE` with Apache-2.0 terms and update README licensing.
- Add `docs/core-v1-resource-schema.md`.
- Add `migrations/001_core_resources.sql` and its reversible down migration.
- Add `docs/phase-1-local-execution-spec.md`.
- Correct example agent entrypoints from `app.main:handle` to `app:handle`.
- Add baseline security, threat-modeling, monitoring, and incident-response guidance.

## Design decisions

1. PostgreSQL is the source of truth for desired resource state.
2. `spec` is desired state; `status` is controller-owned observed state.
3. Project and DatabaseInstance names are unique within their parent scope while active, allowing soft deletion without losing historical rows.
4. Resource mutations and outbox events are committed in one transaction.
5. Phase 1 supports one local Qdrant instance and a trusted-only Python subprocess runner. It does not claim sandbox isolation.
6. The migration stores typed resource fields plus JSONB snapshots so the API can preserve the submitted manifest while strict validation remains at the boundary.

## Migration procedure

Apply the up migration with the project's migration runner or directly in a development database:

```bash
psql "$ARCHON_DB_URL" -v ON_ERROR_STOP=1 -f migrations/001_core_resources.sql
```

To remove the development schema only:

```bash
psql "$ARCHON_DB_URL" -v ON_ERROR_STOP=1 -f migrations/001_core_resources_down.sql
```

The down migration is destructive and must not be used against a database containing tenant data.

## Validation performed

- Checked the patch with `git diff --check`.
- Confirmed no tests or Go toolchain are currently available in the analysis sandbox; CI should run `go test ./...`, `go vet ./...`, and `make fmt-check` once Go is installed.
- Reviewed SQL constraints for project scoping, positive replica/shard counts, allowed lifecycle phases, JSON object fields, foreign-key protection, and outbox indexing.

## Follow-up PRs

- Add Go typed API models matching the resource document and strict validation.
- Add PostgreSQL migration runner and schema version table.
- Add authentication, project authorization, request IDs, timeouts, readiness checks, and graceful shutdown.
- Implement a Qdrant adapter and local reconciler.
- Add idempotency-key handling and outbox publisher retries.
- Add CI, SQL integration tests, manifest conformance tests, and a local `archon dev` command.
