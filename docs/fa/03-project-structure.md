# 03. ساختار پروژه

## درخت کامل حاشیه‌نویسی‌شده

درخت زیر فایل‌های موجود در زمان بررسی را نشان می‌دهد. پوشه‌های تولیدشده/زمان اجرا علامت‌گذاری شده‌اند.
`logs/`، `tmp/` و `.env` به‌صورت محلی وجود دارند اما توسط Git نادیده گرفته می‌شوند.

```text
.
├── .air.toml                         # پیکربندی hot-reload در Air؛ pre_cmd دستور `make gen-docs` را اجرا می‌کند
├── .env                              # فایل env محلی (نادیده‌گرفته‌شده توسط Git؛ اینجا بازتولید نشده)
├── .example.env                      # env نمونه ردیابی‌شده؛ SERVER_URL و AUTH_SECRET_KAY را ندارد
├── .gitignore                        # .env، logs، tmp و مصنوعات تست/build را نادیده می‌گیرد
├── LICENSE                           # مجوز MIT
├── README.md                         # توضیح دوخطی پروژه
├── bin/
│   └── main                          # باینری کامپایل‌شده ۲۷ مگابایتی ردیابی‌شده (توسط این بررسی تولید نشده)
├── cmd/
│   ├── api/
│   │   ├── api.go                    # ساختارهای application/config و شروع http.Server
│   │   ├── auth.go                   # هندلرهای ثبت‌نام + توکن
│   │   ├── feeds.go                  # هندلر فید
│   │   ├── followers.go              # هندلرهای دنبال‌کردن/دنبال‌نکردن
│   │   ├── health.go                 # هندلر سلامت
│   │   ├── main.go                   # نقطه ورود فرآیند، env/config، DI
│   │   ├── middleware.go             # لاگ درخواست، JWT، مالکیت پست
│   │   ├── posts.go                  # هندلرهای CRUD پست
│   │   ├── router.go                 # روتر/میان‌افزار/مسیرهای chi
│   │   └── uesrs.go                  # هندلرهای خواندن کاربر (غلط املایی نام فایل اصلی است)
│   ├── migrate/
│   │   └── migrations/
│   │       ├── 000001_create_users.{up,down}.sql
│   │       ├── 000002_create_posts.{up,down}.sql
│   │       ├── 000003_alter_post_table.{up,down}.sql
│   │       ├── 000004_create_comments.{up,down}.sql
│   │       ├── 000005_alter_comments_table.{up,down}.sql
│   │       ├── 000006_add_followers_table.{up,down}.sql
│   │       ├── 000007_add_indexes.{up,down}.sql
│   │       ├── 000011_add_roles_table.{up,down}.sql
│   │       └── 000012_alter_users_with_roles.{up,down}.sql
│   │       # هیچ cmd/migrate/main.go وجود ندارد. نسخه‌های 000008–000010 وجود ندارند.
│   └── seed/
│       └── main.go                   # نقطه ورود فرمان seed
├── docker-compose.yaml               # فقط PostgreSQL 18.4
├── docs/
│   ├── docs.go                       # قالب تولیدشده Swagger (swag)
│   ├── swagger.json                  # OpenAPI 2.0 JSON تولیدشده (کهنه)
│   └── swagger.yaml                  # OpenAPI 2.0 YAML تولیدشده (کهنه؛ از نظر معنایی همان JSON)
├── go.mod                            # ماژول github.com/behshadrhp/Shadina، Go 1.25.4
├── go.sum                            # checksum وابستگی‌ها
├── internal/
│   ├── auth/
│   │   ├── auth.go                   # رابط Authenticator
│   │   └── jwt.go                    # JWTAuthenticator (HS256)
│   ├── consts/
│   │   └── consts.go                 # تایم‌اوت‌ها، NOT_FOUND، USER_CTX
│   ├── db/
│   │   ├── db.go                     # sql.Open، تنظیمات pool، PingContext
│   │   └── seed.go                   # تولید داده توسعه + Seed()
│   ├── dto/
│   │   ├── auth.go                   # GenerateTokenPayload
│   │   ├── health.go                 # HealthStatusPayload
│   │   ├── post.go                   # CreatePostPayload، UpdatePostPayload
│   │   └── user.go                   # UserRegisterPayload
│   ├── env/
│   │   └── env.go                    # GetString/GetInt
│   ├── response/
│   │   └── response.go               # پاکت Response + کمک‌تابع‌ها
│   ├── storage/
│   │   ├── comments.go               # CommentsStorage + Comment
│   │   ├── followers.go              # FollowersStorage + Follower
│   │   ├── pagination.go             # PaginatedQuery + تجزیه query HTTP
│   │   ├── posts.go                  # PostsStorage + Post/PostWithMetadata
│   │   ├── storage.go                # نمای Storage + NewPostgresql
│   │   └── users.go                  # UsersStorage + User + کمک‌تابع‌های bcrypt
│   └── utils/
│       ├── json.go                   # ReadJson/WriteJson/WriteJsonError
│       └── validator.go              # singleton validator سراسری
├── logs/
│   ├── app.log                       # لاگ‌های زمان اجرا (نادیده‌گرفته‌شده؛ بازتولید نشده)
│   ├── error.log                     # مقصد خطای Zap (نادیده‌گرفته‌شده؛ ۰ بایت در زمان بررسی)
├── makefile                          # اهداف build/migrate/seed/docs/docker
├── scripts/
│   └── init_db.sql                   # در شکل نوشته‌شده PostgreSQL نامعتبر؛ توسط Compose استفاده نمی‌شود
├── tmp/
│   └── build-errors.log              # خروجی خطای build در Air (نادیده‌گرفته‌شده)
└── web/                              # خالی؛ بدون فرانت‌اند
```

## مسئولیت بسته‌ها

| بسته | importهای داخلی | مالک | غیرمالک |
|---|---|---|---|
| `cmd/api` | `docs`، `internal/auth`، `consts`، `db`، `dto`، `env`، `response`، `storage`، `utils` | مسیریابی HTTP، میان‌افزار، هندلرها، سیم‌کشی برنامه | سرویس‌های دامنه، تراکنش‌ها، DTOهای پاسخ غیر از سلامت |
| `cmd/seed` | `internal/db`، `internal/env`، `internal/storage` | نقطه ورود اجرایی seed | ساخت schema (مهاجرت‌ها مالک آن‌اند) |
| `docs` | هیچ داخلی | قالب Swagger تولیدشده و فراداده زمان اجرا | مستندات دست‌نویس API |
| `internal/auth` | هیچ داخلی | رابط توکن + پیاده‌سازی HS256 | جست‌وجوی کاربر، سیاست مجوزدهی، هش رمز |
| `internal/consts` | هیچ داخلی | تایم‌اوت‌ها، خطای sentinel، کلید context | بارگذاری پیکربندی |
| `internal/db` | `internal/consts`، `internal/storage` | اتصال DB، تنظیم pool، تولید seed | متدهای کوئری، مهاجرت‌ها |
| `internal/dto` | هیچ داخلی | payloadهای درخواست، پاسخ سلامت | مدل‌های پاسخ کاربر/پست |
| `internal/env` | هیچ داخلی | کمک‌تابع‌های env | اعتبارسنجی مقادیر پیکربندی |
| `internal/response` | هیچ داخلی | کمک‌تابع‌های پاکت/وضعیت JSON | طبقه‌بندی/لاگ خطا |
| `internal/storage` | `internal/consts` | SQL، ساختارهای ماندگاری، متدهای رمز bcrypt، تجزیه صفحه‌بندی فید | مسیریابی HTTP، مجوزدهی |
| `internal/utils` | هیچ داخلی | مدیریت بدنه JSON، validator سراسری | نگاشت خطا |

## نقاط ورود و فایل‌های اجرایی

