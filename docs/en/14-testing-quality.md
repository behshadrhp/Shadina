# 14. Testing and Quality

## Test inventory

There are **no Go test files** in the repository. A repository-wide search for `*_test.go` returned no
results. There are no mocks, fakes, integration tests, API tests, database tests, golden files, or test
fixtures.

| Test type | Present? | Evidence |
|---|---|---|
| Unit tests | No | No `*_test.go` files |
| Integration tests | No | No test harness/testcontainers |
| HTTP/API tests | No | No `httptest` usage |
| Database tests | No | No test DB setup |
| Migration tests | No | No migration test scripts |
| Load/performance tests | No | No benchmarks |
| Security tests | No | No fuzzing, auth tests, or scanners |
| CI | No | No `.github/`, `.gitlab-ci.yml`, etc. |

## Commands actually executed

All commands below were run on 2026-10-29 in the repository root with Go 1.25.4 and the local PostgreSQL
container available. Results are reported exactly.

| Command | Result | Interpretation |
|---|---|---|
| `go build ./...` | **PASS** (exit 0) | All packages with Go files compile. `cmd/migrate` is skipped because it has no Go files. |
| `go build ./cmd/migrate` | **FAIL** (exit 1) | `no Go files in .../cmd/migrate`; migrations rely on the external CLI. |
| `go vet ./...` | **PASS** (exit 0) | No vet diagnostics. |
| `go test ./...` | **PASS with no tests** | Every package reports `[no test files]`. This is not evidence of correctness. |
| `go test -race ./...` | **PASS with no tests** | No test binaries were built, so the race detector had nothing to execute. |
| `go test -cover ./...` | **0.0% coverage** | Packages with statements report `coverage: 0.0%`; some report `[no test files]`. |
| `go mod verify` | **PASS** | `all modules verified`. |
| `go mod tidy -diff` | **DIFF FOUND** | Shows unused direct dependency `github.com/golang-jwt/jwt v3.2.2+incompatible`; no files were modified. |
| `gofmt -l .` | **PASS** | No files reported as unformatted at review time. |
| `golangci-lint run ./...` | **FAIL** (exit 1) | 12 issues: 9 `errcheck`, 3 `staticcheck`. |
| `make test` | Would run `go test -v ./...` | Not separately needed; same no-test result. |
| `make lint` | **Not run** | It runs `gofmt -s -w .` first, which modifies files. `golangci-lint` was run directly instead. |
| `swag init -g ./api/main.go -d cmd,internal` | **FAIL** (exit 1) | Cannot find type definition `dto.FollowPayload`; Swagger generation is blocked. |
| `make gen-docs` | **FAIL** | Same root cause; the make target invokes the failing `swag init`. |
| `make seed` | **FAIL** | Target points to missing `cmd/migrate/seed/main.go`. |
| `go run ./cmd/seed` | **Not run** | Would insert non-idempotent fixtures into the development DB; not executed during this documentation-only review. |
| `govulncheck ./...` | **Not run** | `govulncheck` is not installed; no dependency vulnerability scan was performed. |
| Integration/API tests | **Not run** | No test suite exists. Manual read-only API checks were performed (below). |

## Lint findings

`golangci-lint run ./...` reported:

**errcheck (9)**

- `cmd/api/health.go`: unchecked `utils.WriteJson` return.
- `cmd/api/main.go`: unchecked `log.Sync`, `db.Close`.
- `cmd/seed/main.go`: unchecked `database.Close`.
- `internal/db/seed.go`: unchecked `tx.Commit`.
- `internal/storage/comments.go`: unchecked `row.Close`.
- `internal/storage/posts.go`: unchecked `row.Close` / `rows.Close` (two findings).
- `internal/storage/users.go`: unchecked `rows.Close`.

**staticcheck (3)**

- `cmd/api/middleware.go`: `SA1029` — built-in string as `context.WithValue` key.
- `internal/consts/consts.go`: `ST1012` — `NOT_FOUND` should be named `ErrFoo`.
- `internal/storage/users.go`: `S1008` — simplify `ComparePassword` to `return err == nil`.

The lint command exits 1, so it is not currently CI-clean. `make lint` would also reformat the tree before
linting.

## Manual runtime checks performed

These were read-only or rolled-back checks against a locally started API and the development database. They
are **not** a substitute for automated tests.

