# Load environment variables from .env
include .env
export

MIGRATIONS_PATH := ./cmd/migrate/migrations


.PHONY: help
help:
	@echo ""
	@echo "Available Commands:"
	@echo ""
	@echo "  make test                      Run all Go tests"
	@echo "  make format                    Format Go source files using gofmt"
	@echo "  make lint                      Run golangci-lint"
	@echo ""
	@echo "Database Migrations:"
	@echo "  make migrate-create NAME=name  Create a new migration"
	@echo "  make migrate-up                Apply all pending migrations"
	@echo "  make migrate-down [N]          Roll back the last migration(s)"
	@echo "  make migrate-force VERSION=n   Force migration version"
	@echo "  make migrate-version           Show current migration version"
	@echo "  make seed                      Seed the database"
	@echo ""
	@echo "Documentation:"
	@echo "  make gen-docs                  Generate Swagger documentation"
	@echo ""
	@echo "Docker:"
	@echo "  make up                        Build and start containers"
	@echo "  make up-back                   Build and start containers in background"
	@echo "  make down                      Stop and remove containers"
	@echo "  make logs                      Follow Docker logs"
	@echo "  make build                     Build Docker images"
	@echo "  make restart                   Restart all containers"
	@echo ""


.PHONY: test
test:
	@go test -v ./...


.PHONY: lint
lint: format
	golangci-lint run ./...


.PHONY: format
format:
	gofmt -s -w .


.PHONY: migrate-create
migrate-create:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DATABASE_ADDR)" up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DATABASE_ADDR)" down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-force
migrate-force:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DATABASE_ADDR)" force $(VERSION)

.PHONY: migrate-version
migrate-version:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DATABASE_ADDR)" version

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

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
