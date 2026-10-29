# Documentation Coverage

**Repository:** `github.com/behshadrhp/Shadina`  
**Commit reviewed:** `6c14149` (`:zap: add: Authorization Layer`)  
**Review date:** 2026-10-29  
**Scope:** the bilingual documentation set in `docs/en`, `docs/fa`, and `docs/reports`

## Persian summary

این گزارش پوشش مستندات را بر اساس کد واقعی، مهاجرت‌ها، پیکربندی و بررسی‌های زمان اجرا اندازه‌گیری می‌کند.
هر ۱۵ مسیر HTTP، ۵ جدول دیتابیس، ۱۱ متغیر محیطی، ۱۸ مهاجرت، ۳۹ کد خطای داخلی و ۳۱ فایل Go پوشش داده
شده‌اند. شکاف‌های باقی‌مانده عمدتاً مربوط به مواردی است که بدون محیط تولید، تست خودکار یا ابزارهای نصب‌نشده
قابل تأیید نبودند: Swagger، Mermaid CLI، تحلیل CVE و رفتار واقعی مسیر ادمین.

## Coverage summary

| Metric | Covered | Total | Coverage | Status |
|---|---:|---:|---:|---|
| HTTP routes registered by the router | 15 | 15 | 100% | Complete |
| HTTP layer Go files (`cmd/api`) | 10 | 10 | 100% | Complete |
| Go source files (incl. generated `docs/docs.go`) | 31 | 31 | 100% | Complete |
| SQL migrations | 18 | 18 | 100% | Complete |
| Database tables | 5 | 5 | 100% | Complete |
| Environment variables read by code | 11 | 11 | 100% | Complete |
| Internal error codes | 39 | 39 | 100% | Complete |
| Mermaid diagrams | 11 | 11 | 100% | Partial (syntax not CLI-validated) |
| English documents | 18 | 18 | 100% | Complete |
| Persian documents | 18 | 18 | 100% | Complete |
| Audit reports | 6 | 6 | 100% | Complete |
| Automated Go tests | 0 | 0 | n/a | Not Implemented |

## Document-by-document coverage

| # | Area | English | Persian | Primary evidence | Status | Known gap |
|---:|---|---|---|---|---|---|
| — | Index and conventions | `docs/en/README.md` | `docs/fa/README.md` | Repository tree, commit metadata | Complete | — |
| 01 | Project overview | `docs/en/01-project-overview.md` | `docs/fa/01-project-overview.md` | `go.mod`, `cmd/`, `internal/`, `web/` | Complete | — |
| 02 | Architecture | `docs/en/02-architecture.md` | `docs/fa/02-architecture.md` | All packages, import graph, wiring in `main.go` | Complete | — |
| 03 | Project structure | `docs/en/03-project-structure.md` | `docs/fa/03-project-structure.md` | Full file inventory | Complete | — |
| 04 | Application flow | `docs/en/04-application-flow.md` | `docs/fa/04-application-flow.md` | `main.go`, `router.go`, `middleware.go` | Complete | — |
| 05 | Database design | `docs/en/05-database-design.md` | `docs/fa/05-database-design.md` | 18 migrations, live schema introspection | Complete | — |
| 06 | Database operations | `docs/en/06-database-operations.md` | `docs/fa/06-database-operations.md` | Migrations, seed code, storage SQL | Complete | Rollback of every migration not executed |
| 07 | API reference | `docs/en/07-api-reference.md` | `docs/fa/07-api-reference.md` | Router + all handlers; read-only runtime calls | Complete | No load/performance testing |
| 08 | Auth and authorization | `docs/en/08-authentication-authorization.md` | `docs/fa/08-authentication-authorization.md` | `internal/auth`, middleware, handlers | Complete | Admin path not exercised (no admin user) |
| 09 | Configuration | `docs/en/09-configuration.md` | `docs/fa/09-configuration.md` | `.env`, `.example.env`, `internal/env`, Compose | Complete | Production config unavailable; secrets not reproduced |
| 10 | Error handling | `docs/en/10-error-handling.md` | `docs/fa/10-error-handling.md` | `internal/response`, `internal/consts`, handlers | Complete | — |
| 11 | Logging and observability | `docs/en/11-logging-observability.md` | `docs/fa/11-logging-observability.md` | `zap` setup, middleware, `logs/` | Complete | No metrics/tracing to verify |
| 12 | Docker and deployment | `docs/en/12-docker-deployment.md` | `docs/fa/12-docker-deployment.md` | `docker-compose.yaml`, `makefile`, `.air.toml` | Partial | No app Dockerfile, no production environment |
| 13 | Development guide | `docs/en/13-development-guide.md` | `docs/fa/13-development-guide.md` | Makefile, env files, migrations, seed | Complete | Commands verified on Linux only |
| 14 | Testing and quality | `docs/en/14-testing-quality.md` | `docs/fa/14-testing-quality.md` | `go build/vet/test`, `golangci-lint`, coverage | Partial | No test suite exists; no CI |
| 15 | Security review | `docs/en/15-security-review.md` | `docs/fa/15-security-review.md` | Static review + limited runtime checks | Partial | No DAST, no `govulncheck`, no production test |
| 16 | Troubleshooting | `docs/en/16-troubleshooting.md` | `docs/fa/16-troubleshooting.md` | Observed failures and source paths | Complete | Only locally reproduced failures |
| 17 | Glossary | `docs/en/17-glossary.md` | `docs/fa/17-glossary.md` | Code terminology | Complete | — |
| — | Diagrams (11 sources) | `docs/en/diagrams/*.mmd` | `docs/fa/diagrams/*.mmd` | Source + schema | Partial | `mmdc` not installed; syntax reviewed manually |
| — | Documentation audit | `docs/reports/documentation-audit.md` | Persian summary only | Repository docs | Complete | — |
| — | Architecture review | `docs/reports/architecture-review.md` | Persian summary only | Source review | Complete | — |
| — | Database review | `docs/reports/database-review.md` | Persian summary only | Schema + query review | Complete | — |
| — | API consistency review | `docs/reports/api-consistency-review.md` | Persian summary only | Swagger vs router/handlers | Complete | Regeneration blocked |
| — | Security review | `docs/reports/security-review.md` | Persian summary only | Static/runtime review | Partial | See SEC-01–SEC-16 |
| — | Documentation coverage | `docs/reports/documentation-coverage.md` | This summary | This report | Complete | — |

