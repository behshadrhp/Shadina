# 07. API Reference

This reference documents the routes that are actually registered in `cmd/api/router.go` and the behavior that
is actually implemented in the handlers. Where the checked-in Swagger contract differs, the difference is
called out explicitly. Examples use placeholder data; IDs and timestamps are illustrative.

## Conventions

| Item | Value |
|---|---|
| Base path | `/v1` (the router mounts `r.Route("/v1", ...)`); `GET /ping` is outside `/v1` |
| Content type | JSON requests and responses, except `/ping` (`text/plain`), Swagger UI assets, and chi's default 404/405 text |
| Authentication | `Authorization: Bearer <JWT>` for all `/v1/users/*` and `/v1/posts/*` routes |
| Public routes | `/ping`, `/v1/health`, `/v1/auth/register`, `/v1/auth/token`, `/v1/swagger/*` |
| Envelope | `{"success": bool, "message": string, "internal_code": int, "data": any}` |
| Error data | Always `null` in the current helpers |
| Success status | All success helpers use HTTP 200; `CreatedResponse` (201) exists but is not used |
| Timestamps | RFC3339 strings produced from PostgreSQL `timestamptz`; second precision |
| IDs | JSON numbers (`int64`) |
| Body limit | 1,048,578 bytes (1 MiB) via `http.MaxBytesReader` |
| Unknown JSON fields | Rejected (`decoder.DisallowUnknownFields`) |

The router accepts both `/v1/users` and `/v1/users/` (and `/v1/posts` / `/v1/posts/`) in local testing.
The canonical paths below omit the trailing slash. Routes with a path parameter accept both
`/v1/posts/{id}` and `/v1/posts/{id}/` when the route is matched; collection/feed routes are less forgiving
(`/v1/users/1/feeds/` returned 404 in testing).

## Route index

| # | Method | Path | Auth | Handler | Purpose |
|---:|---|---|---|---|---|
| 1 | GET | `/ping` | Public | chi `Heartbeat` | Liveness heartbeat; body `.` |
| 2 | POST | `/v1/auth/register` | Public | `userRegistrationHandler` | Create an account |
| 3 | POST | `/v1/auth/token` | Public | `generateTokenHandler` | Exchange email/password for a JWT |
| 4 | GET | `/v1/health` | Public | `healthCheckHandler` | Static health/version payload |
| 5 | GET | `/v1/users` | JWT | `listUsersHandler` | List all users |
| 6 | GET | `/v1/users/{userID}` | JWT | `retrieveUsersHandler` | Retrieve one user |
| 7 | PUT | `/v1/users/{userID}/follow` | JWT | `UserFollowHandler` | Follow target user |
| 8 | PUT | `/v1/users/{userID}/unfollow` | JWT | `UserUnfollowHandler` | Unfollow target user |
| 9 | GET | `/v1/users/{userID}/feeds` | JWT | `listFeedsHandler` | Feed of followed users + own posts |
| 10 | GET | `/v1/posts` | JWT | `listPostsHandler` | List all posts with comments |
| 11 | POST | `/v1/posts` | JWT | `createPostHandler` | Create a post |
| 12 | GET | `/v1/posts/{postID}` | JWT | `retrievePostHandler` | Retrieve one post with comments |
| 13 | PATCH | `/v1/posts/{postID}` | JWT + owner/admin | `updatePostHandler` | Partially update a post |
| 14 | DELETE | `/v1/posts/{postID}` | JWT + owner/admin | `deletePostHandler` | Delete a post |
| 15 | GET | `/v1/swagger/*` | Public | `httpSwagger.Handler` | Swagger UI and `doc.json` |
| — | * | unmatched | — | chi default | Plain-text 404/405, not the JSON envelope |

