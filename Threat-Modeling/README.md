# Threat Modeling

## Phase 1 local topology

| Asset | Threat | Required mitigation |
|---|---|---|
| PostgreSQL metadata | Local credential theft or destructive access | Loopback binding, disposable credentials, least-privilege role, backups |
| Qdrant data | Unauthorized reads or deletion | Loopback binding, project-scoped API mediation, no public port |
| Agent subprocess | Malicious or generated code escaping | Trusted-code-only warning; no production use; future gVisor/microVM boundary |
| API | Unauthorized resource mutation | Authentication and project authorization before Phase 1 production use |
| Outbox events | Lost reconciliation work | Same-transaction write, retryable publisher, observable failure state |

## Future production threats

Model cross-tenant reads, confused-deputy bindings, SSRF/egress abuse, secret exfiltration, sandbox escape, malicious images, supply-chain compromise, replayed credentials, denial of service, and destructive operator actions. Each threat needs an owner, test, monitoring signal, and recovery procedure before production release.
