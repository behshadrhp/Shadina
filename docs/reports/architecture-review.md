# Architecture Review

**Repository:** `github.com/behshadrhp/Shadina`  
**Commit:** `6c14149`  
**Review date:** 2026-10-29  
**Scope:** all Go packages, routes, middleware, storage SQL, configuration, migrations, and deployment files

## Persian summary

معماری فعلی یک مونولیت لایه‌ای کوچک با `net/http`/chi است. نقاط قوت آن سادگی، SQL پارامتری، JWT با
issuer/audience/exp، bcrypt و مرزهای پکیجی نسبتاً روشن است. ضعف‌های اصلی: نبود لایه سرویس و مدل خطای
تایپ‌شده، نشت ساختارهای storage به پاسخ API، باگ مجوزدهی ادمین، کوئری feed با شمارش تکراری و sort بی‌اثر،
نبود graceful shutdown، نبود تست/CI و نبود observability است. پیشنهادهای P0/P1/P2 در ادامه آمده است.

## Current architecture

```text
HTTP client
  → net/http Server (timeouts)
  → chi router
  → global middleware (Heartbeat, RealIP, chi Logger, Recoverer, app Logger)
  → route-level middleware (JWT, post ownership)
  → handler (cmd/api)
  → storage facade (internal/storage)
  → database/sql + lib/pq
  → PostgreSQL
```

Supporting packages: `internal/auth` (JWT), `internal/dto` (request payloads), `internal/response`
(envelope), `internal/utils` (JSON/validator), `internal/env` (env lookup), `internal/consts` (timeouts,
sentinel), `docs` (generated Swagger). `cmd/seed` is a separate executable. `cmd/migrate` contains SQL only;
migrations are applied by an external CLI.

The style is a **layered modular monolith**. There is no service/domain layer, no repository-per-entity, no
ORM, no message broker, no cache, and no external identity provider. Dependency injection is manual in
`cmd/api/main.go`.

## Strengths

1. **Small, readable surface.** The entire API fits in a handful of files; a new developer can trace a request
   without framework magic.
2. **Explicit middleware order.** Authentication and ownership are visible in `router.go`/`middleware.go`.
3. **Parameterized SQL.** No SQL injection was found; search and arrays use parameters.
4. **Reasonable auth primitives.** bcrypt, HS256, issuer/audience/expiry checks, valid-method restriction.
5. **Storage facade.** Handlers depend on `storage.Storage`, which is a useful seam for tests/fakes.
6. **Operational basics.** Server timeouts, DB pool limits, a startup ping, structured zap logs, body limit.
7. **Versioned schema.** golang-migrate migrations with up/down pairs.
8. **Swagger integration exists.** The contract is generated and served by the API (even though stale).

## Problems observed

### Architecture and boundaries

| # | Issue | Evidence | Impact |
|---:|---|---|---|
| A1 | No service/domain layer; handlers orchestrate SQL and business rules | `cmd/api/*.go` call `app.storage` directly | Duplication, hard-to-test logic, mixed concerns. |
| A2 | Storage structs are API response models | `storage.User`/`Post`/`Comment` have JSON tags and are returned directly | DB row shape changes become API breaking changes; zero-value relational fields leak. |
| A3 | Storage package parses HTTP query parameters | `internal/storage/pagination.go` imports `net/http` | Persistence layer depends on transport. |
| A4 | DB bootstrap contains seed fixtures | `internal/db/seed.go` imports `internal/storage` | `internal/db` is not a pure connection package; fixture logic coupled to DB package. |
| A5 | Global validator singleton | `internal/utils/validator.go`, `init()` in main | Hidden dependency; tests must initialize it; nil risk for alternative entry points. |
| A6 | Context key is a string literal | `GetUserCTX` uses `"user"`; staticcheck SA1029 | Collision/fragility; should use a private typed key. |
| A7 | No typed errors | `err.Error()` returned to clients | Inconsistent statuses, info leakage, hard to test. |
| A8 | No transactions for multi-step operations | post delete + comments FK | Partial failures and 500s. |

### Authorization

| # | Issue | Evidence | Impact |
|---:|---|---|---|
| B1 | Admin branch calls `next` then writes 401 | `UserPostAccessMiddleware` | Response corruption; admin actions may execute with a misleading result. |
| B2 | Moderator role is not honored | Only `RoleID == 3` bypass | Role description contradicts behavior. |
| B3 | Hard-coded role IDs | `RoleID: 1`, `RoleID == 3` | Fragile if roles are reseeded/renumbered. |
| B4 | Feed path user not checked | `listFeedsHandler` | Any authenticated user can read any feed. |
| B5 | 401 used for forbidden | ownership denial | Semantically wrong; clients cannot distinguish auth failure from permission denial. |

### Data access and correctness

