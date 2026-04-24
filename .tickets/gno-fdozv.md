---
id: gno-fdozv
status: closed
deps: [gno-drkdd]
links: []
created: 2026-04-24T09:21:32Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: set up release-please

Add release-please configuration for automated releases. Use the Go release type. Conventional commits. release-please should create release PRs that bump the version and update CHANGELOG.md. Configuration file: release-please-config.json + .release-please-manifest.json in the repo root. GitHub Action workflow at .github/workflows/release-please.yml that triggers on push to main.

