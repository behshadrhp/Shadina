# API Consistency Review

**Implementation:** `cmd/api/router.go` and handlers  
**Contract:** `docs/swagger.json` / `docs/swagger.yaml` (OpenAPI 2.0)  
**Review date:** 2026-10-29

## Persian summary

Swagger موجود با پیاده‌سازی هم‌خوان نیست: مسیر feed به‌صورت `/feed` مستند شده اما واقعی `/feeds` است؛
follow/unfollow به‌صورت POST/DELETE با بدنه مستند شده‌اند اما واقعی PUT بدون بدنه‌اند؛ register مستند نشده؛
پاسخ 201/400 با رفتار 200/500 جایگزین شده؛ امنیت در سطح operation تعریف نشده و DTOهای `FollowPayload` و
`UnFollowPayload` وجود ندارند، بنابراین `swag init` شکست می‌خورد. فایل‌های JSON و YAML با هم سازگارند اما هر دو
کهنه‌اند.

## Endpoint matrix: implementation versus Swagger

| # | Method/path (actual) | Auth actual | Documented? | Contract differences |
|---:|---|---|---|---|
| 1 | `GET /ping` | Public | No | Heartbeat middleware; not in Swagger. |
| 2 | `POST /v1/auth/register` | Public | **No** | Missing entirely. |
| 3 | `POST /v1/auth/token` | Public | Yes | Response schema generic object; actual `data` is a string; 500 documented but not produced. |
| 4 | `GET /v1/health` | Public | Yes | Matches; response is `dto.HealthStatusPayload` directly, not envelope. |
| 5 | `GET /v1/users` | JWT | Yes | No operation security; actual `role_id: 0`; no pagination documented/implemented. |
| 6 | `GET /v1/users/{userID}` | JWT | Yes | 400/404 documented; actual 500 for parse/not found; no operation security. |
| 7 | `PUT /v1/users/{userID}/follow` | JWT | Yes (wrong) | Swagger `POST` + required body `dto.FollowPayload`; actual `PUT`, no body; DTO missing. |
| 8 | `PUT /v1/users/{userID}/unfollow` | JWT | Yes (wrong) | Swagger `DELETE` + required body `dto.UnFollowPayload`; actual `PUT`, no body; DTO missing. |
| 9 | `GET /v1/users/{userID}/feeds` | JWT | Yes (wrong) | Swagger `/feed`; actual `/feeds`; `page` documented, `offset` actual; sort ineffective. |
| 10 | `GET /v1/posts` | JWT | Yes | No operation security; no pagination; N+1; `user` zero. |
| 11 | `POST /v1/posts` | JWT | Yes | Swagger 201/400; actual 200; validation/JSON errors 500. |
| 12 | `GET /v1/posts/{postID}` | JWT | Yes | No operation security; matches statuses (404/500). |
| 13 | `PATCH /v1/posts/{postID}` | JWT + owner/admin | Yes | No operation security; admin response bug; `updated_at` not changed. |
| 14 | `DELETE /v1/posts/{postID}` | JWT + owner/admin | Yes | No operation security; comments FK can cause 500. |
| 15 | `GET /v1/swagger/*` | Public | No | UI + doc.json; not documented. |

## Schema and DTO issues

| Issue | Detail |
|---|---|
| Missing DTOs | `dto.FollowPayload` and `dto.UnFollowPayload` are referenced by annotations but absent from `internal/dto`. `swag init` fails with `cannot find type definition`. |
| Register DTO absent from contract | `dto.UserRegisterPayload` exists but the route is not annotated. |
| Response models are storage structs | `storage.User`, `Post`, `Comment` are serialized directly; Swagger includes `storage.User` but not full post/comment response definitions. |
| Token response | Swagger models 200 as a generic object; actual `data` is a string. |
| Generic error response | Swagger uses `response.Response`; actual messages include raw validator/SQL/JWT text. |
| `FollowPayload` body | Actual handlers read no body; body is ignored. |
| Feed response | Swagger uses generic `response.Response`; actual `[]PostWithMetadata` with `comment_count` and partial `user`. |
| Health response | Not wrapped in the standard envelope; documented as `dto.HealthStatusPayload` (correct). |
| Partial user objects | Post list/retrieve return zero-value `user`; comments return only id/username; feed returns username only. |
| `role_id` in user list | Actual 0 due query omission; Swagger schema suggests a real role. |

