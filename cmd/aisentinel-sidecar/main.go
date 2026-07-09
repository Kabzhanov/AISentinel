// AISentinel sidecar — transparent proxy that enforces policy on every MCP
// tools/call between a client and the wrapped MCP server.
//
// Usage:
//
//	aisentinel-sidecar --target python -m my_mcp_server
//	aisentinel-sidecar --policy strict.yaml --target ./my-server
//
// Wraps ANY stdio MCP server. Blocked calls never reach the server;
// the client receives a JSON-RPC error instead.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Kabzhanov/AISentinel/internal/iox"
	"github.com/Kabzhanov/AISentinel/internal/logger"
	"github.com/Kabzhanov/AISentinel/internal/policy"
)

// version is overridden at build time via:
//
//	go build -ldflags "-X main.version=1.2.3"
//
// CI's release job sets it from the git tag (see .github/workflows/ci.yml).
// Local/`go install` builds without that flag report "dev".
var version = "dev"

func main() {
	policyPath := flag.String("policy", "", "policy YAML file (default: $AISENTINEL_POLICY, then ./policies/default.yaml, then built-in default)")
	logDir := flag.String("log-dir", "", "audit log directory (default: $AISENTINEL_LOG_DIR or ~/.aisentinel)")
	dryRun := flag.Bool("dry-run", false, "never block — only log decisions (also via $AISENTINEL_DRY_RUN=1)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Printf("aisentinel-sidecar v%s — by Kabzhanov / BizDNAi\n", version)
		return
	}

	// Positional args = command to wrap (and its arguments).
	target := flag.Args()
	if len(target) == 0 {
		fmt.Fprintln(os.Stderr, "error: command to wrap is required (positional arguments)")
		usage()
		os.Exit(2)
	}

	if *logDir == "" {
		*logDir = os.Getenv("AISENTINEL_LOG_DIR")
	}
	if *logDir == "" {
		home, _ := os.UserHomeDir()
		*logDir = filepath.Join(home, ".aisentinel")
	}
	if !*dryRun {
		if v := os.Getenv("AISENTINEL_DRY_RUN"); v == "1" || v == "true" {
			*dryRun = true
		}
	}

	// Load policy: --policy > $AISENTINEL_POLICY > ./policies/default.yaml
	// (if present) > built-in embedded default. Never fails just because no
	// policy file is reachable — only when an explicit path/env is set but
	// unreadable/invalid.
	eng, policySource, err := policy.Resolve(*policyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sidecar: load policy %s: %v\n", policySource, err)
		os.Exit(1)
	}
	*policyPath = policySource

	// Open audit logger
	if err := os.MkdirAll(*logDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sidecar: mkdir log dir: %v\n", err)
		os.Exit(1)
	}
	logPath := filepath.Join(*logDir, "events-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	// In dry-run mode we still write audit events; we just never block.
	// This is critical for shadow rollouts: see what would have been blocked.
	var log *logger.Logger
	log, err = logger.Open(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sidecar: open log: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	fmt.Fprintf(os.Stderr, "aisentinel-sidecar v%s\n", version)
	fmt.Fprintf(os.Stderr, "  policy: %s (sig=%s)\n", *policyPath, eng.Signature())
	fmt.Fprintf(os.Stderr, "  log:    %s\n", logPath)
	fmt.Fprintf(os.Stderr, "  target: %s\n", strings.Join(target, " "))
	fmt.Fprintf(os.Stderr, "  mode:   %s\n", map[bool]string{true: "dry-run (audit only)", false: "enforce"}[*dryRun])

	// Spawn target subprocess
	cmd := exec.Command(target[0], target[1:]...)
	cmd.Stderr = os.Stderr // pass through server stderr
	inPipe, err := cmd.StdinPipe()
	if err != nil {
		fatal("stdin pipe: %v", err)
	}
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		fatal("stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		fatal("start target: %v", err)
	}

	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup

	// Both the "server → client" copy loop (goroutine B) and the blocked-call
	// JSON-RPC error response written from intercept() (goroutine A) write to
	// os.Stdout. Route both through one lockedLineWriter so a block response
	// can never get interleaved mid-line with a line the target server is
	// forwarding through, and vice versa.
	stdout := iox.NewLockedLineWriter(os.Stdout)

	// Goroutine A: client → (intercept) → server
	wg.Add(1)
	go func() {
		defer wg.Done()
		intercept(os.Stdin, inPipe, stdout, eng, log, *dryRun, "client", "")
	}()

	// Goroutine B: server → client, forwarded line-by-line so writes to
	// os.Stdout stay atomic per line (see stdout comment above).
	wg.Add(1)
	go func() {
		defer wg.Done()
		copyLines(outPipe, stdout)
	}()

	// Wait for either target exit or signal
	doneCh := make(chan error, 1)
	go func() { doneCh <- cmd.Wait() }()

	select {
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		<-doneCh
	case err := <-doneCh:
		if err != nil {
			fmt.Fprintf(os.Stderr, "sidecar: target exited: %v\n", err)
		}
	}

	_ = inPipe.Close()
	wg.Wait()
}

// intercept reads newline-delimited JSON-RPC from src, evaluates each
// tools/call against the policy, and writes the result to dst (the target
// server's stdin). Blocked calls instead write a JSON-RPC error to stdout
// (the client side) via the shared lockedLineWriter, and never reach dst.
func intercept(src io.Reader, dst io.Writer, stdout *iox.LockedLineWriter, eng *policy.Engine, log *logger.Logger,
	dryRun bool, direction string, agentID string) {

	scanner := bufio.NewScanner(src)
	// MCP messages can be large; raise the buffer cap.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	out := bufio.NewWriter(dst)
	defer out.Flush()

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var msg map[string]any
		if err := json.Unmarshal(line, &msg); err != nil {
			// Not JSON or malformed — forward as-is.
			_, _ = out.Write(line)
			_ = out.WriteByte('\n')
			_ = out.Flush()
			continue
		}

		method, _ := msg["method"].(string)
		if method != "tools/call" {
			// Non-tool-call — forward transparently.
			_, _ = out.Write(line)
			_ = out.WriteByte('\n')
			_ = out.Flush()
			continue
		}

		// Parse tool name + arguments
		params, _ := msg["params"].(map[string]any)
		toolName, _ := params["name"].(string)
		toolArgs, _ := params["arguments"].(map[string]any)

		// Evaluate policy
		call := policy.ToolCall{
			ToolName:  toolName,
			ToolArgs:  toolArgs,
			AgentID:   agentID,
			SessionID: os.Getenv("AISENTINEL_SESSION_ID"),
		}
		dec := eng.Check(call)

		// Log every decision
		_ = log.Write(logger.Event{
			EventType:     "pre_tool",
			ToolName:      toolName,
			ToolArgs:      toolArgs,
			Decision:      dec.Decision,
			PolicyMatched: dec.PolicyMatched,
			RiskSignals:   dec.RiskSignals,
			PolicySig:     dec.PolicySig,
			Metadata: map[string]any{
				"jsonrpc_id": msg["id"],
				"direction":  direction,
				"dry_run":    dryRun,
				"reason":     dec.Reason,
				"rule_id":    dec.RuleID,
			},
		})

		// Decide whether to forward
		blocked := dec.Decision == "block" ||
			(dec.Decision == "require_human_approval" && !dryRun)

		if !blocked {
			// Forward to server
			_, _ = out.Write(line)
			_ = out.WriteByte('\n')
			_ = out.Flush()
			continue
		}

		// Blocked — respond with JSON-RPC error to client (NOT to server)
		id := msg["id"]
		errResp := map[string]any{
			"jsonrpc": "2.0",
			"id":      id,
			"error": map[string]any{
				"code":    -32000,
				"message": fmt.Sprintf("aisentinel: blocked by policy (rule=%s): %s", dec.RuleID, dec.Reason),
				"data": map[string]any{
					"decision":       dec.Decision,
					"rule_id":        dec.RuleID,
					"reason":         dec.Reason,
					"policy_sig":     dec.PolicySig,
					"policy_matched": dec.PolicyMatched,
				},
			},
		}
		respBytes, _ := json.Marshal(errResp)
		if err := stdout.WriteLine(respBytes); err != nil {
			fmt.Fprintf(os.Stderr, "sidecar: write block response: %v\n", err)
		}

		fmt.Fprintf(os.Stderr, "  [BLOCK] %s (rule=%s) %s\n", toolName, dec.RuleID, dec.Reason)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "sidecar: read %s: %v\n", direction, err)
	}
}

