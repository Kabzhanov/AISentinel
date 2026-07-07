# Threat Model

AISentinel is built against the **OWASP Top 10 for LLM Applications** (2025)
and **MITRE ATLAS**. This document maps each technique to one or more
AISentinel controls.

## OWASP LLM Top 10 (2025)

| ID | Title | AISentinel control |
|---|---|---|
| LLM01 | Prompt Injection (direct & indirect) | `prompt_guard` (planned v1.1); `pre_tool_guard` blocks sensitive tools after `<untrusted>` markers. |
| LLM02 | Sensitive Information Disclosure | `secret-in-args` rule blocks tokens/passwords in tool_args; `audit-only` mode captures all reads. |
| LLM03 | Supply Chain | `policy_signature` field tracks policy version; future `version --check` pins dependencies. |
| LLM04 | Data and Model Poisoning | Out of scope (training-time). |
| LLM05 | Improper Output Handling | Built-in: callers must treat tool output as untrusted; AISentinel never executes tool output as code. |
| LLM06 | Excessive Agency | **Core use case.** `pre_tool_guard` denies dangerous tool calls unless policy says otherwise. |
| LLM07 | System Prompt Leakage | Out of scope for AISentinel; the policy itself is treated as semi-private. |
| LLM08 | Vector and Embedding Weaknesses | Out of scope. |
| LLM09 | Misinformation | Out of scope. |
| LLM10 | Unbounded Consumption | Default policy caps `SELECT` row counts; rate-limit rules possible. |

## MITRE ATLAS (selected)

| Technique | Description | AISentinel control |
|---|---|---|
| AML.T0010 | ML Supply Chain Compromise | `policy_signature` + audit log; future dependency pinning. |
| AML.T0024 | Exfiltration via Cyber Means | `bash-network` + `lan-deny` + outbound rate-limit rules. |
| AML.T0051 | LLM Prompt Injection | Same as LLM01. |
| AML.T0048 | Erode ML Model Integrity | Out of scope. |

## AISentinel-specific threats

- **Tool description drift** — an MCP server updates its tool description
  to include hidden instructions. AISentinel emits `policy_signature` on
  every event; future `aisentinel scan` will hash tool descriptions at
  startup and alert on drift.
- **Tool squatting** — two MCP servers register the same tool name. The
  client resolves by priority and may pick the wrong one. Out of scope
  for AISentinel (depends on MCP-client behaviour); mitigate with strict
  per-server allowlist.
- **DNS tunneling** — agent beacons to `data-<uuid>.attacker.com`. Default
  policy requires approval for any `Bash` call to `nslookup`/`dig`. Combined
  with a local DNS sinkhole (Pi-hole, dnsmasq), full visibility.
- **Local IPC pivot** — agent hits `127.0.0.1:<port>` of another local
  process (VSCode, Slack). Default policy doesn't block localhost, but
  the user can add a rule.

## What AISentinel does NOT cover

- **Host compromise** — if the host OS is rooted, all bets are off.
- **Compromised LLM provider** — a malicious upstream model can craft
  tool calls that bypass string-matching rules.
- **Side-channels** — timing, power analysis. Out of scope.
- **Cryptographic weaknesses** — out of scope.
- **Physical access** — out of scope.

## See also

- [`SECURITY_AUDIT.md`](SECURITY_AUDIT.md) — the audit that produced this model.
- [`event-schema.md`](event-schema.md) — the event schema used in audit logs.
- AI Trust Index — https://bizdnai.com/index/