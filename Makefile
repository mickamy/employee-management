-include .env
export DATABASE_URL

BUILD_DIR = bin
GOOSE = go tool -modfile=tools/go.mod goose -dir internal/storage/db/migrations
DB_USER = app
DB_NAME = employee_management

.PHONY: build clean test lint compose-up compose-up-d compose-down db-migrate db-create db-drop db-reset new-migration

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/server ./cmd/server

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

test:
	go test ./... -race

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint is not installed"; \
		exit 1; \
	}
	golangci-lint run

compose-up:
	docker compose up --wait

compose-up-d:
	docker compose up -d --wait

compose-down:
	docker compose down

db-migrate:
	$(GOOSE) postgres "$(DATABASE_URL)" up

db-create:
	docker compose exec db createdb -U $(DB_USER) $(DB_NAME)

db-drop:
	docker compose exec db dropdb -U $(DB_USER) --force --if-exists $(DB_NAME)

db-reset: db-drop db-create db-migrate

new-migration:
	$(GOOSE) create $(name) sql
