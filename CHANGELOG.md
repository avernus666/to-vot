# Changelog

## 0.5.0-alpha.1 — AVERNA ALPHA

- Rebranded the project as AVERNA ALPHA and versioned the Mini App package as `0.5.0-alpha.1`.
- Added an organization-scoped product update endpoint and publishing controls (publish/hide) through `PATCH /v1/organizations/{organizationID}/products/{productID}`.
- Added order status management through `PATCH /v1/organizations/{organizationID}/orders/{orderID}`.
- Added guarded order transitions: pending → confirmed/cancelled; confirmed → completed/cancelled. Terminal states cannot be reopened.
- Added management UI actions for hiding/showing products and accepting, completing, or cancelling orders.
- Added unit tests for order status transition rules.

### Verification and limitations

- This is an alpha build. Telegram `initData` verification, bot-to-owner order notifications, online payments, inventory, image uploads, and production hardening are not included.
- Do not use real customer data or accept payments until authentication, security, privacy, and payment flows are independently reviewed and tested.
