package logger

import (
	"strings"
	"testing"

	"github.com/Kabzhanov/AISentinel/internal/secrets"
)

func TestRedactString(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		wantRedacted bool
	}{
		{"aws access key", "AKIAIOSFODNN7EXAMPLE", true},
		{"aws key in sentence", "found key AKIAIOSFODNN7EXAMPLE in commit", true},
		{"openai key", "sk-proj-abcdefghijklmnopqrstuvwxyz0123456789", true},
		{"url with password", "postgres://admin:hunter2pass@db.internal:5432/app", true},
		{"github pat", "github_pat_11ABCDEFG0123456789_abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQR", true},
		{"github classic token", "ghp_1234567890abcdefghijklmnopqrstuvwxyz", true},
		{"slack token", "xoxb-1234567890-abcdefghijklmnop", true},
		{"bearer token", "Authorization: Bearer abcdefghijklmnopqrstuvwx", true},
		{"pem private key", "-----BEGIN RSA PRIVATE KEY-----", true},
		{"clean text", "echo hello world", false},
		{"benign word token", "this token is not a secret shape", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := secrets.RedactString(tc.in)
			gotRedacted := strings.Contains(got, "[REDACTED]")
			if gotRedacted != tc.wantRedacted {
				t.Errorf("RedactString(%q) = %q, redacted=%v want=%v", tc.in, got, gotRedacted, tc.wantRedacted)
			}
			if tc.wantRedacted && got == tc.in {
				t.Errorf("expected input to change after redaction, got unchanged: %q", got)
			}
		})
	}
}

func TestRedactEventToolArgsAndResult(t *testing.T) {
	ev := Event{
		EventType: "pre_tool",
		ToolName:  "Bash",
		ToolArgs: map[string]any{
			"command": "curl -H 'Authorization: Bearer sk-proj-abcdefghijklmnopqrstuvwxyz0123456789' https://api.example.com",
			"nested": map[string]any{
				"conn": "mysql://root:sup3rSecretPW@127.0.0.1:3306/db",
			},
			"list": []any{"clean", "AKIAIOSFODNN7EXAMPLE"},
			"num":  42,
		},
		ToolResult: map[string]any{
			"stdout": "token=ghp_1234567890abcdefghijklmnopqrstuvwxyz",
		},
		Metadata: map[string]any{
			"note": "no secret here",
		},
	}

	Redact(&ev)

	cmd, _ := ev.ToolArgs["command"].(string)
	if strings.Contains(cmd, "sk-proj-") {
		t.Errorf("command still contains raw secret: %q", cmd)
	}
	if !strings.Contains(cmd, "[REDACTED]") {
		t.Errorf("command missing redaction marker: %q", cmd)
	}

	nested, _ := ev.ToolArgs["nested"].(map[string]any)
	conn, _ := nested["conn"].(string)
	if strings.Contains(conn, "sup3rSecretPW") {
		t.Errorf("nested map value still contains raw secret: %q", conn)
	}

	list, _ := ev.ToolArgs["list"].([]any)
	if s, ok := list[1].(string); !ok || strings.Contains(s, "AKIA") {
		t.Errorf("list element still contains raw AWS key: %v", list[1])
	}
	if s, ok := list[0].(string); !ok || s != "clean" {
		t.Errorf("clean list element was altered: %v", list[0])
	}

	if n, ok := ev.ToolArgs["num"].(int); !ok || n != 42 {
		t.Errorf("non-string value was mutated: %v", ev.ToolArgs["num"])
	}

	stdout, _ := ev.ToolResult["stdout"].(string)
	if strings.Contains(stdout, "ghp_1234567890") {
		t.Errorf("tool_result still contains raw secret: %q", stdout)
	}

	if ev.Metadata["note"] != "no secret here" {
		t.Errorf("clean metadata value was altered: %v", ev.Metadata["note"])
	}
}
