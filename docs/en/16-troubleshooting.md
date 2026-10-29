# 16. Troubleshooting

## Startup and environment

| Symptom | Likely cause | Evidence / check | Safe resolution |
|---|---|---|---|
| `error loading .env file` | `.env` missing, unreadable, or process started from another directory | `ls -l .env`, `pwd` | Run from repository root; create `.env` from `.example.env`; do not commit secrets. |
| `make: *** No rule to make target ...` or make fails immediately | Makefile does `include .env`; missing `.env` breaks make | `ls -l .env` | Create `.env` or use direct `go`/`migrate` commands. |
| `dial tcp 127.0.0.1:5435: connect: connection refused` | Code fallback DSN uses 5435/`shadina`; Compose exposes 5454/`shadina_db` | `docker compose ps`; check `DATABASE_ADDR` | Set `DATABASE_ADDR` to the Compose DSN on port 5454. |
| `database stable publish` never appears / process exits | PostgreSQL not running or wrong DSN | `docker compose logs postgresql`; `psql`/`docker exec` | Start Compose; fix DSN. Do not delete volumes. |
| `invalid duration` / startup DB error | `MAX_IDLE_TIMEOUT` is not a Go duration such as `15m` | Inspect `.env` | Use `15m`, `1h`, etc. |
| Permission denied creating `logs/` | Working directory not writable, or existing `logs` ownership mismatch | `ls -ld logs` | Run from a writable directory; fix ownership with approval. |
| API starts but Swagger UI cannot load spec | `docsPath` lacks `/v1` and a scheme | Fetch `/v1/swagger/index.html` and inspect `url:` | Use `/v1/swagger/doc.json` directly; fix `router.go`/`SERVER_URL` in code. |
| `AUTH_SECRET_KAY` seems ignored | Correct spelling `AUTH_SECRET_KEY` is not read; code uses `KAY` | `grep AUTH .env` | Set the misspelled key for now, or fix code and rename consistently. |
| Tokens forged/valid unexpectedly | Env secret absent; hard-coded fallback in use | Check startup env; code fallback | Set a strong secret; remove fallback in code (needs change). |

## Database and migrations

| Symptom | Likely cause | Check | Resolution |
|---|---|---|---|
| `migrate: command not found` | CLI not installed/on PATH | `which migrate` | Install golang-migrate; verify `migrate -version`. |
| `make migrate-up` fails with connection error | Wrong `DATABASE_ADDR` or DB down | `make migrate-version`; `docker compose ps` | Fix DSN/start DB. |
| `Dirty database version` | A previous migration failed midway | `make migrate-version` | Inspect the failed migration; `migrate force` only after understanding state (dangerous). |
| `no Go files in .../cmd/migrate` | Trying `go build ./cmd/migrate` | `find cmd/migrate` | Use `migrate` CLI; there is no Go main. |
| `make migrate-create NAME=x` creates no files | Target does not pass `NAME`; expands without a name | `make -n migrate-create NAME=x` | Use `migrate create -seq -ext sql -dir ./cmd/migrate/migrations <name>`. |
| `make seed` → `stat cmd/migrate/seed/main.go: no such file` | Broken make target | `make -n seed` | Use `go run ./cmd/seed`. |
| Seed fails with duplicate key | Seed is not idempotent | Row counts/unique errors | Use a disposable DB, or make seed idempotent/clean up with approval. |
| Seed completes but login fails | Seed users have empty password hashes | `SELECT count(*) FROM users WHERE password=''::bytea;` (read-only) | Register a new user via `/v1/auth/register`. |
| Seed transaction does not roll back | `BeginTx` is not used by storage methods | Code review `internal/db/seed.go` | Fix code to pass `tx`; do not rely on rollback. |
| `scripts/init_db.sql` fails | `CREATE DATABASE IF NOT EXISTS` is not valid PostgreSQL | Read file | Do not use it; Compose creates the DB; migrations create schema. |
| Migration gap 8–10 is confusing | Files absent | `ls cmd/migrate/migrations` | Not a runtime error; document/renumber in a future change. |
| Delete post returns 500 | Comments FK has no cascade | `\d comments`; FK `fk_post_comment` | Delete comments first in a transaction or add cascade via migration (needs approval). |

## Authentication and authorization

