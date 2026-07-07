"""
AISentinel + LangGraph example.

Prerequisites:
    pip install langgraph langchain-openai
    go install github.com/Kabzhanov/AISentinel/cmd/aisentinel@latest
    aisentinel serve &  # in another terminal, or as a subprocess

This example:
1. Spawns `aisentinel serve` as a subprocess.
2. Connects to it via stdio JSON-RPC (manual implementation; replace with
   `mcp` Python client in production).
3. Gates every tool call through `aisentinel_check_policy`.
4. Logs every decision via `aisentinel_log_event`.

NOTE: This is a demonstration. In production use the official
`mcp` Python package to talk to the AISentinel MCP server.
"""
import json
import subprocess
import sys
from typing import Any


class AisentinelClient:
    """Minimal JSON-RPC client for the AISentinel MCP server over stdio."""

    def __init__(self, policy_path: str = "policies/default.yaml"):
        self.proc = subprocess.Popen(
            ["aisentinel", "serve", "--policy", policy_path],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
            bufsize=1,
        )
        self._req_id = 0
        self._initialize()

    def _initialize(self):
        self._req_id += 1
        req = {
            "jsonrpc": "2.0",
            "id": self._req_id,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {"name": "langgraph-demo", "version": "0.1.0"},
            },
        }
        self._send(req)
        # Send initialized notification
        self._send({"jsonrpc": "2.0", "method": "notifications/initialized"})
        resp = self._recv()
        assert "result" in resp, f"initialize failed: {resp}"
        print(f"✓ Connected to {resp['result']['serverInfo']['name']} "
              f"v{resp['result']['serverInfo']['version']} "
              f"({resp['result']['serverInfo'].get('vendor', '?')})")

    def _send(self, msg: dict):
        self.proc.stdin.write(json.dumps(msg) + "\n")
        self.proc.stdin.flush()

    def _recv(self) -> dict:
        line = self.proc.stdout.readline()
        if not line:
            raise EOFError("aisentinel process exited")
        return json.loads(line)

    def check_policy(self, tool_name: str, tool_args: dict, agent_id: str = "demo-agent") -> dict:
        self._req_id += 1
        req = {
            "jsonrpc": "2.0",
            "id": self._req_id,
            "method": "tools/call",
            "params": {
                "name": "aisentinel_check_policy",
                "arguments": {
                    "tool_name": tool_name,
                    "tool_args": tool_args,
                    "agent_id": agent_id,
                },
            },
        }
        self._send(req)
        resp = self._recv()
        content = resp.get("result", {}).get("content", [])
        if not content:
            return {"decision": "error", "raw": resp}
        return json.loads(content[0]["text"])

    def close(self):
        self.proc.terminate()
        self.proc.wait(timeout=5)


def main():
    client = AisentinelClient("policies/default.yaml")
    try:
        # ---- Demo 1: safe Bash call (allow) ----
        result = client.check_policy("Bash", {"command": "ls -la"})
        print(f"\n[Bash: ls -la]")
        print(f"  decision: {result['decision']}  (reason: {result.get('reason', '—')})")

        # ---- Demo 2: outbound network (require_approval) ----
        result = client.check_policy(
            "Bash",
            {"command": "curl https://webhook.site/abc"},
            agent_id="demo-agent",
        )
        print(f"\n[Bash: curl https://webhook.site/abc]")
        print(f"  decision: {result['decision']}  (reason: {result.get('reason', '—')})")

        # ---- Demo 3: secret in args (block) ----
        result = client.check_policy(
            "Bash",
            {"command": "echo $OPENAI_API_KEY"},
            agent_id="demo-agent",
        )
        print(f"\n[Bash: echo $OPENAI_API_KEY]")
        print(f"  decision: {result['decision']}  (reason: {result.get('reason', '—')})")

        # ---- Demo 4: LAN access (block) ----
        result = client.check_policy(
            "Bash",
            {"command": "ping 192.168.0.1"},
            agent_id="demo-agent",
        )
        print(f"\n[Bash: ping 192.168.0.1]")
        print(f"  decision: {result['decision']}  (reason: {result.get('reason', '—')})")

    finally:
        client.close()


if __name__ == "__main__":
    sys.exit(main())