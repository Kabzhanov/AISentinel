# Security Policy

AISentinel is a security tool — we take reports seriously.

## Reporting a Vulnerability

Please email **kabzhanov@gmail.com** (GPG encouraged but not required).

Please include:
- Description of the vulnerability
- Steps to reproduce (PoC preferred)
- Affected versions
- Your assessment of impact

We aim to acknowledge within **48 hours** and ship a fix within **7 days**
for critical issues.

## What AISentinel Does and Does Not Protect Against

AISentinel is a **runtime policy gate** for AI agents. It does not replace
host-based security (antivirus, EDR), network security (firewalls, IDS),
or identity (SSO, MFA).

It **does** help with:

- Tool-call policy enforcement (block `curl` to attacker.com, etc.)
- Prompt-injection detection (indirect injection via untrusted content)
- Audit logging compatible with [AI Trust Index](https://bizdnai.com/index/)
- Drift detection on tool descriptions (MCP-specific)

It **does not** protect against:

- Compromise of the host running the agent
- Side-channel attacks on shared infrastructure
- A malicious or compromised LLM provider

## Threat Model

See [`docs/threat-model.md`](docs/threat-model.md) for the OWASP LLM Top 10
+ MITRE ATLAS mapping that AISentinel is built against.

## Security Practices in This Repo

- No telemetry, no phone-home, no remote config push.
- All policy files are local; the binary never reads from a remote URL.
- Dependencies pinned via `go.sum`; review `go.mod` changes carefully.
- Audit logs are append-only; the JSONL format is line-oriented and
  machine-parseable.