# AISentinel Security Audit (2026-07-07)

> **Note:** This document is the **internal threat model** that drove AISentinel's
> v1.0 design. See [`threat-model.md`](threat-model.md) for the OWASP/ATLAS
> mapping. This file is the longer-form discussion.

## Why this exists

In 2026 an AI agent typically has:

- Shell access (`Bash` tool)
- File read/write (`Read`, `Write`, `Edit` tools)
- Browser control (CDP, Puppeteer)
- Database access (SQL, 1C OData, Bitrix24 REST)
- Outbound HTTP (`curl`, `requests`)
- E-mail (`Email_send`)
- Calendar / messaging APIs

These are all **legitimate** tools — antivirus does not flag them. Yet they
form a complete attack chain: an attacker who can convince the agent to use
them in combination (via prompt injection) can exfiltrate everything the
agent has access to.

AISentinel closes this gap by **gating every tool call** through a YAML
policy engine and **logging every decision** to an AI-Trust-Index-compatible
JSONL audit log.

## The 10 most-common holes AISentinel blocks

1. **Secret in arguments** — agent passes `OPENAI_API_KEY` to `Bash("echo $KEY")`.
2. **Mass read** — agent does `SELECT * FROM users` and exfiltrates.
3. **Outbound network from Bash** — agent reaches `curl webhook.site/...`.
4. **LAN pivot** — agent hits `192.168.0.1` to attack the local network.
5. **Write to sensitive paths** — `Write("/etc/cron.d/x", ...)`.
6. **Shell history leak** — agent reads `~/.bash_history`.
7. **Hidden-tool description drift** — MCP server updates tool description
   to include instructions.
8. **Indirect prompt injection** — PDF/email tells the agent to call
   `Email_send` to attacker.
9. **DNS tunneling** — agent beacons to a C2 domain.
10. **Tool squatting** — fake MCP server overrides a trusted tool name.

Each of these maps to either a built-in policy rule (see
`policies/default.yaml`) or a future control.

## See also

- [`threat-model.md`](threat-model.md) — the OWASP LLM Top 10 + MITRE ATLAS mapping.
- [`event-schema.md`](event-schema.md) — the event schema.
- https://bizdnai.com/index/ — AI Trust Index, the publishing channel for these audits.