---
id: gno-icdbl
status: closed
deps: []
links: []
created: 2026-04-24T08:39:05Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: project skeleton

Initialize Go module github.com/skorokithakis/gnosis in this repo. Lay out package structure (cmd/gnosis for the binary, internal packages for storage/index/commands as they grow). Add go.mod with modernc.org/sqlite as the only non-stdlib dependency. Add a minimal main.go with a subcommand dispatcher (stdlib flag-based, no cobra) that prints available commands when called with no args or 'help'. Add a Makefile or justfile with build/install targets. Add .gitignore for Go artifacts.

Scope: just the skeleton. No real commands implemented yet — main.go should dispatch to stub handlers that print 'not implemented'. This task exists so all downstream tasks have a stable foundation.

Non-goals: any real functionality, any doctrine text, any storage code.

