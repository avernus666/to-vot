# Operations handbook

## Local environment

Use Docker Compose for PostgreSQL and Redis. Do not use production credentials locally.
Keep `.env` out of version control.

## Database changes

- Create a new numbered migration; do not edit an already-applied migration.
- Prefer additive changes, backfill, then constraint tightening.
- Test upgrade and rollback against a disposable database.
- Take and verify backups before production migrations.

## Deployments

A production deployment needs a pinned container image, secret manager, TLS termination,
database migration job, health probes, resource limits, logs, metrics, alerting, and a rollback
plan. The sample Compose file is for local development, not a production deployment recipe.

## Incident checklist

1. Protect customer data and stop ongoing damage.
2. Identify affected workspaces and event time range.
3. Preserve relevant logs and audit records.
4. Communicate status clearly without exposing secrets.
5. Restore service from a known-good state.
6. Verify data consistency and replay eligible events safely.
7. Write a short post-incident review with corrective actions.
