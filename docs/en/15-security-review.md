# 15. Security Review

This page summarizes the security-relevant behavior observed in the code. The detailed findings, evidence,
and remediation are in [Security review report](../reports/security-review.md).

## Scope and method

- Static review of all Go, SQL, configuration, and Swagger files.
- Read-only runtime checks against a locally started API and the development PostgreSQL database.
- No penetration testing, fuzzing, dependency vulnerability scan (`govulncheck` unavailable), production
  configuration review, or load testing.
- Findings are limited to what the repository and local runtime demonstrate. Absence of a finding is not a
  claim that the system is secure.

## Secure baseline observed

| Control | Implementation |
|---|---|
| Password storage | bcrypt with `bcrypt.DefaultCost`; hashes stored in `bytea`; JSON omits password |
| SQL injection resistance | All queries use `$1`-style parameters; `pq.Array` for arrays; no user-controlled SQL concatenation |
| JWT algorithm restriction | `WithValidMethods(HS256)` plus HMAC key-function check |
| JWT issuer/audience/expiry | `WithIssuer`, `WithAudience`, `WithExpirationRequired`; `nbf` validated by jwt/v5 |
| Request body limit | 1 MiB `MaxBytesReader` |
| Unknown JSON fields | Rejected |
| HTTP server timeouts | Read 10s, write 30s, idle 1m |
| DB pool limits | Max open/idle 30, idle time 15m |
| DB parameterization | Yes, including feed search |

These are solid baselines, but several authorization, configuration, and error-handling issues remain.

## Findings summary

| ID | Severity | Finding | Status |
|---|---|---|---|
| SEC-01 | High | JWT secret has a hard-coded fallback and no startup validation; env key is misspelled `AUTH_SECRET_KAY` | Verified by code/config |
| SEC-02 | High | Admin post-access branch calls `next` and then writes 401, corrupting responses and potentially allowing execution with a misleading result | Verified by code path; not exercised with an admin user |
| SEC-03 | Medium | Raw internal errors (JWT parser, SQL, strconv) are returned to clients | Verified locally |
| SEC-04 | Medium | Invalid/expired JWT returns HTTP 500 instead of 401 | Verified locally |
| SEC-05 | Medium | Feed endpoint does not check that path `userID` matches the authenticated user | Verified locally |
| SEC-06 | Medium | No token revocation, refresh, or logout; 72-hour bearer tokens | Design gap |
| SEC-07 | Medium | No rate limiting or login lockout on register/token | Design gap |
| SEC-08 | Low | Registration lacks email format validation and exposes duplicate/DB errors | Verified by code; duplicate path not exercised |
| SEC-09 | Low | User enumeration signals differ for unknown email vs wrong password | Verified by code |
| SEC-10 | Low | No CORS, TLS, security headers, or request-ID policy in the app | Verified by code/config |
| SEC-11 | Low | `context.WithValue` uses a built-in string key; literal `"user"` in `GetUserCTX` | Verified by staticcheck/code |
| SEC-12 | Low | `response.Respond` calls `log.Fatalf` on JSON encode failure | Verified by code |
| SEC-13 | Low | `ReadHeaderTimeout`/`MaxHeaderBytes` are not set; slowloris hardening incomplete | Verified by code |
| SEC-14 | Info | No SQL injection found; parameterized queries used | Verified by review |
| SEC-15 | Info | Seed users have empty password hashes and cannot log in | Verified in DB/API |
| SEC-16 | Info | Swagger UI is public and its spec URL is misconfigured | Verified by UI config |

## Highest-priority remediation

1. **Require a strong JWT secret at startup.** Rename `AUTH_SECRET_KAY` to `AUTH_SECRET_KEY` (with backward
   compatibility or a migration note), remove the fallback for non-development environments, and fail fast
   if the secret is missing or weak.
2. **Fix `UserPostAccessMiddleware`.** `return` after `next.ServeHTTP`; return 403 for authenticated-but-
   forbidden actions; define moderator/admin permissions explicitly.
3. **Stop returning raw errors.** Introduce typed errors and a single mapper; log internals with a request ID
   and return generic client messages.
4. **Return 401 for invalid tokens.** Do not expose parser details.
5. **Decide and enforce feed access.** If feeds are private, compare path user ID to the token user; if
   public, document it and consider removing JWT from that route.

## Not implemented security features

- Refresh token rotation, revocation, logout, `jti`, device/session management.
- Password reset/change, email verification, MFA.
- Rate limiting, account lockout, CAPTCHA.
- Role/permission management beyond numeric `role_id`.
- Audit logging of authn/authz decisions.
- CORS policy, TLS termination, security headers (CSP/HSTS/X-Content-Type-Options), CSRF strategy for
  cookie-based clients (no cookie auth exists).
- Secrets manager/KMS integration.
- Dependency vulnerability scanning, SAST/DAST, SBOM.
- Input sanitization beyond JSON/validator (output encoding is JSON; no HTML rendering).
- File upload/media handling (not implemented).

## Verification limits

- No admin user existed, so SEC-02 was not reproduced end-to-end.
- No expired-token test was run; behavior follows the same `ValidateToken` error branch as malformed tokens.
- No external attacker model, network boundary, or production secrets were available.
- `govulncheck` was not installed; no claim is made about dependency CVEs.
- JWT secret fallback risk depends on deployment configuration; if a strong env secret is always set, the
  fallback is not used. The code does not enforce that condition.

## Related documents

- [Authentication and authorization](08-authentication-authorization.md)
- [Configuration](09-configuration.md)
- [Error handling](10-error-handling.md)
- [Security review report](../reports/security-review.md)
