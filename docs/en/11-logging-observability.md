# 11. Logging and Observability

## Logging setup

`cmd/api/main.go`:

1. `os.MkdirAll("logs", 0755)` — creates the log directory relative to the process working directory.
2. `zap.NewProductionConfig()` — JSON encoder, production defaults, sampling enabled.
3. `OutputPaths = ["logs/app.log"]`.
4. `ErrorOutputPaths = ["logs/error.log"]`.
5. `zap.Must(...).Sugar()`; `defer log.Sync()`.

`logs/app.log` receives zap entries from `app.logger`. `logs/error.log` is zap's **internal error sink**; it
does not automatically receive application errors. At review time `logs/error.log` was 0 bytes.

The `defer log.Sync()` does not run when `main` exits via `log.Fatal` after `ListenAndServe` returns. The
same is true for `db.Close()`. This can lose buffered log entries on shutdown.

## Request logging

`app.LoggerMiddleware` wraps the response writer with `middleware.NewWrapResponseWriter` and, after the next
handler returns, logs:

| Field | Value |
|---|---|
| `msg` | `http request` |
| `method` | HTTP method |
| `path` | `r.URL.Path` (query string is not logged) |
| `remote_addr` | `r.RemoteAddr` (after chi `RealIP`) |
| `status_code` | captured response status |
| `bytes` | bytes written |
| `duration` | `time.Since(start)` |

It does **not** log:

- the authenticated user ID or username
- a request/correlation ID
- the route pattern or handler name
- request headers (including `Authorization`)
- request/response bodies
- query parameters (good for avoiding token leakage in URLs, but limits debugging)
- errors (handlers return errors without logging them)

Chi's `middleware.Logger` is also registered globally and writes a common-log-format line to stdout. In a
normal terminal run this produces a second, unstructured log stream. In a detached process it may go to the
service manager or be lost.

`middleware.Recoverer` recovers panics and logs a stack trace using the standard logger. Because
`LoggerMiddleware` wraps downstream handlers after `Recoverer`, a panic unwinds past `LoggerMiddleware`'s
post-handler logging code; the custom request log may be missing for panicked requests.

## Log format and level

- Format: zap production JSON, for example:
  ```json
  {"level":"info","ts":1790791373.617005,"caller":"api/main.go:84","msg":"database stable publish","driver":"postgres"}
  ```
- Level: production config default (`info`); no debug/development logger in the API.
- Sampling: zap production sampling is enabled by default; repeated identical entries may be dropped.
- Rotation: **none**. There is no lumberjack/rotation policy, size limit, retention, or compression. The file
  grows until external rotation is configured.
- Output: file only for zap; chi/stdlib logs go to stdout/stderr.
- Directory/file permissions: `logs` is created `0755`; file mode is governed by the process umask.

## What is logged at startup

- `database stable publish` with `driver` (not the DSN).
- `server stable` with `address` (the value of `SERVER_URL`, e.g. `127.0.0.1:8000`) and `environment`.
- Fatal startup errors via `log.Fatalln` (standard log, stderr): `.env` load failure, DB connection failure.

The server log message says `address` but uses `app.config.url` (`SERVER_URL`), not `app.config.addr`
(`SERVER_ADDR`). If they differ, the log is misleading.

## What is not logged

- Authentication successes/failures (except the status code in the request log).
- Authorization denials (only 401 status/bytes).
- SQL errors or slow queries.
- Validation failures details (returned to client, not logged).
- Panic details in the custom zap log (Recoverer uses standard log).
- Audit trail for post/user changes.
- Health-check results or DB latency.
- Migration/seed outcomes (seed uses standard `log`).

## Observability gaps

| Capability | Present? | Notes |
|---|---|---|
| Structured logs | Partial | Zap JSON for request/startup; chi/stdlib unstructured logs elsewhere. |
| Request/correlation ID | No | No middleware, header, or context value. |
| Metrics | No | No Prometheus/StatsD/OTel metrics. |
| Distributed tracing | No | No OpenTelemetry/Jaeger. |
| Error tracking | No | No Sentry/Rollbar integration. |
| Health check | Partial | `/v1/health` is static; no DB/dependency check. |
| Readiness check | No | Same as health. |
| Profiling | No | No `net/http/pprof`. |
| Query logging | No | `lib/pq` errors are returned but not logged centrally. |
| Audit log | No | No record of who changed what. |
| Log rotation | No | External rotation required. |
| Dashboards/alerts | No | No configuration in repository. |

## Health endpoints

| Route | Type | Behavior |
|---|---|---|
| `GET /ping` | Heartbeat | chi middleware; 200 `text/plain` body `.`; bypasses the router. |
| `GET /v1/health` | Application liveness | 200 JSON `{status:"ok",version,environment}`. Does not touch PostgreSQL. |

Neither endpoint is a readiness probe for the database. A deployment can route traffic to an API instance
whose DB is down. See [Docker and deployment](12-docker-deployment.md).

## Sensitive-data handling

Based on code review:

- `Authorization` headers and request bodies are not logged by the custom middleware or chi logger.
- Query strings are not logged by the custom middleware (it uses `URL.Path`), which reduces token leakage
  risk.
- `remote_addr` (client IP) is logged; treat logs as potentially personal data.
- Error responses may contain database/parser details, but those are sent to clients, not written to
  `app.log` by the response helpers.
- Zap's error sink is not an application-error log; do not assume errors appear there.
- No masking/redaction layer exists. If future logging adds headers/body/user fields, add redaction first.

## Operational commands

```bash
# Follow structured request logs (read-only)
tail -f logs/app.log

# Follow database container logs (read-only)
docker compose logs -f postgresql

# Check whether the API is listening
curl -sS http://127.0.0.1:8000/v1/health
curl -sS http://127.0.0.1:8000/ping
```

There is no documented log-rotation command in the repository. Use an external logrotate/container logging
driver in real deployments.

## Recommendations

1. Add a request ID middleware; include it in every log line and error response.
2. Add structured error logging with `zap.Error(err)` and the request ID.
3. Log the authenticated user ID/role on protected routes (not the token).
4. Add DB-aware readiness (`PingContext` with a short timeout) separate from liveness.
5. Add metrics for request count/duration/status, DB pool stats, and error codes.
6. Add OpenTelemetry tracing around HTTP and SQL.
7. Add log rotation/retention or ship logs to a centralized system.
8. Replace `log.Fatal` in the server path with graceful shutdown and log flushing.
9. Avoid `err.Error()` in responses; log it internally instead.
10. Document a retention/redaction policy for IP addresses and user identifiers.

## Related documents

- [Error handling](10-error-handling.md)
- [Docker and deployment](12-docker-deployment.md)
- [Security review](../reports/security-review.md)
- [Troubleshooting](16-troubleshooting.md)
