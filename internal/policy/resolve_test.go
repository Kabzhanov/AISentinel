package policy

import (
	"os"
	"path/filepath"
	"testing"
)

// withCWD temporarily changes the working directory to dir for the duration
// of the test, restoring the original CWD on cleanup. Tests using this must
// not run in parallel with each other (CWD is process-global).
func withCWD(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("restore chdir: %v", err)
		}
	})
}

func TestLoadDefault(t *testing.T) {
	eng, err := LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault: %v", err)
	}
	if eng.Name != "default" {
		t.Errorf("name = %q, want %q", eng.Name, "default")
	}
	if len(eng.Rules) == 0 {
		t.Error("embedded default policy has no rules")
	}
	if eng.Signature() == "" {
		t.Error("signature is empty")
	}
}

func TestResolveExplicitPathMissingIsHardError(t *testing.T) {
	// An explicit path that doesn't exist must always be an error — never a
	// silent fallback to the built-in default.
	_, source, err := Resolve(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatal("expected error for missing explicit policy path, got nil")
	}
	if source == "" {
		t.Error("expected source to be reported even on error")
	}
}

func TestResolveExplicitPathInvalidYAMLIsHardError(t *testing.T) {
	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(badFile, []byte("not: [valid, policy"), 0o644); err != nil {
		t.Fatalf("write bad policy: %v", err)
	}
	if _, _, err := Resolve(badFile); err == nil {
		t.Fatal("expected error for invalid explicit policy YAML, got nil")
	}
}

func TestResolveEnvPathMissingIsHardError(t *testing.T) {
	t.Setenv("AISENTINEL_POLICY", filepath.Join(t.TempDir(), "nope.yaml"))
	if _, _, err := Resolve(""); err == nil {
		t.Fatal("expected error for missing $AISENTINEL_POLICY path, got nil")
	}
}

func TestResolveOnDiskDefaultPreferredOverEmbedded(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "policies"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	onDisk := []byte("version: 1\nname: on-disk-marker\nrules:\n  - id: x\n    match: {}\n    decision: allow\n")
	if err := os.WriteFile(filepath.Join(dir, "policies", "default.yaml"), onDisk, 0o644); err != nil {
		t.Fatalf("write policies/default.yaml: %v", err)
	}
	withCWD(t, dir)

	eng, source, err := Resolve("")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if eng.Name != "on-disk-marker" {
		t.Errorf("name = %q, want %q (expected on-disk policies/default.yaml to win)", eng.Name, "on-disk-marker")
	}
	if source != filepath.Join("policies", "default.yaml") {
		t.Errorf("source = %q, want %q", source, filepath.Join("policies", "default.yaml"))
	}
}

func TestResolveFallsBackToEmbeddedDefault(t *testing.T) {
	// Nothing explicit, no $AISENTINEL_POLICY, and CWD has no policies/
	// directory at all — this is the `go install` + run-from-anywhere case
	// that used to os.Exit(1) before the embedded default was added.
	t.Setenv("AISENTINEL_POLICY", "")
	withCWD(t, t.TempDir())

	eng, source, err := Resolve("")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if eng.Name != "default" {
		t.Errorf("name = %q, want %q (expected embedded default)", eng.Name, "default")
	}
	if source != "built-in default" {
		t.Errorf("source = %q, want %q", source, "built-in default")
	}
}
