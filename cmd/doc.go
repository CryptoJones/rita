// Package cmd defines the RITA command-line interface.
//
// RITA is built on urfave/cli/v2; each user-facing subcommand (import, view,
// delete, list, validate, and so on) is implemented here as a *cli.Command and
// assembled into the application in the repository-root rita.go entrypoint.
//
// The most substantial command is import (see import.go), which walks the
// supplied log directory, deduplicates files by hash(path+mtime), and drives the
// per-day/per-hour pipeline of importer.Import -> analyzer.Analyze ->
// modifier.Modify before recording completion in the MetaDB.
package cmd
