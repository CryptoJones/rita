# Compose environment overrides

These are example Docker Compose **override** files for different environments.
They layer on top of the repository-root `docker-compose.yml` (which defines the
`rita`, `syslog-ng`, and `clickhouse` services).

## Usage

Compose merges files left-to-right; later files win:

```sh
# development
docker compose -f docker-compose.yml -f deploy/compose/docker-compose.dev.yml up

# staging
docker compose -f docker-compose.yml -f deploy/compose/docker-compose.staging.yml up -d

# production (copy and pin versions first)
cp deploy/compose/docker-compose.prod.yml-example deploy/compose/docker-compose.prod.yml
docker compose -f docker-compose.yml -f deploy/compose/docker-compose.prod.yml up -d
```

## What each override changes

| File | Logging | Ports exposed | Resources | Restart |
|---|---|---|---|---|
| `docker-compose.dev.yml` | most verbose (`LOG_LEVEL=0`, `APP_ENV=dev`) | ClickHouse 8123/9000, metrics 2112 | none | none |
| `docker-compose.staging.yml` | info (`LOG_LEVEL=3`) | none | moderate limits | on-failure / unless-stopped |
| `docker-compose.prod.yml-example` | warn (`LOG_LEVEL=4`) | none (DB internal only) | tuned limits + reservations | unless-stopped |

## Notes

- For staging/production, **pin** `RITA_VERSION` and `CLICKHOUSE_VERSION` (e.g. in
  `.env`) instead of relying on `:latest`. The prod example fails fast if they are
  unset.
- The root `docker-compose.override.example.yml` is a separate, auto-merged local
  override for ad-hoc developer tweaks (`cp` it to `docker-compose.override.yml`).
- Set `RITA_METRICS_ADDR` (and map the port) to scrape Prometheus metrics; set
  `OTEL_EXPORTER_OTLP_ENDPOINT` to enable tracing.
