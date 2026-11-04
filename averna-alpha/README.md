<div align="center">

# AVERNA ALPHA

### A Telegram storefront and business workspace.

**One Telegram bot. One Mini App. Two journeys: manage a business and shop from it.**

</div>

---

AVERNA combines a customer-facing storefront and a business workspace inside one Telegram Mini App. The ordinary Telegram bot is the entry point: `/start` sends a button that opens the app.

## Product experience

- **Shopping:** public catalog, cart, customer contact details, order creation.
- **Business management:** owner registration/login, organization workspace, product creation, recent orders and a basic sales summary.

This is an MVP foundation, not a complete payment SaaS. Payments, delivery, refunds, Telegram `initData` authentication, inventory, product editing/images and production security/compliance review are not yet implemented.

## Technology

- **Go** — API, Telegram webhook and business logic.
- **React + TypeScript + Vite** — Telegram Mini App.
- **PostgreSQL** — users, organizations, catalog, customers and orders.
- **Redis** — infrastructure foundation for background jobs.
- **Rust** — analytics service scaffold.
- **Docker Compose** — local infrastructure and app profile.

## Repository map

```text
.
├── apps/
│   ├── api/                 # Go API and Telegram webhook
│   ├── miniapp/             # React Telegram Mini App
│   └── analytics/           # Rust analytics scaffold
├── migrations/              # Versioned PostgreSQL migrations
├── deploy/docker/           # API container
├── docs/                    # Architecture, operations, roadmap
├── api/openapi.yaml         # API contract starter
├── docker-compose.yml
├── Makefile
└── SECURITY.md
```

## Run locally

Requirements: Docker Compose, Go 1.23+, Node.js 22+, npm and Make (optional).

```bash
cp .env.example .env
docker compose up -d postgres redis
make migrate-up
make api
```

In a second terminal:

```bash
cd apps/miniapp
cp .env.example .env
# Set VITE_API_BASE_URL=http://localhost:8080 in apps/miniapp/.env
npm install
npm run dev
```

The Mini App dev server usually listens on `http://localhost:5173`; the API listens on `http://localhost:8080`.

### Docker app profile

After configuring a public HTTPS Mini App URL and Telegram credentials in `.env`:

```bash
docker compose up -d postgres redis
docker compose --profile tools run --rm migrate
docker compose --profile app up --build -d
```

The Mini App is exposed on port `8088`. Nginx serves the frontend and proxies `/v1/`, `/telegram/` and `/health/` to the Go API.

## Connect a Telegram bot

1. Create a bot with the official `@BotFather` and copy its token.
2. Set these values in `.env`:

```env
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_WEBHOOK_SECRET=long_random_secret_with_letters_numbers_dash_or_underscore
MINI_APP_URL=https://your-domain.example/?store=your-store-slug
```

3. Deploy the app on a public HTTPS domain.
4. Register the webhook:

```bash
set -a; . ./.env; set +a

curl -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook" \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"https://your-domain.example/telegram/webhook\",\"secret_token\":\"$TELEGRAM_WEBHOOK_SECRET\",\"allowed_updates\":[\"message\"]}"
```

5. Open the bot and send `/start`. To open a particular shop, share `https://t.me/YOUR_BOT?start=your-store-slug`; the bot passes the validated slug to the Mini App.

Telegram requires HTTPS for Mini Apps. Never commit bot tokens, passwords, webhook secrets or `.env` files.

## API surface

| Method | Endpoint | Purpose |
|---|---|---|
| POST | `/v1/auth/register` | Create owner account and organization |
| POST | `/v1/auth/login` | Owner sign-in |
| GET | `/v1/organizations` | List organizations for current user |
| GET | `/v1/store/{slug}/products` | Public catalog |
| POST | `/v1/store/{slug}/orders` | Create an order |
| GET / POST | `/v1/organizations/{id}/products` | List or create products |
| GET | `/v1/organizations/{id}/orders` | List recent orders |
| POST | `/telegram/webhook` | Telegram update webhook |

The API validates product availability and calculates order totals server-side. Prices are stored in minor currency units. Organization endpoints enforce membership and role checks.

## Checks

```bash
cd apps/api && go test ./... && go vet ./...
cd apps/miniapp && npm install && npm run build
```

## Before real customers or payments

- Validate Telegram Mini App `initData` before using Telegram identity as authentication.
- Add product editing/archive, product images, inventory and order status management.
- Add customer/owner notifications and payment-provider integration.
- Configure shipping, taxes, receipts and refund rules for the market where the shop operates.
- Add rate limits, abuse controls, monitoring, backup/restore tests and a security review.
- Publish privacy information and implement retention/deletion workflows for customer data.

Do not collect real customer data or accept real payments until those controls and local legal requirements are reviewed. Pricing and product-market fit also need validation with real businesses; code alone cannot guarantee sales.
