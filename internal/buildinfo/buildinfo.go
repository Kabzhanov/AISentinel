// Package buildinfo resolves the binary version for display.
package buildinfo

import (
	"runtime/debug"
	"strings"
)

// Resolve returns the best available version string. Preference order:
// the ldflags-injected value (release builds), then the module version
// recorded by `go install module@version`, then the raw fallback ("dev").
func Resolve(ldflagsVersion string) string {
	if ldflagsVersion != "" && ldflagsVersion != "dev" {
		return ldflagsVersion
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			// Module versions carry a "v" prefix (v1.0.7); callers format
			// their own "v%s", so strip it to avoid "vv1.0.7".
			return strings.TrimPrefix(v, "v")
		}
	}
	return ldflagsVersion
}
