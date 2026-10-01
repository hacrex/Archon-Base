# Incident Response

## Suspected credential or secret exposure

1. Disable the affected key or binding.
2. Preserve request IDs, audit records, and relevant timestamps.
3. Rotate the credential at its source; do not paste it into an issue.
4. Inspect access logs for replay or cross-project use.
5. Record the incident and add a regression test.

## Suspected cross-project data access

1. Stop the affected agent or API route.
2. Preserve database audit logs and resource versions.
3. Revoke project credentials and disable bindings.
4. Determine affected projects and data before restoring access.
5. Notify stakeholders according to the deployment's incident policy.

## Ransomware, malware, or destructive operator action

1. Isolate affected hosts and stop automation that can spread the change.
2. Preserve immutable logs and snapshots.
3. Do not delete evidence or blindly restore over the original data.
4. Rebuild from trusted images and restore only after validating backups.
5. Rotate credentials and review supply-chain and access paths.
