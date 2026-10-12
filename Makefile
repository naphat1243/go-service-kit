DATABASE_URL ?= postgres://app:app@localhost:5432/app?sslmode=disable
MIGRATE := docker run --rm -v $(PWD)/internal/repository/migrations:/migrations --network host migrate/migrate -path=/migrations -database "$(DATABASE_URL)"

.PHONY: dev run build test lint migrate-up migrate-down compose-up compose-down

## dev: start postgres, run migrations, then run the API locally
dev: compose-up migrate-up run

run:
	go run ./cmd/api

build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api

test:
	go test ./... -count=1

lint:
	go vet ./...
	gofmt -l . | tee /dev/stderr | wc -l | grep -q '^ *0$$'

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

compose-up:
	docker compose up -d db
	@until docker compose exec db pg_isready -U app >/dev/null 2>&1; do sleep 0.5; done

compose-down:
	docker compose down -v
