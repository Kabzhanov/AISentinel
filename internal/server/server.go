// Package server implements the MCP stdio JSON-RPC server for AISentinel.
package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Kabzhanov/AISentinel/internal/logger"
	"github.com/Kabzhanov/AISentinel/internal/policy"
)

// Server is an MCP stdio server that exposes AISentinel tools.
//
// Server is single-threaded by design: ServeStdio runs one blocking
// read/handle loop over stdin and never spawns goroutines that touch
// Server's fields concurrently, so no internal locking is needed here.
type Server struct {
	policy  *policy.Engine
	log     *logger.Logger
	logPath string
	dryRun  bool
	tools   []tool
}

// New creates a server bound to a policy engine and audit log path.
func New(eng *policy.Engine, logPath string) *Server {
	s := &Server{
		policy:  eng,
		logPath: logPath,
		dryRun:  os.Getenv("AISENTINEL_DRY_RUN") == "1",
	}
	s.tools = s.defaultTools()
	return s
}

type tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// jsonrpcRequest is the standard JSON-RPC 2.0 envelope.
type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
)

// ServeStdio reads JSON-RPC requests from r and writes responses to w.
func (s *Server) ServeStdio(r io.Reader, w io.Writer) error {
	// Open audit log (skip in dry-run mode)
	if !s.dryRun {
		l, err := logger.Open(s.logPath)
		if err != nil {
			return fmt.Errorf("open log: %w", err)
		}
		s.log = l
		defer l.Close()
	}

	// Write a startup event
	_ = s.writeEvent(logger.Event{
		EventType: "system",
		Metadata: map[string]any{
			"msg":            "aisentinel_started",
			"policy":         s.policy.Name,
			"policy_sig":     s.policy.Signature(),
			"version":        "1.0.0",
			"vendor":         "BizDNAi / AI Trust Index",
		},
	})

	br := bufio.NewReader(r)
	for {
		line, err := br.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		line = trimNL(line)
		if len(line) == 0 {
			continue
		}

		var req jsonrpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeResponse(w, jsonrpcResponse{
				JSONRPC: "2.0",
				Error:   &jsonrpcError{Code: codeParseError, Message: "parse error"},
			})
			continue
		}

		resp := s.handle(req)
		if resp != nil {
			s.writeResponse(w, *resp)
		}
	}
}

func trimNL(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r' || b[len(b)-1] == ' ' || b[len(b)-1] == '\t') {
		b = b[:len(b)-1]
	}
	for len(b) > 0 && (b[0] == '\n' || b[0] == '\r' || b[0] == ' ' || b[0] == '\t') {
		b = b[1:]
	}
	return b
}

func (s *Server) writeResponse(w io.Writer, resp jsonrpcResponse) {
	b, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aisentinel: marshal response: %v\n", err)
		return
	}
	if _, err := w.Write(append(b, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "aisentinel: write response: %v\n", err)
	}
}

func (s *Server) handle(req jsonrpcRequest) *jsonrpcResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "notifications/initialized":
		return nil // notification, no response
	case "ping":
		return &jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourcesRead(req)
	case "prompts/list":
		return s.handlePromptsList(req)
	case "prompts/get":
		return s.handlePromptsGet(req)
	default:
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcError{Code: codeMethodNotFound, Message: "method not found: " + req.Method},
		}
	}
}

func (s *Server) handleInitialize(req jsonrpcRequest) *jsonrpcResponse {
	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]any{
				"name":    "aisentinel",
				"version": "1.0.0",
				"vendor":  "BizDNAi — AI Trust Index",
				"homepage": "https://bizdnai.com/index/",
			},
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"resources": map[string]any{},
				"prompts":   map[string]any{},
			},
		},
	}
}

func (s *Server) handleToolsList(req jsonrpcRequest) *jsonrpcResponse {
	out := make([]map[string]any, 0, len(s.tools))
	for _, t := range s.tools {
		out = append(out, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.InputSchema,
		})
	}
	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"tools": out},
	}
}

