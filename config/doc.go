// Package config loads, validates, and supplies RITA's runtime configuration.
//
// Configuration is read from an HJSON file (config.hjson by default) and merged
// with secrets and connection details taken from the environment (DB_ADDRESS,
// CLICKHOUSE_USERNAME, CLICKHOUSE_PASSWORD, LOG_LEVEL, CONFIG_DIR, ...). The
// resulting Config struct is validated with go-playground/validator, including
// custom rules for impact categories, score-threshold ordering, and wildcard
// FQDNs.
//
// The optional config_version field is checked against CurrentConfigVersion so
// that a config written for a newer RITA is rejected with a clear error rather
// than silently misbehaving.
package config
