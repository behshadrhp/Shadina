# 04. Application Flow

## Startup flow

`cmd/api/main.go` is the only long-running entry point. The sequence is:

```text
init()
  └── utils.Validate = validator.New(WithRequiredStructEnabled)

main()
  ├── os.MkdirAll("logs", 0755)
  ├── zap.NewProductionConfig()
  │     ├── OutputPaths       = logs/app.log
  │     └── ErrorOutputPaths  = logs/error.log
  ├── zap.Must(...).Sugar()
  ├── godotenv.Load()                    # .env in current working directory
  ├── env.GetString/GetInt(...)          # config with fallbacks
  ├── db.NewDB(driver, addr, pool...)    # sql.Open + pool + PingContext(10s)
  ├── storage.NewPostgresql(db)
  ├── auth.NewJWTAuthenticator(secret, "Shadina", "Shadina")
  ├── application{config, storage, logger, authenticator}
  ├── app.mount()                        # chi router
  └── app.run(&mux)
        ├── docs.SwaggerInfo.Version/Host/BasePath
        ├── http.Server{ReadTimeout, WriteTimeout, IdleTimeout}
        └── srv.ListenAndServe()
```

Important details:

- `godotenv.Load()` does **not** override already-exported environment variables. Values already present in
  the process environment win.
- If `.env` is missing or unreadable, startup exits with `error loading .env file`; there is no fallback to
  an `.env.example` or a remote secret store.
- `db.NewDB` pings with `consts.CTX_TIMEOUT` (10 seconds). Invalid `MAX_IDLE_TIMEOUT` causes startup failure.
- `ListenAndServe` is called directly. There is no `signal.Notify`, `Server.Shutdown`, or drain logic.
- `log.Fatal(app.run(&mux))` exits the process on any server error; deferred `log.Sync()` and `db.Close()` do
  not run.

## Router and middleware order

`cmd/api/router.go` registers middleware in this exact order:

| Order | Middleware | Scope | Behavior |
|---:|---|---|---|
| 1 | `middleware.Heartbeat("/ping")` | Global | Returns `200 text/plain` with body `.` for GET/HEAD `/ping`; bypasses later handlers. |
| 2 | `middleware.RealIP` | Global | Rewrites `RemoteAddr` from forwarding headers. |
| 3 | `middleware.Logger` | Global | Chi's common-log-format logger to stdout. |
| 4 | `middleware.Recoverer` | Global | Recovers panics and returns 500; logs stack. |
| 5 | `app.LoggerMiddleware` | Global | WrapResponseWriter + zap request log to `logs/app.log`. |
| 6 | `app.JWTAuthMiddleware` | `/v1/users/*`, `/v1/posts/*` | Parses bearer token, loads user, stores `*storage.User` in request context. |
| 7 | `app.UserPostAccessMiddleware` | PATCH/DELETE `/v1/posts/{postID}` | Loads post, checks owner/admin, then calls next. |

The custom logger runs after `Recoverer` in registration order, but because middleware wraps the downstream
handler, `LoggerMiddleware`'s post-handler code does not run when a downstream panic unwinds; recovery is
handled by `Recoverer` outside it. This is a subtle logging/diagnostic gap, not a functional API contract.

## Route tree

```text
/ping                                  GET        (Heartbeat middleware)
/v1
├── /swagger/*                         GET        (Swagger UI + doc.json)
├── /auth
│   ├── /register                      POST       public
│   └── /token                         POST       public
├── /users                             GET        JWT
│   └── /{userID}
│       ├── /                          GET        JWT
│       ├── /follow                    PUT        JWT
│       ├── /unfollow                  PUT        JWT
│       └── /feeds                     GET        JWT
├── /posts                             GET        JWT
│   ├── /                              POST       JWT
│   └── /{postID}
│       ├── /                          GET        JWT
│       ├── /                          DELETE     JWT + owner/admin
│       └── /                          PATCH      JWT + owner/admin
└── /health                            GET        public
```