| فایل اجرایی | فرمان build | یادداشت |
|---|---|---|
| API | `go build -o /tmp/shadina-api ./cmd/api` | تنها فرآیند طولانی‌مدت. |
| Seed | `go run ./cmd/seed` | نیازمند دیتابیس مهاجرت‌شده است. idempotent نیست. |
| مهاجرت | CLI خارجی `migrate` | `go build ./cmd/migrate` شکست می‌خورد: `no Go files`. |
| اجراکننده توسعه Air | `air` | `.air.toml` فایل `./cmd/api/*.go` را build می‌کند، اما `pre_cmd` دستور `make gen-docs` را اجرا می‌کند که اکنون شکست می‌خورد. |
| باینری ردیابی‌شده | `bin/main` | قبلاً commit شده؛ ممکن است نسبت به منبع کهنه باشد. ترجیحاً در مسیر موقت دوباره build کنید. |

## فایل‌های تولیدشده در برابر دست‌نویس

| تولیدشده | دست‌نویس |
|---|---|
| `docs/docs.go` | `cmd/**/*.go` |
| `docs/swagger.json` | `internal/**/*.go` |
| `docs/swagger.yaml` | `cmd/migrate/migrations/*.sql` |
| `bin/main` (مصنوع build) | `makefile`، `docker-compose.yaml`، پیکربندی‌ها |
| `logs/*` در زمان اجرا | `README.md`، `LICENSE`، `scripts/init_db.sql` |

`docs.go` با هدر استاندارد swaggo شروع می‌شود: `Code generated by swaggo/swag. DO NOT EDIT`.
فایل‌های تولیدشده Swagger را دستی ویرایش نکنید؛ حاشیه‌نویسی‌ها/DTOها را اصلاح و دوباره تولید کنید (اکنون مسدود
است، پایین را ببینید).

## مشاهدات سطح فایل

- `cmd/api/uesrs.go` یک غلط املایی در نام فایل است؛ بسته و هندلرها معتبرند.
- `cmd/api/auth.go` هندلرهای ثبت‌نام و توکن را مخلوط می‌کند؛ بسته `internal/service` وجود ندارد.
- `internal/storage/pagination.go` بسته `net/http` را import و query string را تجزیه می‌کند؛ بسته storage صرفاً
  ماندگاری نیست.
- `internal/db/seed.go` بسته `internal/storage` را import می‌کند؛ راه‌اندازی DB و تولید fixture به هم گره خورده‌اند.
- `internal/dto` هیچ ساختار پاسخی برای `User`، `Post` یا `Comment` ندارد؛ ساختارهای storage مستقیماً توسط
  هندلرها سریالایز می‌شوند.
- `docs/swagger.json` و `docs/swagger.yaml` از نظر معنایی با هم سازگارند (با تجزیه و مقایسه کلیدها/مقادیر تأیید
  شد)، اما نسبت به هندلرها کهنه‌اند.
- `scripts/init_db.sql` از `CREATE DATABASE IF NOT EXISTS` استفاده می‌کند که PostgreSQL پشتیبانی نمی‌کند، و قبل
  از انتخاب دیتابیس یک extension می‌سازد. توسط Compose یا makefile ارجاع داده نمی‌شود.
- `web/` خالی است؛ هیچ build فرانت‌اند یا خط لوله دارایی استاتیک وجود ندارد.
- `bin/main` با وجود اینکه مصنوع build است ردیابی می‌شود. `.gitignore` پوشه `bin/` را مستثنی نمی‌کند.
- `logs/` و `tmp/` نادیده گرفته می‌شوند؛ `.env` نادیده گرفته می‌شود. `.example.env` ردیابی می‌شود اما برای
  احراز هویت ناقص است.

## ساختارهای غایب

موارد زیر در مخزن وجود ندارند:

- `cmd/migrate/main.go` (CLI مهاجرت خارجی است)
- `Dockerfile` / ایمیج برنامه
- `.github/`، `.gitlab-ci.yml` یا هر پیکربندی CI
- `*_test.go`، mockها، fixtureها یا راه‌اندازی testcontainers
- تست‌های OpenAPI/قرارداد
- بسته‌های `internal/service`، `internal/domain` یا repository به‌ازای هر موجودیت
- پیکربندی متریک/ردیابی
- `.golangci.yml` (golangci-lint از پیش‌فرض‌ها استفاده می‌کند)

## اسناد مرتبط

- [معماری](02-architecture.md)
- [جریان برنامه](04-application-flow.md)
- [راهنمای توسعه](13-development-guide.md)
- [تست و کیفیت](14-testing-quality.md)
