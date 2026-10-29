# 09. پیکربندی

## مدل بارگذاری

فرمان‌های API و seed مقدار `godotenv.Load()` را بدون نام فایل صدا می‌زنند. این `.env` را از **پوشه کاری جاری**
بارگذاری می‌کند. متغیرهای از قبل موجود در محیط فرآیند را بازنویسی نمی‌کند. اگر `.env` وجود نداشته باشد یا
خوانده نشود، هر دو فرمان با `error loading .env file` خارج می‌شوند.

makefile این کار را می‌کند:

```make
include .env
export
```

بنابراین هر هدف `make` به `.env` قابل خواندن نیاز دارد، حتی اهدافی که فقط راهنما چاپ می‌کنند یا Docker را
اجرا می‌کنند. اعتبارنامه‌های PostgreSQL در فایل Compose از محیط برنامه جدا هستند؛ API مقدار `DATABASE_ADDR`
را از `.env`/محیط فرآیند می‌خواند.

## متغیرهای محیطی استفاده‌شده توسط کد

| متغیر | هدف | الزامی | پیش‌فرض در کد | مثال | منبع |
|---|---|---|---|---|---|
| `SERVER_ADDR` | آدرس شنود HTTP | خیر | `:8000` | `:8000` | `cmd/api/main.go` |
| `SERVER_URL` | میزبان Swagger و ساخت آدرس doc | خیر | `127.0.0.1:8000` | `127.0.0.1:8000` | `cmd/api/main.go`، `router.go`، `api.go` |
| `APP_VERSION` | نسخه نمایش‌داده‌شده در health و Swagger | خیر | `0.0.1` | `0.0.1` | `cmd/api/main.go` |
| `APP_ENVIRONMENT` | برچسب محیط نمایش‌داده‌شده در health | خیر | `development` | `development` | `cmd/api/main.go` |
| `DATABASE_DRIVER` | نام درایور `database/sql` برای API | خیر | `postgres` | `postgres` | `cmd/api/main.go` |
| `DATABASE_ADDR` | DSN PostgreSQL | خیر | `postgres://postgres:postgres@127.0.0.1:5435/shadina?sslmode=disable` | `postgres://user:password@127.0.0.1:5454/shadina_db?sslmode=disable` | `cmd/api/main.go`، `cmd/seed/main.go` |
| `MAX_OPEN_CONNS` | `SetMaxOpenConns` | خیر | `30` | `30` | `cmd/api/main.go`، `cmd/seed/main.go` |
| `MAX_IDLE_CONNS` | `SetMaxIdleConns` | خیر | `30` | `30` | `cmd/api/main.go`، `cmd/seed/main.go` |
| `MAX_IDLE_TIMEOUT` | `SetConnMaxIdleTime` (مدت Go) | خیر | `15m` | `15m` | `cmd/api/main.go`، `cmd/seed/main.go` |
| `AUTH_SECRET_KAY` | راز HMAC JWT (به املا توجه کنید) | **عملاً الزامی**؛ کد fallback دارد | fallback توسعه hard-code در `cmd/api/main.go` | یک مقدار تصادفی طولانی، مثلاً `<32+ random bytes>` | `cmd/api/main.go` |
| `DB_DRIVER` | نام درایور **فقط برای `cmd/seed`** | خیر | `postgres` | `postgres` | `cmd/seed/main.go` |

`AUTH_SECRET_KAY` در زمان بررسی در `.env` یا `.example.env` وجود ندارد. `AUTH_SECRET_KEY` (املای طبیعی)
**خوانده نمی‌شود**. اگر اپراتور متغیر با املای صحیح را تنظیم کند، سرویس بی‌صدا از fallback hard-code استفاده
می‌کند.

`SERVER_URL` در `.env` محلی وجود دارد اما در `.example.env` غایب است. به‌عنوان `host` در Swagger (بدون scheme)
و به‌عنوان پایه آدرس `doc.json` در UI استفاده می‌شود. مسیر واقعی هندلر Swagger برابر
`/v1/swagger/doc.json` است، بنابراین آدرس ساخته‌شده مطابقت ندارد (به [مرجع API](07-api-reference.md) مراجعه
کنید).

### عدم تطابق DSN پیش‌فرض

fallback کد:

```text
postgres://postgres:postgres@127.0.0.1:5435/shadina?sslmode=disable
```

سرویس Compose:

```text
پورت میزبان 5454 -> کانتینر 5432
دیتابیس shadina_db
کاربر user
رمز password
```

فایل ردیابی‌شده `.example.env` از مقادیر Compose استفاده می‌کند:

```text
postgres://user:password@127.0.0.1:5454/shadina_db?sslmode=disable
```

اگر `.env` فاقد DSN باشد، API پورت 5435/دیتابیس `shadina` را امتحان می‌کند و در برابر سرویس Compose ارائه‌شده
شکست می‌خورد. این یک خطای رایج راه‌اندازی محلی است.

## پیکربندی hard-code شده

این مقادیر **قابل پیکربندی از محیط نیستند**:

