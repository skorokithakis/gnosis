---
id: gno-ohxwg
status: closed
deps: []
links: []
created: 2026-04-24T16:22:39Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Consolidate release workflows so binaries build on release

Releases created by release-please (using the default GITHUB_TOKEN) do not trigger the separate release-binaries workflow, so v0.2.0 shipped with no binaries attached.

Fix by merging release-binaries.yml into release-please.yml as a second job that runs after release-please, gated on its release_created output.

Scope:
- Delete .github/workflows/release-binaries.yml.
- In .github/workflows/release-please.yml, expose release-please outputs (release_created, tag_name) from the existing job and add a new 'build' job with:
  - needs: release-please
  - if: needs.release-please.outputs.release_created
  - the existing 4-entry matrix (linux/darwin x amd64/arm64)
  - checkout, setup-go (go-version-file: go.mod), go build of ./cmd/gnosis, gh release upload using needs.release-please.outputs.tag_name

Non-goals:
- No PAT, no goreleaser, no extra archs, no checksums/signing, no Windows.
- Do not change release-please-config.json or manifest.

Caveats:
- The release-please step needs an id so its outputs are addressable.
- Job-level outputs must be declared on release-please job to be visible to the build job.

## Acceptance Criteria

Merging a release-please PR produces a GitHub release with four uploaded binaries named gnosis-<goos>-<goarch>, and release-binaries.yml no longer exists.

