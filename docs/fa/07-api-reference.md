# 07. مرجع API

این مرجع مسیرهایی را مستند می‌کند که واقعاً در `cmd/api/router.go` ثبت شده‌اند و رفتاری را که واقعاً در
هندلرها پیاده‌سازی شده است. هرجا قرارداد Swagger موجود تفاوت دارد، تفاوت صریحاً ذکر می‌شود. مثال‌ها از داده
جانشین استفاده می‌کنند؛ شناسه‌ها و زمان‌ها نمونه‌اند.

## قراردادها

| مورد | مقدار |
|---|---|
| مسیر پایه | `/v1` (روتر `r.Route("/v1", ...)` را نصب می‌کند)؛ `GET /ping` بیرون `/v1` است |
| نوع محتوا | درخواست/پاسخ JSON، به‌جز `/ping` (`text/plain`)، دارایی‌های Swagger UI و متن پیش‌فرض 404/405 در chi |
| احراز هویت | `Authorization: Bearer <JWT>` برای همه مسیرهای `/v1/users/*` و `/v1/posts/*` |
| مسیرهای عمومی | `/ping`، `/v1/health`، `/v1/auth/register`، `/v1/auth/token`، `/v1/swagger/*` |
| پاکت | `{"success": bool, "message": string, "internal_code": int, "data": any}` |
| داده خطا | در کمک‌تابع‌های فعلی همیشه `null` |
| وضعیت موفقیت | همه کمک‌تابع‌های موفقیت HTTP 200 دارند؛ `CreatedResponse` (201) وجود دارد اما استفاده نمی‌شود |
| زمان‌ها | رشته‌های RFC3339 تولیدشده از `timestamptz` در PostgreSQL؛ دقت ثانیه |
| شناسه‌ها | عدد JSON (`int64`) |
| محدودیت بدنه | ۱٬۰۴۸٬۵۷۸ بایت (۱ MiB) با `http.MaxBytesReader` |
| فیلدهای JSON ناشناخته | رد می‌شوند (`decoder.DisallowUnknownFields`) |

روتر در تست محلی هم `/v1/users` و هم `/v1/users/` (و `/v1/posts` / `/v1/posts/`) را می‌پذیرد. مسیرهای متعارف
زیر اسلش انتهایی را حذف می‌کنند. مسیرهای دارای پارامتر مسیر، هنگام تطبیق هم `/v1/posts/{id}` و هم
`/v1/posts/{id}/` را می‌پذیرند؛ مسیرهای مجموعه/فید کمتر انعطاف‌پذیرند (`/v1/users/1/feeds/` در تست 404
برگرداند).

## فهرست مسیرها

| # | متد | مسیر | احراز هویت | هندلر | هدف |
|---:|---|---|---|---|---|
| ۱ | GET | `/ping` | عمومی | `Heartbeat` در chi | ضربان زنده‌بودن؛ بدنه `.` |
| ۲ | POST | `/v1/auth/register` | عمومی | `userRegistrationHandler` | ساخت حساب |
| ۳ | POST | `/v1/auth/token` | عمومی | `generateTokenHandler` | تبدیل ایمیل/رمز به JWT |
| ۴ | GET | `/v1/health` | عمومی | `healthCheckHandler` | payload ثابت سلامت/نسخه |
| ۵ | GET | `/v1/users` | JWT | `listUsersHandler` | فهرست همه کاربران |
| ۶ | GET | `/v1/users/{userID}` | JWT | `retrieveUsersHandler` | دریافت یک کاربر |
| ۷ | PUT | `/v1/users/{userID}/follow` | JWT | `UserFollowHandler` | دنبال‌کردن کاربر هدف |
| ۸ | PUT | `/v1/users/{userID}/unfollow` | JWT | `UserUnfollowHandler` | دنبال‌نکردن کاربر هدف |
| ۹ | GET | `/v1/users/{userID}/feeds` | JWT | `listFeedsHandler` | فید کاربران دنبال‌شده + پست‌های شخصی |
| ۱۰ | GET | `/v1/posts` | JWT | `listPostsHandler` | فهرست همه پست‌ها با نظرها |
| ۱۱ | POST | `/v1/posts` | JWT | `createPostHandler` | ساخت پست |
| ۱۲ | GET | `/v1/posts/{postID}` | JWT | `retrievePostHandler` | دریافت یک پست با نظرها |
| ۱۳ | PATCH | `/v1/posts/{postID}` | JWT + مالک/ادمین | `updatePostHandler` | ویرایش جزئی پست |
| ۱۴ | DELETE | `/v1/posts/{postID}` | JWT + مالک/ادمین | `deletePostHandler` | حذف پست |
| ۱۵ | GET | `/v1/swagger/*` | عمومی | `httpSwagger.Handler` | Swagger UI و `doc.json` |
| — | * | بدون تطبیق | — | پیش‌فرض chi | متن ساده 404/405، نه پاکت JSON |

مسیر `POST /v1/auth/register` در Swagger **وجود ندارد**. مسیرهای feed/follow/unfollow در Swagger با روتر
مطابقت ندارند (به [مقایسه Swagger](#مقایسه-swagger) مراجعه کنید).

## ۱. POST /v1/auth/register

ساخت حساب کاربر. عمومی.

**بدنه درخواست** (`dto.UserRegisterPayload`):

| فیلد | نوع | الزامی | اعتبارسنجی | یادداشت |
|---|---|---:|---|---|
| `email` | string | بله | `required,max=255` | **بدون اعتبارسنجی قالب ایمیل**؛ یکتایی `citext` در DB |
| `username` | string | بله | `required,max=255` | یکتا در DB؛ بدون قاعده قالب/حداقل طول |
| `password` | string | بله | `required,min=8,max=255` | هش bcrypt؛ بیش از ۷۲ بایت bcrypt را شکست می‌دهد (در زمان اجرا تست نشد) |

فیلدهای ناشناخته رد می‌شوند. بدنه خالی 400 با کد `4011` برمی‌گرداند.

**مثال درخواست:**

```http
POST /v1/auth/register HTTP/1.1
Content-Type: application/json

{
  "email": "new.user@example.com",
  "username": "newuser",
  "password": "correct horse battery"
}
```

**پاسخ موفق:** HTTP **200** (نه 201)، کد داخلی `1010`.

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

`password` از JSON حذف می‌شود (`json:"-"`). `role_id` به `1` hard-code شده است.

**کدهای وضعیت و خطاها:**

| وضعیت | کد | شرط |
|---:|---:|---|
| 200 | 1010 | کاربر درج شد |
| 400 | 4011 | JSON نامعتبر / فیلد ناشناخته / بدنه >۱ MiB / EOF |
| 400 | 4012 | شکست اعتبارسنجی (پیام شامل جزئیات validator) |
| 500 | 5018 | شکست bcrypt یا درج دیتابیس (نام کاربری/ایمیل تکراری، FK و غیره)؛ `err.Error()` خام برای خطاهای DB |
| 401 | — | توسط این هندلر استفاده نمی‌شود (با وجود نام کد 4011) |

**اثر دیتابیس:** `INSERT INTO users(username, email, password, role_id) VALUES (...)`.
**میان‌افزار:** فقط لاگ سراسری؛ بدون JWT.
**Swagger:** کاملاً غایب.

## ۲. POST /v1/auth/token

تبدیل ایمیل/رمز به JWT. عمومی.

**بدنه درخواست** (`dto.GenerateTokenPayload`):

| فیلد | نوع | الزامی | اعتبارسنجی |
|---|---|---:|---|
| `email` | string | بله | `required,max=255` |
| `password` | string | بله | `required,min=8,max=255` |

**مثال:**

```http
POST /v1/auth/token HTTP/1.1
Content-Type: application/json

{"email":"new.user@example.com","password":"correct horse battery"}
```

**پاسخ موفق:** HTTP 200، کد داخلی `1011`؛ `data` رشته توکن است، نه شیء.

```json
{
  "success": true,
  "message": "generate token successfully",
  "internal_code": 1011,
  "data": "<jwt-token>"
}
```

**claims توکن:** `sub` (شناسه کاربر)، `exp` (اکنون + ۷۲ ساعت)، `iat`، `nbf`، `iss` = `Shadina`،
`aud` = `Shadina`. با HS256 و `AUTH_SECRET_KAY` (یا fallback hard-code آن) امضا می‌شود. هیچ refresh token،
`jti`، claim نقش یا شناسه ابطال وجود ندارد.

**کدهای وضعیت و خطاها:**

| وضعیت | کد | شرط |
|---:|---:|---|
| 200 | 1011 | توکن تولید شد |
| 400 | 4013 | JSON نامعتبر / فیلد ناشناخته |
| 400 | 4014 | شکست اعتبارسنجی |
| 400 | 4015 | کاربر یافت نشد (`sql.ErrNoRows`) |
| 400 | 4016 | خطای دیگر دیتابیس (پیام عمومی `"invalid request"`) |
| 400 | 4017 | عدم تطابق رمز (`"invalid email or password"`) |
| 400 | 4018 | شکست امضای توکن (پیام عمومی `"invalid request"`) |
| 500 | — | توسط این هندلر تولید نمی‌شود؛ 500 در Swagger نادرست است |

**اثر دیتابیس:** `SELECT ... FROM users WHERE email=$1`؛ بدون نوشتن.
**کاربران seed‌شده:** نمی‌توانند وارد شوند چون رمزشان خالی است (مقایسه bcrypt شکست می‌خورد). ابتدا ثبت‌نام کنید.
**Swagger:** مسیر/متد مطابق است؛ Swagger می‌گوید 500 و پاسخ شیء؛ حداکثر واقعی 400 است و `data` رشته است.

## ۳. GET /v1/health

payload زنده‌بودن عمومی.

**پاسخ:** HTTP 200، `dto.HealthStatusPayload` (در `response.Response` پیچیده نشده).

```json
{"status":"ok","version":"0.0.1","environment":"development"}
```

PostgreSQL را کوئری **نمی‌کند**، مهاجرت‌ها را بررسی نمی‌کند و سلامت وابستگی را گزارش نمی‌دهد. قطعی دیتابیس
همچنان می‌تواند `ok` برگرداند.

## ۴. GET /v1/users

فهرست همه کاربران. JWT لازم است.

**درخواست:** بدون پارامتر query. بدون صفحه‌بندی.

**موفقیت:** HTTP 200، کد `1005`، `data` آرایه‌ای از `storage.User`:

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

**نکته تأییدشده:** `role_id` برای هر سطر فهرست `0` است چون `UsersStorage.List` مقدار `role_id` را انتخاب
نمی‌کند. برای دریافت نقش واقعی از `GET /v1/users/{userID}` استفاده کنید.

**کدهای وضعیت:** 200؛ 401 (توکن غایب/بدشکل یا شکست جست‌وجوی کاربر)؛ 500 کد `5009` برای خطاهای SQL.
`GET /v1/users/{userID}` معنای خطای متفاوتی دارد.

**Swagger:** مسیر/متد مطابق؛ هیچ الزام امنیتی اعلام نشده؛ schema پاسخ `storage.User` را نشان می‌دهد.

## ۵. GET /v1/users/{userID}

دریافت یک کاربر. JWT لازم است.

**پارامتر مسیر:** `userID` عدد صحیح (`int64`).

**موفقیت:** HTTP 200، کد `1006`، `data` یک `storage.User` با `role_id` واقعی:

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

**خطاها:** HTTP 500، کد `5010` هم برای شناسه غیرعددی و هم برای `sql.ErrNoRows` (`"not found"`). هندلر
`consts.NOT_FOUND` را به 404 نگاشت نمی‌کند. با `/v1/users/abc` و `/v1/users/999999` تأیید شد: هر دو 500
برگرداندند. Swagger 400/404 اعلام می‌کند؛ این وضعیت‌ها تولید نمی‌شوند.

**Swagger:** مسیر/متد مطابق؛ امنیت غایب؛ قرارداد خطا نادرست.

## ۶. PUT /v1/users/{userID}/follow

دنبال‌کردن کاربر مشخص‌شده در مسیر. JWT لازم است.

**پارامتر مسیر:** `userID` شناسه کاربر هدف.

**بدنه درخواست:** هیچ. هندلر `ReadJson` را صدا نمی‌زند؛ هر بدنه‌ای نادیده گرفته می‌شود.

**اثر دیتابیس:**

```sql
INSERT INTO followers(user_id, follower_id) VALUES ($1, $2)
-- $1 = کاربر احراز هویت‌شده، $2 = هدف مسیر
```

**موفقیت:** HTTP 200، کد `1007`، `data: null`.

```json
{"success":true,"message":"user followed successfully","internal_code":1007,"data":null}
```

**خطاها:**

| وضعیت | کد | شرط |
|---:|---:|---|
| 200 | 1007 | درج موفق |
| 401 | 4019/4020/4021/4022 | احراز هویت غایب/بدشکل/نامعتبر یا شکست جست‌وجوی کاربر |
| 500 | 5020 | شناسه مسیر عدد صحیح نیست |
| 500 | 5013 | خطای دیتابیس: دنبال‌کردن تکراری (PK)، هدف ناشناخته (FK) و غیره |
| 404 | 4006 | کد در هندلر وجود دارد اما شاخه غیرقابل‌دسترسی است: storage خطاهای خام DB برمی‌گرداند، نه `consts.NOT_FOUND` |

هیچ بررسی دنبال‌کردن خود وجود ندارد. دنبال‌کردن خودتان `(user_id=self, follower_id=self)` درج می‌کند و
موفق می‌شود مگر اینکه جفت از قبل وجود داشته باشد.

**Swagger:** `POST` با بدنه الزامی `dto.FollowPayload` را مستند می‌کند. مسیر واقعی `PUT` است و بدنه را نادیده
می‌گیرد. `dto.FollowPayload` وجود ندارد و تولید Swagger را می‌شکند.

## ۷. PUT /v1/users/{userID}/unfollow

دنبال‌نکردن کاربر مسیر. JWT لازم است.

**بدنه درخواست:** هیچ.

**اثر دیتابیس:**

```sql
DELETE FROM followers WHERE user_id=$1 AND follower_id=$2
```

**موفقیت:** HTTP 200، کد `1008`، `data: null`، حتی اگر رابطه‌ای وجود نداشته باشد. `RowsAffected` بررسی
نمی‌شود، بنابراین `404` هرگز از مسیر storage برگردانده نمی‌شود.

**خطاها:** 401 احراز هویت؛ 500 کد `5021` برای شناسه مسیر غیرعددی؛ 500 کد `5015` برای خطاهای SQL.
کد 404 اعلام‌شده `4008` غیرقابل‌دسترسی است.

**Swagger:** `DELETE` با بدنه (`dto.UnFollowPayload`) را مستند می‌کند؛ واقعیت `PUT` بدون بدنه است و نوع
وجود ندارد.

## ۸. GET /v1/users/{userID}/feeds

بازگرداندن فید یک کاربر. JWT لازم است.

**پارامتر مسیر:** `userID` (int64). هندلر **بررسی نمی‌کند** که با شناسه کاربر احراز هویت‌شده برابر است. هر
توکن معتبر می‌تواند فید هر کاربری را درخواست کند.

**پارامترهای query** (`storage.PaginatedQuery`):

| پارامتر | نوع | پیش‌فرض | اعتبارسنجی | رفتار واقعی |
|---|---|---:|---|---|
| `limit` | int | 20 | ۱..۱۰۰ | مقادیر غیرعددی نامعتبر بی‌صدا نادیده گرفته می‌شوند؛ مقادیر خارج از محدوده اعتبارسنجی را شکست می‌دهند (400). |
| `offset` | int | **10** | >=0 | پیش‌فرض ۱۰ است، نه ۰. مقادیر غیرعددی نامعتبر بی‌صدا نادیده گرفته می‌شوند. |
| `page` | int | — | — | با وجود مستندسازی در Swagger **خوانده نمی‌شود**. |
| `sort` | string | `desc` | `ASC DESC asc desc` | پذیرفته می‌شود اما بی‌اثر است: SQL از `ORDER BY p.created_at, $2` استفاده می‌کند، بنابراین جهت تغییر نمی‌کند. |
| `search` | string | خالی | حداکثر ۱۰۰ | عنوان/محتوا را با `ILIKE` فیلتر می‌کند؛ پارامتری. |
| `since` | string | خالی | هیچ | با `time.DateTime` تجزیه می‌شود؛ در SQL استفاده نمی‌شود. مقادیر نامعتبر خالی می‌شوند. |
| `until` | string | خالی | هیچ | ذخیره می‌شود اما استفاده نمی‌شود؛ تجزیه/قالب‌بندی نمی‌شود. |

**موفقیت:** HTTP 200، کد `1009`، `data` برابر `[]PostWithMetadata`:

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

فقط `user.username` توسط SQL فید پر می‌شود؛ سایر فیلدهای کاربر مقدار صفر دارند. نظرها بارگذاری نمی‌شوند؛
شمارش تجمیعی به‌صورت `comment_count` ارائه می‌شود.

**خطاها:** 400 کد `4009` برای خطاهای تجزیه query (اعتبارسنجی رشته sort/limit/offset در هندلر)؛ 400 کد `4010`
برای شکست `Validate.Struct`؛ 500 کد `5016` برای شناسه مسیر غیرعددی؛ 500 کد `5017` برای خطاهای SQL؛ 401
احراز هویت.

**مشکلات صحت تأییدشده:** تورم شمارش، مرتب‌سازی بی‌اثر، offset پیش‌فرض ۱۰، since/until استفاده‌نشده، دسترسی
بین‌کاربری. به [06. عملیات دیتابیس](06-database-operations.md#تحلیل-کوئری-فید) مراجعه کنید.

**Swagger:** مسیر `/users/{userID}/feed` (مفرد) است در حالی که مسیر واقعی `/feeds` است؛ `page` مستند شده اما
استفاده نمی‌شود؛ امنیت غایب.

## ۹. GET /v1/posts

فهرست همه پست‌ها با نظرها. JWT لازم است.

**درخواست:** بدون پارامتر query، بدون صفحه‌بندی.

**موفقیت:** HTTP 200، کد `1001`، `data` برابر `[]*storage.Post`:

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

`Post.User` یک شیء با مقدار صفر است چون کوئری پست کاربران را join نمی‌کند. `user` نظرها فقط `id` و
`username` را دارد که `CommentsStorage.GetByPostID` پر می‌کند.

**خطاها:** 500 کد `5001` برای کوئری فهرست پست؛ 500 کد `5009` برای کوئری نظرها؛ 401 احراز هویت.
endpoint به‌ازای هر پست یک کوئری نظر اجرا می‌کند (N+1). به [جریان برنامه](04-application-flow.md) مراجعه کنید.

**Swagger:** مسیر/متد مطابق؛ امنیت غایب؛ schema پاسخ عمومی است.

## ۱۰. POST /v1/posts

ساخت پست برای کاربر احراز هویت‌شده. JWT لازم است.

**بدنه درخواست** (`dto.CreatePostPayload`):

| فیلد | نوع | الزامی | اعتبارسنجی |
|---|---|---:|---|
| `title` | string | بله | `required,max=255` |
| `content` | string | بله | `required` (بدون حداکثر) |
| `tags` | string[] | خیر | هیچ |

**مثال:**

```json
{"title":"Hello","content":"World","tags":["intro","go"]}
```

**موفقیت:** HTTP **200** (Swagger می‌گوید 201)، کد `1002`، `data` برابر `storage.Post` ساخته‌شده:

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

**خطاها:**

| وضعیت | کد | شرط |
|---:|---:|---|
| 500 | 5002 | خطای decode JSON (شامل بدنه خالی) |
| 500 | 5006 | شکست اعتبارسنجی (با وجود 400 در Swagger) |
| 500 | 5003 | خطای SQL/DB |
| 401 | — | میان‌افزار JWT |

**اثر دیتابیس:** درج در `posts`؛ `user_id` از context JWT می‌آید، نه بدنه.

**Swagger:** می‌گوید 201 و 400؛ واقعیت 200 و 500 برای این شکست‌ها است. امنیت غایب.

## ۱۱. GET /v1/posts/{postID}

دریافت پست با نظرها. JWT لازم است.

**موفقیت:** HTTP 200، کد `1003`، `data` یک `storage.Post` با `comments` پر و `user` با مقدار صفر.

**خطاها:** 404 کد `4001` برای `sql.ErrNoRows`؛ 500 کد `5004` برای شناسه غیرعددی؛ 500 کد `5005` برای سایر
خطاهای کوئری پست؛ 500 کد `5009` برای خطاهای کوئری نظر؛ 401 احراز هویت. تأیید شد `/v1/posts/999999` پاسخ
404 با `{"success":false,"message":"post not found","internal_code":4001,"data":null}` می‌دهد.

## ۱۲. PATCH /v1/posts/{postID}

ویرایش جزئی پست. JWT + `UserPostAccessMiddleware`.

**بدنه درخواست** (`dto.UpdatePostPayload`):

| فیلد | نوع | الزامی | اعتبارسنجی | یادداشت |
|---|---|---:|---|---|
| `title` | اشاره‌گر string | خیر | `omitempty,max=255` | در صورت وجود اعمال می‌شود |
| `content` | اشاره‌گر string | خیر | `omitempty` | در صورت وجود اعمال می‌شود |

هر دو فیلد اختیاری‌اند؛ یک شیء خالی معتبر است و ویرایشی با مقادیر بدون تغییر انجام می‌دهد. فیلدهای ناشناخته
رد می‌شوند.

**موفقیت:** HTTP 200، کد `1004`، `data: null`.

**قواعد دسترسی (واقعی):**

- ابتدا پست دریافت می‌شود؛ `sql.ErrNoRows` → 404 کد `4002`.
- اگر `user.ID == post.UserID`، هندلر اجرا می‌شود.
- اگر مالک نباشد و `user.RoleID == 3`، هندلر اجرا می‌شود **و سپس میان‌افزار نیز `UnauthorizedResponse` را
  صدا می‌زند**. ویرایش ادمین ممکن است اعمال شود، اما پاسخ خراب می‌شود (ابتدا پاسخ موفقیت، سپس تلاش نوشتن 401).
  این یک نقص تأییدشده در مسیر کد است؛ در زمان اجرا آزمایش نشد چون دیتابیس توسعه هیچ کاربر ادمین ندارد.
- اگر مالک نباشد و ادمین نباشد، 401 کد `4023` (`"access denied"`) برمی‌گرداند. مدیران (`role_id=2`) با وجود
  توضیح نقش، رد می‌شوند.
- مالکان بدون توجه به نقش مجازند.

**سایر خطاها:** 400 کد `4003` برای decode JSON؛ 400 کد `4004` برای اعتبارسنجی؛ 500 کد `5007` برای شناسه مسیر
غیرعددی؛ 500 کد `5008` برای خطاهای SQL دریافت/ویرایش.

**اثر دیتابیس:** `UPDATE posts SET title=$1, content=$2 WHERE id=$3`؛ `updated_at` تغییر نمی‌کند.
**Swagger:** مسیر/متد مطابق؛ امنیت غایب؛ 400/404 مستند شده، اما رفتار ادمین/نقش مستند نشده است.

## ۱۳. DELETE /v1/posts/{postID}

حذف پست. JWT + میان‌افزار مالک/ادمین.

**موفقیت:** HTTP 200، کد `1004`، `data: null`.

**قواعد دسترسی:** همان منطق مالکیت/ادمین و همان باگ پاسخ ادمین در PATCH. غیرمالک/غیرادمین 401 کد `4023`
می‌گیرد.

**خطاها:** 404 کد `4001` اگر پست غایب باشد؛ 500 کد `5007` برای شناسه غیرعددی؛ 500 کد `5008` برای خطای حذف.
حذف پستی که نظر دارد روی `fk_post_comment` شکست می‌خورد و 500 با خطای خام PostgreSQL برمی‌گرداند. یک تست
SQL rollback‌شده نقض FK را تأیید کرد؛ بیشتر پست‌های seed‌شده نظر دارند.

**اثر دیتابیس:** `DELETE FROM posts WHERE id=$1`؛ نظرها حذف نمی‌شوند و آبشار ندارند.
**Swagger:** مسیر/متد مطابق؛ امنیت غایب؛ رفتار FK مستند نشده است.

## ۱۴. endpointهای Swagger

| مسیر | احراز هویت | رفتار |
|---|---|---|
| `GET /v1/swagger/index.html` | عمومی | HTML Swagger UI، وضعیت 200 |
| `GET /v1/swagger/doc.json` | عمومی | JSON تولیدشده Swagger توسط `http-swagger`، وضعیت 200 |
| `GET /v1/swagger/` | عمومی | redirect به `index.html` |

پیکربندی UI به این صورت تولید می‌شود:

```text
url: "127.0.0.1:8000/swagger/doc.json"   # مثال با SERVER_URL=127.0.0.1:8000
```

مسیر واقعی هندلر `/v1/swagger/doc.json` است؛ آدرس پیکربندی‌شده پیشوند `/v1` و scheme ندارد. UI ارائه می‌شود،
اما انتظار می‌رود بارگذاری spec از آن آدرس شکست بخورد. این با بررسی HTML تولیدشده تأیید شد؛ مرورگر headless
اجرا نشد.

## ۱۵. رفتار پیش‌فرض 404/405

مسیرهای بدون تطبیق و متدهای پشتیبانی‌نشده به هندلرهای پیش‌فرض chi می‌رسند. آن‌ها متن ساده برمی‌گردانند
(معمولاً `404 page not found` یا `405 method not allowed`)، **نه** پاکت JSON. مثال‌های تأییدشده:

- `GET /v1/users/1/feed` → 404 (مسیر واقعی `/feeds` است).
- `GET /v1/users/2/follow` → 405 (مسیر `PUT` است؛ GET وجود ندارد).
- `/v1/users/1/feeds/` → 404.

کلاینت‌ها نباید فرض کنند هر خطا JSON است.

## مقایسه Swagger

فایل‌های `docs/swagger.json` و `docs/swagger.yaml` موجود از نظر معنایی با هم یکسان‌اند (در ممیزی تجزیه و
مقایسه شدند)، اما با پیاده‌سازی مطابقت ندارند.

| رفتار واقعی | Swagger | تفاوت |
|---|---|---|
| `POST /v1/auth/register` | غایب | endpoint گمشده. |
| `POST /v1/auth/token` با 200 و `data` رشته | 200/400/500، شیء عمومی | مجموعه وضعیت و schema اشتباه؛ 500 تولید نمی‌شود. |
| `GET /v1/health` | حاضر | مطابق. |
| `GET /v1/users` | حاضر، بدون security | الزام احراز هویت غایب. |
| `GET /v1/users/{userID}` برای not found 500 می‌دهد | 400/404 | قرارداد وضعیت اشتباه. |
| `GET /v1/users/{userID}/feeds` | `/feed`، `page` | مسیر و پارامترهای query اشتباه. |
| `PUT /v1/users/{userID}/follow`، بدون بدنه | `POST`، بدنه الزامی `dto.FollowPayload` | متد/بدنه اشتباه؛ DTO غایب. |
| `PUT /v1/users/{userID}/unfollow`، بدون بدنه | `DELETE`، بدنه الزامی `dto.UnFollowPayload` | متد/بدنه اشتباه؛ DTO غایب. |
| `POST /v1/posts` با 200؛ اعتبارسنجی 500 | 201؛ اعتبارسنجی 400 | قرارداد وضعیت اشتباه. |
| `PATCH/DELETE /v1/posts/{postID}` نیازمند مالک/ادمین | بدون security | سیاست احراز هویت غایب. |
| Swagger UI در `/v1/swagger/*` | مستند نشده | مسیر UI غایب. |
| بدون `security` در سطح operation | فقط تعریف `ApiKeyAuth` | Swagger UI توکن bearer را نمی‌فرستد. |

`docs/swagger.json`/`yaml` همچنین `host` یا `info.version` ندارند چون به‌صورت استاتیک تولید شده‌اند؛
`docs.SwaggerInfo` این‌ها را در زمان اجرا پر می‌کند. `swag init` اکنون در `dto.FollowPayload` شکست می‌خورد،
بنابراین فایل‌های کهنه بدون اصلاح حاشیه‌نویسی‌ها/DTOها قابل تولید مجدد نیستند.

ماتریس کامل در [مرور سازگاری API](../reports/api-consistency-review.md) آمده است.

## اسناد مرتبط

- [احراز هویت و مجوزدهی](08-authentication-authorization.md)
- [مدیریت خطا](10-error-handling.md)
- [عملیات دیتابیس](06-database-operations.md)
- [عیب‌یابی](16-troubleshooting.md)