| Symptom | Cause | Check | Resolution |
|---|---|---|---|
| 401 `authorization header is missing` | No `Authorization` header | Request headers | Add `Authorization: Bearer <token>`. |
| 401 `authorization header is malformed` | Lowercase `bearer`, multiple spaces, extra parts | Header format | Use exactly `Bearer <token>`. |
| 500 `token is malformed...` | Invalid/expired token; middleware maps to 500 | Response code 5019 | Obtain a fresh token; fix middleware to return 401 in code. |
| 401 code 4022 after token | User deleted or DB lookup failed | DB user ID / logs | Ensure user exists; check DB. |
| Admin update/delete returns odd/mixed response | Admin branch writes 401 after handler | Code `UserPostAccessMiddleware` | Fix middleware; do not rely on admin path until fixed. |
| Moderator cannot edit others' posts | Only `RoleID == 3` is allowed | Code | Expected current behavior; fix policy if needed. |
| 401 `access denied` deleting another user's post | Ownership check | Post `user_id` vs token user | Use the owner account; do not bypass. |
| Feed shows another user's posts | Path `userID` is not compared to token | Request `/v1/users/2/feeds` | Product decision; enforce check in code if private. |
| Feed `sort=asc/desc` same order | `$2` in `ORDER BY` is a constant | Compare responses | Fix SQL with a whitelisted sort expression. |
| Feed comment count inflated | Join duplicates rows for own posts | Compare `COUNT` vs `COUNT(DISTINCT)` | Fix query; see Database operations. |

## API behavior

| Symptom | Cause | Check | Resolution |
|---|---|---|---|
| `GET /v1/users/1/feed` 404 | Real route is `/feeds` | Router source | Use `/v1/users/1/feeds`. |
| `GET /v1/users/2/follow` 405 | Follow is `PUT`, not GET | Router source | Use `PUT` with a token. |
| Register returns 200, not 201 | `SuccessResponse` used | Handler source | Contract choice; fix code or docs. |
| Create post validation returns 500 | Handler uses `InternalErrorResponse` | Code | Fix to 400 in code. |
| `GET /v1/users/999999` returns 500 | Handler doesn't map `NOT_FOUND` | Code | Fix to 404 in code. |
| Non-numeric ID returns 500 with `strconv` text | Parse error mapped to internal | Code | Return 400 and a generic message. |
| Duplicate registration returns 500 with DB text | Unique violation not mapped | DB/error text | Pre-check or map unique violation to 409/400; sanitize. |
| Follow duplicate/unknown target returns 500 | PK/FK error not mapped | DB/error text | Map to 409/404; sanitize. |
| Unfollow returns success for nonexistent relationship | `RowsAffected` ignored | Code | Return 404 or document idempotent behavior. |
| Swagger generation fails | Missing `dto.FollowPayload`/`UnFollowPayload` | `swag init` output | Add DTOs or remove body annotations; regenerate. |
| Swagger UI spec fails to load | URL missing `/v1` and scheme | UI HTML `url:` | Use `/v1/swagger/doc.json`; fix router. |
| 404/405 responses are plain text | chi default handlers | `curl -i` unmatched route | Add JSON 404/405 handlers in code. |

## Logging and diagnostics

| Symptom | Cause | Check | Resolution |
|---|---|---|---|
| `logs/error.log` is empty | It is zap's internal error sink, not app errors | Code | Add explicit error logging; don't rely on this file. |
| No request logs after panic | Logger middleware unwound by panic/Recoverer order | Code | Fix middleware order/logging; add recovery logging. |
| Log file grows without rotation | No rotation policy | File size | Configure external logrotate/central logging. |
| `server stable` address looks wrong | Log uses `SERVER_URL`, not `SERVER_ADDR` | Code | Fix log field or config. |
| No request ID in errors | Not implemented | Code | Add request ID middleware and include it in responses/logs. |

## Safe diagnostic commands

```bash
# Read-only
docker compose ps
docker compose logs --tail=100 postgresql
curl -sS -i http://127.0.0.1:8000/v1/health
curl -sS -i http://127.0.0.1:8000/ping
curl -sS -i http://127.0.0.1:8000/v1/swagger/doc.json
go test ./...
go vet ./...
go build ./...
gofmt -l .
golangci-lint run ./...

# Database inspection (read-only)
docker exec shadina-postgres psql -U user -d shadina_db -c '\dt'
docker exec shadina-postgres psql -U user -d shadina_db -c '\d users'
docker exec shadina-postgres psql -U user -d shadina_db -c 'SELECT version, dirty FROM schema_migrations;'

# Avoid unless explicitly approved (destructive/data-changing):
# docker compose down -v
# migrate down
# DELETE/UPDATE/TRUNCATE/DROP
```

Do not paste tokens, passwords, full `.env` contents, or production DSNs into issues/chat.

## Related documents

- [Development guide](13-development-guide.md)
- [Configuration](09-configuration.md)
- [API reference](07-api-reference.md)
- [Database operations](06-database-operations.md)
