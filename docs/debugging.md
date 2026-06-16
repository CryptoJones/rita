# Debugging Guide

How to diagnose failed imports and analysis runs.

## Turn on debug logging

- Pass the global `--debug` / `-d` flag **before** the subcommand:
  `rita --debug import ...`.
- Or set `APP_ENV=dev` in the environment.
- `LOG_LEVEL` (0–6, lower = more verbose) and `LOGGING_ENABLED` control structured
  logging; `SYSLOG_ADDRESS` ships logs to syslog as well.

Debug mode also keeps temporary tables around (they are normally truncated before
each import), which is useful for inspecting intermediate import state.

## Read the structured logs

RITA logs are structured (zerolog). Useful fields to grep for:

- `import_id` — correlate all work for one import.
- `database` / `writer` — which table/writer a message refers to.
- `stage` — batch stage on a write error (`prepare`, `append`, `send`, ...).
- `elapsed_time` — per-phase timing (parsing, seasoning, analysis, modification).

## Inspect import bookkeeping (MetaDB)

RITA records imports in the MetaDB. To see what was imported:

```sql
-- which files were imported, and when
SELECT * FROM metadatabase.files ORDER BY ts DESC LIMIT 50;
-- import start/finish records
SELECT * FROM metadatabase.imports ORDER BY ts DESC LIMIT 50;
```

(Exact table/column names: see `database/tables.go` and the MetaDB helpers.)

## Common investigations

**"Import succeeded but no threats":** confirm conn data actually landed and that
your network is in `filtering.internal_subnets`:

```sql
SELECT count() FROM <database>.conn;
SELECT count() FROM <database>.threat_mixtape;
```

If `conn` is populated but `threat_mixtape` is empty, the filtering config is the
usual culprit (see TROUBLESHOOTING.md → "empty results").

**A write failed mid-import:** RITA now returns write errors instead of calling
`log.Fatal`, so the error propagates up with context (writer name, stage,
batch size). Look for `rita_database_write_errors_total` on the metrics endpoint
and the wrapped error in the logs.

**Re-running an import does nothing:** files are deduped by `hash(path+mtime)`;
see TROUBLESHOOTING.md → "all files were previously imported".

## Re-running cleanly

To start fresh, import into a new database name, or rebuild (`--rebuild`). For
rolling datasets, ensure logs are mounted at a stable real path so dedup keys are
consistent across runs (`rita.sh`).

## Tracing (deep timing)

Set an OTLP endpoint (`OTEL_EXPORTER_OTLP_ENDPOINT=...`) to emit OpenTelemetry
traces from the import/analysis pipeline to a collector (Jaeger, Tempo, etc.) for
span-level timing. Without an endpoint, tracing is a no-op.
