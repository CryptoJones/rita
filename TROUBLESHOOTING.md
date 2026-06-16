# Troubleshooting

Common RITA problems and how to resolve them. Error strings below come from the
RITA source so you can match them against what you see.

## ClickHouse connection failures

**Symptom:** import/analysis exits early with a connection or dial error.

- Confirm `DB_ADDRESS` points at the ClickHouse **native** port (`host:9000`),
  not the HTTP port (8123).
- Confirm `CLICKHOUSE_USERNAME` / `CLICKHOUSE_PASSWORD` are set in `.env`
  (`CLICKHOUSE_PASSWORD` may be empty, but the others are required).
- ClickHouse must be reachable and healthy before RITA runs. With Docker Compose,
  RITA waits for the `clickhouse` healthcheck; if it never becomes healthy, check
  `docker compose logs clickhouse`.
- The driver dial timeout is 120s; a slow/overloaded server can still time out.

## "all files were previously imported"

`ErrAllFilesPreviouslyImported` — every file in the directory was already imported
(RITA dedups by `hash(path + mtime)`). This is expected when re-running the same
import. To re-import intentionally, change the input (or its mtime), use a fresh
database, or use `--rebuild`.

## Duplicate / re-imported rolling logs

RITA keys file dedup on `hash(path + mtime)`, so a rolling log that is rewritten
in place (new mtime) is treated as new, while an identical file is skipped
(`ErrSkippedDuplicateLog`). If imports collide unexpectedly across runs, make sure
logs are mounted at their **real** path (so the path component of the key is
stable) — see `rita.sh`.

## "no valid files found" / empty results

- `ErrNoValidFilesFound` — the directory has no recognized Zeek logs. Check file
  naming/prefixes (conn/dns/http/ssl, and their `open*` variants) and that the
  files are TSV or JSON Zeek logs.
- Results empty even though import "succeeded": you likely didn't include your
  network in `filtering.internal_subnets`. RITA only scores internal↔external
  pairs; if no IPs are classified as internal, nothing is scored. Add your CIDRs
  to `filtering.internal_subnets` in `config.hjson`.

## Permission errors

`ErrInsufficientReadPermissions` (or "file does not have readable permission or
does not exist") — RITA can't read a log file. Check file ownership/permissions
and that the path is mounted readable into the container.

## "import twice" on a non-rolling database

`ErrImportTwiceNonRolling` — you tried to import additional data into a
non-rolling database. Create the database with `--rolling` if you intend to add
to it over time, or import into a new database.

## Config validation errors

On startup RITA validates `config.hjson`. Messages are prefixed with
"encountered an error while reading the config file" and name the failing field.
Common causes:

- `batch_size` outside 25000–2000000.
- Score thresholds not strictly increasing (base < low < medium < high).
- Beacon weights that don't sum to 1.
- `config_version` newer than this RITA build understands — upgrade RITA or
  remove/lower `config_version`.

Use `config.schema.json` in your editor for autocomplete and inline validation.

## Beacon analysis skipped / "could not find min/max timestamps"

`ErrInvalidMinMaxTimestamp` — there wasn't enough timestamped connection data to
establish the analysis time window. Usually means the conn data didn't import
(see "empty results" above) or the dataset is too small/sparse.

## Version-check noise on stdout

The "newer version available" notice is written to **stderr** (so it won't
corrupt CSV/stdout parsing). Set `update_check_enabled: false` in `config.hjson`
to disable the check entirely.
