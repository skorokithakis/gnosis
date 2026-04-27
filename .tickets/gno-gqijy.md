---
id: gno-gqijy
status: open
deps: []
links: []
created: 2026-04-27T01:17:26Z
type: feature
priority: 2
assignee: Stavros Korokithakis
---
# gn export --format markdown

Add 'gn export [--format markdown]' that emits the entire knowledge base as a single human-readable Markdown document on stdout. Useful for generating docs, sharing, or archiving.

Scope: new internal/commands/export.go. Markdown is the only initial format; the --format flag exists for future formats (json-lines just for re-import, html, etc.) — but only markdown is implemented here.

Proposed structure:
- Group by topic (entries appear under each of their topics)
- Within a topic, sorted chronologically
- Each entry shows: heading with id, created/updated dates, related links, body

Open decision: should an entry with N topics appear N times (once per topic) or just under its primary topic with cross-references?

## Acceptance Criteria

gn export --format markdown writes valid Markdown to stdout. Unknown format value errors out. Empty store produces a header-only document or empty output (decision when implementing).

