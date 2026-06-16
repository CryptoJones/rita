-- Opt-in materialized views for RITA (ClickHouse).
--
-- These are NOT applied automatically by RITA's schema setup. They are an
-- operator-applied performance optimization for large datasets, kept separate so
-- a mistake here can never break the import path. Apply them per analysis
-- database AFTER validating against your ClickHouse version, e.g.:
--
--   clickhouse-client --database <rita_db> --multiquery < materialized_views.sql
--
-- ---------------------------------------------------------------------------
-- Accelerate GetNetworkSize (database/db.go).
--
-- GetNetworkSize counts the distinct local IPs seen since a cutoff by UNION-ing
-- DISTINCT src/dst across uconn/openconn/usni/openhttp/udns. On large datasets
-- that scan is expensive. These views pre-aggregate the distinct local IPs per
-- hour as a uniqExact aggregate state, so the count becomes a cheap merge over a
-- small per-hour table.
--
-- Target table holding the partial aggregate states.
CREATE TABLE IF NOT EXISTS network_size_by_hour
(
    hour      DateTime,
    local_ips AggregateFunction(uniqExact, String)
)
ENGINE = AggregatingMergeTree()
ORDER BY hour;

-- One materialized view per (table, side) that contributes local IPs. Each fires
-- on insert into its source table and folds new rows into the per-hour state.
CREATE MATERIALIZED VIEW IF NOT EXISTS network_size_uconn_src_mv TO network_size_by_hour AS
SELECT toStartOfHour(hour) AS hour, uniqExactState(src) AS local_ips
FROM uconn WHERE src_local = true GROUP BY hour;

CREATE MATERIALIZED VIEW IF NOT EXISTS network_size_uconn_dst_mv TO network_size_by_hour AS
SELECT toStartOfHour(hour) AS hour, uniqExactState(dst) AS local_ips
FROM uconn WHERE dst_local = true GROUP BY hour;

CREATE MATERIALIZED VIEW IF NOT EXISTS network_size_usni_src_mv TO network_size_by_hour AS
SELECT toStartOfHour(hour) AS hour, uniqExactState(src) AS local_ips
FROM usni WHERE http = true AND src_local = true GROUP BY hour;

CREATE MATERIALIZED VIEW IF NOT EXISTS network_size_usni_dst_mv TO network_size_by_hour AS
SELECT toStartOfHour(hour) AS hour, uniqExactState(dst) AS local_ips
FROM usni WHERE http = true AND dst_local = true GROUP BY hour;

CREATE MATERIALIZED VIEW IF NOT EXISTS network_size_udns_src_mv TO network_size_by_hour AS
SELECT toStartOfHour(hour) AS hour, uniqExactState(src) AS local_ips
FROM udns WHERE src_local = true GROUP BY hour;

CREATE MATERIALIZED VIEW IF NOT EXISTS network_size_udns_dst_mv TO network_size_by_hour AS
SELECT toStartOfHour(hour) AS hour, uniqExactState(dst) AS local_ips
FROM udns WHERE dst_local = true GROUP BY hour;

-- NOTE: open* tables (openconn/openhttp) have no per-hour partitioning in the
-- same way; include them with their own views if your deployment needs them.
--
-- Once these exist and have been backfilled, GetNetworkSize can be switched to:
--
--   SELECT uniqExactMerge(local_ips) AS network_size
--   FROM network_size_by_hour
--   WHERE hour >= toStartOfHour(fromUnixTimestamp({min_ts:Int64}));
--
-- Backfill existing data after creating the views:
--
--   INSERT INTO network_size_by_hour
--   SELECT toStartOfHour(hour), uniqExactState(src) FROM uconn WHERE src_local GROUP BY hour;
--   -- ...repeat the SELECT half of each view above.
