---
id: gno-grwpg
status: open
deps: []
links: []
created: 2026-04-27T01:17:34Z
type: feature
priority: 2
assignee: Stavros Korokithakis
---
# JSON output flag for search/show/topics/list

Add --json flag to search, show, topics, and list. Emits a stable, documented JSON schema instead of human-formatted text. No ANSI codes in JSON mode.

Scope: each command checks for --json early and branches; shared schema types live in a small json package or inline structs.

Open decisions to settle when starting:
- exact schema per command (entry shape: include rank/snippet for search?)
- whether --json implies no color regardless of TTY (yes, probably)
- topics output: array of {topic, count}

## Acceptance Criteria

Each affected command supports --json and produces valid parseable JSON. Schema is documented (in code comments or help text). Existing text output is unchanged when flag is absent.