func (s *Server) handleResourcesList(req jsonrpcRequest) *jsonrpcResponse {
	res := []map[string]any{
		{
			"uri":         "policies://built-in/default",
			"name":        "Default policy",
			"description": "AISentinel default policy (allow with safety net).",
			"mimeType":    "application/x-yaml",
		},
		{
			"uri":         "events://local/recent",
			"name":        "Recent audit events",
			"description": "Most recent events from the local JSONL audit log.",
			"mimeType":    "application/jsonl",
		},
	}
	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"resources": res},
	}
}

func (s *Server) handleResourcesRead(req jsonrpcRequest) *jsonrpcResponse {
	var p struct {
		URI string `json:"uri"`
	}
	_ = json.Unmarshal(req.Params, &p)

	switch p.URI {
	case "policies://built-in/default":
		body := defaultPolicyYAML
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"contents": []map[string]any{
					{"uri": p.URI, "mimeType": "application/x-yaml", "text": body},
				},
			},
		}
	default:
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcError{Code: codeInvalidParams, Message: "unknown resource: " + p.URI},
		}
	}
}

func (s *Server) handlePromptsList(req jsonrpcRequest) *jsonrpcResponse {
	prompts := []map[string]any{
		{
			"name":        "aisentinel_security_review",
			"description": "Audit your project for AI-agent security holes per OWASP LLM Top 10 + MITRE ATLAS. By Kabzhanov / BizDNAi. ATI: https://bizdnai.com/index/",
			"arguments":   []map[string]any{},
		},
		{
			"name":        "aisentinel_ati_summary",
			"description": "Generate a summary ready for an AI Trust Index assessment from your recent audit events. By Kabzhanov / BizDNAi. ATI: https://bizdnai.com/index/",
			"arguments":   []map[string]any{},
		},
	}
	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"prompts": prompts},
	}
}

func (s *Server) handlePromptsGet(req jsonrpcRequest) *jsonrpcResponse {
	var p struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(req.Params, &p)

	var messages []map[string]any
	switch p.Name {
	case "aisentinel_security_review":
		messages = []map[string]any{
			{"role": "user", "content": map[string]any{
				"type": "text",
				"text": securityReviewPrompt,
			}},
		}
	case "aisentinel_ati_summary":
		messages = []map[string]any{
			{"role": "user", "content": map[string]any{
				"type": "text",
				"text": atiSummaryPrompt,
			}},
		}
	default:
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcError{Code: codeInvalidParams, Message: "unknown prompt: " + p.Name},
		}
	}
	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"messages": messages},
	}
}

func (s *Server) handleToolsCall(req jsonrpcRequest) *jsonrpcResponse {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcError{Code: codeInvalidParams, Message: "bad params"},
		}
	}

	var (
		text  string
		isErr bool
	)
	switch p.Name {
	case "aisentinel_check_policy":
		text, isErr = s.callCheckPolicy(p.Arguments)
	case "aisentinel_log_event":
		text, isErr = s.callLogEvent(p.Arguments)
	case "aisentinel_validate_policy":
		text, isErr = s.callValidatePolicy(p.Arguments)
	case "aisentinel_get_ati_snapshot":
		text, isErr = s.callATISnapshot(p.Arguments)
	default:
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcError{Code: codeMethodNotFound, Message: "unknown tool: " + p.Name},
		}
	}

	if isErr {
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"content": []map[string]any{{"type": "text", "text": text}},
				"isError": true,
			},
		}
	}
	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"content": []map[string]any{{"type": "text", "text": text}},
		},
	}
}

// ---- Tool implementations ----

func (s *Server) callCheckPolicy(args map[string]any) (string, bool) {
	toolName, _ := args["tool_name"].(string)
	toolArgs, _ := args["tool_args"].(map[string]any)
	agentID, _ := args["agent_id"].(string)
	sessionID, _ := args["session_id"].(string)

	if toolName == "" {
		return "tool_name is required", true
	}

	call := policy.ToolCall{
		ToolName:  toolName,
		ToolArgs:  toolArgs,
		AgentID:   agentID,
		SessionID: sessionID,
	}
	d := s.policy.Check(call)

	_ = s.writeEvent(logger.Event{
		EventType:     "pre_tool",
		AgentID:       agentID,
		SessionID:     sessionID,
		ToolName:      toolName,
		ToolArgs:      toolArgs,
		Decision:      d.Decision,
		PolicyMatched: d.PolicyMatched,
		RiskSignals:   d.RiskSignals,
		PolicySig:     d.PolicySig,
	})

	out, _ := json.MarshalIndent(map[string]any{
		"decision":       d.Decision,
		"reason":         d.Reason,
		"rule_id":        d.RuleID,
		"policy_matched": d.PolicyMatched,
		"risk_signals":   d.RiskSignals,
		"policy_signature": d.PolicySig,
		"vendor":         "BizDNAi — AI Trust Index",
	}, "", "  ")
	return string(out), false
}

