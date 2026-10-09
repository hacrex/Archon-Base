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

## Security boundary

The plaintext session token is returned only when a session is created. The database stores only the token hash. Password plaintext is never persisted.

This milestone does **not** yet expose login, logout, member invitations, or privileged API mutations. Those endpoints must be connected to the primitives only after request authentication and project/organization authorization middleware are implemented.

## Next integration steps

1. Add an authentication store for user lookup, credential verification, and session lifecycle.
2. Add `POST /v1/auth/login` and `POST /v1/auth/logout`.
3. Add authenticated principal middleware from `Authorization: Bearer` tokens.
4. Add organization membership lookup and project authorization checks.
5. Add authenticated Admin Panel member invitation and role management.
6. Add session listing, revocation, rate limits, and audit events.
7. Add CSRF protection if browser cookies are introduced.
