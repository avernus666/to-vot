# Architecture

## High-level view

```text
Telegram user
    |
    +---- ordinary Telegram bot (/start)
    |             |
    |             +---- Telegram Mini App (React + TypeScript)
    |                              |
    +---------------------------- HTTPS
                                   |
                         Go API (modular monolith)
     +----------+-----------+------------+
     | Identity | CRM       | Commerce   |
     | Telegram | Booking   | Support    |
     | Billing  | Automation| Integrations|
     +----------+-----------+------------+
                |
        PostgreSQL transaction
          +-------------+
          | domain write|
          | outbox row  |
          +-------------+
                |
          Outbox publisher
                |
           Redis Streams
                |
       Rust analytics consumer
                |
      analytics aggregates/read models
```

## Why a modular monolith

The product needs consistent authorization and transactions across customers, orders, bots,
and subscriptions. A modular Go application keeps these workflows simple while allowing a
separate Rust analytics process where CPU-heavy aggregation or independent scaling is useful.
Service boundaries should follow operational needs, not fashion.

## Layers

- **Transport:** HTTP parsing, validation, status codes, middleware.
- **Application:** use cases, orchestration, transaction boundaries.
- **Domain:** business invariants and domain types, independent of HTTP/SQL.
- **Infrastructure:** PostgreSQL, Redis, Telegram API, external providers.

Transport must not contain business policy. Domain code must not import infrastructure.

## Tenant isolation

Every tenant-owned record must include `organization_id`. Repository methods should require
an explicit tenant context. The authenticated user must be a member of the selected
organization and have the required role. Composite foreign keys prevent references across
organizations. PostgreSQL RLS should later add defense in depth using `SET LOCAL` within each
transaction. RLS policies must be tested with both permitted and forbidden tenant contexts.

## Event reliability

Business changes and outbox records are committed in the same PostgreSQL transaction.
A publisher claims pending rows using `FOR UPDATE SKIP LOCKED`, publishes a stable event ID,
and marks the row published after Redis acknowledges it. Retries use bounded backoff and
dead-letter visibility. Consumers deduplicate by event ID. This is at-least-once delivery.

## Money and currencies

Amounts are stored as integer minor units (`price_minor`) with an explicit ISO currency code.
Never add amounts from different currencies. Currency exponent rules must be centralized and
tested before supporting currencies with non-two-decimal minor units.

## Security boundaries

- Access tokens should be short-lived.
- Refresh tokens should be stored only as hashes and rotated.
- Telegram and integration secrets should be encrypted with a managed key.
- Webhooks must validate provider signatures/secrets and deduplicate provider event IDs.
- Audit records should cover permission changes, exports, billing actions, and secret changes.
- Public endpoints need request limits, timeouts, rate limits, and abuse monitoring.


## Current MVP boundary

The current implementation includes the Telegram `/start` webhook and Mini App launch button,
public catalog reads, guest order creation, owner registration/login, product creation, and
recent order listing. PostgreSQL stores the commerce records transactionally. Redis Streams,
outbox publishing, Rust consumer integration, Telegram `initData` authentication, payment
processing, inventory, and delivery are architectural targets rather than completed workflows.
