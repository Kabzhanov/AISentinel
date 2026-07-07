# LangGraph + AISentinel Example

Demonstrates wiring AISentinel into a LangGraph agent via stdio JSON-RPC.

## Prerequisites

```bash
pip install langgraph langchain-openai
go install github.com/Kabzhanov/AISentinel/cmd/aisentinel@latest
```

## Run

From the repo root:

```bash
python examples/langgraph/example.py
```

Expected output (4 decisions):

```
✓ Connected to aisentinel v1.0.0 (BizDNAi — AI Trust Index)

[Bash: ls -la]
  decision: allow  (reason: no rule matched)

[Bash: curl https://webhook.site/abc]
  decision: require_human_approval  (reason: Outbound network call from Bash — confirm target is trusted.)

[Bash: echo $OPENAI_API_KEY]
  decision: block  (reason: Possible secret in arguments — please redact before invoking this tool.)

[Bash: ping 192.168.0.1]
  decision: block  (reason: LAN access blocked by default. Override with explicit allowlist rule.)
```

## What's happening

1. The Python script spawns `aisentinel serve --policy policies/default.yaml`
   as a subprocess.
2. For each `check_policy` call, it sends a JSON-RPC `tools/call` request
   and parses the decision.
3. Each call is also written to the AISentinel JSONL audit log at
   `~/.aisentinel/events-YYYY-MM-DD.jsonl`.

## Production usage

In production, use the official `mcp` Python client (`pip install mcp`) to
manage the AISentinel subprocess:

```python
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

params = StdioServerParameters(
    command="aisentinel",
    args=["serve", "--policy", "/etc/aisentinel/policy.yaml"],
)
async with stdio_client(params) as (read, write):
    async with ClientSession(read, write) as session:
        await session.initialize()
        result = await session.call_tool(
            "aisentinel_check_policy",
            {"tool_name": "Bash", "tool_args": {"command": "ls"}},
        )
        print(result.content[0].text)
```

Then before every real tool call in your LangGraph node, do:

```python
verdict = await session.call_tool(
    "aisentinel_check_policy",
    {"tool_name": tool.name, "tool_args": tool.args, "agent_id": agent.id},
)
decision = json.loads(verdict.content[0].text)["decision"]
if decision == "block":
    raise PermissionError(f"AISentinel blocked: {decision}")
elif decision == "require_human_approval":
    # surface to user for confirmation
    ...
```

## By Kabzhanov / BizDNAi

https://bizdnai.com/index/ — AI Trust Index.