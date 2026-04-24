---
id: gno-hbxbr
status: closed
deps: []
links: []
created: 2026-04-24T10:49:46Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: minor cleanups from review

Batch of small fixes from code review:

1. internal/storage/storage.go: bufio.Scanner has a default 64KB token limit. A single entry body over that size would break every read. Either raise the buffer limit (scanner.Buffer with a larger size) or switch to bufio.Reader+ReadString('\n'). I lean the scanner.Buffer approach with, say, 10MB limit — simple and adequate for human-written knowledge entries.

2. internal/commands/edit.go: $EDITOR with arguments (e.g. 'code --wait', 'vim -n') currently fails because exec.Command(editor, path) treats the whole string as the program name. Fix: shell-split via strings.Fields or a small splitter; pass first element as program, rest + path as args. Don't invoke via /bin/sh — too fragile.

3. internal/commands/write.go: Write prints to stdout via fmt.Println. Thread a writer parameter like the other commands do. Also go.mod has github.com/mattn/go-isatty marked // indirect but it's imported directly in write.go — run go mod tidy to fix.

4. Delete empty internal/commands/commands.go (1-line placeholder, unused).

5. internal/commands/edit.go: replace slicesEqual helper with slices.Equal from stdlib (Go 1.25 has it).

6. internal/commands/search.go: replace the manual queryArgs join with strings.Join.

## Acceptance Criteria

Entries with multi-megabyte bodies round-trip through the tool. $EDITOR='code --wait' works. go mod tidy is clean. No dead code.

