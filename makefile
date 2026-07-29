# Load environment variables from .env
include .env
export

MIGRATIONS_PATH := ./cmd/migrate/migrations

.PHONY: test
test:
	@go test -v ./...

.PHONY: migrate-create
migrate-create:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_DSN)" up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_DSN)" down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-force
migrate-force:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_DSN)" force $(VERSION)

.PHONY: migrate-version
migrate-version:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_DSN)" version

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./cmd/api/main.go -d cmd,internal
	@swag fmt

.PHONY: up
up:
	docker compose up --build

.PHONY: up-back
up-back:
	docker compose up --build -d

.PHONY: down
down:
	docker compose down

.PHONY: logs
logs:
	docker compose logs -f

.PHONY: build
build:
	docker compose build

.PHONY: restart
restart:
	docker compose down && docker compose up --build

# Prevent make from treating migration names as targets
%:
	@:
