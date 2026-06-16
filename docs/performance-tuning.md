# Performance Tuning

How to tune RITA for different data volumes. Most knobs live in `config.hjson`;
a few are derived from the host's CPU count.

## Batch size

`batch_size` (default 100000; valid 25000–2000000) controls how many rows
`database.BulkWriter` buffers before flushing to ClickHouse.

- **Larger** batches reduce per-insert overhead and improve throughput on big
  imports, at the cost of more memory per writer worker.
- **Smaller** batches lower memory use on constrained sensors but increase insert
  frequency.
- Memory scales roughly with `batch_size × number of writer workers × row size`,
  so raise it gradually and watch RSS.

## Worker counts (CPU)

During import, parser/digester/writer pools are sized from available cores
(`GetAvailableCores` / `SetWorkerCount` in `cmd/import.go`): on machines with
more than 3 cores RITA reserves ~2 for the OS and splits the rest roughly evenly
across parsers, digesters, and writers. On small VMs this self-limits; on large
sensors it scales out automatically.

## Query execution time

`max_query_execution_time` (default 600s) maps to ClickHouse's
`max_execution_time`. Increase it for very large datasets where analysis queries
are legitimately long; decrease it to fail fast in latency-sensitive setups.

## Connection pool

`database.ConnectToDB` configures `MaxOpenConns: 50`, `MaxIdleConns: 50`,
`ConnMaxLifetime: 1h`. The default 50 is generous for a single RITA process; if
you run many concurrent imports against one ClickHouse, ensure the server's
`max_connections` comfortably exceeds the sum of all RITA pools. Lower these for
a shared/contended ClickHouse.

## Insert rate limiting

Inserts are throttled by a shared `rate.Limiter` (≈5 batches/sec). ClickHouse
prefers fewer, larger inserts; if you raise `batch_size` substantially you
generally do **not** need to touch the rate limit.

## Beacon / strobe thresholds

- `scoring.beacon.unique_connection_threshold` gates which pairs are even
  considered for (relatively expensive) beacon analysis.
- Pairs with `Count >= 86400` (one conn/sec/day) are treated as strobes and skip
  beacon analysis entirely, which bounds analysis cost on noisy pairs.

## ClickHouse server

For large deployments, give ClickHouse adequate RAM (aggregations are
memory-hungry), fast local disk for `/var/lib/clickhouse`, and consider
materialized views for the heaviest repeated rollups. See
`docs/architecture.md` for how the schema is used.

## Measuring

Build and run the benchmarks for hot helpers:

```sh
go test ./util/ ./analysis/ -run '^$' -bench . -benchmem
```

Enable the metrics endpoint (`RITA_METRICS_ADDR=:2112`) to scrape import/analysis
durations and counts (`rita_import_duration_seconds`, `rita_import_runs_total`,
`rita_analysis_runs_total`, `rita_database_write_errors_total`).
