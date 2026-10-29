# Security Review

**Repository:** `github.com/behshadrhp/Shadina`  
**Commit:** `6c14149`  
**Review date:** 2026-10-29  
**Method:** static code/config review + limited read-only runtime checks  
**Not performed:** penetration test, fuzzing, DAST/SAST product scan, dependency CVE scan, production review

## Persian summary

مهم‌ترین یافته‌ها: (۱) کلید امضای JWT در صورت نبود متغیر محیطی، مقدار پیش‌فرض ضعیف و هاردکد دارد و نام متغیر
`AUTH_SECRET_KAY` غلط نوشته شده است؛ (۲) در `UserPostAccessMiddleware` مسیر ادمین پس از اجرای handler دوباره
پاسخ 401 می‌نویسد؛ (۳) خطاهای خام JWT/SQL/strconv به کلاینت برگردانده می‌شوند؛ (۴) توکن نامعتبر 500 برمی‌گرداند
نه 401؛ (۵) feed دسترسی کاربران دیگر را محدود نمی‌کند. هیچ تزریق SQL پیدا نشد و کوئری‌ها پارامتری هستند. این
بررسی، تست نفوذ یا اسکن آسیب‌پذیری وابستگی نیست.

## Secure controls observed

| Control | Evidence | Assessment |
|---|---|---|
| Password hashing | `bcrypt.GenerateFromPassword(..., bcrypt.DefaultCost)` | Good baseline; cost fixed. |
| Password not serialized | `Password string json:"-"` | Good. |
| SQL parameterization | `$1` placeholders throughout; `pq.Array` for arrays | No SQL injection found. |
| JWT algorithm restriction | `WithValidMethods(HS256)` + HMAC key check | Good. |
| JWT issuer/audience/expiry | `WithIssuer`, `WithAudience`, `WithExpirationRequired` | Good; `nbf` validated by jwt/v5. |
| Request body limit | `http.MaxBytesReader` 1 MiB | Good. |
| Unknown fields rejected | `DisallowUnknownFields` | Good. |
| HTTP timeouts | read 10s, write 30s, idle 1m | Partial; no ReadHeaderTimeout. |
| DB pool limits | 30 open/30 idle/15m idle | Good baseline. |
| Error envelope | Uniform JSON for handler errors | Good, but leaks messages. |
| No auth header logging | Custom logger logs method/path/status/bytes/duration | Good; no token leakage in logs. |

## Findings

### SEC-01 — JWT secret fallback and misspelled environment key

**Severity:** High  
**Status:** Verified by code and configuration review  
**Evidence:** `cmd/api/main.go`:
`secretKey: env.GetString("AUTH_SECRET_KAY", <hard-coded fallback>)`; `.env` and `.example.env` do not define
`AUTH_SECRET_KAY`/`AUTH_SECRET_KEY`; no startup validation.  
**Condition:** Deployment does not set the (misspelled) variable, or sets `AUTH_SECRET_KEY` expecting it to
work. The service starts with the fallback.  
**Impact:** Anyone with source access can sign tokens with the fallback and impersonate arbitrary `sub`
values, including admins (subject to user existence/role in DB).  
**Remediation:** Rename to `AUTH_SECRET_KEY`; require a strong value at startup; remove the fallback outside
explicit development mode; rotate the secret; invalidate old tokens.  
**Verification:** Code path confirmed; no token forgery was attempted.

### SEC-02 — Admin post-access middleware writes after `next`

**Severity:** High  
**Status:** Verified by code inspection; not exercised with an admin user  
**Evidence:** `cmd/api/middleware.go`, `UserPostAccessMiddleware`:

```go
if user.ID != post.UserID {
    if user.RoleID == 3 {
        next.ServeHTTP(w, r)
    }
    response.UnauthorizedResponse(w, "access denied", 4023)
    return
}
```

**Condition:** A user with `role_id=3` updates/deletes another user's post.  
**Impact:** The handler executes (operation may succeed), then the middleware attempts a 401 after the
response has started. The first status wins; the body may contain concatenated JSON; clients see a
misleading result. This is an authorization/response-integrity defect.  
**Remediation:** `next.ServeHTTP(w, r); return` inside the admin branch; return 403 for authenticated-but-
forbidden; define moderator behavior; add tests.  
**Verification:** No admin user existed in the development DB; not reproduced end-to-end.

### SEC-03 — Raw internal errors returned to clients

