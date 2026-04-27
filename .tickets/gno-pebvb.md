---
id: gno-pebvb
status: open
deps: []
links: []
created: 2026-04-27T01:16:29Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Add gn serve for web browsing

Add a 'gn serve' command that starts a minimal HTTP server rendering the knowledge base as HTML. The server should list all topics on the root path, show entries by topic at '/topic/<name>', and show individual entries at '/entry/<id>'. Related entry links should be clickable. Use only the Go standard library (net/http, html/template) to avoid new dependencies. Default to a sensible port (e.g. 7331) with a --port flag. Read-only access; no editing through the web interface.

## Acceptance Criteria

A user can run 'gn serve', open the printed URL in a browser, click through topics, and read entries with clickable links to related entries. No external dependencies beyond the Go standard library.

