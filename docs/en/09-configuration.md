# 09. Configuration

## Loading model

The API and seed commands call `godotenv.Load()` with no filename. That loads `.env` from the **current
working directory**. It does not override variables already present in the process environment. If `.env`
is missing or unreadable, both commands exit with `error loading .env file`.

The makefile does:

```make
include .env
export
```

Therefore every `make` target requires a readable `.env`, even targets that only print help or run Docker.
The Compose file's PostgreSQL credentials are separate from the application's environment; the API reads
`DATABASE_ADDR` from `.env`/the process environment.

## Environment variables used by the code

| Variable | Purpose | Required | Default in code | Example | Source |
|---|---|---|---|---|---|
| `SERVER_ADDR` | HTTP listen address | No | `:8000` | `:8000` | `cmd/api/main.go` |
| `SERVER_URL` | Swagger host and doc URL construction | No | `127.0.0.1:8000` | `127.0.0.1:8000` | `cmd/api/main.go`, `router.go`, `api.go` |
| `APP_VERSION` | Version shown by health and Swagger | No | `0.0.1` | `0.0.1` | `cmd/api/main.go` |
| `APP_ENVIRONMENT` | Environment label shown by health | No | `development` | `development` | `cmd/api/main.go` |
| `DATABASE_DRIVER` | `database/sql` driver name for the API | No | `postgres` | `postgres` | `cmd/api/main.go` |
| `DATABASE_ADDR` | PostgreSQL DSN | No | `postgres://postgres:postgres@127.0.0.1:5435/shadina?sslmode=disable` | `postgres://user:password@127.0.0.1:5454/shadina_db?sslmode=disable` | `cmd/api/main.go`, `cmd/seed/main.go` |
| `MAX_OPEN_CONNS` | `SetMaxOpenConns` | No | `30` | `30` | `cmd/api/main.go`, `cmd/seed/main.go` |
| `MAX_IDLE_CONNS` | `SetMaxIdleConns` | No | `30` | `30` | `cmd/api/main.go`, `cmd/seed/main.go` |
| `MAX_IDLE_TIMEOUT` | `SetConnMaxIdleTime` (Go duration) | No | `15m` | `15m` | `cmd/api/main.go`, `cmd/seed/main.go` |
| `AUTH_SECRET_KAY` | JWT HMAC secret (note spelling) | **Effectively required**; code has a fallback | Hard-coded development fallback in `cmd/api/main.go` | a long random value, e.g. `<32+ random bytes>` | `cmd/api/main.go` |
| `DB_DRIVER` | Driver name used **only by `cmd/seed`** | No | `postgres` | `postgres` | `cmd/seed/main.go` |

`AUTH_SECRET_KAY` is not present in `.env` or `.example.env` at review time. `AUTH_SECRET_KEY` (the natural
spelling) is **not read**. If an operator sets the correctly spelled variable, the service silently uses the
hard-coded fallback.

`SERVER_URL` is present in the local `.env` but absent from `.example.env`. It is used as Swagger's `host`
(without a scheme) and as the base of the UI's `doc.json` URL. The actual Swagger handler path is
`/v1/swagger/doc.json`, so the constructed URL does not match (see [API reference](07-api-reference.md)).

### Default DSN mismatch

The code fallback is:

```text
postgres://postgres:postgres@127.0.0.1:5435/shadina?sslmode=disable
```

The Compose service is:

```text
host port 5454 -> container 5432
database shadina_db
user user
password password
```

The tracked `.example.env` uses the Compose values:

```text
postgres://user:password@127.0.0.1:5454/shadina_db?sslmode=disable
```

If `.env` is missing the DSN, the API tries port 5435/database `shadina` and fails against the provided
Compose service. This is a common local setup error.

## Hard-coded configuration

These values are **not** environment-configurable:

| Setting | Value | Location |
|---|---|---|
| JWT lifetime | 72 hours | `cmd/api/main.go` (`time.Hour * 24 * 3`) |
| JWT issuer | `Shadina` | `cmd/api/main.go` |
| JWT audience | `Shadina` | `cmd/api/main.go` |
| HTTP read timeout | 10 seconds | `internal/consts/consts.go` |
| HTTP write timeout | 30 seconds | `internal/consts/consts.go` |
| HTTP idle timeout | 1 minute | `internal/consts/consts.go` |
| DB/context timeout | 10 seconds | `internal/consts/consts.go` |
| Request body limit | 1,048,578 bytes | `internal/utils/json.go` |
| Log paths | `logs/app.log`, `logs/error.log` | `cmd/api/main.go` |
| Log directory mode | `0755` | `cmd/api/main.go` |
| bcrypt cost | `bcrypt.DefaultCost` | `internal/storage/users.go` |
| Swagger base path | `/v1/` at runtime; `/v1` in annotation | `cmd/api/api.go`, `cmd/api/main.go` |

There is no `ReadHeaderTimeout` or `MaxHeaderBytes` setting. There is no CORS configuration.

## Example `.env` (placeholders only)

Do not copy real secrets into documentation or version control. A development file can look like:

```dotenv
# SERVER
SERVER_ADDR=:8000
SERVER_URL=127.0.0.1:8000
APP_VERSION=0.0.1
APP_ENVIRONMENT=development

# DATABASE (matches docker-compose.yaml)
DATABASE_DRIVER=postgres
DATABASE_ADDR=postgres://user:password@127.0.0.1:5454/shadina_db?sslmode=disable
MAX_OPEN_CONNS=30
MAX_IDLE_CONNS=30
MAX_IDLE_TIMEOUT=15m

# AUTH: typo is intentional in the current code; use a strong random value
AUTH_SECRET_KAY=<replace-with-a-long-random-secret>
```

The values `user`/`password` above are the Compose development defaults, not production credentials. Use a
secret manager or environment injection in real deployments.

## Configuration validation

`internal/env/env.go` is intentionally minimal:

- `GetString` returns the fallback only when the variable is **absent**; an empty string is returned as-is.
- `GetInt` returns the fallback if the value is absent or cannot be parsed with `strconv.Atoi`.
- There is no range validation for pool sizes.
- `MAX_IDLE_TIMEOUT` is parsed by `db.NewDB`; an invalid duration causes startup failure.
- There is no validation that `AUTH_SECRET_KAY` is present, long enough, or not the fallback.
- There is no validation that `DATABASE_ADDR` points to a reachable database until `PingContext`.
- There is no validation of `SERVER_ADDR`.

## Configuration precedence

1. Process environment (highest; `godotenv.Load` does not overwrite).
2. `.env` in the current working directory.
3. Code fallbacks.

The Compose file sets only `POSTGRES_*` variables for the database container; it does not inject application
variables into the API because there is no API container.

## Secrets handling

- `.env` is ignored by Git (`.gitignore` line `.env`).
- `.example.env` is tracked and contains development database credentials only; it contains no JWT secret.
- `docker-compose.yaml` contains development database credentials; do not reuse them outside local
  development.
- Logs do not include the `Authorization` header, request bodies, or passwords in the code as reviewed.
- Error responses can include raw database/parser text; they are not a secret store, but they can leak
  schema/internal details.
- No secret manager, vault, sealed secret, or KMS integration exists.
- `bin/main` is a tracked binary; ensure it does not embed secrets at build time (none were configured in the
  reviewed code).

## Related documents

- [Development guide](13-development-guide.md)
- [Docker and deployment](12-docker-deployment.md)
- [Authentication and authorization](08-authentication-authorization.md)
- [Security review](../reports/security-review.md)