| تنظیم | مقدار | مکان |
|---|---|---|
| عمر JWT | ۷۲ ساعت | `cmd/api/main.go` (`time.Hour * 24 * 3`) |
| صادرکننده JWT | `Shadina` | `cmd/api/main.go` |
| مخاطب JWT | `Shadina` | `cmd/api/main.go` |
| تایم‌اوت خواندن HTTP | ۱۰ ثانیه | `internal/consts/consts.go` |
| تایم‌اوت نوشتن HTTP | ۳۰ ثانیه | `internal/consts/consts.go` |
| تایم‌اوت بیکاری HTTP | ۱ دقیقه | `internal/consts/consts.go` |
| تایم‌اوت DB/context | ۱۰ ثانیه | `internal/consts/consts.go` |
| محدودیت بدنه درخواست | ۱٬۰۴۸٬۵۷۸ بایت | `internal/utils/json.go` |
| مسیرهای لاگ | `logs/app.log`، `logs/error.log` | `cmd/api/main.go` |
| مجوز پوشه لاگ | `0755` | `cmd/api/main.go` |
| هزینه bcrypt | `bcrypt.DefaultCost` | `internal/storage/users.go` |
| مسیر پایه Swagger | `/v1/` در زمان اجرا؛ `/v1` در حاشیه‌نویسی | `cmd/api/api.go`، `cmd/api/main.go` |

هیچ تنظیم `ReadHeaderTimeout` یا `MaxHeaderBytes` وجود ندارد. هیچ پیکربندی CORS وجود ندارد.

## نمونه `.env` (فقط جانشین)

رازهای واقعی را در مستندات یا کنترل نسخه کپی نکنید. یک فایل توسعه می‌تواند این‌گونه باشد:

```dotenv
# SERVER
SERVER_ADDR=:8000
SERVER_URL=127.0.0.1:8000
APP_VERSION=0.0.1
APP_ENVIRONMENT=development

# DATABASE (مطابق docker-compose.yaml)
DATABASE_DRIVER=postgres
DATABASE_ADDR=postgres://user:password@127.0.0.1:5454/shadina_db?sslmode=disable
MAX_OPEN_CONNS=30
MAX_IDLE_CONNS=30
MAX_IDLE_TIMEOUT=15m

# AUTH: غلط املایی در کد فعلی عمدی است؛ از مقدار تصادفی قوی استفاده کنید
AUTH_SECRET_KAY=<replace-with-a-long-random-secret>
```

مقادیر `user`/`password` بالا پیش‌فرض‌های توسعه Compose هستند، نه اعتبارنامه تولید. در استقرار واقعی از
secret manager یا تزریق محیط استفاده کنید.

## اعتبارسنجی پیکربندی

`internal/env/env.go` عمداً حداقلی است:

- `GetString` فقط وقتی fallback را برمی‌گرداند که متغیر **غایب** باشد؛ رشته خالی همان‌طور که هست برگردانده
  می‌شود.
- `GetInt` اگر مقدار غایب یا با `strconv.Atoi` قابل تجزیه نباشد fallback را برمی‌گرداند.
- هیچ اعتبارسنجی محدوده برای اندازه pool وجود ندارد.
- `MAX_IDLE_TIMEOUT` توسط `db.NewDB` تجزیه می‌شود؛ مدت نامعتبر باعث شکست راه‌اندازی می‌شود.
- هیچ اعتبارسنجی وجود ندارد که `AUTH_SECRET_KAY` حاضر، به‌اندازه کافی طولانی یا غیر-fallback باشد.
- هیچ اعتبارسنجی وجود ندارد که `DATABASE_ADDR` به دیتابیس قابل دسترس اشاره کند تا `PingContext`.
- هیچ اعتبارسنجی برای `SERVER_ADDR` وجود ندارد.

## اولویت پیکربندی

1. محیط فرآیند (بالاترین؛ `godotenv.Load` بازنویسی نمی‌کند).
2. `.env` در پوشه کاری جاری.
3. fallbackهای کد.

فایل Compose فقط متغیرهای `POSTGRES_*` را برای کانتینر دیتابیس تنظیم می‌کند؛ متغیرهای برنامه را به API تزریق
نمی‌کند چون کانتینر API وجود ندارد.

## مدیریت رازها

- `.env` توسط Git نادیده گرفته می‌شود (خط `.env` در `.gitignore`).
- `.example.env` ردیابی می‌شود و فقط اعتبارنامه‌های دیتابیس توسعه را دارد؛ هیچ راز JWT ندارد.
- `docker-compose.yaml` اعتبارنامه‌های دیتابیس توسعه را دارد؛ خارج از توسعه محلی از آن‌ها استفاده نکنید.
- لاگ‌ها در کد بررسی‌شده هدر `Authorization`، بدنه درخواست یا رمزها را شامل نمی‌شوند.
- پاسخ‌های خطا می‌توانند متن خام دیتابیس/parser را شامل شوند؛ آن‌ها مخزن راز نیستند، اما می‌توانند جزئیات
  schema/داخلی را افشا کنند.
- هیچ secret manager، vault، sealed secret یا یکپارچه‌سازی KMS وجود ندارد.
- `bin/main` یک باینری ردیابی‌شده است؛ مطمئن شوید در زمان build راز جاسازی نمی‌کند (در کد بررسی‌شده هیچ
  رازی پیکربندی نشده بود).

## اسناد مرتبط

- [راهنمای توسعه](13-development-guide.md)
- [Docker و استقرار](12-docker-deployment.md)
- [احراز هویت و مجوزدهی](08-authentication-authorization.md)
- [مرور امنیتی](../reports/security-review.md)
