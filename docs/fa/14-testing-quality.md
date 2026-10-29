# 14. تست و کیفیت

## فهرست تست

هیچ **فایل تست Go** در مخزن وجود ندارد. جست‌وجوی کل مخزن برای `*_test.go` نتیجه‌ای برنگرداند. هیچ mock، fake،
تست یکپارچگی، تست API، تست دیتابیس، golden file یا fixture تست وجود ندارد.

| نوع تست | حاضر؟ | شاهد |
|---|---|---|
| تست واحد | خیر | بدون فایل `*_test.go` |
| تست یکپارچگی | خیر | بدون test harness/testcontainers |
| تست HTTP/API | خیر | بدون استفاده از `httptest` |
| تست دیتابیس | خیر | بدون راه‌اندازی DB تست |
| تست مهاجرت | خیر | بدون اسکریپت تست مهاجرت |
| تست بار/کارایی | خیر | بدون benchmark |
| تست امنیت | خیر | بدون fuzzing، تست auth یا scanner |
| CI | خیر | بدون `.github/`، `.gitlab-ci.yml` و غیره |

## فرمان‌های واقعاً اجراشده

همه فرمان‌های زیر در ۲۰۲۶-۱۰-۲۹ در ریشه مخزن با Go 1.25.4 و کانتینر PostgreSQL محلی در دسترس اجرا شدند.
نتایج دقیقاً گزارش می‌شوند.

| فرمان | نتیجه | تفسیر |
|---|---|---|
| `go build ./...` | **موفق** (خروج ۰) | همه بسته‌های دارای فایل Go کامپایل می‌شوند. `cmd/migrate` چون فایل Go ندارد رد می‌شود. |
| `go build ./cmd/migrate` | **شکست** (خروج ۱) | `no Go files in .../cmd/migrate`؛ مهاجرت‌ها به CLI خارجی وابسته‌اند. |
| `go vet ./...` | **موفق** (خروج ۰) | بدون تشخیص vet. |
| `go test ./...` | **موفق بدون تست** | هر بسته `[no test files]` گزارش می‌کند. این شاهد صحت نیست. |
| `go test -race ./...` | **موفق بدون تست** | هیچ باینری تستی ساخته نشد، بنابراین race detector چیزی برای اجرا نداشت. |
| `go test -cover ./...` | **پوشش ۰.۰٪** | بسته‌های دارای statement مقدار `coverage: 0.0%` گزارش می‌کنند؛ برخی `[no test files]`. |
| `go mod verify` | **موفق** | `all modules verified`. |
| `go mod tidy -diff` | **diff یافت شد** | وابستگی مستقیم استفاده‌نشده `github.com/golang-jwt/jwt v3.2.2+incompatible` را نشان می‌دهد؛ هیچ فایلی تغییر نکرد. |
| `gofmt -l .` | **موفق** | در زمان بررسی هیچ فایل غیرقالب‌بندی‌شده گزارش نشد. |
| `golangci-lint run ./...` | **شکست** (خروج ۱) | ۱۲ مشکل: ۹ `errcheck`، ۳ `staticcheck`. |
| `make test` | `go test -v ./...` اجرا می‌کرد | جداگانه لازم نبود؛ همان نتیجه بدون تست. |
| `make lint` | **اجرا نشد** | ابتدا `gofmt -s -w .` را اجرا می‌کند که فایل‌ها را تغییر می‌دهد. به‌جای آن `golangci-lint` مستقیم اجرا شد. |
| `swag init -g ./api/main.go -d cmd,internal` | **شکست** (خروج ۱) | تعریف نوع `dto.FollowPayload` پیدا نمی‌شود؛ تولید Swagger مسدود است. |
| `make gen-docs` | **شکست** | همان علت ریشه‌ای؛ هدف make دستور شکست‌خورده `swag init` را صدا می‌زند. |
| `make seed` | **شکست** | هدف به `cmd/migrate/seed/main.go` گمشده اشاره می‌کند. |
| `go run ./cmd/seed` | **اجرا نشد** | fixture غیر-idempotent در DB توسعه درج می‌کرد؛ در این بررسی فقط-مستندات اجرا نشد. |
| `govulncheck ./...` | **اجرا نشد** | `govulncheck` نصب نیست؛ هیچ اسکن آسیب‌پذیری وابستگی انجام نشد. |
| تست یکپارچگی/API | **اجرا نشد** | هیچ مجموعه تستی وجود ندارد. بررسی‌های دستی فقط-خواندنی API انجام شد (پایین). |

