# 13. Development Guide

## Prerequisites

| Tool | Version / requirement | Why |
|---|---|---|
| Go | `1.25.4` (`go.mod`); review used `go1.25.4` | Build/run/test |
| Docker Engine + Compose | Any recent version; review used Docker 29.8.0 / Compose v5.5.1 | Local PostgreSQL |
| golang-migrate CLI | Review used `4.16.2` | Apply/revert migrations |
| `make` | GNU make | Convenience targets |
| `swag` | CLI `1.16.4` during review; module `1.16.6` | Swagger generation (currently blocked) |
| `golangci-lint` | Optional; review used `2.15.0` | Linting |
| `air` | Optional | Hot reload; pre-command currently fails |
| `psql` / Docker exec | Optional | DB inspection |

No Node.js, Java, Python, or frontend toolchain is required for the backend. `web/` is empty.

## 1. Get the repository and inspect the environment

```bash
git clone <repository-url>
cd Shadina
git status
ls -l .env .example.env
```

If `.env` already exists, **do not overwrite it**. It is ignored by Git and may contain local values. Read it
and add only missing keys.

## 2. Create/update `.env`

If needed:

```bash
cp .example.env .env
```

Then ensure these values are present:

```dotenv
SERVER_ADDR=:8000
SERVER_URL=127.0.0.1:8000
APP_VERSION=0.0.1
APP_ENVIRONMENT=development

DATABASE_DRIVER=postgres
DATABASE_ADDR=postgres://user:password@127.0.0.1:5454/shadina_db?sslmode=disable
MAX_OPEN_CONNS=30
MAX_IDLE_CONNS=30
MAX_IDLE_TIMEOUT=15m

# Typo is required by the current code. Use a strong random value.
AUTH_SECRET_KAY=<replace-with-a-long-random-secret>
```

Notes:

- The code reads `AUTH_SECRET_KAY`, not `AUTH_SECRET_KEY`.
- `.example.env` lacks `SERVER_URL` and the auth secret.
- The code's fallback DSN uses port **5435** and database `shadina`; the Compose service uses **5454** and
  `shadina_db`. Use the Compose values for local development.
- Do not commit `.env` or real secrets.

## 3. Start PostgreSQL

```bash
docker compose up -d
# or: make up-back
```

Check:

```bash
docker compose ps
docker compose logs postgresql
```

The container is named `shadina-postgres`; PostgreSQL is exposed on host port **5454**.

## 4. Apply migrations

```bash
make migrate-up
```

Prerequisites: `migrate` on `PATH`, `.env` readable, PostgreSQL reachable. This creates/alters tables and
indexes and writes `schema_migrations`. It is safe on a fresh development DB but always review before running
against data you care about.

Check the version:

```bash
make migrate-version
```

Expected local state at review: version `12`, `dirty = false`.

To create a migration (the make target is broken):

```bash
migrate create -seq -ext sql -dir ./cmd/migrate/migrations add_some_change
```

Edit the generated `.up.sql`/`.down.sql`, review, then `make migrate-up`. Do not run migrations solely for
documentation.

## 5. Seed development data

The make target is broken:

```bash
make seed
# => go run cmd/migrate/seed/main.go
# => stat cmd/migrate/seed/main.go: no such file or directory
```

Use the real command:

```bash
go run ./cmd/seed
```

This inserts 100 users, 200 posts, and 500 comments. It is **not idempotent**: rerunning fails on duplicate
users and may leave partial data. It does not create followers. Seeded users have empty password hashes and
**cannot log in**.

If you need a login-capable account, register one through the API (below) or insert a user with a real bcrypt
hash using a controlled script. Do not manually edit production data.

## 6. Run the API

```bash
go run ./cmd/api
```

The process:

- loads `.env` from the current directory;
- creates `logs/`;
- connects to PostgreSQL and pings it;
- listens on `SERVER_ADDR` (default `:8000`);
- logs to `logs/app.log`.

Verify:

```bash
curl -sS http://127.0.0.1:8000/v1/health
curl -sS http://127.0.0.1:8000/ping
```

`/v1/health` returns JSON and does not check the DB. `/ping` returns `.`.

## 7. Register and authenticate a user

The seeded accounts cannot log in. Create a user:

