// Package database manages RITA's connection to ClickHouse and all schema and
// data access.
//
// It owns connection setup and pooling (ConnectToDB), DDL for the analysis and
// snapshot tables (tables.go), the MetaDB that tracks imports, and a generic
// concurrent BulkWriter (writer.go) used to stream rows into ClickHouse in
// rate-limited batches.
//
// BulkWriter runs a pool of worker goroutines coordinated through an
// errgroup.Group; write errors are propagated back to the caller from Close()
// rather than terminating the process, so the caller decides how to react.
package database
