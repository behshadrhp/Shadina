# 16. عیب‌یابی

## راه‌اندازی و محیط

| نشانه | علت احتمالی | شاهد / بررسی | راه‌حل ایمن |
|---|---|---|---|
| `error loading .env file` | `.env` غایب، غیرقابل خواندن یا اجرای فرآیند از پوشه دیگر | `ls -l .env`، `pwd` | از ریشه مخزن اجرا کنید؛ `.env` را از `.example.env` بسازید؛ رازها را commit نکنید. |
| `make: *** No rule to make target ...` یا شکست فوری make | Makefile دستور `include .env` دارد؛ نبود `.env` make را می‌شکند | `ls -l .env` | `.env` بسازید یا از فرمان‌های مستقیم `go`/`migrate` استفاده کنید. |
| `dial tcp 127.0.0.1:5435: connect: connection refused` | DSN fallback کد از 5435/`shadina` استفاده می‌کند؛ Compose پورت 5454/`shadina_db` را ارائه می‌کند | `docker compose ps`؛ بررسی `DATABASE_ADDR` | `DATABASE_ADDR` را به DSN Compose روی پورت 5454 تنظیم کنید. |
| `database stable publish` هرگز ظاهر نمی‌شود / فرآیند خارج می‌شود | PostgreSQL اجرا نمی‌شود یا DSN اشتباه است | `docker compose logs postgresql`؛ `psql`/`docker exec` | Compose را شروع کنید؛ DSN را اصلاح کنید. volumeها را حذف نکنید. |
| `invalid duration` / خطای DB راه‌اندازی | `MAX_IDLE_TIMEOUT` یک مدت Go مانند `15m` نیست | بررسی `.env` | از `15m`، `1h` و غیره استفاده کنید. |
| Permission denied در ساخت `logs/` | پوشه کاری قابل نوشتن نیست یا مالکیت `logs` موجود ناسازگار است | `ls -ld logs` | از پوشه قابل نوشتن اجرا کنید؛ مالکیت را با تأیید اصلاح کنید. |
| API شروع می‌شود اما Swagger UI spec را بارگذاری نمی‌کند | `docsPath` فاقد `/v1` و scheme است | `/v1/swagger/index.html` را بگیرید و `url:` را بررسی کنید | مستقیماً از `/v1/swagger/doc.json` استفاده کنید؛ `router.go`/`SERVER_URL` را در کد اصلاح کنید. |
| `AUTH_SECRET_KAY` نادیده گرفته می‌شود | املای صحیح `AUTH_SECRET_KEY` خوانده نمی‌شود؛ کد `KAY` را می‌خواند | `grep AUTH .env` | فعلاً کلید غلط‌املا را تنظیم کنید، یا کد را اصلاح و به‌طور سازگار تغییر نام دهید. |
| توکن‌ها به‌طور غیرمنتظره جعل/معتبرند | راز env غایب؛ fallback hard-code در استفاده است | بررسی env راه‌اندازی؛ fallback کد | راز قوی تنظیم کنید؛ fallback را در کد حذف کنید (نیازمند تغییر). |

## دیتابیس و مهاجرت‌ها

