---
id: gno-ifqay
status: closed
deps: []
links: []
created: 2026-07-10T21:44:28Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Add Go test CI workflow

Add an independent GitHub Actions workflow for pushes and pull requests. Use actions/checkout@v4, actions/setup-go@v5 with go-version-file: go.mod, and run go test ./.... Do not alter release automation or add linting.

## Acceptance Criteria

Pushes and pull requests run the repository Go test suite using the version declared in go.mod.


## Notes

**2026-07-10T21:45:14Z**

ready for implementation
