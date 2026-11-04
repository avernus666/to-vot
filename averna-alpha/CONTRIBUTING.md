# Contributing

## Development workflow

1. Create a focused branch.
2. Keep changes small and explain the business reason.
3. Add or update tests.
4. Update API/event documentation when contracts change.
5. Run `make fmt`, `make test`, and `make lint`.
6. Never commit secrets or real customer data.

## Code standards

- Prefer explicit dependencies and small interfaces at boundaries.
- Keep HTTP handlers thin; business rules belong in application/domain packages.
- Validate input at transport boundaries and enforce invariants in domain logic.
- Return typed/sentinel errors from application code; map them to HTTP in transport.
- Use context-aware I/O and bounded timeouts.
- Avoid global mutable state.
- Do not add a new service or dependency without a clear reason.
- Tests must not depend on execution order or external production services.

## Pull request checklist

- [ ] Scope is focused and documented.
- [ ] Tests cover normal and failure paths.
- [ ] Tenant isolation has been considered.
- [ ] Logs contain no secrets or sensitive payloads.
- [ ] Migration is reversible where practical.
- [ ] API contract and README are current.
