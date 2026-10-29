# 13. راهنمای توسعه

## پیش‌نیازها

| ابزار | نسخه / نیازمندی | چرا |
|---|---|---|
| Go | `1.25.4` (`go.mod`)؛ بررسی با `go1.25.4` | Build/run/test |
| Docker Engine + Compose | هر نسخه اخیر؛ بررسی با Docker 29.8.0 / Compose v5.5.1 | PostgreSQL محلی |
| CLI golang-migrate | بررسی با `4.16.2` | اعمال/برگشت مهاجرت‌ها |
| `make` | GNU make | اهداف راحتی |
| `swag` | CLI `1.16.4` در زمان بررسی؛ ماژول `1.16.6` | تولید Swagger (اکنون مسدود) |
| `golangci-lint` | اختیاری؛ بررسی با `2.15.0` | لینت |
| `air` | اختیاری | hot reload؛ pre-command اکنون شکست می‌خورد |
| `psql` / Docker exec | اختیاری | بررسی DB |

برای بک‌اند به Node.js، Java، Python یا زنجیره ابزار فرانت‌اند نیازی نیست. `web/` خالی است.

## ۱. دریافت مخزن و بررسی محیط

```bash
git clone <repository-url>
cd Shadina
git status
ls -l .env .example.env
```

اگر `.env` از قبل وجود دارد، **آن را بازنویسی نکنید**. توسط Git نادیده گرفته می‌شود و ممکن است مقادیر محلی
داشته باشد. آن را بخوانید و فقط کلیدهای گمشده را اضافه کنید.

## ۲. ساخت/به‌روزرسانی `.env`

در صورت نیاز:

```bash
cp .example.env .env
```

سپس مطمئن شوید این مقادیر حاضرند:

```dotenv
SERVER_ADDR=:8000
SERVER_URL=127.0.0.1:8000
APP_VERSION=0.0.1
APP_ENVIRONMENT=development

DATABASE_DRIVER=postgres
DATABASE_ADDR=postgres://user:password@127.0.0.1:5454/shadina_db?sslmode=disable
MAX_OPEN_CONNS=30
MAX_IDLE_CONNS=30
MAX_IDLE_TIMEOUT=15m

# غلط املایی توسط کد فعلی الزامی است. از مقدار تصادفی قوی استفاده کنید.
AUTH_SECRET_KAY=<replace-with-a-long-random-secret>
```

یادداشت‌ها:

- کد `AUTH_SECRET_KAY` را می‌خواند، نه `AUTH_SECRET_KEY`.
- `.example.env` فاقد `SERVER_URL` و راز احراز هویت است.
- DSN fallback کد از پورت **5435** و دیتابیس `shadina` استفاده می‌کند؛ سرویس Compose از **5454** و
  `shadina_db`. برای توسعه محلی از مقادیر Compose استفاده کنید.
- `.env` یا رازهای واقعی را commit نکنید.

## ۳. شروع PostgreSQL

```bash
docker compose up -d
# یا: make up-back
```

بررسی:

```bash
docker compose ps
docker compose logs postgresql
```

نام کانتینر `shadina-postgres` است؛ PostgreSQL روی پورت میزبان **5454** ارائه می‌شود.

## ۴. اعمال مهاجرت‌ها

```bash
make migrate-up
```

پیش‌نیازها: `migrate` در `PATH`، `.env` قابل خواندن، PostgreSQL قابل دسترس. این جدول‌ها و ایندکس‌ها را
می‌سازد/تغییر می‌دهد و `schema_migrations` را می‌نویسد. روی DB توسعه تازه ایمن است اما همیشه قبل از اجرا روی
داده‌ای که برایتان مهم است بررسی کنید.

بررسی نسخه:

```bash
make migrate-version
```

وضعیت محلی مورد انتظار در زمان بررسی: نسخه `12`، `dirty = false`.

برای ساخت مهاجرت (هدف make شکسته است):

```bash
migrate create -seq -ext sql -dir ./cmd/migrate/migrations add_some_change
```

فایل‌های `.up.sql`/`.down.sql` تولیدشده را ویرایش، بررسی و سپس `make migrate-up` کنید. مهاجرت‌ها را فقط برای
مستندسازی اجرا نکنید.

