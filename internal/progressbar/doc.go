// Package progressbar renders progress feedback for long-running RITA operations
// such as imports and analysis.
//
// It provides terminal progress indicators (built on the Bubble Tea ecosystem)
// that report how far the current stage has advanced, driven by the per-batch
// progress counts emitted by database.BulkWriter and the importer/analysis
// pipelines.
package progressbar