// copyLines forwards newline-delimited output from src to dst one whole
// line at a time via dst.WriteLine, instead of io.Copy's raw byte-stream
// copy. This keeps every line dst writes atomic with respect to any other
// concurrent WriteLine caller sharing the same underlying writer (here,
// intercept's blocked-call JSON-RPC error response also targets stdout) —
// io.Copy gives no such guarantee and could interleave partial JSON-RPC
// lines from the two sources.
func copyLines(src io.Reader, dst *iox.LockedLineWriter) {
	scanner := bufio.NewScanner(src)
	// Same increased buffer cap as intercept's scanner — MCP messages can be
	// large (e.g. tool results with big payloads).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		if err := dst.WriteLine(scanner.Bytes()); err != nil {
			fmt.Fprintf(os.Stderr, "sidecar: write stdout: %v\n", err)
			return
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "sidecar: read target stdout: %v\n", err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `aisentinel-sidecar v%s — transparent policy proxy for stdio MCP servers

Usage:
  aisentinel-sidecar [flags] CMD [ARGS...]
  aisentinel-sidecar --policy FILE CMD [ARGS...]

CMD [ARGS...] is the stdio MCP server to wrap. Every tools/call is checked
against the policy BEFORE reaching the server. Blocked calls return a
JSON-RPC error to the client. Every decision is written to the JSONL audit log.

Examples:
  # Wrap the example echo server
  aisentinel-sidecar ./bin/echo-server

  # Wrap any MCP server with strict policy
  aisentinel-sidecar --policy policies/strict.yaml python -m my_mcp

Environment:
  AISENTINEL_POLICY       default policy file (overridden by --policy)
  AISENTINEL_LOG_DIR      default log directory (overridden by --log-dir)
  AISENTINEL_DRY_RUN      "1" to never block — only log decisions
  AISENTINEL_SESSION_ID   optional session identifier in audit events

Flags:
`, version)
	flag.PrintDefaults()
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "sidecar: "+format+"\n", args...)
	os.Exit(1)
}
