# Security policy

## Reporting

Do not publish exploitable vulnerabilities in public issues. Until a private security
contact is configured, report findings privately to the repository maintainer.

## Required controls before production

- Argon2id password hashing with reviewed parameters and per-password salts.
- Short-lived access tokens and hashed, rotated, revocable refresh tokens.
- Server-side authorization on every tenant-scoped operation.
- PostgreSQL RLS as defense in depth, with transaction-local tenant context.
- Encryption at rest for Telegram bot credentials and integration secrets.
- Webhook secret validation, payload size limits, replay/deduplication protection.
- Rate limiting and abuse controls for public endpoints.
- Strict CORS policy, secure cookies where cookies are used, and CSRF protection.
- Secret management outside source control and container images.
- Dependency scanning, pinned/reviewed upgrades, and SBOM generation.
- Encrypted backups and periodically tested restoration.
- Audit events for sensitive administrative operations.
- Data export/deletion and retention workflows designed before customer onboarding.

## Never

- Commit credentials, private keys, tokens, or production `.env` files.
- Log passwords, authorization headers, full payment data, or bot tokens.
- Trust `organization_id` supplied by a client without checking membership.
- Treat Redis as the durable source of truth.
- Claim GDPR compliance without a documented legal and technical review.
