# Documentation Audit

**Repository:** `github.com/behshadrhp/Shadina`  
**Commit reviewed:** `6c14149` (`:zap: add: Authorization Layer`)  
**Review date:** 2026-10-29  
**Method:** source review, SQL/config review, read-only local runtime checks, generated-contract comparison

## Persian summary

پیش از این ممیزی، مستندات پروژه تقریباً منحصر به یک `README.md` دوخطی و فایل‌های تولیدشده Swagger بود. مستندات
معماری، دیتابیس، API، احراز هویت، پیکربندی، استقرار و تست وجود نداشت. Swagger موجود با کد ناسازگار است و فرایند
تولید مجدد آن به دلیل ارجاع به DTOهای حذف‌شده شکست می‌خورد. این مجموعه مستندات بر اساس کد واقعی ایجاد شده و
موارد تأییدنشده/ناقص صریحاً علامت‌گذاری شده‌اند.

## Existing documentation inventory

| Artifact | Content | Assessment |
|---|---|---|
| `README.md` | `# Shadina` + one sentence | Minimal; no setup, architecture, API, or DB information. |
| `docs/docs.go` | Generated swaggo template | Generated; do not edit manually. Runtime fills host/version via `SwaggerInfo`. |
| `docs/swagger.json` | OpenAPI 2.0, 9 paths, 8 definitions | Stale; generated from older annotations; no host/version; no operation security. |
| `docs/swagger.yaml` | Same contract in YAML | Semantically consistent with JSON (verified by parsing/comparing); equally stale. |
| `makefile` help text | Lists targets | De facto command documentation; several targets are broken. |
| `.air.toml` | Air config | Operational config; `pre_cmd` invokes failing docs generation. |
| Handler annotations | Swagger comments | Incomplete/inconsistent; reference missing DTOs. |
| `LICENSE` | MIT | Not technical documentation. |
| `logs/`, `tmp/` | Runtime artifacts | Ignored; not documentation sources. |
| `web/` | Empty | No frontend docs needed. |

No architecture decision records (ADRs), database docs, API guide, auth docs, runbooks, or test docs existed
before this set. No CI documentation existed because there is no CI.

## Missing documentation (before this work)

- Project purpose, capabilities, and non-capabilities.
- Architecture/component/package responsibilities and dependency graph.
- Startup/request lifecycle and middleware order.
- Database schema, constraints, indexes, relationships, ERD.
- Migration workflow, rollback caveats, seed behavior.
- API reference with real request/response examples and status codes.
- Authentication/authorization behavior and failure modes.
- Environment variables, defaults, secret handling, and the `AUTH_SECRET_KAY` typo.
- Error envelope and internal code registry.
- Logging/observability behavior and gaps.
- Docker/deployment reality (DB-only Compose, no app image).
- Development setup, troubleshooting, testing status.
- Security review and known risks.

## Documentation inconsistencies found