## Status-code contract

| Endpoint | Swagger | Actual |
|---|---|---|
| Register | not documented | 200; 400; 500 |
| Token | 200/400/500 | 200; 400 only |
| Create post | 201/400/500 | 200; 500 (decode/validation/DB) |
| User retrieve | 200/400/404/500 | 200; 500 (parse/not found/DB) |
| Feed | 200/400/500 | 200; 400; 500 |
| Follow | 200/400/404/500 | 200; 401; 500 (404 dead) |
| Unfollow | 200/400/404/500 | 200; 401; 500 (404 dead) |
| Post update | 200/400/404/500 | 200; 400; 404; 500 |
| Post delete | 200/404/500 | 200; 404; 500 |
| Auth middleware | not documented | 401; 500 for invalid token |

## Security definitions

- `securityDefinitions.ApiKeyAuth` exists: header `Authorization`, type `apiKey`.
- No operation has `"security": [{"ApiKeyAuth": []}]`, and there is no global `security`.
- Swagger UI will not attach the bearer token to protected operations.
- The actual scheme expects `Bearer <token>`, not a raw API key. An `apiKey` definition named
  `Authorization` is technically usable if the value includes `Bearer ...`, but this should be explicit.

## JSON/YAML consistency

`docs/swagger.json` and `docs/swagger.yaml` were parsed and compared recursively during the audit. No
semantic differences were found. Both:

- use `swagger: "2.0"`;
- have `basePath: /v1`;
- have 9 paths and 8 definitions;
- lack `host` and `info.version` (filled at runtime by `docs.SwaggerInfo`);
- lack operation-level security;
- reference missing DTOs.

They are consistent with each other but inconsistent with the code.

## Recommended fixes

**Contract-only (no behavior change)**

1. Add `dto.FollowPayload`/`dto.UnFollowPayload` **or** remove the body parameters from annotations. Removing
   them matches the current handlers and unblocks generation.
2. Change feed annotation to `/users/{userID}/feeds`; document `offset` instead of `page` (or add `page`
   support in code, which is a behavior change).
3. Change follow/unfollow annotations to `PUT` and remove required bodies.
4. Add `POST /auth/register` annotation and `dto.UserRegisterPayload` schema.
5. Add `@Security ApiKeyAuth` to protected operations (or a global security with public overrides).
6. Correct create-post success to 200 (or change code to 201, a behavior change requiring approval).
7. Correct user-retrieve error statuses to 500 (or fix code to 400/404).
8. Add response schemas for post/comment/feed/token.
9. Regenerate with `swag init` and verify no diff.

**Code changes requiring approval**

10. Return 400 for create-post decode/validation; 201 for creation.
11. Return 404 for missing users; 400 for non-numeric IDs.
12. Return 401 for invalid tokens; 403 for forbidden ownership.
13. Fix admin middleware double-write.
14. Fix feed count/sort/offset and enforce/justify path user access.
15. Add JSON 404/405 handlers.

## Verification

- Route table extracted from `router.go`.
- Handler behavior reviewed in `auth.go`, `uesrs.go`, `posts.go`, `followers.go`, `feeds.go`, `middleware.go`.
- Runtime checks: health, swagger UI/doc.json, missing/invalid token, seeded login, users list, posts list,
  feed sort/access, wrong paths, user not found, create-post validation, delete-other-post denial.
- `swag init` run in a temporary copy: exit 1 at `dto.FollowPayload`.
- No browser-based Swagger UI test; UI config inspected.

## Related documents

- [API reference](../en/07-api-reference.md)
- [Documentation audit](documentation-audit.md)
- [Architecture review](architecture-review.md)
