# Identity and sessions (v0.3)

## Implemented

- Email normalization and basic input validation.
- Argon2id password hashes with random per-password salts.
- Opaque random bearer tokens; only SHA-256 token digests are stored.
- 15-minute access-token lifetime.
- 30-day refresh-token lifetime.
- Refresh-token rotation in a database transaction.
- Session revocation on logout.
- Registration transaction creates user, organization, owner membership, and session.
- Authenticated `/v1/me` and `/v1/organizations` endpoints.

## Important limitations before public launch

- Add IP/account-based rate limits and bot protection to registration/login/refresh.
- Add email verification and secure account recovery.
- Add refresh-token reuse detection and session/device management.
- Add password reset flow with one-time, hashed, expiring tokens.
- Add integration tests using a disposable PostgreSQL database.
- Add database RLS and tests for cross-tenant access.
- Review Argon2id parameters against deployment memory/latency budgets.
- Consider a managed identity provider if operating authentication becomes a distraction.

Access tokens are opaque and database-backed in this foundation. This enables immediate
revocation but adds a database lookup to authenticated requests. A later signed-token design
would need a documented revocation strategy and key rotation plan.