| # | Issue | Evidence | Impact |
|---:|---|---|---|
| C1 | Feed comment counts inflated | join duplicates; local post 19: 16 vs distinct 4 | Wrong API data. |
| C2 | Feed sort ineffective | `ORDER BY p.created_at, $2` | Sort parameter is ignored. |
| C3 | Feed default offset 10 | `CreatePaginatedQuery(20, 10, ...)` | First page skips rows; Swagger says page 1. |
| C4 | Feed inner join on followers | own posts require a follower row | Empty feed for users with no follower graph. |
| C5 | N+1 comment queries | `listPostsHandler` loop | 204 queries for 203 posts. |
| C6 | Post delete blocked by comments FK | `fk_post_comment` no cascade | 500 for posts with comments. |
| C7 | `updated_at` never updated | update SQL omits it; no trigger | Stale timestamps. |
| C8 | User list omits `role_id` | `UsersStorage.List` select | `role_id: 0` in API. |
| C9 | Unfollow ignores `RowsAffected` | `Follow`/`Unfollow` | Success for no-op; 404 branch dead. |
| C10 | Seed transaction is a no-op | `BeginTx` not passed to stores | No atomicity; partial seed data. |
| C11 | Seed users have empty passwords | `generateUsers` | Cannot log in; misleading fixtures. |

### Reliability, observability, operations

| # | Issue | Evidence | Impact |
|---:|---|---|---|
| D1 | No graceful shutdown | `ListenAndServe` + `log.Fatal` | Dropped requests; defers skipped. |
| D2 | Health does not check DB | `healthCheckHandler` | DB outage still reports `ok`. |
| D3 | No request IDs/metrics/tracing | no code | Hard to debug/operate. |
| D4 | Error log is not an app error log | zap error sink; 0 bytes | Errors invisible. |
| D5 | No tests/CI | no `*_test.go`; no CI | Regressions unguarded. |
| D6 | Swagger generation blocked | missing DTOs | Contract drifts further. |
| D7 | Makefile targets broken | seed/create/docs | Developer friction. |
| D8 | No app Dockerfile | Compose DB only | No reproducible deployment. |
| D9 | `bin/main` tracked | Git inventory | Stale binary risk; repo bloat. |

## Coupling and dependency analysis

- `cmd/api` → `docs`, `internal/auth`, `consts`, `db`, `dto`, `env`, `response`, `storage`, `utils`.
- `cmd/seed` → `db`, `env`, `storage`.
- `internal/db` → `consts`, `storage` (seed coupling).
- `internal/storage` → `consts`, `lib/pq`, `bcrypt`.
- `internal/auth` → `golang-jwt/jwt/v5`.
- `internal/utils` → `validator/v10`.
- `docs` → `swaggo/swag`.

No import cycles. The highest internal fan-in is `internal/storage` and `internal/consts`; the highest fan-out
is `cmd/api`. Coupling is acceptable for a small monolith, but A2/A3/A4 blur layer boundaries.

## Scalability and evolution risks

1. **API contract tied to DB structs.** Adding/removing a column changes JSON, especially zero-value
   relational fields.
2. **Unbounded list endpoints.** `/v1/posts` and `/v1/users` return all rows; memory/latency grow linearly.
3. **N+1 comment loading.** Post list latency scales with post count.
4. **Feed SQL complexity.** Correctness and performance are both risky; no query tests.
5. **No pagination/filter consistency.** Only feed has pagination, and it is partly broken.
6. **Auth middleware bug.** Admin operations are unreliable.
7. **No transactions.** Multi-step operations cannot be made atomic.
8. **No tests/CI.** Refactoring is unsafe.
9. **No observability.** Production incidents are hard to diagnose.
10. **Swagger drift.** Clients cannot trust the contract.

## Prioritized recommendations

### P0 — correctness/security (do first)

1. Fix `UserPostAccessMiddleware`: `return` after `next`; use 403 for forbidden; define moderator policy.
2. Require a strong JWT secret; rename the env key; remove the fallback outside development.
3. Fix feed SQL: `LEFT JOIN`/`EXISTS`, `COUNT(DISTINCT c.id)`, whitelisted sort direction, correct offset
   default, enforce/justify path user access.
4. Decide post-delete semantics (cascade vs transactional delete vs 409) and implement it.
5. Map user/follow not-found and validation errors consistently; stop returning raw errors.
6. Fix Swagger generation (DTOs/annotations) and align status codes/paths.

### P1 — reliability/testability

7. Introduce a service layer (or at least typed application errors and response DTOs) without over-engineering.
8. Add graceful shutdown, DB-aware readiness, request IDs, and structured error logging.
9. Add tests: unit for auth/pagination/response; integration with PostgreSQL; HTTP tests for every route;
   migration up/down tests.
10. Add CI: build, vet, test -race, coverage, lint, Swagger diff.
11. Fix makefile seed/create/docs targets; remove tracked `bin/main`.
12. Add pagination to posts/users and batch comment loading.

### P2 — scale/operations

13. Add indexes based on measured query plans (`posts.created_at`, `followers(follower_id,user_id)`, FK
    indexes as needed); remove redundant/unused indexes.
14. Add metrics/tracing/log rotation and dashboards.
15. Add an application Dockerfile, migration job, TLS/reverse proxy, secret injection, and backups.
16. Consider soft delete/audit columns if product requirements demand them.
17. Add rate limiting/refresh-token strategy if the API is exposed beyond a trusted client.

## Verification limits

- No load/performance tests were run; scalability claims are structural, not benchmarked.
- No production topology exists to review.
- No automated tests exist; correctness findings are from code/SQL inspection plus limited runtime checks.
- Admin middleware behavior was not exercised end-to-end because no admin user existed.

## Related documents

- [Architecture](../en/02-architecture.md)
- [Database review](database-review.md)
- [API consistency review](api-consistency-review.md)
- [Security review](security-review.md)