## ۵. Seed داده توسعه

هدف make شکسته است:

```bash
make seed
# => go run cmd/migrate/seed/main.go
# => stat cmd/migrate/seed/main.go: no such file or directory
```

از فرمان واقعی استفاده کنید:

```bash
go run ./cmd/seed
```

این ۱۰۰ کاربر، ۲۰۰ پست و ۵۰۰ نظر درج می‌کند. **idempotent نیست**: اجرای مجدد روی کاربران تکراری شکست
می‌خورد و ممکن است داده ناقص باقی بگذارد. دنبال‌کننده نمی‌سازد. کاربران seed‌شده هش رمز خالی دارند و
**نمی‌توانند وارد شوند**.

اگر حساب قابل ورود نیاز دارید، از طریق API ثبت‌نام کنید (پایین) یا با یک اسکریپت کنترل‌شده کاربری با هش
bcrypt واقعی درج کنید. داده تولید را دستی ویرایش نکنید.

## ۶. اجرای API

```bash
go run ./cmd/api
```

فرآیند:

- `.env` را از پوشه جاری بارگذاری می‌کند؛
- `logs/` را می‌سازد؛
- به PostgreSQL وصل و ping می‌کند؛
- روی `SERVER_ADDR` (پیش‌فرض `:8000`) گوش می‌دهد؛
- در `logs/app.log` لاگ می‌کند.

بررسی:

```bash
curl -sS http://127.0.0.1:8000/v1/health
curl -sS http://127.0.0.1:8000/ping
```

`/v1/health` JSON برمی‌گرداند و DB را بررسی نمی‌کند. `/ping` مقدار `.` برمی‌گرداند.

## ۷. ثبت‌نام و احراز هویت کاربر

حساب‌های seed‌شده نمی‌توانند وارد شوند. یک کاربر بسازید:

```bash
curl -sS -X POST http://127.0.0.1:8000/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"dev@example.com","username":"devuser","password":"correct horse battery"}'
```

انتظار: HTTP 200، کد `1010`، داده کاربر با `role_id: 1`.

دریافت توکن:

```bash
curl -sS -X POST http://127.0.0.1:8000/v1/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"email":"dev@example.com","password":"correct horse battery"}'
```

استفاده:

```bash
TOKEN='<paste-token>'
curl -sS http://127.0.0.1:8000/v1/users \
  -H "Authorization: Bearer $TOKEN"
```

یادداشت‌ها:

- توکن ۷۲ ساعت معتبر است.
- هیچ endpoint refresh/logout وجود ندارد.
- توکن‌های نامعتبر اکنون HTTP 500 برمی‌گردانند، نه 401.
- از ایمیل/نام کاربری یکتا استفاده کنید؛ موارد تکراری HTTP 500 با متن خام DB برمی‌گردانند.

## ۸. Build

```bash
# توصیه‌شده: bin/main ردیابی‌شده را بازنویسی نکنید
go build -o /tmp/shadina-api ./cmd/api

# ساخت فرمان seed
go build -o /tmp/shadina-seed ./cmd/seed

# ساخت همه بسته‌ها؛ cmd/migrate رد می‌شود (بدون فایل Go)
go build ./...
```

`go build ./cmd/migrate` شکست می‌خورد؛ از CLI `migrate` استفاده کنید. `bin/main` از قبل ردیابی می‌شود و
ممکن است کهنه باشد.

## ۹. اجرای بررسی‌ها

```bash
go test ./...            # بدون تست؛ هر بسته [no test files] گزارش می‌کند
go test -race ./...      # بدون تست
go test -cover ./...     # 0.0% برای بسته‌های دارای statement
go vet ./...             # موفق
go mod verify            # همه ماژول‌ها تأیید شدند
golangci-lint run ./...  # اکنون با ۱۲ مشکل شکست می‌خورد
```

`make test` دستور `go test -v ./...` را اجرا می‌کند؛ با کد ۰ خارج می‌شود اما تستی اجرا نمی‌کند. `make lint`
ابتدا `gofmt -s -w .` را اجرا می‌کند که **فایل‌ها را تغییر می‌دهد**، سپس لینت می‌کند. اگر فقط می‌خواهید
قالب‌بندی را بررسی کنید، ابتدا `gofmt -l .` را اجرا کنید.

