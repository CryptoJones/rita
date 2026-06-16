// Package logger provides RITA's process-wide structured logger.
//
// It wraps rs/zerolog and exposes a shared logger via GetLogger(). Output and
// verbosity are controlled by the environment: LOGGING_ENABLED toggles logging,
// LOG_LEVEL selects the zerolog level, and SYSLOG_ADDRESS, when set, routes logs
// to a syslog endpoint in addition to (or instead of) the console.
package logger
