# Product and engineering roadmap

Roadmap items are targets, not delivery guarantees. Completed MVP work is separated from production release gates.

## Phase 0 — Validate the problem

- Interview 15–20 owners in one candidate segment.
- Map how they currently publish products, receive orders and contact buyers.
- Recruit 3–5 design partners and measure whether AVERNA reduces manual work.

## Phase 1 — Secure foundation

- [x] Go API, PostgreSQL schema, Docker Compose and health endpoints
- [x] Owner registration/login and rotating token sessions
- [x] Organization membership and role checks
- [ ] Database row-level security and tenant-isolation integration tests
- [ ] Rate limiting, email verification, password recovery and audit review

## Phase 2 — Telegram commerce MVP

- [x] Telegram webhook secret validation
- [x] `/start` response with Mini App launch button
- [x] React + TypeScript Mini App shell
- [x] Public product catalog and shopping cart
- [x] Server-side order price calculation and transactional order creation
- [x] Business workspace with product creation and recent orders
- [ ] Telegram `initData` verification and secure account linking
- [ ] Bot notifications to owners and buyers
- [ ] Product editing, archiving, images and inventory
- [ ] Order status transitions and customer history

## Phase 3 — Safe commercial pilot

- [ ] Pilot with a small number of real shops after privacy/security review
- [ ] Payment-provider integration, refunds and reconciliation
- [ ] Shipping, tax, receipts and country-specific checkout rules
- [ ] Customer data export, retention and deletion workflows
- [ ] Monitoring, alerting, backups and restore drills
- [ ] Terms, privacy notice and support process

## Phase 4 — Expansion

- [ ] CRM notes and activity timeline
- [ ] Staff invitations and granular permissions
- [ ] Automation triggers and scheduled jobs
- [ ] Rust analytics ingestion and reporting
- [ ] Integrations and additional messaging channels
- [ ] Subscription billing and usage limits

## Release gates

Do not call AVERNA production-ready until authorization and tenant isolation, webhook retries, migrations/rollbacks, backups/restores, privacy controls, payment state transitions, monitoring and expected-load failure tests have been verified.

## Business metrics

Track activation, time-to-first-value, catalog-to-order conversion, order completion, weekly active workspaces, retention, support burden, gross margin and acquisition cost. A feature is not successful merely because it shipped.


## AVERNA ALPHA 0.5.0-alpha.1

Completed in this alpha: product publishing controls and guarded order status management in the owner workspace.

Next priorities: Telegram `initData` verification, reliable owner order notifications via an outbox/worker, product inventory, image uploads, and security review before any live commerce use.