## ۱۰. Swagger / مستندات API

```bash
make gen-docs
# => swag init -g ./api/main.go -d cmd,internal && swag fmt
```

این اکنون با خطای زیر شکست می‌خورد:

```text
ParseComment error ... dto.FollowPayload ... cannot find type definition
```

فایل‌های Swagger موجود کهنه‌اند. برای تولید مجدد، ابتدا حاشیه‌نویسی‌های follow/unfollow را اصلاح یا DTOهای
گمشده را اضافه کنید، سپس `swag init` را اجرا و diff را بررسی کنید. `docs/docs.go` را دستی ویرایش نکنید.

Swagger UI در این آدرس‌ها ارائه می‌شود:

```text
http://127.0.0.1:8000/v1/swagger/index.html
http://127.0.0.1:8000/v1/swagger/doc.json
```

آدرس spec پیکربندی‌شده در UI اشتباه است (فاقد `/v1` و scheme)؛ به [مرجع API](07-api-reference.md) مراجعه
کنید.

## ۱۱. Hot reload (اختیاری)

`.air.toml` فایل `./cmd/api/*.go` را به `./bin/main` build و اجرا می‌کند، با `pre_cmd = ["make gen-docs"]`.
چون `make gen-docs` اکنون شکست می‌خورد، `air` ممکن است شروع نشود یا خطاهای build را در
`tmp/build-errors.log` لاگ کند. تا اصلاح حاشیه‌نویسی‌های Swagger، `go run ./cmd/api` را ترجیح دهید.

## ۱۲. جریان کاری رایج توسعه

1. DTO درخواست را در `internal/dto` اضافه کنید (در صورت نیاز).
2. متدهای storage را در `internal/storage` اضافه/توسعه دهید؛ SQL را پارامتری نگه دارید.
3. هندلر را در `cmd/api` اضافه کنید؛ از کمک‌تابع‌های `response` به‌طور سازگار استفاده کنید.
4. مسیر و میان‌افزار را در `cmd/api/router.go` ثبت کنید.
5. حاشیه‌نویسی‌های Swagger را اضافه کنید؛ مطمئن شوید DTOهای ارجاع‌شده وجود دارند.
6. تست اضافه کنید (هنوز هیچ test harness وجود ندارد).
7. `gofmt`، `go vet`، `go test` و `golangci-lint` را اجرا کنید.
8. اثرهای DB و نگاشت خطا را بررسی کنید.

## ۱۳. بازنشانی دیتابیس (مخرب؛ بدون تأیید اجرا نکنید)

این فرمان‌ها داده را حذف می‌کنند:

```bash
# کانتینرها را متوقف/حذف می‌کند اما volume نام‌دار را نگه می‌دارد:
docker compose down

# volume دیتابیس را هم حذف می‌کند (از دست رفتن داده):
docker compose down -v

# همه مهاجرت‌ها را برمی‌گرداند (اشیای schema را حذف می‌کند):
make migrate-down
```

فقط در محیط‌های محلی دورانداختنی استفاده کنید و ابتدا با تیم تأیید کنید.

## ۱۴. لینک‌های سریع عیب‌یابی

- `.env` غایب → `error loading .env file`
- Connection refused → پورت 5454 در برابر fallback کد 5435 و وضعیت کانتینر را بررسی کنید
- `make seed` شکست می‌خورد → از `go run ./cmd/seed` استفاده کنید
- ورود کاربر seed شکست می‌خورد → رمز خالی؛ کاربر جدید ثبت‌نام کنید
- تولید Swagger شکست می‌خورد → `dto.FollowPayload` گمشده
- توکن نامعتبر 500 برمی‌گرداند → رفتار فعلی میان‌افزار
- حذف پست 500 برمی‌گرداند → FK نظرها آبشار ندارد

فهرست کامل: [عیب‌یابی](16-troubleshooting.md).

## اسناد مرتبط

- [پیکربندی](09-configuration.md)
- [Docker و استقرار](12-docker-deployment.md)
- [عملیات دیتابیس](06-database-operations.md)
- [تست و کیفیت](14-testing-quality.md)
