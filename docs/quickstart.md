# Quickstart

Three steps from clone to protected agent.

## 1. Install

```bash
go install github.com/Kabzhanov/AISentinel/cmd/aisentinel@latest
```

Verify:

```bash
aisentinel version
# AISentinel v1.0.0 — by Kabzhanov / BizDNAi / AI Trust Index
```

## 2. Pick a policy

AISentinel ships three:

| Policy | Use when |
|---|---|
| `policies/default.yaml` | Normal agents — allow with a safety net. |
| `policies/strict.yaml` | Sensitive agents — deny by default, whitelist only. |
| `policies/audit-only.yaml` | First rollout — log everything, block nothing. |

Validate yours:

```bash
aisentinel validate-policy policies/default.yaml
```

## 3. Wire it up

### Claude Code / Claude Desktop / Cursor / Cline

Add to your MCP config (e.g. `~/.claude/mcp.json` or `~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "aisentinel": {
      "command": "aisentinel",
      "args": ["serve", "--policy", "/absolute/path/to/policies/default.yaml"]
    }
  }
}
```

Restart your client. From now on, every tool call is gated.

### Standalone (CLI test)

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"curl","version":"1"}}}' | aisentinel serve
```

This should respond with `result.serverInfo.vendor = "BizDNAi — AI Trust Index"`.

## 4. Inspect the audit log

```bash
aisentinel events --last 20
```

Default log directory is `~/.aisentinel/events-YYYY-MM-DD.jsonl`. Override
with `AISENTINEL_LOG_DIR=/path/to/logs`.

## 5. Shadow mode

Want to test a policy without blocking anything?

```bash
AISENTINEL_DRY_RUN=1 aisentinel serve --policy policies/default.yaml
```

## What's next?

- Read [docs/threat-model.md](threat-model.md) for what AISentinel covers.
- Customise your policy — see [docs/event-schema.md](event-schema.md) for
  what events look like, then craft matchers.
- When ready for SaaS / HTTP transport, watch the v1.2 release.