# Key Dependencies

This document explains why RITA's major dependencies were chosen. See `go.mod`
for the full list and exact versions.

## Core

- **ClickHouse Go driver — `github.com/ClickHouse/clickhouse-go/v2`**
  RITA stores and analyzes very large volumes of connection records. ClickHouse
  is a column-oriented OLAP database purpose-built for fast aggregations over
  billions of rows, which is exactly the access pattern of beacon/threat
  analysis. The official v2 driver gives us native-protocol batch inserts (used
  by `database.BulkWriter`), server-side parameter binding, and LZ4 compression.

- **urfave/cli v2 — `github.com/urfave/cli/v2`**
  Lightweight CLI framework used to define RITA's subcommands (`import`, `view`,
  `delete`, ...) in the `cmd/` package and assembled in `rita.go`. Chosen for its
  small surface and simple command/flag model. (Note: `spf13/cobra` appears only
  as an *indirect* dependency; RITA itself does not use it.)

- **Bubble Tea — `github.com/charmbracelet/bubbletea`** (with **Lip Gloss** and
  **Bubbles**)
  Powers the interactive terminal UI in `viewer/`. Bubble Tea's Elm-style
  model/update/view architecture makes a sortable, drill-down threat table
  maintainable, and the Charm ecosystem provides ready-made table, viewport, and
  styling components.

- **mpb v8 — `github.com/vbauerster/mpb/v8`**
  Multi-progress-bar rendering for the long-running import and seasoning stages
  (`progressbar/`, `importer/`).

## Configuration & logging

- **HJSON — `github.com/hjson/hjson-go/v4`**
  RITA's config (`config.hjson`) is human-edited. HJSON allows comments, relaxed
  quoting, and trailing commas, making the config far friendlier to operators
  than strict JSON while remaining JSON-compatible.

- **validator/v10 — `github.com/go-playground/validator/v10`**
  Declarative struct-tag validation of the parsed config, plus custom rules
  (impact categories, score-threshold ordering, wildcard FQDNs) in
  `config.NewValidator`.

- **zerolog — `github.com/rs/zerolog`**
  Zero-allocation structured logging (`logger/`). Structured fields make import
  IDs, database names, and stages greppable; it also supports syslog output for
  containerized deployments.

- **godotenv — `github.com/joho/godotenv`**
  Loads connection/secret settings from `.env` so infrastructure config stays
  out of the HJSON application config and out of source control.

## Utilities

- **afero — `github.com/spf13/afero`**
  Filesystem abstraction. Import/validation code takes an `afero.Fs` so tests can
  run against an in-memory filesystem instead of touching disk.

- **google/uuid**, **blang/semver**, **google/go-github**
  Network-ID handling, version comparison, and the GitHub release "newer version
  available" check, respectively.

- **golang.org/x/sync (errgroup)** and **golang.org/x/time/rate**
  `errgroup` coordinates worker pools and propagates the first error (used in
  `BulkWriter`, importer, and analysis); `rate.Limiter` throttles batch inserts
  to a rate ClickHouse is comfortable with.

## Testing

- **testcontainers-go — `github.com/testcontainers/testcontainers-go`**
  Spins up real ClickHouse containers for integration/database tests so they
  exercise actual SQL and schema rather than mocks. This is why the `database`,
  `integration`, and `cmd` test suites require Docker.

- **go.uber.org/mock (gomock)**
  Generates mocks (e.g. `MockDatabase`) for unit tests that need to drive code
  paths without a live database. Tool dependency is pinned in `tools.go`;
  regenerate with `go generate ./...`.

## Observability

- **Prometheus client — `github.com/prometheus/client_golang`**
  Exposes RITA metrics (import/analysis counters, durations, write errors) on an
  opt-in `/metrics` endpoint (`metrics/`).

- **OpenTelemetry — `go.opentelemetry.io/otel` (+ SDK and OTLP/HTTP exporter)**
  Optional distributed tracing (`telemetry/`), enabled only when an OTLP endpoint
  is configured via the standard `OTEL_*` environment variables.
