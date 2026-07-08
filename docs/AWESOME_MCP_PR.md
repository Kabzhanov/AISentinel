# PR to punkpeye/awesome-mcp-servers

## Goal
Add AISentinel to the awesome-mcp-servers list so it becomes discoverable to the broader MCP community.

## Repo
https://github.com/punkpeye/awesome-mcp-servers

## Suggested insertion point
The README is NOT strictly alphabetical — it groups by section type (Frameworks, Servers, Community, etc.). Place AISentinel at the **start of the Servers / Frameworks section** (around line 137 in the current README, just before the `- [Aganium/agenium]` entry or right after, in the local-stdio-servers cluster).

## Line to add

```markdown
- [Kabzhanov/AISentinel](https://github.com/Kabzhanov/AISentinel) 📇 🏠 🍎 🪟 🐧 - Open-source security, control, and observability layer for AI agents. Pre-tool policy enforcement (allow / block / require_human_approval), append-only JSONL audit logs, YAML-based policy engine. Wraps any stdio MCP server via drop-in sidecar proxy. Apache 2.0. By BizDNAi — creators of AI Trust Index.
```

## How to submit (3 ways)

### Way 1 — `gh` CLI (recommended, agent-friendly)
The awesome-mcp-servers project explicitly supports agent PRs: add `🤖🤖🤖` to the end of the PR title for fast-track merging.

```bash
gh auth login                            # one-time, OAuth flow
gh repo fork punkpeye/awesome-mcp-servers --clone --remote
cd awesome-mcp-servers
# edit README.md — paste the line above near the start of the Servers section
git checkout -b add-aisentinel
git add README.md
git commit -m "Add AISentinel — security layer for AI agents"
git push -u origin add-aisentinel
gh pr create --title "Add AISentinel — security layer for AI agents 🤖🤖🤖" --body "Adds AISentinel, an Apache 2.0 MCP server that wraps any stdio MCP server with policy enforcement and audit logging. Built by Kabzhanov / BizDNAi."
```

### Way 2 — Web UI (no gh)
1. Open https://github.com/punkpeye/awesome-mcp-servers
2. Click "Fork"
3. In the fork, edit `README.md` and paste the line above
4. Click "Propose changes" → "Create pull request"
5. Title: `Add AISentinel — security layer for AI agents 🤖🤖🤖`

### Way 3 — patch file
If you want to review the diff first:

```bash
curl -sL https://raw.githubusercontent.com/punkpeye/awesome-mcp-servers/main/README.md > original.md
# Insert the line after line 137 (before "- [Aganium/agenium]")
sed -i '137a - [Kabzhanov/AISentinel](https://github.com/Kabzhanov/AISentinel) 📇 🏠 🍎 🪟 🐧 - Open-source security, control, and observability layer for AI agents. Pre-tool policy enforcement (allow / block / require_human_approval), append-only JSONL audit logs, YAML-based policy engine. Wraps any stdio MCP server via drop-in sidecar proxy. Apache 2.0. By BizDNAi — creators of AI Trust Index.' original.md
diff original.md <(curl -sL https://raw.githubusercontent.com/punkpeye/awesome-mcp-servers/main/README.md) | head -5
```

## Badge legend used
- 📇 = local stdio MCP server
- 🏠 = runs locally
- 🍎 🪟 🐧 = macOS / Windows / Linux
- (no ☁️ — we don't have a hosted cloud endpoint yet)