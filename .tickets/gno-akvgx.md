---
id: gno-akvgx
status: closed
deps: []
links: []
created: 2026-05-04T03:45:52Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Add non-interactive text edit to gn edit

Currently `gn edit <id>` always launches $EDITOR. Add a non-interactive path that replaces only the text body, keeping topics and related untouched.

Behaviour:
- `gn edit <id>`                     → editor (unchanged)
- `gn edit <id> "new text"`           → replace text body, save, no editor
- `echo "new text" | gn edit <id>`    → same as above via stdin (only when stdin is not a TTY)
- If positional text is provided, ignore stdin (match `gn write` precedence)
- Trim text; empty after trim is an error (same message style as `write`)
- If new text equals existing text, print 'no changes' and exit 0 (match editor path)
- Use Store.Update for the atomic write, same as the editor path, so concurrent appends are preserved
- UpdatedAt set to time.Now().UTC()

Non-goals:
- No `--text`, `--topics`, `--related` flags
- No partial-update of topics/related non-interactively (use editor for that)

Update the usage string in cmd/gn (and any help text in internal/doctrine if edit is mentioned) to show the new forms. Add tests next to edit.go covering: positional arg, stdin, no-change case, empty text error, positional takes precedence over stdin.

## Acceptance Criteria

gn edit <id> "text" updates the entry without launching an editor; piped stdin works the same way; existing editor behavior is preserved when no text arg and stdin is a TTY.

