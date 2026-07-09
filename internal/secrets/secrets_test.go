package secrets

import (
	"strings"
	"testing"
)

func TestRedactString(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"aws access key", "export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"},
		{"openai key", "OPENAI_API_KEY=sk-abcdefghijklmnopqrstuvwxyz012345"},
		{"password in url", "psql postgres://user:p4ssw0rd@db.example.com:5432/app"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !ContainsSecret(tc.in) {
				t.Fatalf("ContainsSecret(%q) = false, want true", tc.in)
			}
			got := RedactString(tc.in)
			if got == tc.in {
				t.Fatalf("RedactString did not change input: %q", tc.in)
			}
			if !strings.Contains(got, Redacted) {
				t.Fatalf("RedactString(%q) = %q, missing %q marker", tc.in, got, Redacted)
			}
		})
	}
}

func TestRedactStringLeavesCleanTextAlone(t *testing.T) {
	in := "echo hello world; ls -la /tmp"
	if ContainsSecret(in) {
		t.Fatalf("ContainsSecret(%q) = true, want false", in)
	}
	if got := RedactString(in); got != in {
		t.Fatalf("RedactString altered clean text: %q -> %q", in, got)
	}
}
