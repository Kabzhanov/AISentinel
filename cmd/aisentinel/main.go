// AISentinel — Security, control, and observability layer for AI agents.
//
// Usage:
//
//	aisentinel serve                  # MCP stdio server
//	aisentinel validate-policy <file> # Validate a YAML policy
//	aisentinel version                # Print version info
//	aisentinel policies               # List built-in policies
//
// Part of the AISentinel project by Kabzhanov / BizDNAi.
// Licensed under Apache 2.0.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Kabzhanov/AISentinel/internal/policy"
	"github.com/Kabzhanov/AISentinel/internal/server"
)

const (
	version   = "1.0.6"
	policyDir = "policies"
	banner    = "AISentinel v" + version + " — by Kabzhanov / BizDNAi / AI Trust Index"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "serve":
		err = runServe(args)
	case "serve-http":
		fmt.Fprintln(os.Stderr, "streamable-HTTP transport: not yet implemented in v1.0 — use `serve` (stdio) for now")
		err = nil
	case "validate-policy":
		err = runValidatePolicy(args)
	case "version":
		fmt.Println(banner)
		err = nil
	case "policies":
		err = runListPolicies()
	case "events":
		err = runEvents(args)
	case "scan":
		fmt.Fprintln(os.Stderr, "`scan` command: coming in v1.1 — see docs/SECURITY_AUDIT.md")
		err = nil
	case "-h", "--help", "help":
		usage()
		err = nil
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `%s

Usage:
  aisentinel serve [--policy FILE]      Start MCP stdio server
  aisentinel validate-policy FILE       Validate a policy YAML
  aisentinel policies                   List available built-in policies
  aisentinel events [--last N]          Tail audit log (JSONL)
  aisentinel version                    Print version

Environment:
  AISENTINEL_POLICY     Default policy file path
  AISENTINEL_LOG_DIR    Audit log directory (default: ~/.aisentinel)
  AISENTINEL_DRY_RUN    If "1", never block — only log decisions

Docs: https://github.com/Kabzhanov/AISentinel
`, banner)
}

func runServe(args []string) error {
	explicitPath := explicitPolicyFlag(args)

	eng, policyPath, err := policy.Resolve(explicitPath)
	if err != nil {
		return fmt.Errorf("load policy %s: %w", policyPath, err)
	}

	logDir := os.Getenv("AISENTINEL_LOG_DIR")
	if logDir == "" {
		home, _ := os.UserHomeDir()
		logDir = filepath.Join(home, ".aisentinel")
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	srv := server.New(eng, filepath.Join(logDir, "events-"+today()+".jsonl"))
	fmt.Fprintf(os.Stderr, "%s\n  policy: %s\n  log:    %s\n", banner, policyPath, logDir)

	return srv.ServeStdio(os.Stdin, os.Stdout)
}

func runValidatePolicy(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: aisentinel validate-policy FILE")
	}
	path := args[0]
	eng, err := policy.LoadFromFile(path)
	if err != nil {
		return fmt.Errorf("invalid: %w", err)
	}
	out := map[string]any{
		"valid":     true,
		"file":      path,
		"name":      eng.Name,
		"version":   eng.Version,
		"rules":     len(eng.Rules),
		"actions":   ruleActions(eng),
		"signature": eng.Signature(),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func runListPolicies() error {
	entries, err := os.ReadDir(policyDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no built-in policies dir (%s): %v\n", policyDir, err)
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && (strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml")) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	fmt.Println("Built-in policies (relative to ./policies):")
	for _, n := range names {
		fmt.Printf("  - %s\n", n)
	}
	return nil
}

func runEvents(args []string) error {
	limit := 20
	for i, a := range args {
		if a == "--last" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &limit)
		}
	}
	logDir := os.Getenv("AISENTINEL_LOG_DIR")
	if logDir == "" {
		home, _ := os.UserHomeDir()
		logDir = filepath.Join(home, ".aisentinel")
	}
	files, err := filepath.Glob(filepath.Join(logDir, "events-*.jsonl"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "no event logs in %s\n", logDir)
		return nil
	}
	sort.Strings(files)
	last := files[len(files)-1]
	f, err := os.Open(last)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	count := 0
	ring := make([]map[string]any, 0, limit)
	for {
		var ev map[string]any
		if err := dec.Decode(&ev); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		ring = append(ring, ev)
		if len(ring) > limit {
			ring = ring[1:]
		}
		count++
	}
	fmt.Fprintf(os.Stderr, "showing last %d of %d events from %s\n", len(ring), count, last)
	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "  ")
	for _, ev := range ring {
		_ = out.Encode(ev)
	}
	return nil
}

// explicitPolicyFlag extracts an explicit --policy value from args, if any.
// It does NOT consult $AISENTINEL_POLICY or fall back to a default path —
// that resolution order is centralized in policy.Resolve.
func explicitPolicyFlag(args []string) string {
	for i, a := range args {
		if a == "--policy" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func ruleActions(eng *policy.Engine) map[string]int {
	m := map[string]int{}
	for _, r := range eng.Rules {
		m[r.Decision]++
	}
	return m
}

func today() string {
	return server.Today()
}