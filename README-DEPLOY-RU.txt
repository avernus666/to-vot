AVERNA — обновлённый пакет для Render

Что изменено:
- Go config читает PORT от Render и слушает :PORT; локальный HTTP_ADDR остаётся поддержан.
- Корневой Dockerfile генерирует go.sum командой go mod tidy, запускает go test и собирает бинарник.
- deploy/docker/api.Dockerfile обновлён для сборки из исходной структуры репозитория.

Render Web Service:
- Runtime: Docker
- Dockerfile Path: ./Dockerfile
- Docker Context: .
- Branch: main

Не добавляйте секреты в репозиторий. DATABASE_URL, TELEGRAM_BOT_TOKEN, TELEGRAM_WEBHOOK_SECRET и MINI_APP_URL задаются в Environment на Render.
