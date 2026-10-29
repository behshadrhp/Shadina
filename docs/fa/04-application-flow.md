# 04. جریان برنامه

## جریان راه‌اندازی

`cmd/api/main.go` تنها نقطه ورود طولانی‌مدت است. توالی به این صورت است:

```text
init()
  └── utils.Validate = validator.New(WithRequiredStructEnabled)

main()
  ├── os.MkdirAll("logs", 0755)
  ├── zap.NewProductionConfig()
  │     ├── OutputPaths       = logs/app.log
  │     └── ErrorOutputPaths  = logs/error.log
  ├── zap.Must(...).Sugar()
  ├── godotenv.Load()                    # .env در پوشه کاری جاری
  ├── env.GetString/GetInt(...)          # پیکربندی با fallbackها
  ├── db.NewDB(driver, addr, pool...)    # sql.Open + pool + PingContext(10s)
  ├── storage.NewPostgresql(db)
  ├── auth.NewJWTAuthenticator(secret, "Shadina", "Shadina")
  ├── application{config, storage, logger, authenticator}
  ├── app.mount()                        # روتر chi
  └── app.run(&mux)
        ├── docs.SwaggerInfo.Version/Host/BasePath
        ├── http.Server{ReadTimeout, WriteTimeout, IdleTimeout}
        └── srv.ListenAndServe()
```

جزئیات مهم:

- `godotenv.Load()` متغیرهای محیطی از قبل export‌شده را **بازنویسی نمی‌کند**. مقادیری که از قبل در محیط فرآیند
  وجود دارند اولویت دارند.
- اگر `.env` وجود نداشته باشد یا خوانده نشود، راه‌اندازی با `error loading .env file` خارج می‌شود؛ هیچ fallback
  به `.env.example` یا مخزن راز راه دور وجود ندارد.
- `db.NewDB` با `consts.CTX_TIMEOUT` (۱۰ ثانیه) ping می‌کند. مقدار نامعتبر `MAX_IDLE_TIMEOUT` باعث شکست
  راه‌اندازی می‌شود.
- `ListenAndServe` مستقیماً فراخوانی می‌شود. هیچ `signal.Notify`، `Server.Shutdown` یا منطق drain وجود ندارد.
- `log.Fatal(app.run(&mux))` در هر خطای سرور فرآیند را خارج می‌کند؛ `log.Sync()` و `db.Close()` معوق اجرا
  نمی‌شوند.

## روتر و ترتیب میان‌افزار

`cmd/api/router.go` میان‌افزار را به این ترتیب دقیق ثبت می‌کند:

| ترتیب | میان‌افزار | دامنه | رفتار |
|---:|---|---|---|
| ۱ | `middleware.Heartbeat("/ping")` | سراسری | برای GET/HEAD روی `/ping` پاسخ `200 text/plain` با بدنه `.` برمی‌گرداند؛ هندلرهای بعدی را دور می‌زند. |
| ۲ | `middleware.RealIP` | سراسری | `RemoteAddr` را از هدرهای forwarding بازنویسی می‌کند. |
| ۳ | `middleware.Logger` | سراسری | لاگر common-log-format در chi به stdout. |
| ۴ | `middleware.Recoverer` | سراسری | panicها را بازیابی و 500 برمی‌گرداند؛ stack را لاگ می‌کند. |
| ۵ | `app.LoggerMiddleware` | سراسری | WrapResponseWriter + لاگ درخواست zap به `logs/app.log`. |
| ۶ | `app.JWTAuthMiddleware` | `/v1/users/*`، `/v1/posts/*` | توکن bearer را تجزیه، کاربر را بارگذاری و `*storage.User` را در context درخواست ذخیره می‌کند. |
| ۷ | `app.UserPostAccessMiddleware` | PATCH/DELETE `/v1/posts/{postID}` | پست را بارگذاری، مالک/ادمین را بررسی و سپس next را صدا می‌زند. |

