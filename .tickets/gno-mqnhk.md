---
id: gno-mqnhk
status: closed
deps: [gno-gicyu]
links: []
created: 2026-04-24T16:58:38Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Colorize topics and write/rm output

Three small changes in one ticket since they are trivial and all just apply existing termcolor helpers.

1. internal/commands/topics.go: bold the count column, cyan the topic name.

2. internal/commands/write.go (line ~131): the newly-created ID printed after a successful write should be colorized via ColorizeID. Load all IDs (including the just-written one) before printing.

3. internal/commands/rm.go (line ~82): each removed ID printed to stdout should be colorized. Since the entries are already gone by the time we print, capture the full IDs list *before* deletion and pass it to the colorizer. The warning on stderr stays plain.

## Acceptance Criteria

All three commands produce colored output on a TTY, plain output when piped. Tests pass.