| نشانه | علت احتمالی | بررسی | راه‌حل |
|---|---|---|---|
| `migrate: command not found` | CLI نصب نیست/در PATH نیست | `which migrate` | golang-migrate را نصب کنید؛ `migrate -version` را بررسی کنید. |
| `make migrate-up` با خطای اتصال شکست می‌خورد | `DATABASE_ADDR` اشتباه یا DB خاموش | `make migrate-version`؛ `docker compose ps` | DSN را اصلاح/DB را شروع کنید. |
| `Dirty database version` | مهاجرت قبلی نیمه‌کاره شکست خورده | `make migrate-version` | مهاجرت شکست‌خورده را بررسی کنید؛ `migrate force` فقط بعد از فهم وضعیت (خطرناک). |
| `no Go files in .../cmd/migrate` | تلاش برای `go build ./cmd/migrate` | `find cmd/migrate` | از CLI `migrate` استفاده کنید؛ main Go وجود ندارد. |
| `make migrate-create NAME=x` فایلی نمی‌سازد | هدف `NAME` را پاس نمی‌دهد؛ بدون نام باز می‌شود | `make -n migrate-create NAME=x` | از `migrate create -seq -ext sql -dir ./cmd/migrate/migrations <name>` استفاده کنید. |
| `make seed` → `stat cmd/migrate/seed/main.go: no such file` | هدف make شکسته | `make -n seed` | از `go run ./cmd/seed` استفاده کنید. |
| seed با کلید تکراری شکست می‌خورد | seed idempotent نیست | شمارش سطر/خطاهای یکتایی | از DB دورانداختنی استفاده کنید، یا seed را idempotent/پاک‌سازی با تأیید کنید. |
| seed کامل می‌شود اما ورود شکست می‌خورد | کاربران seed هش رمز خالی دارند | `SELECT count(*) FROM users WHERE password=''::bytea;` (فقط خواندنی) | کاربر جدید از `/v1/auth/register` ثبت‌نام کنید. |
| تراکنش seed rollback نمی‌شود | `BeginTx` توسط متدهای storage استفاده نمی‌شود | مرور کد `internal/db/seed.go` | کد را برای پاس دادن `tx` اصلاح کنید؛ به rollback اتکا نکنید. |
| `scripts/init_db.sql` شکست می‌خورد | `CREATE DATABASE IF NOT EXISTS` در PostgreSQL معتبر نیست | فایل را بخوانید | از آن استفاده نکنید؛ Compose دیتابیس را می‌سازد؛ مهاجرت‌ها schema را می‌سازند. |
| شکاف مهاجرت ۸–۱۰ گیج‌کننده است | فایل‌ها غایب | `ls cmd/migrate/migrations` | خطای زمان اجرا نیست؛ در تغییر آینده مستند/شماره‌گذاری مجدد کنید. |
| حذف پست 500 برمی‌گرداند | FK نظرها آبشار ندارد | `\d comments`؛ FK `fk_post_comment` | ابتدا نظرها را در یک تراکنش حذف کنید یا با مهاجرت cascade اضافه کنید (نیازمند تأیید). |

## احراز هویت و مجوزدهی

| نشانه | علت | بررسی | راه‌حل |
|---|---|---|---|
| 401 `authorization header is missing` | هدر `Authorization` وجود ندارد | هدرهای درخواست | `Authorization: Bearer <token>` اضافه کنید. |
| 401 `authorization header is malformed` | `bearer` با حروف کوچک، فاصله‌های متعدد، بخش‌های اضافی | قالب هدر | دقیقاً `Bearer <token>` استفاده کنید. |
| 500 `token is malformed...` | توکن نامعتبر/منقضی؛ میان‌افزار به 500 نگاشت می‌کند | کد پاسخ 5019 | توکن تازه بگیرید؛ میان‌افزار را در کد برای 401 اصلاح کنید. |
| 401 کد 4022 پس از توکن | کاربر حذف شده یا جست‌وجوی DB شکست خورده | شناسه کاربر DB / لاگ‌ها | مطمئن شوید کاربر وجود دارد؛ DB را بررسی کنید. |
| ویرایش/حذف ادمین پاسخ عجیب/مخلوط برمی‌گرداند | شاخه ادمین بعد از هندلر 401 می‌نویسد | کد `UserPostAccessMiddleware` | میان‌افزار را اصلاح کنید؛ تا اصلاح به مسیر ادمین اتکا نکنید. |
| مدیر نمی‌تواند پست دیگران را ویرایش کند | فقط `RoleID == 3` مجاز است | کد | رفتار فعلی مورد انتظار؛ در صورت نیاز سیاست را اصلاح کنید. |
| 401 `access denied` هنگام حذف پست کاربر دیگر | بررسی مالکیت | `user_id` پست در برابر کاربر توکن | از حساب مالک استفاده کنید؛ دور نزنید. |
| فید پست‌های کاربر دیگر را نشان می‌دهد | `userID` مسیر با توکن مقایسه نمی‌شود | درخواست `/v1/users/2/feeds` | تصمیم محصول؛ اگر خصوصی است بررسی را در کد اعمال کنید. |
| فید `sort=asc/desc` ترتیب یکسان | `$2` در `ORDER BY` ثابت است | مقایسه پاسخ‌ها | SQL را با عبارت مرتب‌سازی لیست‌سفید اصلاح کنید. |
| شمارش نظر فید متورم است | join سطرها را برای پست‌های شخصی تکرار می‌کند | مقایسه `COUNT` با `COUNT(DISTINCT)` | کوئری را اصلاح کنید؛ به عملیات دیتابیس مراجعه کنید. |

## رفتار API

