# 1. Use ClickHouse as the analytics datastore

- Status: Accepted
- Date: 2026-06-15

## Context

RITA ingests and analyzes very large volumes of network connection records
(conn/dns/http/ssl), then runs heavy aggregations over them — per host-pair
connection counts, timestamp/data-size histograms for beacon scoring, first-seen
tracking, and threat scoring. The workload is overwhelmingly **append-heavy
writes followed by large analytical reads**, with almost no row-level updates.

Earlier RITA generations used a document database, which struggled with the
aggregation-heavy analytical queries at scale.

## Decision

Use **ClickHouse** as RITA's datastore, accessed through the official
`clickhouse-go/v2` native driver.

- Writes go through a concurrent, rate-limited batch writer (`database.BulkWriter`).
- Analysis is expressed as ClickHouse SQL (`analysis/`, `database/tables.go`),
  leaning on aggregate functions (`groupArray`, `*State`/`*Merge` combinators)
  to compute histograms and per-pair rollups server-side.

## Consequences

- **Positive:** Column storage + vectorized execution make the analytical
  queries fast over billions of rows; native batch insert is efficient; LZ4
  compression keeps storage reasonable.
- **Positive:** SQL keeps analysis logic declarative and inspectable.
- **Negative:** ClickHouse is an operational dependency that must be deployed and
  tuned (connection pool, `max_execution_time`, memory).
- **Negative:** Tests need a real ClickHouse instance — provided via
  testcontainers — so the `database`/`integration`/`cmd` suites require Docker.
- **Negative:** ClickHouse's eventual-merge model (e.g. `ReplacingMergeTree`)
  requires care around when data is fully merged before reads.
