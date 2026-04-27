---
id: gno-lrigx
status: open
deps: []
links: []
created: 2026-04-27T01:16:26Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Improve inter-entry navigation in show

Enhance the 'gn show <id>' output so related entries are more useful. Inline the first sentence (or first 120 characters) of each related entry beneath the 'related:' line, indented. Add a 'gn show <id> --follow' flag that, when present, prints the full contents of every directly related entry after the main entry, separated by horizontal rules. This turns isolated entries into a traversable knowledge graph without requiring multiple CLI invocations.

## Acceptance Criteria

'gn show <id>' on an entry with related entries prints a preview snippet for each related entry. 'gn show <id> --follow' prints the full related entries inline.

