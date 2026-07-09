// Package policies embeds the built-in default policy so that AISentinel
// binaries work out of the box even when installed via `go install` and run
// outside of a clone of this repository (i.e. no policies/ directory on disk
// relative to the current working directory).
package policies

import _ "embed"

// Default is the built-in default policy YAML (policies/default.yaml),
// embedded at build time.
//
//go:embed default.yaml
var Default []byte
