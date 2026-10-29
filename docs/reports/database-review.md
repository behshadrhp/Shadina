# Database Review

**Database:** PostgreSQL 18.4  
**Schema source:** `cmd/migrate/migrations/*.sql`  
**Verified state:** `schema_migrations.version = 12`, `dirty = false`  
**Extensions:** `citext`, `pg_trgm`, `plpgsql`

## Persian summary

طرح دیتابیس از نظر پایه‌ای سالم است: کلیدهای خارجی، یکتایی ایمیل/نام کاربری، timestamptz، ایندکس‌های trigram
و GIN برای tags وجود دارد. مشکلات مهم: حذف پست دارای کامنت به‌دلیل نبود cascade شکست می‌خورد؛ پرس‌وجوی feed
تعداد کامنت‌ها را به‌دلیل join تکراری چند برابر می‌کند و sort بی‌اثر است؛ seed در تراکنش واقعی اجرا نمی‌شود و
کاربران seed رمز خالی دارند؛ `updated_at` به‌روزرسانی نمی‌شود؛ چند ایندکس بلااستفاده/تکراری است؛ `username`
ایندکس تکراری دارد و `scripts/init_db.sql` معتبر نیست.

## Schema and relationships

| Table | Rows/columns | PK | Unique | FKs |
|---|---|---|---|---|
| `roles` | id, name, level, description | id | name | — |
| `users` | id, username, email, password, created_at, updated_at, role_id | id | username, email (citext) | role_id → roles.id |
| `posts` | id, user_id, title, content, tags, created_at, updated_at | id | — | user_id → users.id (`fk_user_post`) |
| `comments` | id, user_id, post_id, content, created_at, updated_at | id | — | user_id → users.id (`fk_user_comment`), post_id → posts.id (`fk_post_comment`) |
| `followers` | user_id, follower_id, created_at | (user_id, follower_id) | — | user_id/follower_id → users.id, both `ON DELETE CASCADE` |

`schema_migrations` is managed by golang-migrate.

All relationships are database-enforced. There are no application-only relations.

### Followers direction

`followers.user_id` = acting follower; `followers.follower_id` = followed user. The naming is inverted from
common expectations and should be documented/renamed in a future migration.

## Constraints and indexes

### Verified indexes

| Table | Index | Type | Assessment |
|---|---|---|---|
| users | `users_pkey` | btree (id) | Needed. |
| users | `users_username_key` | unique btree (username) | Needed. |
| users | `users_email_key` | unique btree (email) | Needed. |
| users | `idx_users_username` | btree (username) | **Redundant** with unique constraint index. |
| users | `users_role_id_fkey` | FK only | No index on `role_id`; add only if queries need it. |
| posts | `posts_pkey` | btree (id) | Needed. |
| posts | `idx_posts_user_id` | btree (user_id) | Useful for user posts/feed. |
| posts | `idx_posts_title` | GIN (title gin_trgm_ops) | Supports feed title ILIKE. |
| posts | `idx_posts_tags` | GIN (tags) | **Unused** by current queries. |
| posts | — | none on created_at | Feed orders by `created_at`; consider after measurement. |
| comments | `comments_pkey` | btree (id) | Needed. |
| comments | `idx_comments_post_id` | btree (post_id) | Supports `GetByPostID`. |
| comments | `idx_comments_content` | GIN (content gin_trgm_ops) | **Unused**; no comment search. |
| comments | — | none on user_id | Not currently needed. |
| followers | `followers_pkey` | btree (user_id, follower_id) | Supports prefix `user_id`; join on `follower_id` may not use it. |

### Constraint observations

- `posts.user_id` and `comments.post_id` have no `ON DELETE` action; deletes fail rather than cascade.
- `followers` cascades on both user FKs.
- `users.role_id` has no `ON DELETE`; deleting a role in use fails.
- `posts.tags` is nullable; no default/`NOT NULL`.
- No check constraints for non-empty titles/content, role level ranges, or self-follow.
- No `updated_at` trigger.
- No soft-delete/audit columns.

## Query review

| Query | Assessment |
|---|---|
| `UsersStorage.Create` | Straight insert; unique/FK errors returned raw. |
| `UsersStorage.List` | Omits `role_id` (bug); no ordering/pagination. |
| `UsersStorage.GetByID/GetByEmail` | Indexed point lookups; good. |
| `PostsStorage.Create` | Insert with array; good. |
| `PostsStorage.List` | Full scan, no ordering/pagination. |
| `PostsStorage.GetByID` | PK lookup; good. |
| `PostsStorage.Delete` | PK delete; FK comments block; no transaction. |
| `PostsStorage.Update` | PK update; omits `updated_at`; no rows-affected check. |
| `PostsStorage.GetFeed` | Correctness issues; see below. |
| `CommentsStorage.Create` | Insert; returns only id/created_at (updated_at zero in struct). |
| `CommentsStorage.GetByPostID` | Uses post_id index + users PK join; good. |
| `FollowersStorage.Follow` | PK/unique; duplicate/unknown target raw errors. |
| `FollowersStorage.Unfollow` | PK delete; ignores zero rows. |