## API route coverage

Every route registered in `cmd/api/router.go` is documented with method, path, authentication requirement,
request payload, response envelope, status codes, and observed quirks.

| Method | Path | Auth | Documented in |
|---|---|---|---|
| `GET` | `/ping` | Public | 07, 04 |
| `GET` | `/v1/health` | Public | 07 |
| `POST` | `/v1/auth/register` | Public | 07, 08 |
| `POST` | `/v1/auth/token` | Public | 07, 08 |
| `GET` | `/v1/users` | JWT | 07 |
| `GET` | `/v1/users/{userID}` | JWT | 07 |
| `PUT` | `/v1/users/{userID}/follow` | JWT | 07, 08 |
| `PUT` | `/v1/users/{userID}/unfollow` | JWT | 07, 08 |
| `GET` | `/v1/users/{userID}/feeds` | JWT | 07, 06 |
| `GET` | `/v1/posts` | JWT | 07 |
| `POST` | `/v1/posts` | JWT | 07 |
| `GET` | `/v1/posts/{postID}` | JWT | 07 |
| `PATCH` | `/v1/posts/{postID}` | JWT + owner/admin | 07, 08 |
| `DELETE` | `/v1/posts/{postID}` | JWT + owner/admin | 07, 08 |
| `GET` | `/v1/swagger/*` | Public | 07, 12 |

`GET /v1/swagger/*` is a subtree serving the UI and the JSON document; it is counted once.

## Database object coverage

| Object | Documented | Evidence |
|---|---|---|
| `roles` | Yes | Migration 1; live schema |
| `users` | Yes | Migration 1; storage queries |
| `posts` | Yes | Migration 1; storage queries |
| `comments` | Yes | Migration 4; storage queries |
| `followers` | Yes | Migration 6; follow SQL |
| Schema version table (`schema_migrations`) | Yes | golang-migrate behavior |
| Indexes and constraints | Yes | Live schema introspection and migrations enumerated in 05/06 |
| Foreign-key behavior | Yes | Rolled-back deletion test |

## Environment variable coverage

