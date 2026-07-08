# AISentinel — Distribution Submission Guide

This file lists the remaining distribution channels that need a human action to complete.
Everything else (code, GitHub repo, releases pipeline, Docker pipeline, MCP Registry publish pipeline, bizdnai.com landing block) has been set up automatically.

## ✅ Already done automatically

| Channel | Status | Where |
|---|---|---|
| GitHub repo | ✅ live | https://github.com/Kabzhanov/AISentinel |
| GitHub Releases (cross-platform binaries) | ✅ pipeline ready, will trigger on tag push | `.github/workflows/ci.yml` |
| GitHub Container Registry (GHCR) docker images | ✅ pipeline ready | `.github/workflows/docker.yml` |
| MCP Registry publish pipeline | ✅ ready, needs `MCP_PUBLISHER_GITHUB_TOKEN` secret | `.github/workflows/publish-mcp-registry.yml` |
| bizdnai.com landing block | ✅ live | https://bizdnai.com/index/ (block "AISentinel MCP Server") |
| server.json (MCP Registry format) | ✅ updated to 2025-12-11 schema | `server.json` |

## ❌ Requires Рашид's manual action

### 1. punkpeye/awesome-mcp-servers PR
**File with the ready-to-paste line and step-by-step:** `docs/AWESOME_MCP_PR.md`

**Recommended:** `gh repo fork punkpeye/awesome-mcp-servers --clone`, edit README, push branch, `gh pr create --title "...🤖🤖🤖"` (agent fast-track marker in title).

### 2. MCP Registry — set secret and finish publish
**File with instructions:** `docs/SUBMISSION_GUIDE.md` (this file)

The publish-mcp-registry.yml workflow is already wired. To complete:
1. Go to https://github.com/Kabzhanov/AISentinel/settings/secrets/actions
2. Add a new secret: `MCP_PUBLISHER_GITHUB_TOKEN` = a PAT owned by the same GitHub account (Kabzhanov) with `modelcontextprotocol:publish` scope. Generate at https://github.com/settings/personal-access-tokens.
3. The next push of a v* tag will publish automatically. (Or trigger manually via Actions → run the publish-mcp-registry workflow on the latest release.)

### 3. punkpeye/awesome-mcp-servers PR — optional but high-traffic
See `docs/AWESOME_MCP_PR.md`.

### 4. mcp.so self-submit
Web form: https://mcp.so/server/submit

Required fields:
- Name: AISentinel
- GitHub: https://github.com/Kabzhanov/AISentinel
- Description: Open-source security, control, and observability layer for AI agents. Pre-tool policy enforcement + append-only audit logs + YAML policy engine. Sidecar mode wraps any stdio MCP server with one command. Apache 2.0. By BizDNAi.
- Category: Security
- Tags: security, audit, policy, agent-runtime

### 5. glama.ai/mcp/servers self-submit
Web form: https://glama.ai/mcp/servers/submit

Use the same fields as above. Optional but glama.ai is a popular MCP directory.

### 6. Homebrew tap
**File with formula + steps:** `docs/HOMEBREW_TAP.md`

Create new public repo `kabzhanov/homebrew-tap`, add `Formula/aisentinel.rb`, push. macOS users install with `brew install kabzhanov/tap/aisentinel`.

### 7. Docker Hub
If you want a Docker Hub mirror in addition to GHCR:
```bash
docker login                 # to docker.io/kabzhanov
docker buildx build --platform linux/amd64,linux/arm64 \
  -t kabzhanov/aisentinel:v1.0.3 -t kabzhanov/aisentinel:latest --push .
docker buildx build --platform linux/amd64,linux/arm64 \
  --target sidecar \
  -t kabzhanov/aisentinel-sidecar:v1.0.3 -t kabzhanov/aisentinel-sidecar:latest --push .
```
(Not strictly needed because GHCR images are public at `ghcr.io/kabzhanov/aisentinel`.)

## Timeline summary
- **v1.0.0 (intial MVP)**: code only, no releases
- **v1.0.1**: sidecar + new cross-platform release pipeline
- **v1.0.2**: server.json → MCP Registry schema + publish workflow + awesome-mcp-servers PR text
- **v1.0.3 (latest)**: + Dockerfile + docker-compose + Docker GHCR pipeline
- **Once secret is set, the next v1.0.4 tag will publish to MCP Registry automatically**.
