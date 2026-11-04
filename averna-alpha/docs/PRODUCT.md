# Product brief

## Product definition

AVERNA is a Telegram-first commerce and business-management platform. A single Telegram bot opens one Mini App with two journeys:

- **Buyer:** browse a shop catalog, manage a cart, submit contact details and place an order.
- **Business owner:** create a workspace, add products, review recent orders and see a basic overview.

## Problem hypothesis

Small shops often combine Telegram conversations, social media posts, spreadsheets and manual order tracking. AVERNA tests whether a compact storefront plus a simple owner workspace can reduce missed orders and repetitive administration.

## Initial customer hypothesis

Start with small online shops that already sell through Telegram or social channels. Avoid trying to serve every type of business in the first pilot.

## Product principles

- The buyer should complete a simple order without installing another app.
- The owner should publish a first product and receive a first order quickly.
- The server, not the browser, calculates prices and validates product availability.
- Every organization must be isolated from every other organization.
- Customer contact data must be protected, retained only as needed and deletable/exportable.
- Payments, shipping and taxes must be explicit and country-aware; never imply an order is paid when no payment provider confirmed it.

## Current MVP boundary

Implemented: Telegram `/start` webhook, Mini App launch, public catalog, cart, guest checkout/order creation, owner registration/login, product creation, recent-order list.

Not implemented: Telegram `initData` login, product editing/images, inventory, order status controls, owner/buyer notifications, payment processing, delivery integrations, full CRM and production compliance/security review.
