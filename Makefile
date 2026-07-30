include .env
export ENV MODULE_ROOT DATABASE_URL DATABASE_WRITER_URL DATABASE_READER_URL

BUILD_DIR = bin
GOOSE = go tool -modfile=tools/go.mod goose -dir internal/infra/storage/db/migrate/sql
DB_USER = app
DB_NAME = employee_management

.PHONY: build \
		clean \
		test \
		lint \
		lint-fix \
		compose-up \
		compose-up-d \
		compose-down \
		db-create \
		db-drop \
		db-migrate \
		db-reset \
		gen \
		gen-buf \
		gen-go \
		gen-injector \
		gen-sqlc \
		new-migration \

.env:
	envsubst < .env.example > $@

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/server ./cmd/server

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./... -race

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint is not installed"; \
		exit 1; \
	}
	golangci-lint run

lint-fix:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint is not installed"; \
		exit 1; \
	}
	golangci-lint run --fix

compose-up:
	docker compose up --build

compose-up-d:
	docker compose up -d --build --wait

compose-down:
	docker compose down

db-create:
	docker compose exec -T db psql -U $(DB_USER) -d postgres -tAc \
		"SELECT 1 FROM pg_database WHERE datname = '$(DB_NAME)'" | grep -q 1 || \
		docker compose exec -T db createdb -U $(DB_USER) $(DB_NAME)

db-drop:
	docker compose exec -T db dropdb -U $(DB_USER) --force --if-exists $(DB_NAME)

# Applies migrations to whatever DATABASE_URL points at — this is the
# deployment path. Destructive targets (db-create / db-drop) stay local-only
# via docker compose exec.
db-migrate:
	$(GOOSE) postgres "$(DATABASE_URL)" up

db-reset: db-drop db-create db-migrate

gen: gen-buf gen-go gen-injector gen-sqlc

gen-buf:
	@command -v buf >/dev/null 2>&1 || { \
		echo "buf is not installed"; \
		exit 1; \
	}
	buf generate

gen-go:
	go generate ./...

gen-injector:
	go tool -modfile=tools/go.mod injector ./...

gen-sqlc:
	go tool -modfile=tools/go.mod sqlc generate

frontend-dev:
	cd frontend && npm run dev

frontend-typecheck:
	cd frontend && npm run typecheck

new-migration:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make new-migration name=migration_name"; \
		exit 1; \
	fi
	$(GOOSE) create $(name) sql
