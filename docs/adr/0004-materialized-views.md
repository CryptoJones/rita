# 4. Materialized views are opt-in, not part of the base schema

- Status: Accepted
- Date: 2026-06-15

## Context

Some analysis queries — notably `GetNetworkSize` (`database/db.go`), which counts
distinct local IPs by UNION-ing DISTINCT src/dst across several tables — get
expensive on large datasets. ClickHouse materialized views can pre-aggregate this
work at insert time and turn the query into a cheap merge.

However, a materialized view fires on **every insert** into its source table. A
wrong or mistuned MV in the base schema could slow down or break imports for
every user, and MV correctness is hard to verify without a representative
ClickHouse instance and dataset.

## Decision

Ship materialized views as an **opt-in operator-applied migration**
(`deploy/clickhouse/materialized_views.sql`) rather than as part of the
automatically-created schema (`database/tables.go`).

- The base schema and import path are unchanged, so the MVs cannot affect users
  who do not opt in.
- The file documents how to apply, backfill, and switch `GetNetworkSize` to read
  from the aggregate table.
- The SQL is marked "validate against your ClickHouse version before use."

## Consequences

- **Positive:** Zero risk to the import path; large deployments get a documented
  fast path; small deployments are unaffected.
- **Positive:** Keeps a clear separation between RITA-managed schema and
  operator-managed performance tuning.
- **Negative:** The optimization is not on by default, so it must be discovered
  and applied; and switching `GetNetworkSize` to use it is a follow-up code
  change that should be made and tested against a real ClickHouse.
- **Negative:** The MVs and the base tables can drift; the migration must be
  kept in sync with schema changes in `database/tables.go`.
