# Best Practices

- Bind local services to `127.0.0.1` by default; expose externally only through an authenticated gateway.
- Never commit real credentials. Development Compose credentials are disposable and must not be reused.
- Pin container images by version or digest and scan them in CI.
- Treat the local subprocess runner as trusted-only; use gVisor or microVMs before executing untrusted or generated code.
- Apply least privilege to database users, service accounts, API keys, filesystem mounts, and egress.
- Keep PostgreSQL, Qdrant, and NATS off public interfaces.
- Require backups and restore drills before production claims.
- Record request IDs, authentication decisions, resource mutations, controller transitions, and secret access without logging secret values.
