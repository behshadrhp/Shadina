# 17. Glossary

Terms are listed in English with the Persian equivalent used in the Persian documentation. Names of packages,
types, functions, and environment variables remain in their original form.

| Term | فارسی | Definition |
|---|---|---|
| API | رابط برنامه‌نویسی کاربردی | The HTTP/JSON interface exposed by `cmd/api`. |
| Handler | هندلر | An `http.HandlerFunc` in `cmd/api` that implements an endpoint. |
| Middleware | میان‌افزار | A wrapper around handlers; Shadina uses request logging, JWT auth, and post ownership checks. |
| Router | مسیریاب | chi router in `cmd/api/router.go` mapping methods/paths to handlers. |
| DTO | شیء انتقال داده | Request payload structs in `internal/dto`. Shadina does not use DTOs for responses. |
| Storage | لایه ذخیره‌سازی | `internal/storage`; SQL methods and persistence structs. |
| Repository | مخزن | A persistence abstraction. Shadina has a `Storage` facade, not per-entity repositories. |
| Service layer | لایه سرویس | A business-logic layer. **Not present**; handlers call storage directly. |
| Dependency Injection | تزریق وابستگی | Manual construction/wiring in `cmd/api/main.go`. |
| Monolith | یکپارچه | A single deployable API process containing all backend logic. |
| Layered architecture | معماری لایه‌ای | Transport → handlers → storage → DB, with supporting packages. |
| Migration | مهاجرت پایگاه داده | Versioned SQL changes in `cmd/migrate/migrations`, applied by golang-migrate. |
| Seed | داده اولیه | Development fixture generation in `internal/db/seed.go` / `cmd/seed`. |
| Schema | طرحواره | PostgreSQL table/column/constraint definitions. |
| ERD | نمودار موجودیت-رابطه | Entity Relationship Diagram; see `diagrams/database-erd.mmd`. |
| Primary Key (PK) | کلید اصلی | Unique row identifier; e.g. `users.id`, composite `followers(user_id, follower_id)`. |
| Foreign Key (FK) | کلید خارجی | Referential constraint between tables. |
| Unique Constraint | محدودیت یکتایی | Prevents duplicate values; e.g. `users.email`. |
| Index | ایندکس | B-tree/GIN structure to speed lookups; see `000007_add_indexes`. |
| GIN | جین | PostgreSQL inverted index, used for trigram/array indexes. |
| Trigram (`pg_trgm`) | تری‌گرام | Extension supporting `ILIKE '%...%'` indexes. |
| `citext` | سایتی‌اکست | Case-insensitive text type used for email. |
| Transaction | تراکنش | Atomic unit of DB work. Seed attempts one but does not use it correctly; request handlers have none. |
| Isolation level | سطح جداسازی | PostgreSQL transaction isolation; not explicitly configured. |
| Connection pool | استخر اتصال | `database/sql` pool configured with max open/idle and idle time. |
| N+1 query | کوئری N+1 | One query per parent row; `GET /v1/posts` loads comments per post. |
| Full table scan | اسکن کامل جدول | Reading all rows; possible for content search/feed ordering without indexes. |
| Soft delete | حذف نرم | Marking rows deleted instead of removing them. **Not implemented**. |
| JWT | توکن وب JSON | Signed token used for authentication. |
| Claim | ادعا | JWT field such as `sub`, `exp`, `iss`, `aud`. |
| `sub` | شناسه موضوع | Subject claim; Shadina stores the user ID. |
| Issuer (`iss`) | صادرکننده | Token issuer; `Shadina`. |
| Audience (`aud`) | مخاطب | Intended token audience; `Shadina`. |
| HS256 | اچ‌اس-۲۵۶ | HMAC-SHA256 JWT signing algorithm. |
| Bearer token | توکن حامل | `Authorization: Bearer <jwt>` scheme. |
| Refresh token | توکن تازه‌سازی | Long-lived token to obtain new access tokens. **Not implemented**. |
| Revocation | ابطال | Invalidating issued tokens. **Not implemented**. |
| bcrypt | بی‌کریپت | Password hashing function; `DefaultCost` used. |
| Authentication | احراز هویت | Proving identity (login/token validation). |
| Authorization | مجوزدهی | Deciding what an authenticated user may do (roles/ownership). |
| Role | نقش | `user`, `moderator`, `admin` seeded in `roles`. |
| Ownership check | بررسی مالکیت | `UserPostAccessMiddleware` compares `user.ID` with `post.UserID`. |
| 401 Unauthorized | ۴۰۱ | Missing/invalid authentication (or, incorrectly, some malformed tokens). |
| 403 Forbidden | ۴۰۳ | Authenticated but not allowed; Shadina uses 401 instead. |
| 404 Not Found | ۴۰۴ | Resource absent; inconsistently used for users. |
| 500 Internal Server Error | ۵۰۰ | Server/storage/JWT errors; sometimes used for client input errors. |
| Error envelope | پوشش خطا | `{success, message, internal_code, data}` response. |
| Internal code | کد داخلی | Application-specific code (e.g. 4017), not an HTTP status. |
| Correlation ID | شناسه همبستگی | Request identifier for tracing. **Not implemented**. |
| Request ID | شناسه درخواست | Same concept; not implemented. |
| Observability | قابلیت مشاهده | Logs, metrics, traces, health. Only partial logging exists. |
| Metrics | معیارها | Numeric telemetry (latency, errors). **Not implemented**. |
| Tracing | رهگیری | Distributed request tracing. **Not implemented**. |
| Liveness | زنده‌بودن | Is the process running? `/v1/health` is static. |
| Readiness | آمادگی | Can the process serve dependencies? Not implemented. |
| Heartbeat | ضربان | `/ping` middleware returning `.`. |
| Swagger / OpenAPI | سواگر / اوپن‌ای‌پی‌آی | API contract and UI. Swagger 2.0 files are stale. |
| CORS | سی‌اوآر‌اس | Cross-Origin Resource Sharing; not configured. |
| TLS | تی‌ال‌اس | Transport encryption; not configured in the app. |
| Graceful shutdown | خاموشی تدریجی | Draining in-flight requests; not implemented. |
| Rate limiting | محدودیت نرخ | Throttling requests; not implemented. |
| IDOR | دسترسی مستقیم به شناسه شیء | Accessing another user's object by ID; relevant to feed access. |
| SQL injection | تزریق SQL | Injecting SQL via input; queries are parameterized. |
| Parameterized query | کوئری پارامتری | `$1` placeholders; used throughout. |
| Prepared statement | دستور آماده | Reusable parameterized statement; `database/sql` handles execution. |
| `database/sql` | پکیج دیتابیس | Go standard DB abstraction. |
| `lib/pq` | کتابخانه پی‌کیو | PostgreSQL driver. |
| `sql.ErrNoRows` | خطای بدون ردیف | No row found; mapped to `consts.NOT_FOUND` in some stores. |
| Sentinel error | خطای نشانه | Package-level error value such as `consts.NOT_FOUND`. |
| Rollback | بازگردانی | Reverting a migration/transaction; see destructive warnings. |
| Dirty migration | مهاجرت ناسالم | golang-migrate state after a failed migration. |
| Idempotent | خودتکرارپذیر | Safe to run repeatedly; seed is **not** idempotent. |
| `.env` | فایل محیطی | Local environment file loaded by godotenv. |
| Environment variable | متغیر محیطی | Process-level configuration. |
| Secret | راز | Sensitive value such as JWT secret/DB password. Do not commit. |
| DSN | رشته اتصال داده | Database connection string. |
| Connection string | رشته اتصال | Same as DSN. |
| Development environment | محیط توسعه | Local setup described in the development guide. |
| Production environment | محیط عملیاتی | Real deployment; not provided by the repository. |
| Docker Compose | داکر کامپوز | Local container orchestration; only PostgreSQL here. |
| Volume | والیوم | Persistent Docker storage for PostgreSQL. |
| Container | کانتینر | Isolated runtime unit. |
| Migration CLI | ابزار مهاجرت | External `migrate` binary. |
| Swag | سواگ | Swagger generator; generation currently fails. |

## Related documents

- [Project overview](01-project-overview.md)
- [Architecture](02-architecture.md)
- [API reference](07-api-reference.md)
- [Database design](05-database-design.md)
- [Security review](15-security-review.md)