Chi accepts both `/v1/users` and `/v1/users/` for the user collection in local testing; the router source
declares `r.Get("/")` inside `r.Route("/users", ...)`. The same was observed for `/v1/posts`. The canonical
documented paths omit the trailing slash.

## Request lifecycle by endpoint class

### 1. Public registration

```text
POST /v1/auth/register
  → ReadJson (1 MiB max, DisallowUnknownFields)
  → Validate.Struct(UserRegisterPayload)
  → bcrypt.GenerateFromPassword(password, DefaultCost)
  → UsersStorage.Create (INSERT ... RETURNING id, created_at, updated_at)
  → SuccessResponse 200 {success, message, internal_code:1010, data: User}
```

`User.Password` has `json:"-"`, so the hash is not serialized. `RoleID` is hard-coded to `1`.
Read errors return `400` code `4011`; validation returns `400` code `4012`; hashing/insert errors return
`500` code `5018` with the raw error text for insert failures.

### 2. Public token issuance

```text
POST /v1/auth/token
  → ReadJson + Validate
  → UsersStorage.GetByEmail (SELECT ... WHERE email=$1)
  → User.ComparePassword (bcrypt)
  → claims: sub=user.ID, exp=now+72h, iat, nbf, iss="Shadina", aud="Shadina"
  → JWTAuthenticator.GenerateToken (HS256)
  → SuccessResponse 200 {data: "<token string>"}
```

Failures are mostly mapped to `400` with codes `4013`–`4018`; database errors are returned as generic
`"invalid request"` with `400` (not 500). See [07. API reference](07-api-reference.md).

### 3. Protected request with JWT

```text
GET /v1/users
  → JWTAuthMiddleware
      → require Authorization header
      → require exactly "Bearer <token>" (single space, case-sensitive "Bearer")
      → ValidateToken: HS256, issuer, audience, exp required
      → claims["sub"] formatted with "%.f" and parsed as int64
      → UsersStorage.GetByID
      → context.WithValue(ctx, consts.USER_CTX, user)
  → handler
```

Authentication failures:

| Condition | Status | Internal code |
|---|---|---|
| Header missing | 401 | 4019 |
| Header malformed / not `Bearer ` | 401 | 4020 |
| Token invalid, expired, wrong issuer/audience/signature | **500** | 5019 |
| `sub` cannot be parsed | 401 | 4021 |
| User lookup fails | 401 | 4022 |

The 500-on-invalid-token behavior is implemented; it is not a documentation mistake.

### 4. Follow / unfollow

```text
PUT /v1/users/{userID}/follow
  → JWT middleware
  → parse path userID (int64)
  → FollowersStorage.Follow(ctx, authenticatedUser.ID, targetID)
  → INSERT INTO followers(user_id, follower_id) VALUES ($1,$2)
  → SuccessResponse 200, data null
```

