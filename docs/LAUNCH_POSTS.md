# AISentinel — готовые тексты для запуска (copy-paste)

Логотип/демо-ассеты: https://raw.githubusercontent.com/Kabzhanov/AISentinel/main/assets/aisentinel-logo-400.png

---

## 1. Show HN
Форма: https://news.ycombinator.com/submit
Лучшее время: вт–чт, 08:00–10:00 ET. Автор весь день в комментариях.

**Title:**
`Show HN: AISentinel – a policy/audit sidecar for MCP AI agents (Go, Apache-2.0)`

**URL:** `https://github.com/Kabzhanov/AISentinel`

**Text (первый коммент от автора):**
```
AI agents now get shell, browser, filesystem and API keys through MCP tools.
Antivirus can't see them — they're legitimate processes. One indirect prompt
injection in a PDF or email can pivot into data exfiltration via ordinary tools
like Bash or Email_send.

AISentinel is a small Go sidecar that sits between the agent and its MCP tools.
Every tool call is checked against a YAML policy BEFORE it runs:
allow / block / require_human_approval. Everything is written to an append-only
JSONL audit log. It wraps ANY stdio MCP server with one command — no code change:

  aisentinel-sidecar --policy strict.yaml ./your-mcp-server

Example policy rule (block outbound curl to raw IPs):
  - id: no-raw-ip-exfil
    match: { tool_name: Bash, tool_args_regex: "10\\.|192\\.168\\.|172\\.(1[6-9]|2\\d|3[01])\\." }
    action: block

It's Apache-2.0, single static binary, also ships as .mcpb, Docker (GHCR) and a
Homebrew tap. Built by our team (BizDNAi, behind the AI Trust Index).

Honest limitations: policy is regex/tool-name matching, not semantic; it's a
guardrail + audit layer, not a sandbox. Feedback on the policy model welcome.
```

---

## 2. r/mcp
Форма: https://www.reddit.com/r/mcp/submit (Text post)

**Title:** `AISentinel: an open-source policy + audit sidecar that wraps any stdio MCP server`

**Body:**
```
I kept wanting a way to see and control what my MCP agents actually DO at the
tool-call level, so we built AISentinel (Apache-2.0, Go).

It's a drop-in sidecar: `aisentinel-sidecar --policy strict.yaml ./your-mcp-server`.
Every tool call is evaluated against a YAML policy before execution —
allow / block / require_human_approval — and logged to an append-only JSONL trail.

Use cases:
- Block destructive calls (rm -rf, DROP TABLE, curl to raw IPs) from a
  prompt-injected agent.
- require_human_approval on sensitive tools (payments, email send).
- Tamper-evident audit log for "what did the agent run, and with whom".

Repo: https://github.com/Kabzhanov/AISentinel
It's already in the official MCP Registry (io.github.Kabzhanov/aisentinel).
Would love feedback on the policy schema — what rules would you want built in?
```

---

## 3. r/LocalLLaMA
Форма: https://www.reddit.com/r/LocalLLaMA/submit

**Title:** `Runtime guardrails for local agents: a policy/audit sidecar for MCP tools (Apache-2.0)`

**Body:**
```
If you run agents on self-hosted models with real tools (shell, files, HTTP),
you've probably worried about a bad tool call — from a jailbreak or an indirect
prompt injection in scraped content.

AISentinel is a Go sidecar that enforces a YAML policy on every MCP tool call
before it executes (allow/block/require_human_approval) and writes an append-only
audit log. Wraps any stdio MCP server, no code change, single static binary,
runs fully local — no cloud dependency.

  aisentinel-sidecar --policy strict.yaml ./your-mcp-server

Repo: https://github.com/Kabzhanov/AISentinel (Apache-2.0)
Demo idea it's built for: agent tries `rm -rf` under injection → blocked, logged.
Curious what policies people would want for local agent setups.
```

---

## 4. Сценарий 2-мин видео-демо (Show HN / YouTube / X GIF)
Показывает главный хук: prompt-injection → заблокированный destructive-вызов + audit-log.

- **0:00–0:15** — экран: обычный MCP-агент с shell-tool. «Агенты получили реальную власть через инструменты.»
- **0:15–0:35** — запускаем БЕЗ AISentinel: агент под инъекцией выполняет `rm -rf` / `curl attacker.com` → срабатывает. «Антивирус этого не видит — это легитимный процесс.»
- **0:35–1:00** — оборачиваем одной командой: `aisentinel-sidecar --policy strict.yaml ./mcp-server`. Показываем policy-правило (block по regex).
- **1:00–1:30** — повторяем ту же атаку → **BLOCKED**, tool-call не выполнился. Показываем строку в JSONL audit-log.
- **1:30–1:50** — `require_human_approval` на «оплате»: агент ждёт подтверждения.
- **1:50–2:00** — «Apache-2.0, один бинарь, оборачивает любой stdio MCP-сервер. github.com/Kabzhanov/AISentinel». Логотип BizDNAi + AI Trust Index.

Команды для записи — из README (Option B/E, policies/strict.yaml).
