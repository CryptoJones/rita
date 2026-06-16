# Environment Variables

RITA reads two categories of environment variables:

1. **Application config** — loaded from a `.env` file (via `godotenv`) in the
   working directory. The base `.env` is **required**; RITA exits on startup if
   it is missing. These configure the database connection, logging, and config
   location.
2. **Observability/infra** — read directly from the process environment (not
   `.env`), so they can be set by the container/orchestrator. These are optional.

## Application config (`.env`)

| Variable | Purpose | Default | Read by |
|---|---|---|---|
| `DB_ADDRESS` | ClickHouse `host:port` (native protocol). | `db:9000` (compose) | `config.setEnv` |
| `CLICKHOUSE_USERNAME` | ClickHouse user. | `default` | `config.setEnv` |
| `CLICKHOUSE_PASSWORD` | ClickHouse password (may be empty). | _(empty)_ | `config.setEnv` |
| `CONFIG_DIR` | Directory containing `config.hjson`, the HTTP extensions CSV, and threat-intel feed files. | `/etc/rita` (compose) | `config` |
| `CONFIG_FILE` | Config file name/path. | `config.hjson` | compose/CLI |
| `LOG_LEVEL` | zerolog level as an integer (0–6; lower = more verbose). | — (required when logging) | `config.setEnv`, `logger` |
| `LOGGING_ENABLED` | Toggles file/structured logging. | — | `logger` |
| `APP_ENV` | When `dev`, enables debug logging/behavior (same as `--debug`). | _(unset)_ | `rita.go`, `logger` |
| `SYSLOG_ADDRESS` | If set, also ships logs to this syslog endpoint. | _(unset)_ | `logger` |

> Secrets (`CLICKHOUSE_PASSWORD`) are intentionally kept out of `config.hjson`
> and the JSON-serialized config (the field is `json:"-"`). Prefer Docker/K8s
> secrets over committing `.env`.

## Observability / infrastructure (process env)

| Variable | Purpose | Default | Read by |
|---|---|---|---|
| `RITA_METRICS_ADDR` | If set (e.g. `:2112`), starts the Prometheus `/metrics` + `/healthz` server on this address. | _(unset → disabled)_ | `rita.go` / `metrics` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Enables OpenTelemetry tracing via OTLP/HTTP to this collector endpoint. | _(unset → tracing disabled)_ | `telemetry` |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Traces-specific OTLP endpoint (alternative to the general one). | _(unset)_ | `telemetry` |
| `OTEL_*` | All standard OpenTelemetry SDK env vars (headers, timeouts, sampling, service attributes) are honored by the OTLP exporter. | per OTel spec | `telemetry` |

## Docker Compose / build variables

These are consumed by `docker-compose.yml` / `.prod.yml` and the build, not by
the Go code directly:

| Variable | Purpose | Default |
|---|---|---|
| `RITA_VERSION` | RITA image tag (`ghcr.io/activecm/rita:${RITA_VERSION}`). | `latest` |
| `CLICKHOUSE_VERSION` | ClickHouse server image tag. | pinned in compose |
| `TZ` | Container timezone. | `Etc/UTC` |
| `APP_LOGS` / `LOGS` | Host paths mounted for log input/output. | per compose |
| `BUILDPLATFORM` | Build platform for multi-arch images. | Docker-provided |
