.PHONY: api test lint fmt tidy migrate-up migrate-down infra-up infra-down analytics miniapp-dev miniapp-build app-up

api:
	set -a; [ ! -f .env ] || . ./.env; set +a; cd apps/api && go run ./cmd/api

test:
	cd apps/api && go test ./...

lint:
	cd apps/api && go vet ./...
	cd apps/analytics && cargo check

fmt:
	cd apps/api && gofmt -w .
	cd apps/analytics && cargo fmt

tidy:
	cd apps/api && go mod tidy

infra-up:
	docker compose up -d postgres redis

infra-down:
	docker compose down

migrate-up:
	docker compose --profile tools run --rm migrate

migrate-down:
	docker compose --profile tools run --rm migrate -path=/migrations -database=postgres://averna:local-dev-only@postgres:5432/averna?sslmode=disable down 1

analytics:
	cd apps/analytics && cargo run

miniapp-dev:
	cd apps/miniapp && npm install && npm run dev

miniapp-build:
	cd apps/miniapp && npm install && npm run build

app-up:
	docker compose --profile app up --build -d