لاگر سفارشی بعد از `Recoverer` در ترتیب ثبت اجرا می‌شود، اما چون میان‌افزار هندلر پایین‌دست را می‌پیچد، کد
پس از هندلر در `LoggerMiddleware` وقتی panic پایین‌دست باز می‌شود اجرا نمی‌شود؛ بازیابی توسط `Recoverer` در
بیرون آن انجام می‌شود. این یک شکاف ظریف لاگ/عیب‌یابی است، نه نقض قرارداد عملکردی API.

## درخت مسیرها

```text
/ping                                  GET        (میان‌افزار Heartbeat)
/v1
├── /swagger/*                         GET        (Swagger UI + doc.json)
├── /auth
│   ├── /register                      POST       عمومی
│   └── /token                         POST       عمومی
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
│       ├── /                          DELETE     JWT + مالک/ادمین
│       └── /                          PATCH      JWT + مالک/ادمین
└── /health                            GET        عمومی
```

Chi در تست محلی هم `/v1/users` و هم `/v1/users/` را برای مجموعه کاربران می‌پذیرد؛ منبع روتر `r.Get("/")` را
داخل `r.Route("/users", ...)` اعلام می‌کند. همین رفتار برای `/v1/posts` مشاهده شد. مسیرهای متعارف مستندشده
اسلش انتهایی را حذف می‌کنند.

## چرخه عمر درخواست بر اساس دسته endpoint

### ۱. ثبت‌نام عمومی

```text
POST /v1/auth/register
  → ReadJson (حداکثر ۱ MiB، DisallowUnknownFields)
  → Validate.Struct(UserRegisterPayload)
  → bcrypt.GenerateFromPassword(password, DefaultCost)
  → UsersStorage.Create (INSERT ... RETURNING id, created_at, updated_at)
  → SuccessResponse 200 {success, message, internal_code:1010, data: User}
```

`User.Password` تگ `json:"-"` دارد، بنابراین هش سریالایز نمی‌شود. `RoleID` به `1` hard-code شده است.
خطاهای خواندن `400` با کد `4011`؛ اعتبارسنجی `400` با کد `4012`؛ خطاهای هش/درج `500` با کد `5018` و متن
خام خطا برای شکست درج برمی‌گردانند.

### ۲. صدور توکن عمومی

```text
POST /v1/auth/token
  → ReadJson + Validate
  → UsersStorage.GetByEmail (SELECT ... WHERE email=$1)
  → User.ComparePassword (bcrypt)
  → claims: sub=user.ID، exp=now+72h، iat، nbf، iss="Shadina"، aud="Shadina"
  → JWTAuthenticator.GenerateToken (HS256)
  → SuccessResponse 200 {data: "<token string>"}
```

شکست‌ها عمدتاً به `400` با کدهای `4013`–`4018` نگاشت می‌شوند؛ خطاهای دیتابیس به‌صورت عمومی
`"invalid request"` با `400` (نه 500) برگردانده می‌شوند. به [07. مرجع API](07-api-reference.md) مراجعه کنید.

### ۳. درخواست محافظت‌شده با JWT

```text
GET /v1/users
  → JWTAuthMiddleware
      → نیاز به هدر Authorization
      → نیاز به دقیقاً "Bearer <token>" (یک فاصله، حساس به حروف "Bearer")
      → ValidateToken: HS256، issuer، audience، exp الزامی
      → claims["sub"] با "%.f" قالب‌بندی و به int64 تجزیه می‌شود
      → UsersStorage.GetByID
      → context.WithValue(ctx, consts.USER_CTX, user)
  → هندلر
```

شکست‌های احراز هویت:

| شرط | وضعیت | کد داخلی |
|---|---|---|
| هدر غایب | 401 | 4019 |
| هدر بدشکل / نه `Bearer ` | 401 | 4020 |
| توکن نامعتبر، منقضی، issuer/audience/امضای اشتباه | **500** | 5019 |
| `sub` قابل تجزیه نیست | 401 | 4021 |
| جست‌وجوی کاربر شکست می‌خورد | 401 | 4022 |

رفتار 500 برای توکن نامعتبر پیاده‌سازی شده است؛ این یک اشتباه مستندسازی نیست.

### ۴. دنبال‌کردن / دنبال‌نکردن

