// Minimal MCP echo server for testing aisentinel-sidecar.
// Echoes any tools/call argument back as the tool result.
//
// Usage: echo-server
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	fmt.Fprintln(os.Stderr, "echo-server: started")

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req map[string]any
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		id := req["id"]
		method, _ := req["method"].(string)

		var resp map[string]any
		switch method {
		case "initialize":
			resp = map[string]any{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]any{
					"protocolVersion": "2024-11-05",
					"capabilities":    map[string]any{"tools": map[string]any{}},
					"serverInfo":      map[string]any{"name": "echo-server", "version": "0.1.0"},
				},
			}

		case "notifications/initialized":
			continue // no response

		case "tools/list":
			resp = map[string]any{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]any{
					"tools": []map[string]any{
						{"name": "echo", "description": "Echo arguments back", "inputSchema": map[string]any{"type": "object"}},
						{"name": "Bash", "description": "Run a shell command (test for policy)", "inputSchema": map[string]any{"type": "object"}},
						{"name": "Read", "description": "Read a file (test for policy)", "inputSchema": map[string]any{"type": "object"}},
					},
				},
			}

		case "tools/call":
			params, _ := req["params"].(map[string]any)
			name, _ := params["name"].(string)
			args, _ := params["arguments"].(map[string]any)
			resp = map[string]any{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]any{
					"content": []map[string]any{
						{"type": "text", "text": fmt.Sprintf("echo-server: tool=%s args=%v", name, args)},
					},
				},
			}

		default:
			resp = map[string]any{
				"jsonrpc": "2.0",
				"id":      id,
				"error":   map[string]any{"code": -32601, "message": "method not found: " + method},
			}
		}

		bytes, _ := json.Marshal(resp)
		out.Write(bytes)
		out.WriteByte('\n')
		out.Flush()
	}
}
