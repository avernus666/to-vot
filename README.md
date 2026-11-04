# AVERNA ALPHA — Render deployment

This repository package contains the Go API and React Mini App source archive.
Render services:
- API: Docker Web Service using the root Dockerfile.
- Mini App: Static Site; build command `unzip -q averna-source.zip -d . && cd averna-source/apps/miniapp && npm install --no-audit --no-fund && VITE_API_BASE_URL=https://<API-SERVICE>.onrender.com npm run build`; publish path `averna-source/apps/miniapp/dist`.

Required API environment:
- `APP_ENV=production`
- `DATABASE_URL` (Neon pooled connection string, SSL required)
- `TELEGRAM_BOT_TOKEN` (add privately in Render)
- `TELEGRAM_WEBHOOK_SECRET` (random secret, add privately in Render)
- `MINI_APP_URL` (Static Site URL)
- `LOG_LEVEL=info`

After the API is deployed, run migrations once against Neon using the SQL files in `averna-source/migrations` in ascending order. The API does not automatically apply migrations.
