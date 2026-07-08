# AISentinel — Distribution & Adoption Checklist

Status snapshot (2026-07-08). Автоматизируемое — сделано. Ниже — что осталось,
с готовым copy-paste контентом. Логотип 400×400 уже в репо:
**https://raw.githubusercontent.com/Kabzhanov/AISentinel/main/assets/aisentinel-logo-400.png**

## ✅ Уже живо (действий не требует)
- Официальный **MCP Registry** — `io.github.Kabzhanov/aisentinel` v1.0.4 (OIDC-publish).
- **glama.ai** — проиндексирован (slug `AISentinel`).
- **awesome-mcp-servers** — PR #9630 (OPEN, MERGEABLE).
- **Homebrew tap** — `brew install Kabzhanov/tap/aisentinel` (formula v1.0.4).
- GitHub Releases (.mcpb) + GHCR docker + bizdnai.com/index блок.

---

## ⏳ Каталоги — нужен браузер (claim, не submit)
Большинство САМИ подтягивают из офиц. реестра — задача «застолбить владение», 15 мин суммарно:

| Каталог | Действие | URL |
|---|---|---|
| **PulseMCP** | проверить листинг → заполнить форму если нет | https://www.pulsemcp.com/submit |
| **glama.ai** | войти через GitHub → Claim (апгрейд тира) | https://glama.ai (найти AISentinel) |
| **Smithery** | войти GitHub OAuth → claim/publish; в репо нужен `smithery.yaml` | https://smithery.ai |
| **mcp.so** | web-форма (auto-import из реестра вероятен) | https://mcp.so/submit |
| **mcpservers.org** | web-форма | https://mcpservers.org/submit |

Пропустить: **mcp.run** (требует репаковки в WASM-servlet), **Composio** (курируется, нет self-submit).

---

## ⏳ Клиентские маркетплейсы (появиться ВНУТРИ инструмента — самое ценное)

### Cline Marketplace — GitHub issue (готово к отправке)
Открыть: https://github.com/cline/mcp-marketplace/issues/new?template=mcp-server-submission.yml
Поля:
- **GitHub Repository URL:** `https://github.com/Kabzhanov/AISentinel`
- **Logo (400×400 PNG):** `https://raw.githubusercontent.com/Kabzhanov/AISentinel/main/assets/aisentinel-logo-400.png`
- **Additional info:** `AISentinel is a security/policy/audit sidecar for MCP agents. Wraps any stdio MCP server (aisentinel-sidecar <cmd>) or runs standalone (aisentinel serve). Pre-tool policy enforcement (allow/block/require_human_approval) + append-only JSONL audit. Apache-2.0. By BizDNAi (AI Trust Index). Install: prebuilt .mcpb binaries, go install, Docker (ghcr.io/kabzhanov/aisentinel), or brew install Kabzhanov/tap/aisentinel.`
- ⚠️ **Два обязательных чекбокса** — аттестация, что установку через Cline протестировали (только README/llms-install.md) и что сервер stable. Поставить может только человек, реально прогнавший Cline. **← нужен ты.**

### Cursor — форма + install-badge
- Форма: https://cursor.directory/plugins/new
- Config для юзера (в README уже есть): `{"command":"aisentinel","args":["serve","--policy","<path>/policies/default.yaml"]}`

### VS Code / Copilot — кормится из GitHub MCP Registry (уже покрыто офиц. реестром)
### Zed — покрыт офиц. реестром, действий нет.
### Continue.dev — опц., публикация «block» в https://continue.dev/hub
### Claude Desktop — отдельного сабмита нет; .mcpb-бандл в Releases = one-click install.

---

## ⏳ Docker MCP Catalog (git PR — под ревью Docker, нужна валидация)
Репо: https://github.com/docker/mcp-registry — форк, добавить `servers/aisentinel/server.yaml`:
```yaml
name: AISentinel
image: mcp/aisentinel        # Docker собирает под своим namespace после ревью
type: server
meta:
  category: security
  tags: [security, policy, audit, agent-runtime, guardrails]
about:
  title: AISentinel
  description: Security, control & observability layer for AI agents — pre-tool policy enforcement + audit.
  icon: https://raw.githubusercontent.com/Kabzhanov/AISentinel/main/assets/aisentinel-logo-400.png
source:
  project: https://github.com/Kabzhanov/AISentinel
  branch: main
  directory: cmd/aisentinel
```
⚠️ Требует прохождения их `task validate` + сборки образа их пайплайном — сначала проверить локально, потом PR.

---

## 📣 Adoption — каналы awareness (листинг ≠ пользователи; нужен нарратив+демо)

| Канал | Как | Усилие | Отдача |
|---|---|---|---|
| **Show HN** | "Show HN: AISentinel – policy/audit sidecar for MCP agents", вт–чт 8–10 ET, автор в комментах весь день | средн | высокая (спайк) |
| **r/mcp** | пост «проблема→решение», без маркетинга | низк | высокая |
| **r/LocalLLaMA** | угол «безопасность агентов для self-hosted», демо-ссылка | низк | высокая |
| **X/Twitter** | реплаи в MCP-треды, #MCP #AIagents #AIsecurity, GIF-демо | низк | средн |
| **Discord** (MCP official / Anthropic / Cursor) | #showcase/#servers | низк | средн |
| **YouTube** | 2-мин скринкаст: заблокированный `rm -rf` + audit-log | средн | высокая (переиспользуется) |
| **dev.to / Medium** | «Add a security layer to any MCP server in 5 min» | средн | средн (SEO) |
| **LinkedIn (BizDNAi)** | governance/enterprise, связка с AI Trust Index | низк | средн (B2B) |
| **Product Hunt** | нужны ассеты+хантер+день запуска | высок | средн |

### Хуки (security-нарратив)
1. **«Firewall для AI-агентов»** — агент под prompt-injection пытается `rm -rf`/drop БД → заблокировано. Виральное демо.
2. **Audit-trail для комплаенса** — «что агент запускал и с кем говорил», EU AI Act / закон РК об ИИ №230-VIII.
3. **Связка с AI Trust Index** — слой-исполнитель, дающий измеримые trust-сигналы (G-контроли).
4. **«Оборачивает любой stdio MCP-сервер без единой строки кода»** — лид каждого поста.
