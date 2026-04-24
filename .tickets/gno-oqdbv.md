---
id: gno-oqdbv
status: closed
deps: [gno-ebehg]
links: []
created: 2026-04-24T23:36:33Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Detect terminal width in main and pass to Show

In cmd/gn/main.go, before calling commands.Show, determine the wrap width and pass it as the new int parameter.

Rules:
- If os.Stdout is a TTY, get the terminal width and use min(width, 80).
- If not a TTY, pass 0 (no wrapping).
- If width detection fails for any reason, pass 0.

Use github.com/mattn/go-isatty (already a direct dependency) for TTY detection and golang.org/x/term for GetSize. Add golang.org/x/term to go.mod via 'go get golang.org/x/term'; run 'go mod tidy'.

Only the show command needs the width; do not change other commands.

## Acceptance Criteria

go build succeeds; go.mod/go.sum updated; gn show wraps to terminal width (cap 80) on a TTY and emits unwrapped output when piped.