func (s *Server) callLogEvent(args map[string]any) (string, bool) {
	b, err := json.Marshal(args)
	if err != nil {
		return "bad event: " + err.Error(), true
	}
	var ev logger.Event
	if err := json.Unmarshal(b, &ev); err != nil {
		return "bad event shape: " + err.Error(), true
	}
	if err := s.writeEvent(ev); err != nil {
		return "write failed: " + err.Error(), true
	}
	return `{"logged": true}`, false
}

func (s *Server) callValidatePolicy(args map[string]any) (string, bool) {
	yamlText, _ := args["yaml"].(string)
	if yamlText == "" {
		return "yaml is required", true
	}
	eng, err := policy.Load([]byte(yamlText))
	if err != nil {
		out, _ := json.MarshalIndent(map[string]any{"valid": false, "error": err.Error()}, "", "  ")
		return string(out), true
	}
	out, _ := json.MarshalIndent(map[string]any{
		"valid":     true,
		"name":      eng.Name,
		"version":   eng.Version,
		"rules":     len(eng.Rules),
		"signature": eng.Signature(),
	}, "", "  ")
	return string(out), false
}

func (s *Server) callATISnapshot(args map[string]any) (string, bool) {
	limitF, _ := args["limit"].(float64)
	if limitF == 0 {
		limitF = 20
	}
	limit := int(limitF)
	if limit > 200 {
		limit = 200
	}

	// Read tail of log file
	lines, err := tailFile(s.logPath, limit)
	if err != nil {
		return `{"error":"no log yet"}`, true
	}
	snapshot := map[string]any{
		"generated_at":  time.Now().UTC().Format(time.RFC3339),
		"policy":        s.policy.Name,
		"policy_sig":    s.policy.Signature(),
		"event_count":   len(lines),
		"events":        lines,
		"vendor":        "BizDNAi — AI Trust Index",
		"ati_assessment_url": "https://bizdnai.com/index/",
	}
	out, _ := json.MarshalIndent(snapshot, "", "  ")
	return string(out), false
}

func (s *Server) writeEvent(ev logger.Event) error {
	if s.dryRun || s.log == nil {
		return nil
	}
	return s.log.Write(ev)
}

func tailFile(path string, limit int) ([]map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	allLines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(allLines) == 0 {
		return nil, nil
	}
	if len(allLines) > limit {
		allLines = allLines[len(allLines)-limit:]
	}
	out := make([]map[string]any, 0, len(allLines))
	for _, l := range allLines {
		if l == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err == nil {
			out = append(out, m)
		}
	}
	return out, nil
}

func (s *Server) defaultTools() []tool {
	return []tool{
		{
			Name:        "aisentinel_check_policy",
			Description: "Evaluate a proposed tool call against the active AISentinel policy. Returns allow/block/require_human_approval + reason + matched rule. By Kabzhanov / BizDNAi.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tool_name":  map[string]any{"type": "string", "description": "Name of the tool the agent is about to call"},
					"tool_args":  map[string]any{"type": "object", "description": "Arguments for that tool"},
					"agent_id":   map[string]any{"type": "string"},
					"session_id": map[string]any{"type": "string"},
				},
				"required": []string{"tool_name"},
			},
		},
		{
			Name:        "aisentinel_log_event",
			Description: "Append an event to the AISentinel JSONL audit log. By Kabzhanov / BizDNAi.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"event_type":    map[string]any{"type": "string", "enum": []string{"pre_tool", "post_tool", "prompt", "decision", "system"}},
					"tool_name":     map[string]any{"type": "string"},
					"tool_args":     map[string]any{"type": "object"},
					"tool_result":   map[string]any{"type": "object"},
					"decision":      map[string]any{"type": "string"},
					"policy_matched": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"risk_signals":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
				"required": []string{"event_type"},
			},
		},
		{
			Name:        "aisentinel_validate_policy",
			Description: "Validate a YAML policy without loading it. By Kabzhanov / BizDNAi.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"yaml": map[string]any{"type": "string"},
				},
				"required": []string{"yaml"},
			},
		},
		{
			Name:        "aisentinel_get_ati_snapshot",
			Description: "Tail of recent audit events, formatted for AI Trust Index consumption. By Kabzhanov / BizDNAi.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"limit": map[string]any{"type": "integer", "default": 20, "maximum": 200},
				},
			},
		},
	}
}