### Feed query defects (verified)

1. **Inflated comment count:** local post 19: `COUNT(c.id)=16`, `COUNT(DISTINCT c.id)=4`.
2. **Sort ignored:** `ORDER BY p.created_at, $2` treats `$2` as a constant; asc/desc returned same order.
3. **Default offset 10**, not 0; `page` is not parsed.
4. **Inner join on followers** can hide own posts when the table is empty.
5. **`since`/`until` unused.**
6. **Path user not authorized** against token user.
7. **`search` content** lacks a trigram index on `posts.content`; title has one.

### N+1 and full scans

- `GET /v1/posts`: 1 post query + N comment queries. 203 posts → up to 204 queries.
- `GET /v1/users`: full table, no pagination.
- `GET /v1/posts`: full table, no pagination.
- Feed content search can scan posts due no content trigram index.

## Transactions and concurrency

- No request-level transactions.
- Seed calls `BeginTx` but storage methods use `*sql.DB`, so inserts are outside the transaction; commit is
  empty, rollback does not undo.
- No explicit isolation level; PostgreSQL default (READ COMMITTED) applies.
- `PostsStorage.Update` fetch-then-update in the handler has a race: a concurrent delete between fetch and
  update yields a successful response without updating. Rows affected is not checked.
- No deadlock evidence was observed, but no concurrency tests exist.
- No `SELECT ... FOR UPDATE`/optimistic locking.

## Data integrity risks

| Risk | Severity | Evidence | Mitigation |
|---|---|---|---|
| Empty seed passwords | Medium (data quality) | 100 users `password=''::bytea` | Hash passwords or mark fixtures non-loginable; use register for test users. |
| Post delete blocked by comments | High (functional) | FK `fk_post_comment`; rolled-back delete failed | Cascade, transactional delete, or 409. |
| Feed count inflation | High (data correctness) | 16 vs 4 on post 19 | `COUNT(DISTINCT c.id)` + correct join. |
| Seed non-atomic | Medium | tx not used | Pass `tx` to stores; check errors; idempotency. |
| `updated_at` stale | Medium | update SQL omits it | Trigger or explicit `NOW()`. |
| Role ID hardcoding | Medium | registration/admin checks | Lookup by name; stable seed IDs. |
| Duplicate follow 500 | Medium | PK violation raw | Pre-check/map 409. |
| Unknown follow target 500 | Medium | FK violation raw | Map 404. |
| Invalid `init_db.sql` | Low (tooling) | `CREATE DATABASE IF NOT EXISTS` | Remove/fix; use Compose/migrations. |
| Migration gap 8–10 | Low (history) | files absent | Renumber/document. |
| Redundant `idx_users_username` | Low (write overhead) | duplicate index | Drop in migration after review. |
| Unused indexes | Low (write overhead) | tags/comments content | Confirm usage, then drop. |

## Optimization recommendations

**P0**

1. Fix feed count/join/sort/offset and add tests with multiple followers/comments.
2. Choose post-delete semantics and implement FK/cascade/transaction.
3. Check `RowsAffected` for update/delete/unfollow.
4. Update `updated_at` via trigger or SQL.
5. Select `role_id` in user list; add ordering/pagination.

**P1**

6. Make seed transactional/idempotent; hash or invalidate fixture passwords.
7. Add indexes after `EXPLAIN (ANALYZE, BUFFERS)` evidence:
   - `posts(created_at)` for feed ordering,
   - `followers(follower_id, user_id)` for feed joins,
   - `users(role_id)` if role filtering is added,
   - `comments(user_id)` if user-comment queries are added.
8. Remove `idx_users_username`; evaluate unused GIN indexes.
9. Add migration tests (up/down on disposable DB) and a CI check.
10. Add pagination to posts/users.

**P2**

11. Add soft delete/audit columns if required.
12. Consider partitioning/archival for large posts/comments.
13. Add backup/PITR and restore drills; not in repository.

## Verification limits

- No `EXPLAIN ANALYZE` benchmark was run; index recommendations are structural and marked for measurement.
- No load/concurrency tests exist.
- The live DB is a development dataset (107 users/203 posts/500 comments/4 followers), not production.
- Migration rollback was not executed (would alter the local schema); down files were reviewed statically.

## Related documents

- [Database design](../en/05-database-design.md)
- [Database operations](../en/06-database-operations.md)
- [Architecture review](architecture-review.md)
