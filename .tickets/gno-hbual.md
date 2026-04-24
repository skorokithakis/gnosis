---
id: gno-hbual
status: closed
deps: [gno-fdozv]
links: []
created: 2026-04-24T09:21:36Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: GitHub Actions workflow for release binaries

Add a GitHub Actions workflow that builds Go binaries on release creation (triggered by release-please tags). Build matrix: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64. Use goreleaser. Upload binaries as release assets. The workflow should trigger on tag push (v*) to align with release-please's tag creation.

Depends on release-please being set up first.