## یافته‌های لینت

`golangci-lint run ./...` گزارش داد:

**errcheck (۹)**

- `cmd/api/health.go`: بازگشت `utils.WriteJson` بررسی نشده.
- `cmd/api/main.go`: `log.Sync`، `db.Close` بررسی نشده.
- `cmd/seed/main.go`: `database.Close` بررسی نشده.
- `internal/db/seed.go`: `tx.Commit` بررسی نشده.
- `internal/storage/comments.go`: `row.Close` بررسی نشده.
- `internal/storage/posts.go`: `row.Close` / `rows.Close` بررسی نشده (دو یافته).
- `internal/storage/users.go`: `rows.Close` بررسی نشده.

**staticcheck (۳)**

- `cmd/api/middleware.go`: `SA1029` — رشته built-in به‌عنوان کلید `context.WithValue`.
- `internal/consts/consts.go`: `ST1012` — `NOT_FOUND` باید `ErrFoo` نام بگیرد.
- `internal/storage/users.go`: `S1008` — ساده‌سازی `ComparePassword` به `return err == nil`.

فرمان lint با کد ۱ خارج می‌شود، بنابراین اکنون CI-clean نیست. `make lint` همچنین درخت را قبل از lint دوباره
قالب‌بندی می‌کند.

## بررسی‌های دستی زمان اجرا

این‌ها بررسی‌های فقط-خواندنی یا rollback‌شده روی API محلی و دیتابیس توسعه بودند. آن‌ها **جایگزین تست
خودکار نیستند**.

| بررسی | نتیجه مشاهده‌شده |
|---|---|
| `GET /v1/health` | 200 `{"status":"ok","version":"0.0.1","environment":"development"}` |
| `GET /ping` | 200، `text/plain`، بدنه `.` |
| `GET /v1/swagger/index.html` | 200 HTML |
| `GET /v1/swagger/doc.json` | 200 JSON |
| `GET /v1/users` بدون توکن | 401 پاکت |
| `GET /v1/users` با توکن نامعتبر | 500 پاکت با خطای parser |
| `POST /v1/auth/token` با کاربر seed | 400 کد 4017 (رمز خالی) |
| `GET /v1/users` با توکن معتبر | 200؛ `role_id` برابر ۰ |
| `GET /v1/posts` با توکن معتبر | 200؛ ۲۰۳ پست در DB توسعه |
| فید `sort=asc` در برابر `sort=desc` | ترتیب یکسان |
| `GET /v1/users/2/feeds` با توکن کاربر ۱ | 200 (بدون بررسی مالکیت) |
| `GET /v1/users/1/feed` | 404 (مسیر واقعی `/feeds` است) |
| `GET /v1/users/2/follow` | 405 (مسیر PUT است) |
| `GET /v1/users/999999` | 500 کد 5010 (`not found`) |
| `GET /v1/users/abc` | 500 کد 5010 (خطای خام `strconv`) |
| `DELETE /v1/posts/{other-user-post}` به‌عنوان مالک پست دیگر | 401 کد 4023 |
| اعتبارسنجی ثبت‌نام | 400 کد 4012 با متن validator |
| فیلد ناشناخته ثبت‌نام | 400 کد 4011 |
| ساخت پست بدون عنوان | 500 کد 5006 |
| ساخت پست با بدنه خالی | 500 کد 5002 (`EOF`) |
| حذف پست دارای نظر (تراکنش SQL، rollback‌شده) | نقض FK `fk_post_comment`؛ تعداد سطر بدون تغییر |
| مقایسه شمارش فید | پست ۱۹: `COUNT(c.id)=16` در برابر `COUNT(DISTINCT c.id)=4` |

## چه چیزی تست نشد

- احراز هویت با کاربر ادمین واقعی (هیچ‌کدام وجود نداشت)؛ نقص میان‌افزار ادمین فقط با بررسی کد تأیید شد.
- موفقیت ثبت‌نام سرتاسری (کاربر درج می‌کرد؛ انجام نشد).
- مسیرهای موفق ساخت/ویرایش/حذف پست (داده را تغییر می‌داد؛ انجام نشد).
- مسیرهای موفق/خطای follow/unfollow (داده را تغییر می‌داد؛ انجام نشد).
- درخواست‌های همزمان، race condition، deadlock یا اتمام pool.
- رفتار کارایی/بار endpointهای فید/فهرست.
- پیکربندی تولید، TLS، CORS، reverse proxy یا استقرار کانتینر.
- آسیب‌پذیری وابستگی‌ها (`govulncheck` در دسترس نبود).
- بارگذاری spec در Swagger UI مرورگر (پیکربندی بررسی شد؛ مرورگر headless اجرا نشد).
- مهاجرت up/down روی دیتابیس دورانداختنی (schema از قبل نسخه ۱۲ بود).

## ارزیابی کیفیت

Build و vet موفق‌اند و کد تمیز کامپایل می‌شود. این تمام شاهد خودکار است. نبود تست یعنی موارد زیر تأیید نشده‌اند:

- ماتریس مجوزدهی (مالک/ادمین/مدیر/دیگر)
- قراردادهای وضعیت/کد خطا
- رفتار صفحه‌بندی/مرتب‌سازی/جست‌وجو
- ایمنی rollback مهاجرت
- idempotency/اتمی‌بودن seed
- صحت SQL تحت شکل‌های داده (باگ شمارش فید نشان می‌دهد چرا مهم است)
- حالت‌های مرزی JWT (منقضی، audience اشتباه، `sub` بدشکل، اختلاف ساعت)
- پایداری schema JSON

## برنامه تست پیشنهادی

### تست واحد (سریع، بدون DB)

- `internal/env`: مقادیر حاضر/غایب/نامعتبر.
- `internal/auth`: تولید توکن، توکن معتبر، منقضی، issuer/audience اشتباه، الگوریتم اشتباه، توکن بدشکل.
- `internal/storage/pagination`: پیش‌فرض‌ها، مقادیر نامعتبر، sort/search/since/until.
- `internal/response`: کدهای وضعیت، پاکت، رفتار encode (ابتدا `log.Fatalf` را جایگزین کنید).
- هش/مقایسه رمز `storage.User` (شامل ورودی بیش از ۷۲ بایت).
- `UserPostAccessMiddleware`: مالک، ادمین، مدیر، دیگر، پست غایب (با storage/auth جعلی).

### تست یکپارچگی (PostgreSQL)

- اجرای مهاجرت‌ها روی دیتابیس دورانداختنی (testcontainers یا schema اختصاصی).
- seed یک fixture شناخته‌شده، نه seed تولید.
- تست هر متد storage، شامل نقض FK و دنبال‌کردن تکراری.
- تست فید با چند دنبال‌کننده/نظر؛ تأیید شمارش متمایز و ترتیب مرتب‌سازی.
- تست حذف پست با/بدون نظر.

### تست HTTP/API

- `httptest.NewServer` با روتر واقعی و DB تست/storage جعلی.
- پوشش هر endpoint، الزام auth، کد وضعیت، کد داخلی و شکل JSON.
- تست‌های قرارداد که رفتار هندلر را با سند OpenAPI تولیدشده مقایسه کنند.

### CI

- `gofmt -l`، `go vet`، `go test -race`، `go test -cover`، `golangci-lint`.
- مهاجرت up/down روی DB دورانداختنی.
- `swag init` و شکست در صورت diff.
- `govulncheck` و به‌روزرسانی وابستگی‌ها.
- آستانه پوشش پس از وجود اولین تست‌ها.

## اسناد مرتبط

- [راهنمای توسعه](13-development-guide.md)
- [مرجع API](07-api-reference.md)
- [عملیات دیتابیس](06-database-operations.md)
- [مرور امنیتی](../reports/security-review.md)