```bash
curl -sS -X POST http://127.0.0.1:8000/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"dev@example.com","username":"devuser","password":"correct horse battery"}'
```

Expected: HTTP 200, code `1010`, user data with `role_id: 1`.

Obtain a token:

```bash
curl -sS -X POST http://127.0.0.1:8000/v1/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"email":"dev@example.com","password":"correct horse battery"}'
```

Use it:

```bash
TOKEN='<paste-token>'
curl -sS http://127.0.0.1:8000/v1/users \
  -H "Authorization: Bearer $TOKEN"
```

Notes:

- The token is valid for 72 hours.
- There is no refresh/logout endpoint.
- Invalid tokens currently return HTTP 500, not 401.
- Use a unique email/username; duplicates return HTTP 500 with raw DB text.

## 8. Build

```bash
# Recommended: do not overwrite the tracked bin/main
go build -o /tmp/shadina-api ./cmd/api

# Build the seed command
go build -o /tmp/shadina-seed ./cmd/seed

# Build all packages; cmd/migrate is skipped (no Go files)
go build ./...
```

`go build ./cmd/migrate` fails; use the `migrate` CLI. `bin/main` is already tracked and may be stale.

## 9. Run checks

```bash
go test ./...            # no tests; every package reports [no test files]
go test -race ./...      # no tests
go test -cover ./...     # 0.0% for packages with statements
go vet ./...             # passes
go mod verify            # all modules verified
golangci-lint run ./...  # currently fails with 12 issues
```

`make test` runs `go test -v ./...`; it exits 0 but runs no tests. `make lint` first runs `gofmt -s -w .`,
which **modifies files**, then lints. Run `gofmt -l .` first if you only want to check formatting.

## 10. Swagger / API docs

```bash
make gen-docs
# => swag init -g ./api/main.go -d cmd,internal && swag fmt
```

This currently fails with:

```text
ParseComment error ... dto.FollowPayload ... cannot find type definition
```

The checked-in Swagger files are stale. To regenerate, first fix the follow/unfollow annotations or add the
missing DTOs, then run `swag init` and review the diff. Do not hand-edit `docs/docs.go`.

Swagger UI is served at:

```text
http://127.0.0.1:8000/v1/swagger/index.html
http://127.0.0.1:8000/v1/swagger/doc.json
```

The UI's configured spec URL is wrong (missing `/v1` and scheme); see [API reference](07-api-reference.md).

## 11. Hot reload (optional)

`.air.toml` builds `./cmd/api/*.go` to `./bin/main` and runs it, with `pre_cmd = ["make gen-docs"]`. Because
`make gen-docs` currently fails, `air` may not start or may log build errors to `tmp/build-errors.log`. Until
the Swagger annotations are fixed, prefer `go run ./cmd/api`.

## 12. Common development workflow

1. Add request DTO in `internal/dto` (if needed).
2. Add/extend storage methods in `internal/storage`; keep SQL parameterized.
3. Add handler in `cmd/api`; use `response` helpers consistently.
4. Register route and middleware in `cmd/api/router.go`.
5. Add Swagger annotations; ensure referenced DTOs exist.
6. Add tests (there is no test harness yet).
7. Run `gofmt`, `go vet`, `go test`, and `golangci-lint`.
8. Review DB effects and error mappings.

## 13. Database reset (destructive; do not run without approval)

These commands delete data:

```bash
# Stops and removes containers but keeps the named volume:
docker compose down

# Deletes the database volume too (DATA LOSS):
docker compose down -v

# Rolls back all migrations (drops schema objects):
make migrate-down
```

Use only on disposable local environments, and confirm with the team first.

## 14. Troubleshooting quick links

- `.env` missing → `error loading .env file`
- Connection refused → check port 5454 vs code fallback 5435, container status
- `make seed` fails → use `go run ./cmd/seed`
- Seeded login fails → empty password; register a new user
- Swagger generation fails → missing `dto.FollowPayload`
- Invalid token returns 500 → current middleware behavior
- Delete post returns 500 → comments FK has no cascade

Full list: [Troubleshooting](16-troubleshooting.md).

## Related documents

- [Configuration](09-configuration.md)
- [Docker and deployment](12-docker-deployment.md)
- [Database operations](06-database-operations.md)
- [Testing and quality](14-testing-quality.md)