The `POST /v1/auth/register` route is **not present in Swagger**. The Swagger paths for feed/follow/unfollow
do not match the router (see [Swagger comparison](#swagger-comparison)).

## 1. POST /v1/auth/register

Create a user account. Public.

**Request body** (`dto.UserRegisterPayload`):

| Field | Type | Required | Validation | Notes |
|---|---|---:|---|---|
| `email` | string | Yes | `required,max=255` | **No email-format validation**; `citext` uniqueness at DB |
| `username` | string | Yes | `required,max=255` | Unique at DB; no format/min-length rule |
| `password` | string | Yes | `required,min=8,max=255` | bcrypt hashed; >72 bytes will fail bcrypt (not tested at runtime) |

Unknown fields are rejected. Empty body returns 400 code `4011`.

**Example request:**

```http
POST /v1/auth/register HTTP/1.1
Content-Type: application/json

{
  "email": "new.user@example.com",
  "username": "newuser",
  "password": "correct horse battery"
}
```

**Success response:** HTTP **200** (not 201), internal code `1010`.

```json
{
  "success": true,
  "message": "user registered successfully",
  "internal_code": 1010,
  "data": {
    "id": 108,
    "username": "newuser",
    "email": "new.user@example.com",
    "role_id": 1,
    "created_at": "2026-10-29T12:00:00Z",
    "updated_at": "2026-10-29T12:00:00Z"
  }
}
```

`password` is omitted from JSON (`json:"-"`). `role_id` is hard-coded to `1`.

**Status codes and errors:**

| Status | Code | Condition |
|---:|---:|---|
| 200 | 1010 | User inserted |
| 400 | 4011 | Invalid JSON / unknown field / body >1 MiB / EOF |
| 400 | 4012 | Validation failure (message contains validator details) |
| 500 | 5018 | bcrypt failure or database insert failure (duplicate username/email, FK, etc.); raw `err.Error()` for DB failures |
| 401 | — | Not used by this handler (despite the 4011 code name) |

**Database effect:** `INSERT INTO users(username, email, password, role_id) VALUES (...)`.
**Middleware:** global logging only; no JWT.
**Swagger:** missing entirely.

## 2. POST /v1/auth/token

Exchange email/password for a JWT. Public.

**Request body** (`dto.GenerateTokenPayload`):

| Field | Type | Required | Validation |
|---|---|---:|---|
| `email` | string | Yes | `required,max=255` |
| `password` | string | Yes | `required,min=8,max=255` |

**Example:**

```http
POST /v1/auth/token HTTP/1.1
Content-Type: application/json

{"email":"new.user@example.com","password":"correct horse battery"}
```

**Success response:** HTTP 200, internal code `1011`; `data` is the token string, not an object.

```json
{
  "success": true,
  "message": "generate token successfully",
  "internal_code": 1011,
  "data": "<jwt-token>"
}
```

**Token claims:** `sub` (user ID), `exp` (now + 72h), `iat`, `nbf`, `iss` = `Shadina`,
`aud` = `Shadina`. Signed HS256 with `AUTH_SECRET_KAY` (or its hard-coded fallback). There is no refresh
token, `jti`, role claim, or revocation identifier.

**Status codes and errors:**

| Status | Code | Condition |
|---:|---:|---|
| 200 | 1011 | Token generated |
| 400 | 4013 | Invalid JSON / unknown field |
| 400 | 4014 | Validation failure |
| 400 | 4015 | User not found (`sql.ErrNoRows`) |
| 400 | 4016 | Other database error (message generic `"invalid request"`) |
| 400 | 4017 | Password mismatch (`"invalid email or password"`) |
| 400 | 4018 | Token signing failure (message generic `"invalid request"`) |
| 500 | — | Not produced by this handler; Swagger's 500 is inaccurate |

**Database effect:** `SELECT ... FROM users WHERE email=$1`; no writes.
**Seeded users:** cannot log in because their password is empty (bcrypt comparison fails). Register first.
**Swagger:** path/method match; Swagger says 500 and an object response; actual maximum is 400 and `data` is a
string.

## 3. GET /v1/health

Public liveness payload.

**Response:** HTTP 200, `dto.HealthStatusPayload` (not wrapped in `response.Response`).

```json
{"status":"ok","version":"0.0.1","environment":"development"}
```

It does **not** query PostgreSQL, check migrations, or report dependency health. A database outage can still
return `ok`.

## 4. GET /v1/users

List all users. JWT required.

**Request:** no query parameters. No pagination.

**Success:** HTTP 200, code `1005`, `data` is an array of `storage.User`:

```json
{
  "success": true,
  "message": "users retrieved successfully",
  "internal_code": 1005,
  "data": [
    {
      "id": 1,
      "username": "alice0",
      "email": "alice0@example.com",
      "role_id": 0,
      "created_at": "2026-09-18T20:37:18Z",
      "updated_at": "2026-09-18T20:37:18Z"
    }
  ]
}
```

**Verified quirk:** `role_id` is `0` for every list row because `UsersStorage.List` does not select
`role_id`. Use `GET /v1/users/{userID}` to get the real role.

**Status codes:** 200; 401 (missing/malformed token or user lookup failure); 500 code `5009` for SQL errors.
`GET /v1/users/{userID}` has different error semantics.

**Swagger:** path/method match; no security requirement is declared; response schema shows `storage.User`.

## 5. GET /v1/users/{userID}

Retrieve one user. JWT required.

**Path parameter:** `userID` integer (`int64`).

**Success:** HTTP 200, code `1006`, `data` is a `storage.User` with the real `role_id`:

```json
{
  "success": true,
  "message": "user retrieved successfully",
  "internal_code": 1006,
  "data": {
    "id": 1,
    "username": "alice0",
    "email": "alice0@example.com",
    "role_id": 1,
    "created_at": "2026-09-18T20:37:18Z",
    "updated_at": "2026-09-18T20:37:18Z"
  }
}
```

**Errors:** HTTP 500, code `5010` for both a non-numeric ID and `sql.ErrNoRows` (`"not found"`). The handler
does not map `consts.NOT_FOUND` to 404. Verified with `/v1/users/abc` and `/v1/users/999999`:
both returned 500. Swagger declares 400/404; those statuses are not produced.

**Swagger:** path/method match; security missing; error contract inaccurate.

## 6. PUT /v1/users/{userID}/follow

Follow the user identified by the path. JWT required.

**Path parameter:** `userID` target user ID.

**Request body:** none. The handler does not call `ReadJson`; any body is ignored.

**Database effect:**

```sql
INSERT INTO followers(user_id, follower_id) VALUES ($1, $2)
-- $1 = authenticated user, $2 = path target
```

**Success:** HTTP 200, code `1007`, `data: null`.

```json
{"success":true,"message":"user followed successfully","internal_code":1007,"data":null}
```

**Errors:**

| Status | Code | Condition |
|---:|---:|---|
| 200 | 1007 | Insert succeeded |
| 401 | 4019/4020/4021/4022 | Missing/malformed/invalid auth or user lookup failure |
| 500 | 5020 | Path ID is not an integer |
| 500 | 5013 | Database error: duplicate follow (PK), unknown target (FK), etc. |
| 404 | 4006 | Code exists in the handler but the branch is unreachable: storage returns raw DB errors, not `consts.NOT_FOUND` |

There is no self-follow check. Following yourself inserts `(user_id=self, follower_id=self)` and succeeds
unless the pair already exists.

**Swagger:** documents `POST` with a required `dto.FollowPayload` body. The actual route is `PUT` and ignores
the body. `dto.FollowPayload` does not exist, which breaks Swagger generation.

## 7. PUT /v1/users/{userID}/unfollow

Unfollow the path user. JWT required.

**Request body:** none.

**Database effect:**

```sql
DELETE FROM followers WHERE user_id=$1 AND follower_id=$2
```

**Success:** HTTP 200, code `1008`, `data: null`, even if no relationship existed. `RowsAffected` is not
checked, so `404` is never returned by the storage path.

**Errors:** 401 auth; 500 code `5021` for a non-numeric path ID; 500 code `5015` for SQL errors.
The declared 404 code `4008` is unreachable.

**Swagger:** documents `DELETE` with a body (`dto.UnFollowPayload`); actual is `PUT` with no body, and the
type is missing.

## 8. GET /v1/users/{userID}/feeds

Return the feed for a user. JWT required.

**Path parameter:** `userID` (int64). The handler does **not** check that it equals the authenticated user's
ID. Any valid token can request any user's feed.

**Query parameters** (`storage.PaginatedQuery`):

| Parameter | Type | Default | Validation | Actual behavior |
|---|---|---:|---|---|
| `limit` | int | 20 | 1..100 | Invalid non-numeric values are silently ignored; out-of-range values fail validation (400). |
| `offset` | int | **10** | >=0 | Default is 10, not 0. Invalid non-numeric values are silently ignored. |
| `page` | int | — | — | **Not read** despite Swagger documenting it. |
| `sort` | string | `desc` | `ASC DESC asc desc` | Accepted but ineffective: SQL uses `ORDER BY p.created_at, $2`, so direction does not change. |
| `search` | string | empty | max 100 | Filters title/content with `ILIKE`; parameterized. |
| `since` | string | empty | none | Parsed with `time.DateTime`; unused in SQL. Invalid values become empty. |
| `until` | string | empty | none | Stored but unused; not parsed/formatted. |

**Success:** HTTP 200, code `1009`, `data` is `[]PostWithMetadata`:

```json
{
  "success": true,
  "message": "feeds retrieved successfully",
  "internal_code": 1009,
  "data": [
    {
      "id": 201,
      "user_id": 1,
      "title": "Valhala",
      "content": "Hello world",
      "tags": ["joice", "tensa"],
      "created_at": "2026-10-17T07:31:18Z",
      "updated_at": "2026-10-17T07:31:18Z",
      "user": {
        "id": 0,
        "username": "alice0",
        "email": "",
        "role_id": 0,
        "created_at": "0001-01-01T00:00:00Z",
        "updated_at": "0001-01-01T00:00:00Z"
      },
      "comments": null,
      "comment_count": 0
    }
  ]
}
```

Only `user.username` is populated by the feed SQL; other user fields are zero values. Comments are not
loaded; the aggregate count is exposed as `comment_count`.

**Errors:** 400 code `4009` is declared but **unreachable** (`Parse` always returns `nil`; invalid numeric
values are ignored); 400 code `4010` for `Validate.Struct` failures (invalid sort, out-of-range limit/offset);
500 code `5016` for a non-numeric path ID; 500 code `5017` for SQL errors; 401 auth.

**Verified correctness issues:** count inflation, ineffective sort, default offset 10, unused since/until,
cross-user access. See [06. Database operations](06-database-operations.md#feed-query-analysis).

**Swagger:** path is `/users/{userID}/feed` (singular) while the actual route is `/feeds`; `page` is
documented but unused; security is missing.

## 9. GET /v1/posts

List all posts with comments. JWT required.

**Request:** no query parameters, no pagination.

**Success:** HTTP 200, code `1001`, `data` is `[]*storage.Post`:

```json
{
  "success": true,
  "message": "posts retrieved successfully",
  "internal_code": 1001,
  "data": [
    {
      "id": 201,
      "user_id": 1,
      "title": "Valhala",
      "content": "Hello world",
      "tags": ["joice", "tensa"],
      "created_at": "2026-10-17T07:31:18Z",
      "updated_at": "2026-10-17T07:31:18Z",
      "user": {
        "id": 0,
        "username": "",
        "email": "",
        "role_id": 0,
        "created_at": "0001-01-01T00:00:00Z",
        "updated_at": "0001-01-01T00:00:00Z"
      },
      "comments": [
        {
          "id": 1,
          "user_id": 2,
          "post_id": 201,
          "user": {
            "id": 2,
            "username": "bob1",
            "email": "",
            "role_id": 0,
            "created_at": "0001-01-01T00:00:00Z",
            "updated_at": "0001-01-01T00:00:00Z"
          },
          "content": "Great post!",
          "created_at": "2026-09-18T20:37:20Z",
          "updated_at": "2026-09-18T20:37:20Z"
        }
      ]
    }
  ]
}
```

`Post.User` is a zero-value object because the post query does not join users. Comments' `user` only has
`id` and `username` populated by `CommentsStorage.GetByPostID`.

**Errors:** 500 code `5001` for the post list query; 500 code `5009` for a comments query; 401 auth.
The endpoint performs one comments query per post (N+1). See [Application flow](04-application-flow.md).

**Swagger:** path/method match; security missing; response schema is generic.

## 10. POST /v1/posts

Create a post for the authenticated user. JWT required.

**Request body** (`dto.CreatePostPayload`):

| Field | Type | Required | Validation |
|---|---|---:|---|
| `title` | string | Yes | `required,max=255` |
| `content` | string | Yes | `required` (no max) |
| `tags` | string[] | No | none |

**Example:**

```json
{"title":"Hello","content":"World","tags":["intro","go"]}
```

**Success:** HTTP **200** (Swagger says 201), code `1002`, `data` is the created `storage.Post`:

```json
{
  "success": true,
  "message": "post created successfully",
  "internal_code": 1002,
  "data": {
    "id": 206,
    "user_id": 1,
    "title": "Hello",
    "content": "World",
    "tags": ["intro", "go"],
    "created_at": "2026-10-29T12:00:00Z",
    "updated_at": "2026-10-29T12:00:00Z",
    "user": {
      "id": 0,
      "username": "",
      "email": "",
      "role_id": 0,
      "created_at": "0001-01-01T00:00:00Z",
      "updated_at": "0001-01-01T00:00:00Z"
    },
    "comments": null
  }
}
```

**Errors:**

| Status | Code | Condition |
|---:|---:|---|
| 500 | 5002 | JSON decode error (including empty body) |
| 500 | 5006 | Validation failure (despite Swagger's 400) |
| 500 | 5003 | SQL/DB error |
| 401 | — | JWT middleware |

**Database effect:** insert into `posts`; `user_id` comes from the JWT context, not the body.

**Swagger:** says 201 and 400; actual is 200 and 500 for these failures. Security missing.

## 11. GET /v1/posts/{postID}

Retrieve a post with comments. JWT required.

**Success:** HTTP 200, code `1003`, `data` is a `storage.Post` with `comments` populated; `user` zero-value.

**Errors:** 404 code `4001` for `sql.ErrNoRows`; 500 code `5004` for a non-numeric ID; 500 code `5005` for
other post query errors; 500 code `5009` for comment query errors; 401 auth. Verified `/v1/posts/999999`
returns 404 with `{"success":false,"message":"post not found","internal_code":4001,"data":null}`.

## 12. PATCH /v1/posts/{postID}

Partially update a post. JWT + `UserPostAccessMiddleware`.

**Request body** (`dto.UpdatePostPayload`):

| Field | Type | Required | Validation | Notes |
|---|---|---:|---|---|
| `title` | string pointer | No | `omitempty,max=255` | Applied if present |
| `content` | string pointer | No | `omitempty` | Applied if present |

Both fields are optional; an empty object is valid and performs an update with unchanged values. Unknown
fields are rejected.

**Success:** HTTP 200, code `1004`, `data: null`.

**Access rules (actual):**

- The post is fetched first; `sql.ErrNoRows` → 404 code `4002`.
- If `user.ID == post.UserID`, the handler runs.
- If not the owner and `user.RoleID == 3`, the handler runs **and then the middleware also calls
  `UnauthorizedResponse`**. The admin update may be applied, but the response is corrupted (first success
  response, then a 401 write attempt). This is a verified code-path defect; it was not exercised at runtime
  because the development DB has no admin user.
- If not the owner and not admin, returns 401 code `4023` (`"access denied"`). Moderators (`role_id=2`) are
  denied despite the role description.
- Owners are allowed regardless of role.

**Other errors:** 400 code `4003` for JSON decode; 400 code `4004` for validation; 500 code `5007` for a
non-numeric path ID; 500 code `5008` for fetch/update SQL errors.

**Database effect:** `UPDATE posts SET title=$1, content=$2 WHERE id=$3`; `updated_at` is not changed.
**Swagger:** path/method match; security missing; 400/404 documented, but the admin/role behavior is not.

## 13. DELETE /v1/posts/{postID}

Delete a post. JWT + owner/admin middleware.

**Success:** HTTP 200, code `1004`, `data: null`.

**Access rules:** same ownership/admin logic and same admin response bug as PATCH. Non-owner/non-admin gets
401 code `4023`.

**Errors:** 404 code `4001` if the post is absent; 500 code `5007` for a non-numeric ID; 500 code `5008` for
the delete error. Deleting a post that has comments fails on `fk_post_comment` and returns 500 with the raw
PostgreSQL error. A rolled-back SQL test confirmed the FK violation; most seeded posts have comments.

**Database effect:** `DELETE FROM posts WHERE id=$1`; comments are not deleted and have no cascade.
**Swagger:** path/method match; security missing; the FK behavior is not documented.

## 14. Swagger endpoints

| Route | Auth | Behavior |
|---|---|---|
| `GET /v1/swagger/index.html` | Public | Swagger UI HTML, status 200 |
| `GET /v1/swagger/doc.json` | Public | Generated Swagger JSON served by `http-swagger`, status 200 |
| `GET /v1/swagger/` | Public | Redirects to `index.html` |

The UI config is generated as:

```text
url: "127.0.0.1:8000/swagger/doc.json"   # example with SERVER_URL=127.0.0.1:8000
```

The actual handler path is `/v1/swagger/doc.json`; the configured URL has no `/v1` prefix and no scheme.
The UI is served, but loading the spec from that URL is expected to fail. This was verified by inspecting the
generated HTML; no headless browser was run.

## 15. Default 404/405 behavior

Unmatched routes and unsupported methods fall through to chi's default handlers. They return plain text
(commonly `404 page not found` or `405 method not allowed`), **not** the JSON envelope. Examples verified:

- `GET /v1/users/1/feed` → 404 (the real route is `/feeds`).
- `GET /v1/users/2/follow` → 405 (the route is `PUT`; there is no GET).
- `/v1/users/1/feeds/` → 404.

Clients must not assume every error is JSON.

## Swagger comparison

The checked-in `docs/swagger.json` and `docs/swagger.yaml` are semantically identical to each other (parsed
and compared during the audit), but they do not match the implementation.

| Actual behavior | Swagger | Difference |
|---|---|---|
| `POST /v1/auth/register` | absent | Missing endpoint. |
| `POST /v1/auth/token` returns 200 with `data` string | 200/400/500, generic object | Status set and schema wrong; 500 not produced. |
| `GET /v1/health` | present | Matches. |
| `GET /v1/users` | present, no security | Auth requirement missing. |
| `GET /v1/users/{userID}` returns 500 for not found | 400/404 | Status contract wrong. |
| `GET /v1/users/{userID}/feeds` | `/feed`, `page` | Path and query parameters wrong. |
| `PUT /v1/users/{userID}/follow`, no body | `POST`, required body `dto.FollowPayload` | Method/body wrong; DTO missing. |
| `PUT /v1/users/{userID}/unfollow`, no body | `DELETE`, required body `dto.UnFollowPayload` | Method/body wrong; DTO missing. |
| `POST /v1/posts` returns 200; validation 500 | 201; validation 400 | Status contract wrong. |
| `PATCH/DELETE /v1/posts/{postID}` require owner/admin | no security | Auth policy missing. |
| Swagger UI serves at `/v1/swagger/*` | not documented | UI route absent. |
| No operation-level `security` | `ApiKeyAuth` definition only | Swagger UI will not send the bearer token. |

`docs/swagger.json`/`yaml` also have no `host` or `info.version` because they are generated statically;
`docs.SwaggerInfo` fills these at runtime. `swag init` currently fails at `dto.FollowPayload`, so the stale
files cannot be regenerated without fixing the annotations/DTOs.

See [API consistency review](../reports/api-consistency-review.md) for the full matrix.

## Related documents

- [Authentication and authorization](08-authentication-authorization.md)
- [Error handling](10-error-handling.md)
- [Database operations](06-database-operations.md)
- [Troubleshooting](16-troubleshooting.md)