| Check | Observed result |
|---|---|
| `GET /v1/health` | 200 `{"status":"ok","version":"0.0.1","environment":"development"}` |
| `GET /ping` | 200, `text/plain`, body `.` |
| `GET /v1/swagger/index.html` | 200 HTML |
| `GET /v1/swagger/doc.json` | 200 JSON |
| `GET /v1/users` without token | 401 envelope |
| `GET /v1/users` with invalid token | 500 envelope with parser error |
| `POST /v1/auth/token` with seeded user | 400 code 4017 (empty password) |
| `GET /v1/users` with valid token | 200; `role_id` is 0 |
| `GET /v1/posts` with valid token | 200; 203 posts in dev DB |
| Feed `sort=asc` vs `sort=desc` | Same order |
| `GET /v1/users/2/feeds` with user 1 token | 200 (no ownership check) |
| `GET /v1/users/1/feed` | 404 (real path is `/feeds`) |
| `GET /v1/users/2/follow` | 405 (route is PUT) |
| `GET /v1/users/999999` | 500 code 5010 (`not found`) |
| `GET /v1/users/abc` | 500 code 5010 (raw `strconv` error) |
| `DELETE /v1/posts/{other-user-post}` as owner of another post | 401 code 4023 |
| Registration validation | 400 code 4012 with validator text |
| Registration unknown field | 400 code 4011 |
| Create post missing title | 500 code 5006 |
| Create post empty body | 500 code 5002 (`EOF`) |
| Delete post with comments (SQL transaction, rolled back) | FK violation `fk_post_comment`; row count unchanged |
| Feed count comparison | post 19: `COUNT(c.id)=16` vs `COUNT(DISTINCT c.id)=4` |

## What was not tested

- Authentication with a real admin user (none existed); the admin middleware defect was verified by code
  inspection only.
- Registration success end-to-end (would insert a user; not performed).
- Post create/update/delete success paths (would mutate data; not performed).
- Follow/unfollow success/error paths (would mutate data; not performed).
- Concurrent requests, race conditions, deadlocks, or pool exhaustion.
- Performance/load behavior of feed/list endpoints.
- Production configuration, TLS, CORS, reverse proxy, or container deployment.
- Dependency vulnerabilities (`govulncheck` unavailable).
- Swagger UI spec loading in a browser (config inspected; no headless browser run).
- Migration up/down on a disposable database (schema was already at version 12).

## Quality assessment

Build and vet pass, and the code compiles cleanly. That is the extent of the automated evidence. The absence
of tests means the following are unverified:

- authorization matrix (owner/admin/moderator/other)
- error status/code contracts
- pagination/sort/search behavior
- migration rollback safety
- seed idempotency/atomicity
- SQL correctness under data shapes (the feed count bug shows why this matters)
- JWT edge cases (expired, wrong audience, malformed `sub`, clock skew)
- JSON schema stability

## Recommended test plan

### Unit tests (fast, no DB)

- `internal/env`: present/absent/invalid values.
- `internal/auth`: token generation, valid token, expired token, wrong issuer/audience, wrong algorithm,
  malformed token.
- `internal/storage/pagination`: defaults, invalid values, sort/search/since/until.
- `internal/response`: status codes, envelope, encode behavior (replace `log.Fatalf` first).
- `storage.User` password hash/compare (including >72-byte input).
- `UserPostAccessMiddleware`: owner, admin, moderator, other, missing post (use a fake storage/auth).

### Integration tests (PostgreSQL)

- Run migrations against a disposable database (testcontainers or a dedicated schema).
- Seed a known fixture, not the production seed.
- Test each storage method, including FK violations and duplicate follows.
- Test feed with multiple followers/comments; assert distinct counts and sort order.
- Test post deletion with/without comments.

### HTTP/API tests

- `httptest.NewServer` with the real router and a test DB/fake storage.
- Cover every endpoint, auth requirement, status code, internal code, and JSON shape.
- Add contract tests comparing handler behavior to the generated OpenAPI document.

### CI

- `gofmt -l`, `go vet`, `go test -race`, `go test -cover`, `golangci-lint`.
- Migration up/down on a disposable DB.
- `swag init` and fail on diff.
- `govulncheck` and dependency updates.
- Coverage threshold after the first tests exist.

## Related documents

- [Development guide](13-development-guide.md)
- [API reference](07-api-reference.md)
- [Database operations](06-database-operations.md)
- [Security review](../reports/security-review.md)
