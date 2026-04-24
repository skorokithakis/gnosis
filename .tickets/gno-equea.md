---
id: gno-equea
status: closed
deps: [gno-gicyu]
links: []
created: 2026-04-24T16:58:33Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Colorize show output

Update internal/commands/show.go.

In printEntry:
- Load all IDs once (pass them in, or have showByID/showByTopic pass them through). Simplest: change printEntry signature to accept allIDs []string.
- Colorize entry.ID via ColorizeID.
- Colorize each topic (inside the [Topic1, Topic2] brackets) in cyan. Brackets and commas stay plain.
- Colorize the createdDate and updatedDate in yellow. Keep 'created' / 'updated' labels plain (or dim — pick dim for consistency with 'Related:').
- For the Related line: 'Related:' label dim, each related ID colorized via ColorizeID.
- Body text stays plain.

In showByTopic header 'Topic: <name>  (<n> entries)': color <name> cyan, count bold.

Update showByID and showByTopic to pass the full IDs list down.

## Acceptance Criteria

gn show <id> and gn show <topic> render with the documented color scheme on a TTY. Existing tests pass unchanged.