**Severity:** Medium  
**Status:** Verified locally  
**Evidence:** `response.InternalErrorResponse(w, err.Error(), ...)` in handlers and middleware; examples from
runtime:
`token is malformed: token contains an invalid number of segments` (500/5019),
`strconv.ParseInt: parsing "abc": invalid syntax` (500/5010),
`not found` (500/5010), `EOF` (500/5002), validator field names (400/4012).  
**Condition:** Invalid input, invalid token, DB errors, validation failures.  
**Impact:** Leaks parser/DB/schema details and implementation behavior; aids attackers and confuses clients.  
**Remediation:** Typed errors, central mapper, generic client messages + request ID, internal logging.  
**Verification:** Reproduced with malformed token/non-numeric ID/empty body.

### SEC-04 — Invalid or expired JWT returns HTTP 500

**Severity:** Medium  
**Status:** Verified locally  
**Evidence:** `JWTAuthMiddleware` calls `response.InternalErrorResponse` for `ValidateToken` errors (code
5019).  
**Impact:** Authentication failures are misclassified as server errors; clients may retry; monitoring noise;
parser details leaked.  
**Remediation:** Return 401 with a generic `invalid or expired token` message; keep details in logs.  
**Verification:** `Authorization: Bearer not-a-token` → 500/5019.

### SEC-05 — Feed endpoint does not enforce path-user ownership

**Severity:** Medium (product-dependent)  
**Status:** Verified locally  
**Evidence:** `listFeedsHandler` parses `{userID}` and calls `GetFeed` without comparing to
`GetUserCTX(r).ID`. Any valid token can request `/v1/users/{otherID}/feeds`; tested with user 1 token and
user 2 path → 200.  
**Impact:** Potential horizontal access to another user's feed. If feeds are intended public, this is not a
vulnerability but must be documented; if private, it is an IDOR.  
**Remediation:** Compare path ID to token user, or explicitly document public feeds and remove auth if
appropriate. Add authorization tests.  
**Verification:** Runtime 200 for another user's feed.

### SEC-06 — No token revocation/refresh/logout

**Severity:** Medium  
**Status:** Design gap  
**Evidence:** No refresh endpoint, no `jti`, no session/blacklist table, no logout route; 72-hour expiry.  
**Impact:** Stolen tokens remain valid until expiry; no way to revoke on password change/compromise.  
**Remediation:** Add refresh rotation/revocation, short-lived access tokens, `jti`/session store, logout.  
**Verification:** Code/routes reviewed.

### SEC-07 — No rate limiting or login lockout

**Severity:** Medium  
**Status:** Design gap  
**Evidence:** No rate-limit middleware; `/v1/auth/token` and `/v1/auth/register` are unauthenticated.  
**Impact:** Online password guessing, credential stuffing, registration abuse, resource exhaustion.  
**Remediation:** Per-IP/account rate limits, exponential backoff/lockout, CAPTCHA if public, monitoring.  
**Verification:** No load/abuse testing performed.

### SEC-08 — Registration validation and duplicate-error exposure

**Severity:** Low  
**Status:** Verified by code; duplicate path not exercised  
**Evidence:** `UserRegisterPayload.Email` has `required,max=255`, no `email` validator; duplicate
username/email causes a DB unique violation returned as 500 with `err.Error()`.  
**Impact:** Invalid emails accepted; duplicate registration reveals DB details; user enumeration possible.  
**Remediation:** Add `email` validation, normalize, pre-check or map unique violations to 409/400 with a
generic message.  
**Verification:** Validation/unknown-field behavior tested; duplicate insert not performed.

### SEC-09 — User enumeration signals

**Severity:** Low  
**Status:** Verified by code  
**Evidence:** Unknown email → 400 code 4015 message `invalid request`; wrong password → 400 code 4017
`invalid email or password`; DB error → 4016. Different codes/messages.  
**Impact:** An attacker can distinguish existing emails from non-existing ones (subject to timing/rate
limits).  
**Remediation:** Return a single generic credential error/code for all login failures; equalize timing where
feasible.  
**Verification:** Code review; no timing measurement.

### SEC-10 — Missing CORS/TLS/security headers/request policy

**Severity:** Low  
**Status:** Verified by code/config  
**Evidence:** No CORS middleware, no TLS config, no security headers, no request ID; app intended behind a
proxy but no proxy config in repo.  
**Impact:** Browser cross-origin clients fail; production security depends entirely on external
infrastructure; no defense-in-depth.  
**Remediation:** Add CORS allowlist, TLS/reverse proxy, HSTS/CSP/X-Content-Type-Options, request IDs, header
limits.  
**Verification:** Deployment review; no production environment.

### SEC-11 — String context key / literal key

