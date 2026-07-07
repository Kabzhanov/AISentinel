package policy

import (
	"strings"
	"testing"
)

const samplePolicy = `
version: 1
name: test
rules:
  - id: block-secret
    match: { tool_args_regex: "(?i)(api[_-]?key|secret|token)" }
    decision: block
    reason: "secret in args"
  - id: lan
    match:
      tool_name: Bash
      tool_args_regex: "192\\.168\\."
    decision: block
  - id: safe
    match: { tool_name_regex: "^Read$" }
    decision: allow
  - id: catch-all
    match: {}
    decision: log_only
`

func TestLoadValid(t *testing.T) {
	eng, err := Load([]byte(samplePolicy))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if eng.Name != "test" {
		t.Errorf("name = %q", eng.Name)
	}
	if len(eng.Rules) != 4 {
		t.Errorf("rules = %d, want 4", len(eng.Rules))
	}
	if eng.Signature() == "" {
		t.Errorf("signature is empty")
	}
}

func TestLoadInvalidDecision(t *testing.T) {
	bad := `
version: 1
name: bad
rules:
  - id: x
    match: {}
    decision: maybe
`
	if _, err := Load([]byte(bad)); err == nil {
		t.Fatal("expected error for unknown decision")
	}
}

func TestLoadInvalidRegex(t *testing.T) {
	bad := `
version: 1
name: bad
rules:
  - id: x
    match: { tool_args_regex: "[unclosed" }
    decision: block
`
	if _, err := Load([]byte(bad)); err == nil {
		t.Fatal("expected error for bad regex")
	}
}

func TestCheckFirstMatchWins(t *testing.T) {
	eng, _ := Load([]byte(samplePolicy))

	tests := []struct {
		name    string
		call    ToolCall
		want    string
		ruleID  string
	}{
		{
			name: "secret blocked",
			call: ToolCall{ToolName: "Bash", ToolArgs: map[string]any{"command": "echo $TOKEN"}},
			want: "block", ruleID: "block-secret",
		},
		{
			name: "lan blocked",
			call: ToolCall{ToolName: "Bash", ToolArgs: map[string]any{"command": "curl 192.168.1.1"}},
			want: "block", ruleID: "lan",
		},
		{
			name: "read allowed",
			call: ToolCall{ToolName: "Read", ToolArgs: map[string]any{"path": "/etc/hosts"}},
			want: "allow", ruleID: "safe",
		},
		{
			name: "fallthrough to catch-all",
			call: ToolCall{ToolName: "UnknownTool", ToolArgs: map[string]any{}},
			want: "log_only", ruleID: "catch-all",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := eng.Check(tc.call)
			if d.Decision != tc.want {
				t.Errorf("decision = %q, want %q", d.Decision, tc.want)
			}
			if d.RuleID != tc.ruleID {
				t.Errorf("rule_id = %q, want %q", d.RuleID, tc.ruleID)
			}
			if d.PolicySig == "" {
				t.Errorf("policy_signature is empty")
			}
		})
	}
}

func TestCheckNoMatchDefaultsToAllow(t *testing.T) {
	// Policy without a catch-all rule
	eng, _ := Load([]byte(`
version: 1
name: strict-no-fallback
rules:
  - id: only-secret
    match: { tool_args_regex: "secret" }
    decision: block
`))
	d := eng.Check(ToolCall{ToolName: "Read", ToolArgs: map[string]any{"path": "/x"}})
	if d.Decision != "allow" {
		t.Errorf("expected default allow, got %q", d.Decision)
	}
	if !strings.Contains(d.Reason, "no rule matched") {
		t.Errorf("reason = %q, want 'no rule matched'", d.Reason)
	}
}

func TestSignatureStable(t *testing.T) {
	eng1, _ := Load([]byte(samplePolicy))
	eng2, _ := Load([]byte(samplePolicy))
	if eng1.Signature() != eng2.Signature() {
		t.Errorf("signature should be stable: %s != %s", eng1.Signature(), eng2.Signature())
	}
}