No request body is read, despite Swagger declaring a required body. The table's `user_id` column is the
acting follower and `follower_id` is the followed user. The names are counterintuitive; see
[05. Database design](05-database-design.md#followers-and-direction).

`unfollow` deletes by the same pair and does not check `RowsAffected`, so a missing relationship still
returns 200.

### 5. Feed

```text
GET /v1/users/{userID}/feeds?limit=&offset=&sort=&search=&since=&until=
  → JWT middleware (does not check path userID == authenticated user)
  → parse path userID
  → CreatePaginatedQuery(20, 10, "desc")
  → Parse query parameters (HTTP concerns live in internal/storage/pagination.go)
  → Validate.Struct(pq)
  → PostsStorage.GetFeed(ctx, pathUserID, pq)
  → SuccessResponse 200, data: []PostWithMetadata
```

Actual defaults are `limit=20`, `offset=10`, `sort=desc`. The `page` query parameter in Swagger is not read.
Invalid numeric `limit`/`offset` are silently ignored (`Parse` returns no error); invalid `sort` is caught
by validation. `since`/`until` are parsed/stored but not used in SQL. `sort` is passed as `$2` in
`ORDER BY p.created_at, $2`; PostgreSQL treats it as a constant, so ascending/descending return the same
order in practice (verified locally).

### 6. Post list and retrieve

```text
GET /v1/posts
  → JWT
  → PostsStorage.List (SELECT all posts, no ORDER BY, no pagination)
  → for each post: CommentsStorage.GetByPostID
  → SuccessResponse 200 data: []*Post
```

This is a classic N+1 query pattern: one query for all posts plus one comment query per post (203 posts in
the development DB → up to 204 queries). `Post.User` is not loaded by `List` or `GetByID`; responses contain
a zero-value `user` object. `Comments` are loaded and embedded.

### 7. Post create/update/delete

- **Create:** JWT user becomes `post.UserID`; `CreatePostPayload` is validated; insert returns ID and
  timestamps; success is HTTP **200** (Swagger says 201).
- **Update:** JWT + `UserPostAccessMiddleware`; fetch post, require owner or `RoleID == 3`, read/validate
  optional title/content, mutate, `UPDATE posts SET title=$1, content=$2 WHERE id=$3`. `updated_at` is not
  changed. The admin branch has a response bug (see below).
- **Delete:** JWT + owner/admin middleware; `DELETE FROM posts WHERE id=$1`. Comments reference posts with a
  foreign key that has no `ON DELETE CASCADE`, so deleting a post that has comments fails at the database and
  is returned as HTTP 500. A transaction-rollback test against the local DB confirmed the FK violation.

### 8. Health and Swagger

- `GET /v1/health` builds `dto.HealthStatusPayload{status:"ok", version, environment}` and writes it
  directly with `utils.WriteJson`. It does not check PostgreSQL.
- `GET /ping` is intercepted by chi Heartbeat and returns `.`.
- `GET /v1/swagger/*` is served by `http-swagger/v2`. `index.html` is 200, and
  `/v1/swagger/doc.json` is served by the handler. The configured UI spec URL is
  `SERVER_URL + "/swagger/doc.json"` (for example `127.0.0.1:8000/swagger/doc.json`), which lacks the
  `/v1` prefix and a URL scheme; the correct handler path is `/v1/swagger/doc.json`. The UI is served, but
  loading the spec from the configured URL is expected to fail. This was confirmed by reading the generated
  UI config, not by a headless browser run.

## Data transformations

| Stage | Shape |
|---|---|
| HTTP JSON | Request DTO (`dto.*Payload`) |
| Handler | Domain/storage struct (`storage.User`, `storage.Post`, `storage.Comment`) |
| Storage | SQL parameters; `pq.Array` for tags |
| Database rows | Scanned into storage structs |
| Response | `response.Response{Success, Message, InternalCode, Data}` |
| JSON | `data` is a struct, array, string, or `null` |

There is no response mapper. JSON field names come directly from struct tags on storage types. This is why
zero-value relational fields (`user`, `comments`) appear in responses.

## Context and cancellation

- Every handler uses `r.Context()`; `database/sql` calls receive it.
- `FollowersStorage`, `CommentsStorage.Create`, and `PostsStorage.GetFeed` additionally wrap the context
  with `context.WithTimeout(ctx, consts.CTX_TIMEOUT)` (10 seconds).
- `UsersStorage`, `PostsStorage.Create/List/GetByID/Delete/Update`, and `CommentsStorage.GetByPostID` use the
  request context without an explicit DB timeout. They still inherit client cancellation and server
  timeouts.
- `db.NewDB` uses a 10-second startup ping context.

## Shutdown behavior

There is no graceful shutdown path. On `SIGINT`/`SIGTERM`, the Go runtime terminates the process; in-flight
requests are not drained, and deferred DB/log cleanup is skipped. This matters for deployment and for log
flushing. See [12. Docker and deployment](12-docker-deployment.md) and
[Architecture review](../reports/architecture-review.md).

## Related documents

- [Architecture](02-architecture.md)
- [API reference](07-api-reference.md)
- [Authentication and authorization](08-authentication-authorization.md)
- [Database operations](06-database-operations.md)
- [Logging and observability](11-logging-observability.md)
