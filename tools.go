//go:build tools
// +build tools

// Package tools pins build-time tool dependencies (currently mockgen) so they
// are tracked in go.mod and reproducible across machines and CI. The "tools"
// build tag keeps this file out of normal builds, so nothing here is ever
// compiled into the rita binary.
//
// Regenerate mocks with: go generate ./...
package tools

import (
	_ "go.uber.org/mock/mockgen"
)