```text
PUT /v1/users/{userID}/follow
  → میان‌افزار JWT
  → تجزیه userID مسیر (int64)
  → FollowersStorage.Follow(ctx, authenticatedUser.ID, targetID)
  → INSERT INTO followers(user_id, follower_id) VALUES ($1,$2)
  → SuccessResponse 200، data null
```

هیچ بدنه درخواستی خوانده نمی‌شود، هرچند Swagger بدنه الزامی اعلام کرده است. ستون `user_id` جدول، دنبال‌کننده
عمل‌کننده و `follower_id` کاربر دنبال‌شده است. نام‌ها غیرشهودی‌اند؛ به
[05. طراحی دیتابیس](05-database-design.md) مراجعه کنید.

`unfollow` با همان جفت حذف می‌کند و `RowsAffected` را بررسی نمی‌کند، بنابراین رابطه غایب همچنان 200 برمی‌گرداند.

### ۵. فید

```text
GET /v1/users/{userID}/feeds?limit=&offset=&sort=&search=&since=&until=
  → میان‌افزار JWT (بررسی نمی‌کند که userID مسیر == کاربر احراز هویت‌شده)
  → تجزیه userID مسیر
  → CreatePaginatedQuery(20, 10, "desc")
  → تجزیه پارامترهای query (دغدغه‌های HTTP در internal/storage/pagination.go)
  → Validate.Struct(pq)
  → PostsStorage.GetFeed(ctx, pathUserID, pq)
  → SuccessResponse 200، data: []PostWithMetadata
```

پیش‌فرض‌های واقعی `limit=20`، `offset=10`، `sort=desc` هستند. پارامتر `page` در Swagger خوانده نمی‌شود.
`limit`/`offset` عددی نامعتبر بی‌صدا نادیده گرفته می‌شوند (`Parse` خطا برنمی‌گرداند)؛ `sort` نامعتبر با
اعتبارسنجی گرفته می‌شود. `since`/`until` تجزیه و ذخیره می‌شوند اما در SQL استفاده نمی‌شوند. `sort` به‌صورت
`$2` در `ORDER BY p.created_at, $2` پاس می‌شود؛ PostgreSQL آن را ثابت می‌داند، بنابراین صعودی/نزولی در عمل
همان ترتیب را برمی‌گردانند (محلی تأیید شد).

### ۶. فهرست و دریافت پست

```text
GET /v1/posts
  → JWT
  → PostsStorage.List (SELECT همه پست‌ها، بدون ORDER BY، بدون صفحه‌بندی)
  → برای هر پست: CommentsStorage.GetByPostID
  → SuccessResponse 200 data: []*Post
```

این الگوی کلاسیک کوئری N+1 است: یک کوئری برای همه پست‌ها به‌علاوه یک کوئری نظر برای هر پست (۲۰۳ پست در
دیتابیس توسعه → تا ۲۰۴ کوئری). `Post.User` توسط `List` یا `GetByID` بارگذاری نمی‌شود؛ پاسخ‌ها یک شیء
`user` با مقدار صفر دارند. `Comments` بارگذاری و جاسازی می‌شوند.

### ۷. ساخت/ویرایش/حذف پست

- **ساخت:** کاربر JWT تبدیل به `post.UserID` می‌شود؛ `CreatePostPayload` اعتبارسنجی می‌شود؛ درج شناسه و
  زمان‌ها را برمی‌گرداند؛ موفقیت HTTP **200** است (Swagger می‌گوید 201).
- **ویرایش:** JWT + `UserPostAccessMiddleware`؛ دریافت پست، نیاز به مالک یا `RoleID == 3`، خواندن/اعتبارسنجی
  عنوان/محتوا اختیاری، تغییر، `UPDATE posts SET title=$1, content=$2 WHERE id=$3`. مقدار `updated_at` تغییر
  نمی‌کند. شاخه ادمین باگ پاسخ دارد (پایین را ببینید).
