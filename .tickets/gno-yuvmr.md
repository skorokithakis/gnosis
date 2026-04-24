---
id: gno-yuvmr
status: closed
deps: []
links: []
created: 2026-04-24T16:51:23Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Switch release workflow to goreleaser

Replace the hand-rolled matrix build in `.github/workflows/release-please.yml`
with goreleaser, so releases ship archived artifacts + checksums instead of
bare binaries.

## Scope

- Add `.goreleaser.yaml` at repo root.
- Builds: `./cmd/gn` for linux/darwin × amd64/arm64. Output binary name `gn`.
- Archives: goreleaser defaults (tar.gz, name
  `{ProjectName}_{Version}_{Os}_{Arch}`). Include `README.md`, `LICENSE`,
  `CHANGELOG.md` in each archive.
- Produce `checksums.txt`.
- Disable goreleaser's own changelog generation and release-note body —
  release-please already creates the GitHub release and owns the notes.
  Goreleaser should append artifacts to the existing release
  (`release.mode: append`).
- In the workflow, replace the `build` matrix job with a single job that runs
  `goreleaser/goreleaser-action@v6` with `args: release --clean`, gated on
  `release_created == 'true'`. Pass the release-please tag to goreleaser via
  `GORELEASER_CURRENT_TAG` env (workflow runs on `main`, not on the tag ref).

## Non-goals

- No homebrew tap, scoop, deb/rpm, docker, SBOM, or signing.
- No windows build (keep current OS/arch set).
- No source archive customization (defaults fine).

## Caveats

- `release.mode: append` is required; the default would fail because
  release-please already created the release.
- Goreleaser requires a clean git state and a tag; the checkout step must use
  `fetch-depth: 0` and the job must set `GORELEASER_CURRENT_TAG` to
  `needs.release-please.outputs.tag_name`.