| # | Inconsistency | Evidence |
|---:|---|---|
| 1 | Swagger documents `/users/{userID}/feed`; actual route is `/feeds` | `cmd/api/feeds.go` annotation vs `cmd/api/router.go` |
| 2 | Swagger documents follow as `POST` with required body; actual is `PUT` with no body | `cmd/api/followers.go` annotation vs handler/router |
| 3 | Swagger documents unfollow as `DELETE` with body; actual is `PUT` with no body | Same |
| 4 | Swagger references `dto.FollowPayload` and `dto.UnFollowPayload`, which do not exist | `internal/dto` contains no such types; `swag init` fails |
| 5 | `POST /v1/auth/register` is not documented in Swagger | Router has the route; no annotation in `auth.go` |
| 6 | Swagger says create post returns 201 and 400; handler returns 200 and 500 for validation/JSON errors | `cmd/api/posts.go` |
| 7 | Swagger says user retrieve returns 400/404; handler returns 500 for parse/not-found | `cmd/api/uesrs.go` |
| 8 | Swagger declares `ApiKeyAuth` but no operation-level `security` | `docs/swagger.json`; no `@Security` annotations |
| 9 | Swagger static files lack `host`/`info.version`; runtime fills them | `docs/swagger.json`, `cmd/api/api.go` |
| 10 | Swagger UI spec URL lacks `/v1` and a scheme | `cmd/api/router.go`; generated UI HTML `url:` |
| 11 | Makefile help says `migrate-create NAME=...`; target does not pass name | `makefile`; `make -n` output |
| 12 | Makefile `seed` points to non-existent `cmd/migrate/seed/main.go` | `makefile`; `go run` error |
| 13 | Makefile `gen-docs` cannot succeed with current annotations | `swag init` error |
| 14 | `.example.env` lacks `SERVER_URL` and `AUTH_SECRET_KAY` | File comparison |
| 15 | Code fallback DSN (5435/`shadina`) differs from Compose/example (5454/`shadina_db`) | `cmd/api/main.go`, `.example.env`, compose |
| 16 | README implies a complete project; no setup docs | `README.md` |
| 17 | Swagger says feed supports `page`; code reads `offset` and defaults offset to 10 | `cmd/api/feeds.go`, `internal/storage/pagination.go` |
| 18 | Comments are described as part of posts; no comment write API exists | Code; no comment routes |

## Repository coverage

| Area | Files reviewed | Status | Evidence/notes |
|---|---|---|---|
| Entry points | `cmd/api/main.go`, `cmd/seed/main.go` | Complete | Traced startup/seed. |
| HTTP/router | `cmd/api/router.go`, `api.go` | Complete | Route tree and server config verified. |
| Handlers/middleware | all `cmd/api/*.go` | Complete | Behavior and error paths reviewed. |
| DTO/response/utils | `internal/dto`, `internal/response`, `internal/utils` | Complete | Envelope/codes documented. |
| Auth | `internal/auth` | Complete | JWT claims/validation documented. |
| Storage/SQL | `internal/storage` | Complete | Queries mapped; feed issues verified. |
| DB bootstrap/seed | `internal/db` | Complete | Seed issues documented. |
| Migrations | 18 SQL files | Complete | Inventory and live schema compared. |
| Swagger | `docs/` | Partial | Stale; generation blocked. |
| Configuration | `.env`, `.example.env`, compose, makefile, `.air.toml`, env package | Complete | Secrets not reproduced. |
| Docker/deployment | compose, makefile | Partial | DB-only; no app image/prod deployment. |
| Tests/quality | whole repo | Partial | No tests; build/vet/lint results documented. |
| Security | whole repo | Partial | Static review + limited runtime checks; no DAST/CVE scan. |
| Logging/observability | main, middleware, response | Complete | Gaps documented. |
| Web | `web/` | Not Implemented | Empty. |
| CI/CD | none | Not Implemented | No config found. |

## Structural issues in documentation

1. Swagger is the only machine-readable contract, but it is stale and cannot currently be regenerated.
2. There is no single source for error codes; codes are scattered and collide.
3. There is no documentation for the unusual `followers` column semantics.
4. Environment documentation is incomplete and contains a misspelled variable.
5. Generated and handwritten docs are mixed without a warning (except the generated header in `docs.go`).

## Recommendations

1. Keep the new `docs/fa` and `docs/en` sets as the human documentation source; link them from the root
   `README.md`.
2. Fix the follow/unfollow annotations and regenerate Swagger; add `@Security` and the register endpoint.
3. Add a CI check that regenerated Swagger matches the committed files.
4. Document the correct migration/seed commands and fix the makefile.
5. Add request/response examples generated from tests, not hand-maintained only.
6. Mark deprecated/stale docs explicitly if they cannot be updated immediately.

## Related documents

- [Documentation coverage](documentation-coverage.md)
- [API consistency review](api-consistency-review.md)
- [Architecture review](architecture-review.md)