**Severity:** Low  
**Status:** Verified by staticcheck  
**Evidence:** `context.WithValue(ctx, consts.USER_CTX, user)` where `USER_CTX = "user"`;
`GetUserCTX` uses literal `"user"`. `staticcheck SA1029`.  
**Impact:** Key collision with other packages; fragility; not directly exploitable in current code.  
**Remediation:** Use a private typed key and a single accessor.  
**Verification:** Lint/code review.

### SEC-12 — `log.Fatalf` on response encoding failure

**Severity:** Low  
**Status:** Verified by code  
**Evidence:** `response.Respond` calls `log.Fatalf` if `json.NewEncoder(w).Encode` fails.  
**Impact:** A response encoding error terminates the process (availability). Current payloads are simple, so
likelihood is low.  
**Remediation:** Log and return a 500; do not exit from a request path.  
**Verification:** Code review.

### SEC-13 — Incomplete slowloris hardening

**Severity:** Low  
**Status:** Verified by code  
**Evidence:** `http.Server` sets `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, but no `ReadHeaderTimeout` or
`MaxHeaderBytes`.  
**Impact:** Header-slow attacks may consume connections; mitigated by ReadTimeout to some degree.  
**Remediation:** Set `ReadHeaderTimeout` and `MaxHeaderBytes`; deploy behind a hardened proxy.  
**Verification:** Code review.

### SEC-14 — No SQL injection found

**Severity:** Informational  
**Status:** Verified by review  
**Evidence:** All storage queries use `$n` parameters; `pq.Array` for tags; no string concatenation of user
input into SQL. The feed `ORDER BY $2` is a parameter (safe, but ineffective).  
**Impact:** None found in reviewed paths.  
**Remediation:** Keep parameterization; add query tests; never concatenate sort/column identifiers.  
**Verification:** Static review; no fuzzing.

### SEC-15 — Seeded users have empty passwords

**Severity:** Informational (data quality)  
**Status:** Verified in DB/API  
**Evidence:** 100 rows with `password = ''::bytea`; login with seeded user returns 400/4017.  
**Impact:** Fixtures cannot authenticate; may cause misleading test conclusions. Not a direct auth bypass
because bcrypt comparison fails.  
**Remediation:** Hash generated passwords or mark fixtures non-loginable; use register for test accounts.  
**Verification:** DB count + runtime login.

### SEC-16 — Public Swagger UI with misconfigured spec URL

**Severity:** Informational  
**Status:** Verified by UI config  
**Evidence:** `/v1/swagger/*` requires no auth; UI `url` is `SERVER_URL + "/swagger/doc.json"` (missing
`/v1` and scheme); doc endpoint itself is `/v1/swagger/doc.json`.  
**Impact:** Documentation exposure (often acceptable); UI may fail to load the spec. Not a direct
vulnerability.  
**Remediation:** Restrict docs in production if desired; fix URL construction; add security definitions.  
**Verification:** HTML inspected; no browser run.

## Severity rationale

- **High:** Can lead to impersonation/authorization bypass or materially corrupt authorization responses.
- **Medium:** Information disclosure, auth contract confusion, missing abuse controls, or horizontal access
  depending on product intent.
- **Low:** Defense-in-depth/quality issues with limited direct exploitability.
- **Informational:** Observations, secure controls, or data-quality notes.

## Conditions and non-findings

- No evidence of SQL injection, command injection, path traversal, SSRF, or deserialization RCE in the
  reviewed code.
- No file upload/HTML rendering exists; XSS surface is limited to JSON clients.
- No cookies/session auth; classic CSRF is not applicable to the current bearer-token API.
- No external service integrations to review.
- Dependency vulnerabilities were **not** scanned (`govulncheck` unavailable); do not infer absence of CVEs.
- No production secrets or infrastructure were available; SEC-01 impact depends on deployment configuration.

## Priority remediation plan

1. SEC-01: strong required secret, rename key, remove fallback.
2. SEC-02: fix admin middleware; add authorization tests.
3. SEC-03/04: typed errors, 401 for invalid tokens, generic client messages.
4. SEC-05: enforce or document feed access.
5. SEC-06/07: rate limiting, refresh/revocation strategy.
6. SEC-08/09: email validation and uniform login errors.
7. SEC-10/13: proxy/TLS/CORS/header/timeout hardening.
8. Add SAST, `govulncheck`, dependency updates, and a pre-production security checklist.

## Related documents

- [Authentication and authorization](../en/08-authentication-authorization.md)
- [Configuration](../en/09-configuration.md)
- [Error handling](../en/10-error-handling.md)
- [Security summary](../en/15-security-review.md)
