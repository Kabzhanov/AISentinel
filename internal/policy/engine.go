// Package policy loads and evaluates YAML policy files for AISentinel.
package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Engine is the loaded policy.
type Engine struct {
	Version int    `yaml:"version"`
	Name    string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Rules   []Rule `yaml:"rules"`
}

// Rule matches a tool call and decides what to do.
type Rule struct {
	ID       string         `yaml:"id"`
	Match    Match          `yaml:"match"`
	Decision string         `yaml:"decision"` // allow | block | require_human_approval | log_only
	Reason   string         `yaml:"reason,omitempty"`
	Metadata map[string]any `yaml:"metadata,omitempty"`

	// compiled cache (not serialized)
	toolNameRe  *regexp.Regexp
	argsRe      *regexp.Regexp
}

// Match is the rule matcher.
type Match struct {
	ToolName        string `yaml:"tool_name,omitempty"`
	ToolNameRegex   string `yaml:"tool_name_regex,omitempty"`
	ToolArgsRegex   string `yaml:"tool_args_regex,omitempty"`
	ToolArgsContains string `yaml:"tool_args_contains,omitempty"`
}

// LoadFromFile parses a YAML policy file.
func LoadFromFile(path string) (*Engine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Load(data)
}

// Load parses policy YAML from bytes.
func Load(data []byte) (*Engine, error) {
	var e Engine
	if err := yaml.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("yaml parse: %w", err)
	}
	if e.Version == 0 {
		e.Version = 1
	}
	if e.Name == "" {
		e.Name = "unnamed"
	}
	if len(e.Rules) == 0 {
		return nil, fmt.Errorf("policy has no rules")
	}
	// Compile regexes
	for i := range e.Rules {
		r := &e.Rules[i]
		if r.Match.ToolNameRegex != "" {
			re, err := regexp.Compile(r.Match.ToolNameRegex)
			if err != nil {
				return nil, fmt.Errorf("rule %d (%s): bad tool_name_regex: %w", i, r.ID, err)
			}
			r.toolNameRe = re
		}
		if r.Match.ToolArgsRegex != "" {
			re, err := regexp.Compile(r.Match.ToolArgsRegex)
			if err != nil {
				return nil, fmt.Errorf("rule %d (%s): bad tool_args_regex: %w", i, r.ID, err)
			}
			r.argsRe = re
		}
		if r.Decision == "" {
			return nil, fmt.Errorf("rule %d (%s): decision is required", i, r.ID)
		}
		switch r.Decision {
		case "allow", "block", "require_human_approval", "log_only":
		default:
			return nil, fmt.Errorf("rule %d (%s): unknown decision %q", i, r.ID, r.Decision)
		}
	}
	return &e, nil
}

// Signature returns a SHA-256 fingerprint of the policy (for change detection).
func (e *Engine) Signature() string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d|%s|", e.Version, e.Name)))
	for _, r := range e.Rules {
		h.Write([]byte(r.ID + "|" + r.Decision + "|" + r.Match.ToolName + "|" +
			r.Match.ToolNameRegex + "|" + r.Match.ToolArgsRegex + "|" +
			r.Match.ToolArgsContains + "|"))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Decision is what we return from Check.
type Decision struct {
	Decision      string   `json:"decision"`
	Reason        string   `json:"reason,omitempty"`
	RuleID        string   `json:"rule_id,omitempty"`
	PolicyMatched []string `json:"policy_matched"`
	RiskSignals   []string `json:"risk_signals"`
	PolicySig     string   `json:"policy_signature"`
}

// ToolCall describes the call being checked.
type ToolCall struct {
	ToolName string         `json:"tool_name"`
	ToolArgs map[string]any `json:"tool_args"`
	AgentID  string         `json:"agent_id,omitempty"`
	SessionID string        `json:"session_id,omitempty"`
}

// argsAsString flattens args to a string for regex matching.
func argsAsString(args map[string]any) string {
	if args == nil {
		return ""
	}
	var b strings.Builder
	for k, v := range args {
		fmt.Fprintf(&b, "%s=%v\n", k, v)
	}
	return b.String()
}

// Check evaluates rules in order; first match wins. Returns default allow if
// no rule matches (last-rule is treated as fallback if Decision=="allow" but
// we follow strict first-match semantics).
func (e *Engine) Check(call ToolCall) Decision {
	matched := []string{}
	signals := []string{}

	for _, r := range e.Rules {
		if !matchRule(r, call) {
			continue
		}
		matched = append(matched, r.ID)
		if r.Decision == "block" || r.Decision == "require_human_approval" {
			signals = append(signals, "rule_matched:"+r.ID)
		}
		return Decision{
			Decision:      r.Decision,
			Reason:        r.Reason,
			RuleID:        r.ID,
			PolicyMatched: matched,
			RiskSignals:   signals,
			PolicySig:     e.Signature(),
		}
	}

	// Default: no rule matched → allow with no signal.
	return Decision{
		Decision:      "allow",
		Reason:        "no rule matched",
		PolicyMatched: matched,
		RiskSignals:   signals,
		PolicySig:     e.Signature(),
	}
}

func matchRule(r Rule, call ToolCall) bool {
	// Tool name match
	if r.Match.ToolName != "" && r.Match.ToolName != call.ToolName {
		return false
	}
	if r.toolNameRe != nil && !r.toolNameRe.MatchString(call.ToolName) {
		return false
	}
	// Args match
	args := argsAsString(call.ToolArgs)
	if r.Match.ToolArgsContains != "" && !strings.Contains(args, r.Match.ToolArgsContains) {
		return false
	}
	if r.argsRe != nil && !r.argsRe.MatchString(args) {
		return false
	}
	// If no matchers specified, rule applies to everything (broad rule)
	if r.Match.ToolName == "" && r.toolNameRe == nil &&
		r.Match.ToolArgsRegex == "" && r.Match.ToolArgsContains == "" {
		return true
	}
	// At least one matcher specified and matched
	return r.Match.ToolName != "" || r.toolNameRe != nil ||
		r.Match.ToolArgsContains != "" || r.argsRe != nil
}