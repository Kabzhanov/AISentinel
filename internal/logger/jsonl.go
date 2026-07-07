// Package logger writes JSONL audit logs for AISentinel.
package logger

import (
	"encoding/json"
	"os"
	"sync"
	"time"
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

// Write appends one event as a single JSON line.
func (l *Logger) Write(ev Event) error {
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