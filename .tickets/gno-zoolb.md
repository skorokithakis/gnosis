---
id: gno-zoolb
status: closed
deps: []
links: []
created: 2026-04-24T23:00:01Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Add Homebrew tap to goreleaser

Add a `brews` section to `.goreleaser.yaml` so each release automatically publishes a formula to `github.com/skorokithakis/homebrew-tap`.

Configuration:
- Formula name: `gnosis`
- Binary name: `gn` (already the case)
- Tap repo: `skorokithakis/homebrew-tap` (already created, currently empty)
- Description and homepage should match the project (see README.md / go.mod)
- License: project is MIT (verify against repo state — note the README no longer mentions a license)

Goreleaser needs a GitHub token with write access to the tap repo. The token is typically passed via env var (e.g. `HOMEBREW_TAP_GITHUB_TOKEN`) and configured in the GitHub Actions release workflow as a secret. Document the required secret name in the goreleaser config or a brief comment so the maintainer knows what to set in repo secrets.

Non-goals:
- Don't submit to homebrew-core.
- Don't add Scoop, AUR, nixpkgs, or other package managers.
- Don't change the release-please / goreleaser flow beyond adding the brews block.

Verify the goreleaser config is valid (`goreleaser check` or equivalent) but don't cut a release.