- **حذف:** JWT + میان‌افزار مالک/ادمین؛ `DELETE FROM posts WHERE id=$1`. نظرها با کلید خارجی بدون
  `ON DELETE CASCADE` به پست ارجاع می‌دهند، بنابراین حذف پستی که نظر دارد در دیتابیس شکست می‌خورد و به‌صورت
  HTTP 500 برگردانده می‌شود. یک تست rollback تراکنش روی DB محلی نقض FK را تأیید کرد.

### ۸. سلامت و Swagger

- `GET /v1/health` مقدار `dto.HealthStatusPayload{status:"ok", version, environment}` را می‌سازد و مستقیماً با
  `utils.WriteJson` می‌نویسد. PostgreSQL را بررسی نمی‌کند.
- `GET /ping` توسط Heartbeat در chi رهگیری و `.` برگردانده می‌شود.
- `GET /v1/swagger/*` توسط `http-swagger/v2` ارائه می‌شود. `index.html` وضعیت 200 دارد و
  `/v1/swagger/doc.json` توسط هندلر ارائه می‌شود. آدرس spec پیکربندی‌شده در UI برابر
  `SERVER_URL + "/swagger/doc.json"` است (مثلاً `127.0.0.1:8000/swagger/doc.json`) که پیشوند `/v1` و scheme
  ندارد؛ مسیر صحیح هندلر `/v1/swagger/doc.json` است. UI ارائه می‌شود، اما انتظار می‌رود بارگذاری spec از آدرس
  پیکربندی‌شده شکست بخورد. این با خواندن پیکربندی تولیدشده UI تأیید شد، نه با اجرای مرورگر headless.

## تبدیل داده‌ها

| مرحله | شکل |
|---|---|
| JSON ورودی HTTP | DTO درخواست (`dto.*Payload`) |
| هندلر | ساختار دامنه/storage (`storage.User`، `storage.Post`، `storage.Comment`) |
| Storage | پارامترهای SQL؛ `pq.Array` برای برچسب‌ها |
| سطرهای دیتابیس | Scan به ساختارهای storage |
| پاسخ | `response.Response{Success, Message, InternalCode, Data}` |
| JSON | `data` یک ساختار، آرایه، رشته یا `null` است |

هیچ نگاشت‌کننده پاسخ وجود ندارد. نام فیلدهای JSON مستقیماً از تگ‌های struct روی انواع storage می‌آیند. به
همین دلیل فیلدهای رابطه‌ای با مقدار صفر (`user`، `comments`) در پاسخ‌ها ظاهر می‌شوند.

## Context و لغو

- هر هندلر از `r.Context()` استفاده می‌کند؛ فراخوانی‌های `database/sql` آن را دریافت می‌کنند.
- `FollowersStorage`، `CommentsStorage.Create` و `PostsStorage.GetFeed` علاوه بر این، context را با
  `context.WithTimeout(ctx, consts.CTX_TIMEOUT)` (۱۰ ثانیه) می‌پیچند.
- `UsersStorage`، `PostsStorage.Create/List/GetByID/Delete/Update` و `CommentsStorage.GetByPostID` از context
  درخواست بدون تایم‌اوت صریح DB استفاده می‌کنند. آن‌ها همچنان لغو کلاینت و تایم‌اوت‌های سرور را به ارث می‌برند.
- `db.NewDB` از context ping راه‌اندازی ۱۰ ثانیه‌ای استفاده می‌کند.

## رفتار خاموشی

هیچ مسیر خاموشی آرام وجود ندارد. با `SIGINT`/`SIGTERM`، runtime Go فرآیند را پایان می‌دهد؛ درخواست‌های در
جریان تخلیه نمی‌شوند و پاک‌سازی معوق DB/لاگ اجرا نمی‌شود. این برای استقرار و flush لاگ اهمیت دارد. به
[12. Docker و استقرار](12-docker-deployment.md) و [مرور معماری](../reports/architecture-review.md) مراجعه کنید.

## اسناد مرتبط

- [معماری](02-architecture.md)
- [مرجع API](07-api-reference.md)
- [احراز هویت و مجوزدهی](08-authentication-authorization.md)
- [عملیات دیتابیس](06-database-operations.md)
- [لاگ‌نویسی و مشاهده‌پذیری](11-logging-observability.md)
