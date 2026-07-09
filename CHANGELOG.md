# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [1.0.7] - 2026-07-09

### Fixed
- MCP `serverInfo`/system events now report the real binary version (was
  hardcoded `1.0.0`), and `resources/read` of `policies://built-in/default`
  serves the embedded policy instead of a stale inline mirror.
- Version resolution falls back to `runtime/debug.ReadBuildInfo()`, so
  `go install ...@v1.0.7` binaries report their module version instead of
  `dev` (ldflags still win for release builds).
- `go install` + running either binary outside a clone of this repo no longer
  fails with `os.Exit(1)` when no `--policy`/`$AISENTINEL_POLICY` is given:
  the default policy is now embedded into the binary (`policies` package,
  `go:embed default.yaml`) and used as the last resort in the resolution
  order `--policy` > `$AISENTINEL_POLICY` > `./policies/default.yaml` (if
  present) > built-in default. An explicit `--policy`/`$AISENTINEL_POLICY`
  that fails to load is still a hard error, as before.
- Audit logs (`internal/server`'s `aisentinel_log_event` and the sidecar's
  `intercept`) no longer write raw secret values to the JSONL trail.
  `Logger.Write` now redacts known secret-value shapes (AWS access keys,
  OpenAI/GitHub/Slack tokens, Bearer tokens, URL-embedded credentials, PEM
  private keys) from `tool_args`, `tool_result`, `metadata`, and `prompt`
  before writing.
- Fixed a stdout write race in `aisentinel-sidecar`: the server→client copy
  goroutine and the blocked-call JSON-RPC error response were two
  unsynchronized writers on the same `os.Stdout`, which could interleave
  partial JSON-RPC lines. Both now write through a shared
  `internal/iox.LockedLineWriter`, guaranteeing line-atomic writes.
- Removed an unused `sync.Mutex` field on `server.Server` (the stdio server
  loop is single-threaded; confirmed via grep that the field was never
  locked).

### Changed
- `version` in both `cmd/aisentinel` and `cmd/aisentinel-sidecar` is now set
  via `-ldflags "-X main.version=..."` at build time instead of a hardcoded
  Go constant. Builds without that flag (e.g. `go install ...@latest`,
  local `go build`) now correctly report `dev`. CI's release job derives
  the version from the pushed git tag.
- `gofmt -w .` applied across the module; CI now has a `gofmt -l .` gate so
  formatting drift is caught automatically.

### Added
- `internal/secrets`: shared secret-detection patterns used by both the
  audit-log redaction pass and (available for reuse by) the policy engine's
  `secret-in-args` rule, so the two don't drift apart.
- `internal/iox.LockedLineWriter`: reusable mutex-protected line writer.
- Tests: secret redaction (AWS key, OpenAI `sk-`, URL credentials, GitHub/
  Slack tokens, PEM keys), embedded default-policy load/validity,
  `policy.Resolve` resolution order (explicit-missing → error,
  nothing-set-and-no-file → embedded default, on-disk preferred over
  embedded), and a concurrent `-race` test for `LockedLineWriter`.

## [1.0.6] - 2026-07-09
- Add `policies/planet-prod.yaml` (Planet canon policy: own_party allow,
  B24/1C writes require human approval, LAN/docker-bridge block, Morozov
  partner_invoice explicit allow); bump version in both binaries.
- Bump `server.json` (MCPB URLs + SHA256 for 1.0.6); add
  `bundles/aisentinel{,-sidecar}/manifest.json`; CI packs MCPB bundles on
  tag push.
- CI fixes: install `mcpb` via npm (no GitHub Release binaries upstream);
  checkout source before packing/flattening so `bundles/`/`dist/` survive
  `actions/checkout@v4`'s default `git clean -ffdx`.

## [1.0.5] - 2026-07-08
- Block destructive commands by default (`rm -rf`, `dd of=/dev/...`,
  `mkfs`, fork bomb, `DROP`/`TRUNCATE`); sync the built-in default-policy
  MCP resource with those rules.
- Dockerfile: MCP server image is now the default build target (was the
  sidecar); `CMD serve`.
- README: add live policy-enforcement demo GIF.
- CI: trigger the release build on `v*` tag push (was branch-only, so tags
  never actually built); build GHCR images on tag push + manual dispatch,
  always tag `latest`.
- Docs: distribution & adoption checklist, ready-to-paste launch posts,
  400x400 logo for marketplace submissions.

## [1.0.4] - 2026-07-08
- Bump to v1.0.4 with the real `mcpb` bundle SHA256 for MCP Registry
  publishing.
- Add distribution submission guides for remaining manual channels.

## [1.0.3] - 2026-07-08
- Add `Dockerfile`, `docker-compose.yml`, and a Docker workflow publishing
  to GHCR.

## [1.0.2] - 2026-07-08
- Update `server.json` to the MCP Registry `2025-12-11` schema; add the
  registry publish workflow.

## [1.0.1] - 2026-07-08
- Add `aisentinel-sidecar`: transparent policy proxy for any stdio MCP
  server.
- Add the `mcp-test` policy (allow MCP reads, block 1C writes by default).
- Expand the release workflow to a cross-platform build matrix; document
  sidecar installation.

## [1.0.0] - 2026-07-08
- Initial MVP: MCP stdio server, YAML policy engine, JSONL audit log.
