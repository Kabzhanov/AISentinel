// Package logger writes JSONL audit logs for AISentinel.
package logger

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/Kabzhanov/AISentinel/internal/secrets"
)

// Logger appends events as JSON lines.
type Logger struct {
	mu   sync.Mutex
	file *os.File
}

// Event is the standard event schema. Mirrors docs/event-schema.md.
type Event struct {
	EventID       string         `json:"event_id"`
	Timestamp     string         `json:"timestamp"`
	AgentID       string         `json:"agent_id,omitempty"`
	SessionID     string         `json:"session_id,omitempty"`
	EventType     string         `json:"event_type"` // pre_tool | post_tool | prompt | decision | system
	Prompt        string         `json:"prompt,omitempty"`
	ToolName      string         `json:"tool_name,omitempty"`
	ToolArgs      map[string]any `json:"tool_args,omitempty"`
	ToolResult    map[string]any `json:"tool_result,omitempty"`
	Decision      string         `json:"decision,omitempty"`
	PolicyMatched []string       `json:"policy_matched,omitempty"`
	RiskSignals   []string       `json:"risk_signals,omitempty"`
	PolicySig     string         `json:"policy_signature,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// Open opens (or creates) a JSONL file in append mode.
func Open(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	return &Logger{file: f}, nil
}

// Write appends one event as a single JSON line. Before writing, it redacts
// known secret-value shapes (AWS keys, OpenAI/GitHub/Slack tokens, URL
// credentials, PEM private keys, ...) from ToolArgs, ToolResult, and
// Metadata so raw secrets never land on disk in the audit trail.
func (l *Logger) Write(ev Event) error {
	Redact(&ev)
	if ev.Timestamp == "" {
		ev.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if ev.EventID == "" {
		ev.EventID = newEventID()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil // dry-run
	}
	enc := json.NewEncoder(l.file)
	if err := enc.Encode(ev); err != nil {
		return err
	}
	return nil
}

// Redact walks ev.ToolArgs, ev.ToolResult, and ev.Metadata (including
// nested maps/slices/strings — arbitrary JSON-ish `any` values) and
// replaces any string value that matches a known secret pattern with
// secrets.Redacted. It mutates ev in place.
//
// This is the single point where audit-log redaction happens; both the
// `aisentinel serve` MCP server (internal/server) and the sidecar
// (cmd/aisentinel-sidecar) write events through Logger.Write, which calls
// Redact automatically — callers don't need to redact manually.
func Redact(ev *Event) {
	if ev == nil {
		return
	}
	ev.Prompt = secrets.RedactString(ev.Prompt)
	ev.ToolArgs = redactMap(ev.ToolArgs)
	ev.ToolResult = redactMap(ev.ToolResult)
	ev.Metadata = redactMap(ev.Metadata)
}

// redactMap returns a copy of m with all string values (recursively, through
// nested maps and slices) passed through secrets.RedactString.
func redactMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = redactValue(v)
	}
	return out
}

func redactValue(v any) any {
	switch t := v.(type) {
	case string:
		return secrets.RedactString(t)
	case map[string]any:
		return redactMap(t)
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = redactValue(e)
		}
		return out
	default:
		return v
	}
}

// Close flushes and closes the file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}

// DryRun returns a no-op logger.
func DryRun() *Logger { return &Logger{} }

var evCounter uint64

func newEventID() string {
	evCounter++
	return time.Now().UTC().Format("20060102T150405.000000000Z") + "-" + itoa(evCounter)
}

func itoa(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}