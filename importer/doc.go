// Package importer ingests network logs (Zeek TSV/JSON and compatible formats)
// into ClickHouse.
//
// The importer fans work across a pool of goroutines: files are walked and
// deduplicated by hash(path+mtime), parsed by digester workers into typed
// entries, and streamed into the database through database.BulkWriter instances
// (one per log type: conn, dns, http, ssl, ...). Unbuffered channels between the
// stages provide natural backpressure so that slow database writes throttle the
// parsers rather than letting memory grow unbounded.
package importer
