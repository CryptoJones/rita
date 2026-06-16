// Package viewer implements RITA's terminal user interface for browsing analysis
// results.
//
// It is built on the Bubble Tea framework (with Lip Gloss styling) and renders
// the scored threat data for a database as an interactive, sortable table with
// drill-down detail views. The viewer is a read-only consumer of data produced
// by the importer and analysis packages.
package viewer
