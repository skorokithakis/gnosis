---
id: gno-houzr
status: open
deps: []
links: []
created: 2026-04-27T01:17:07Z
type: feature
priority: 2
assignee: Stavros Korokithakis
---
# gn search --topic <topic>

Restrict search to entries carrying a given topic. Implementation note: FTS5 column filter on the topics column, combined with the user's query (e.g. 'topics:keymaster-token-auth AND <query>').

Scope: extend internal/commands/search.go to accept --topic flag and prepend an FTS5 column qualifier. Topic value is normalized via storage.NormalizeTopic before being injected.

Caveats: must not break sanitizeQuery's existing logic for hyphenated bare queries — the user-supplied query is still sanitized; only the topic-filter clause we inject uses operator syntax.

## Acceptance Criteria

gn search --topic foo bar returns only entries that have topic 'foo' AND whose text/topics match 'bar'. Topic name is normalized (FooTopic == foo-topic). Unknown topic returns zero results without error.