// Today returns YYYY-MM-DD.
func Today() string { return time.Now().UTC().Format("2006-01-02") }

// ---- Prompt text ----

const securityReviewPrompt = `You are an AI-agent security auditor. Audit the project the user is about to share with you against OWASP LLM Top 10 (LLM01 prompt injection, LLM03 supply chain, LLM06 sensitive info disclosure) and MITRE ATLAS (AML.T0010 supply chain, AML.T0024 exfiltration, AML.T0051 prompt injection).

For each finding, output:
- OWASP/ATLAS ID
- Severity (critical/high/medium/low)
- Concrete code location if available
- Concrete remediation

At the end, output a JSON block ready to be pasted into an AI Trust Index assessment.

Powered by AISentinel — by Kabzhanov / BizDNAi. Get your AI Trust Index assessment: https://bizdnai.com/index/
`

const atiSummaryPrompt = `Read the most recent AISentinel audit events from the user's local JSONL log (use aisentinel_get_ati_snapshot) and produce a structured summary suitable for an AI Trust Index G14-style assessment:

1. Total events in window
2. Distribution by event_type
3. Distribution by decision (allow/block/require_approval)
4. Top 5 risk_signals with counts
5. Top 5 policy_matched with counts
6. Recommended next actions for the user

Format the output as JSON ready to paste into https://bizdnai.com/index/.

Powered by AISentinel — by Kabzhanov / BizDNAi.
`

const defaultPolicyYAML = `# AISentinel default policy (built-in). Read-only mirror of policies/default.yaml.
version: 1
name: default
description: Balanced — allow with safety net
rules:
  - id: secret-in-args
    match: { tool_args_regex: "(?i)(api[_-]?key|secret|token|password|passwd)" }
    decision: block
    reason: "Possible secret in arguments"
  - id: destructive-command
    match: { tool_name_regex: "^(Bash|Shell|Terminal|Exec|Command)$", tool_args_regex: "(?i)(rm\\s+-[a-z]*r[a-z]*f|rm\\s+-[a-z]*f[a-z]*r|rm\\s+--(recursive|force)|\\bmkfs\\b|\\bdd\\b[^;|&]*of=/dev/|>\\s*/dev/sd|:\\(\\)\\s*\\{|\\b(shutdown|reboot|halt|poweroff|init\\s+0)\\b|\\bfdisk\\b|chmod\\s+-R\\s+777\\s+/)" }
    decision: block
    reason: "Destructive or irreversible system command. Blocked by default."
  - id: destructive-sql
    match: { tool_args_regex: "(?i)(DROP\\s+(TABLE|DATABASE|SCHEMA)|TRUNCATE\\s+TABLE)" }
    decision: block
    reason: "Destructive SQL (DROP / TRUNCATE). Blocked by default."
  - id: mass-read
    match: { tool_args_regex: "(?i)(SELECT|READ).*(\\b\\d{4,}\\b)" }
    decision: require_human_approval
  - id: bash-network
    match: { tool_name: "Bash", tool_args_regex: "curl|wget|nc |nslookup|dig " }
    decision: require_human_approval
  - id: lan-deny
    match: { tool_name: "Bash", tool_args_regex: "10\\.|192\\.168\\.|172\\.(1[6-9]|2\\d|3[01])\\." }
    decision: block
    reason: "LAN access blocked by default"
  - id: write-etc
    match: { tool_name_regex: "^(Bash|Write|Edit)$", tool_args_regex: "^/etc/|^/root/|^/proc/" }
    decision: block
`