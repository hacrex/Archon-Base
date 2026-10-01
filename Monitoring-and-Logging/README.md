# Monitoring and Logging

## Required signals

- API request count, latency, status code, request ID, and authenticated project.
- PostgreSQL connection/transaction failures and migration version.
- Resource transitions by kind, phase, generation, and reason.
- Outbox backlog, publish attempts, oldest unpublished event, and last error.
- Qdrant readiness, reconciliation retries, collection creation failures, and query latency.
- Agent invocation count, duration, timeout, exit code, and model/vector dependency errors.

## Logging rules

Use structured JSON logs. Never log API keys, secret values, prompt contents, raw embeddings, signed URLs, or full request bodies by default. Support redaction and configurable retention. Keep audit events separate from diagnostic logs.

## Port security

Default local ports are loopback-only. In production, expose only the gateway; keep PostgreSQL, NATS, Qdrant, and internal runtimes on private networks with explicit firewall and NetworkPolicy rules. Alert on unexpected listeners and repeated authentication failures.
