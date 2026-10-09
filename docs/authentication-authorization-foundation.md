# Authentication and Authorization Foundation

Archon Base now has the first security primitives for user and organization management.

## Implemented

- Adaptive bcrypt password hashing with a 12-character minimum policy.
- Opaque 32-byte session tokens.
- SHA-256 token hashes for database persistence.
- Session expiration and revocation model.
- Organization roles:
  - `owner`
  - `admin`
  - `operator`
  - `developer`
  - `viewer`
- Role hierarchy checks for future authorization middleware.
- PostgreSQL credential and session tables through migration `003_auth_sessions.sql`.
- PostgreSQL authentication repository methods for credential verification, session lookup, and revocation.
- `POST /v1/auth/login` and `POST /v1/auth/logout`.
- `GET /v1/auth/me` for the authenticated principal and organization membership.
- Bearer-token middleware protecting project and database resource routes.

## Security boundary

The plaintext session token is returned only when a session is created. The database stores only the token hash. Password plaintext is never persisted.

Login, logout, current-principal lookup, and bearer authentication are now implemented. Member invitations, role mutations, and full project-scoped authorization are still intentionally deferred until their audit and policy contracts are complete.

## Next integration steps

1. Add project authorization checks using the authenticated organization membership.
2. Add authenticated Admin Panel member invitation and role management.
3. Add session listing, revocation UI, rate limits, and audit events.
4. Add CSRF protection if browser cookies are introduced.