| Variable | Read in | Default / fallback documented | Secret handling |
|---|---|---|---|
| `SERVER_ADDR` | `cmd/api/main.go` | `:8000` | Not secret |
| `SERVER_URL` | `cmd/api/main.go`, `router.go`, `api.go` | `127.0.0.1:8000` | Not secret |
| `APP_VERSION` | `cmd/api/main.go` | `0.0.1` | Not secret |
| `APP_ENVIRONMENT` | `cmd/api/main.go` | `development` | Not secret |
| `DATABASE_DRIVER` | `cmd/api/main.go` | `postgres` | Not secret |
| `DATABASE_ADDR` | `cmd/api/main.go`, `cmd/seed/main.go` | fallback DSN (port 5435/`shadina`) | Value never reproduced beyond the documented code fallback |
| `MAX_OPEN_CONNS` | `cmd/api/main.go`, `cmd/seed/main.go` | `30` | Not secret |
| `MAX_IDLE_CONNS` | `cmd/api/main.go`, `cmd/seed/main.go` | `30` | Not secret |
| `MAX_IDLE_TIMEOUT` | `cmd/api/main.go`, `cmd/seed/main.go` | `15m` | Not secret |
| `AUTH_SECRET_KAY` | `cmd/api/main.go` | hard-coded fallback | Name (not value) documented; value never reproduced |
| `DB_DRIVER` | `cmd/seed/main.go` | `postgres` | Not secret |

The Compose file also defines `POSTGRES_*` variables for the database container; these are documented in
[Configuration](../en/09-configuration.md) and [Docker and deployment](../en/12-docker-deployment.md).

## Error-code coverage

All 39 distinct error codes and 11 success codes passed to the response helpers at handler/middleware call
sites are listed in [Error handling](../en/10-error-handling.md), including collisions such as `1004`,
`5004`, `5008`, and `5018`. (`internal/consts` defines shared values such as `NOT_FOUND` and timeouts, not
the numeric response codes.) HTTP status mapping, validation failures, JSON failures, and storage failures
are documented per endpoint.

## Verification coverage

| Check | Performed | Result |
|---|---|---|
| `go build ./...` | Yes | Pass |
| `go vet ./...` | Yes | Pass |
| `go test ./...` | Yes | Pass; no test files |
| `go test -race ./...` | Yes | Pass; no test files |
| `go test -cover ./...` | Yes | 0.0% coverage |
| `golangci-lint run ./...` | Yes | 12 issues (9 `errcheck`, 3 `staticcheck`) |
| `go mod tidy -diff` | Yes | Diff: unused `github.com/golang-jwt/jwt` v3 would be removed |
| `swag init` | Yes | Failed: missing `dto.FollowPayload` / `dto.UnFollowPayload` |
| `go build ./cmd/migrate` | Yes | Failed: no Go files |
| `make seed` | Yes | Failed: missing `cmd/migrate/seed/main.go` |
| Mermaid CLI validation | No | `mmdc` not installed |
| Browser Swagger UI test | No | Not performed |
| `govulncheck` | No | Tool unavailable |
| DAST / penetration test | No | Out of scope |
| Production deployment check | No | No environment/credentials |

## Intentional exclusions

- **Secrets and PII:** `.env` values, JWT secret, database password, and user records are not reproduced.
- **Generated Swagger:** `docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml` were not modified; the
  regeneration failure is reported instead of patching generated files.
- **Business-logic changes:** findings are recorded as recommendations in the reports; no application code,
  API, schema, or migration was changed.
- **Database writes:** runtime checks were read-only except for a transaction that was explicitly rolled back.
- **Frontend:** `web/` is empty, so no UI documentation exists.

## Remaining documentation gaps

1. **Swagger contract:** cannot be regenerated until the missing DTO references are fixed. The checked-in
   contract is stale and lacks operation-level security.
2. **Mermaid rendering:** diagrams are syntactically reviewed but not rendered with `mmdc`.
3. **Admin authorization path:** the double-write bug is documented from code; no admin account exists to
   reproduce it over HTTP.
4. **Production behavior:** TLS, reverse proxy, backups, scaling, and secret management are not implemented in
   the repository and therefore cannot be documented as existing behavior.
5. **Dependency vulnerabilities:** no CVE scan was run; `go.mod` includes an unused JWT v3 dependency.
6. **Performance:** no benchmarks or load tests exist; N+1 and full-table scan risks are documented from code.

## Related documents

- [Documentation audit](documentation-audit.md)
- [Architecture review](architecture-review.md)
- [Database review](database-review.md)
- [API consistency review](api-consistency-review.md)
- [Security review](security-review.md)
