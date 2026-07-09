// Package buildinfo resolves the binary version for display.
package buildinfo

import "runtime/debug"

// Resolve returns the best available version string. Preference order:
// the ldflags-injected value (release builds), then the module version
// recorded by `go install module@version`, then the raw fallback ("dev").
func Resolve(ldflagsVersion string) string {
	if ldflagsVersion != "" && ldflagsVersion != "dev" {
		return ldflagsVersion
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return ldflagsVersion
}
