// Package util provides cross-cutting helpers shared across RITA.
//
// It includes the FixedString type (a 16-byte hash key used as ClickHouse row
// identifiers), subnet parsing and private/public IP classification, network-ID
// derivation, filesystem validation helpers built on spf13/afero (so they can be
// exercised against an in-memory filesystem in tests), FQDN validation, sorting
// helpers, timestamp validation, and the GitHub release version check used by the
// update notifier.
package util
