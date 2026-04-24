---
id: gno-ebehg
status: closed
deps: [gno-kishl]
links: []
created: 2026-04-24T23:36:24Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Restructure gn show entry header and wrap body

Change printEntry in internal/commands/show.go to emit a multi-line label/value header followed by a blank line and the wrapped body.

New per-entry format:
  id: <colored unique-prefix id>
  topics: <cyan topic>, <cyan topic>
  related: <colored id>, <colored id>      (only when entry.Related is non-empty)
  created: <yellow date>
  updated: <yellow date>
  
  <body, wrapped via the textwrap helper>

Labels (id:/topics:/related:/created:/updated:) use termcolor.Dim, matching the existing dim styling for 'created'/'updated'/'Related:'. All-lowercase labels with trailing colon.

Also lowercase the topic-mode banner from 'Topic:' to 'topic:' in showByTopic. Keep its layout otherwise unchanged ('topic: <name>  (N entries)').

Show takes a new wrapWidth int parameter; printEntry forwards it to the wrap helper. wrapWidth <= 0 means no wrapping.

Update internal/commands/show_test.go to match the new format. Existing assertions to fix include the bracketed topics list ('[a, b]' becomes 'topics: a, b'), the 'Related:' substring (now 'related:'), and the 'Topic:' prefix check in dispatch tests (now 'topic:'). Update Show call sites in tests to pass 0 for wrapWidth. Add at least one test asserting body wrapping kicks in when wrapWidth>0 and is a no-op when wrapWidth==0.

Also update internal/commands/resolve_test.go Show call sites for the new signature.

## Acceptance Criteria

go test ./... passes. gn show output matches the new layout.