| نشانه | علت | بررسی | راه‌حل |
|---|---|---|---|
| `GET /v1/users/1/feed` 404 | مسیر واقعی `/feeds` است | منبع روتر | از `/v1/users/1/feeds` استفاده کنید. |
| `GET /v1/users/2/follow` 405 | follow متد `PUT` است، نه GET | منبع روتر | از `PUT` با توکن استفاده کنید. |
| ثبت‌نام 200 برمی‌گرداند، نه 201 | `SuccessResponse` استفاده شده | منبع هندلر | انتخاب قرارداد؛ کد یا مستندات را اصلاح کنید. |
| اعتبارسنجی ساخت پست 500 برمی‌گرداند | هندلر از `InternalErrorResponse` استفاده می‌کند | کد | در کد به 400 اصلاح کنید. |
| `GET /v1/users/999999` 500 برمی‌گرداند | هندلر `NOT_FOUND` را نگاشت نمی‌کند | کد | در کد به 404 اصلاح کنید. |
| شناسه غیرعددی 500 با متن `strconv` برمی‌گرداند | خطای تجزیه به internal نگاشت شده | کد | 400 و پیام عمومی برگردانید. |
| ثبت‌نام تکراری 500 با متن DB برمی‌گرداند | نقض یکتایی نگاشت نشده | DB/متن خطا | پیش‌بررسی یا نگاشت نقض یکتایی به 409/400؛ پاک‌سازی. |
| follow تکراری/هدف ناشناخته 500 برمی‌گرداند | خطای PK/FK نگاشت نشده | DB/متن خطا | به 409/404 نگاشت؛ پاک‌سازی. |
| unfollow برای رابطه ناموجود موفقیت برمی‌گرداند | `RowsAffected` نادیده گرفته می‌شود | کد | 404 برگردانید یا رفتار idempotent را مستند کنید. |
| تولید Swagger شکست می‌خورد | `dto.FollowPayload`/`UnFollowPayload` گمشده | خروجی `swag init` | DTOها را اضافه یا حاشیه‌نویسی بدنه را حذف کنید؛ دوباره تولید کنید. |
| Swagger UI spec را بارگذاری نمی‌کند | URL فاقد `/v1` و scheme است | `url:` در HTML UI | از `/v1/swagger/doc.json` استفاده کنید؛ روتر را اصلاح کنید. |
| پاسخ‌های 404/405 متن ساده‌اند | هندلرهای پیش‌فرض chi | `curl -i` مسیر بدون تطبیق | هندلرهای JSON 404/405 در کد اضافه کنید. |

## لاگ و عیب‌یابی

| نشانه | علت | بررسی | راه‌حل |
|---|---|---|---|
| `logs/error.log` خالی است | مقصد خطای داخلی zap است، نه خطاهای برنامه | کد | لاگ خطای صریح اضافه کنید؛ به این فایل اتکا نکنید. |
| پس از panic لاگ درخواست وجود ندارد | میان‌افزار لاگ توسط panic/ترتیب Recoverer باز می‌شود | کد | ترتیب/لاگ میان‌افزار را اصلاح کنید؛ لاگ recovery اضافه کنید. |
| فایل لاگ بدون چرخش رشد می‌کند | بدون سیاست چرخش | اندازه فایل | logrotate خارجی/لاگ متمرکز پیکربندی کنید. |
| آدرس `server stable` اشتباه به نظر می‌رسد | لاگ از `SERVER_URL` استفاده می‌کند، نه `SERVER_ADDR` | کد | فیلد لاگ یا پیکربندی را اصلاح کنید. |
| بدون request ID در خطاها | پیاده‌سازی نشده | کد | میان‌افزار request ID اضافه و در پاسخ/لاگ بگنجانید. |

## فرمان‌های تشخیصی ایمن

```bash
# فقط خواندنی
docker compose ps
docker compose logs --tail=100 postgresql
curl -sS -i http://127.0.0.1:8000/v1/health
curl -sS -i http://127.0.0.1:8000/ping
curl -sS -i http://127.0.0.1:8000/v1/swagger/doc.json
go test ./...
go vet ./...
go build ./...
gofmt -l .
golangci-lint run ./...

# بررسی دیتابیس (فقط خواندنی)
docker exec shadina-postgres psql -U user -d shadina_db -c '\dt'
docker exec shadina-postgres psql -U user -d shadina_db -c '\d users'
docker exec shadina-postgres psql -U user -d shadina_db -c 'SELECT version, dirty FROM schema_migrations;'

# مگر با تأیید صریح استفاده نکنید (مخرب/تغییردهنده داده):
# docker compose down -v
# migrate down
# DELETE/UPDATE/TRUNCATE/DROP
```

توکن‌ها، رمزها، محتوای کامل `.env` یا DSNهای تولید را در issue/چت نچسبانید.

## اسناد مرتبط

- [راهنمای توسعه](13-development-guide.md)
- [پیکربندی](09-configuration.md)
- [مرجع API](07-api-reference.md)
- [عملیات دیتابیس](06-database-operations.md)
