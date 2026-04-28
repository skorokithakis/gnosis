# AGENTS.md

Orientation notes for agents working in this repo.

## What this is

`gnosis` (binary: `gn`) is a small Go CLI that stores a per-repo knowledge
base of "why" entries — decisions, rejected alternatives, constraints.
Entries are JSONL on disk with a SQLite FTS5 search index.

## Layout

- `cmd/gn/main.go` — thin CLI dispatcher; one `case` per subcommand.
- `internal/commands/` — one file per subcommand, plus `resolve.go`
  for ID-prefix resolution. Each command takes a `*storage.Store` and
  an `io.Writer` so tests can capture output.
- `internal/storage/` — JSONL store, entry type, ID generation, topic
  normalization, repo-root discovery.
- `internal/index/` — FTS5 index wrapper.
- `internal/doctrine/` — embedded help text (`gn help`, `gn help plan`,
  `gn help review`).
- `internal/paths/`, `internal/termcolor/`, `internal/textwrap/` —
  small support packages; purpose matches the name.

## Conventions

- IDs: 6-char lowercase, alphabet excludes ambiguous letters (see
  `idAlphabetSet` in `resolve.go`).
- Topics: stored normalized (lowercase, dashes). `NormalizeTopic` is
  the single source of truth.
- Dispatch rule in `show`: target ≤ 6 chars is an ID prefix, otherwise
  a topic name.
- Color output via `internal/termcolor`, gated on TTY detection in
  `cmd/gn/main.go`. Do not add color writes outside `termcolor`; do
  route new colored output through it.
- Tests live next to their code; `go test ./...` runs them.

## Commands

- Build: `make build` → `./gn`.
- Test: `make test` (or `go test ./...`).
- Install: `make install`.

## Agent workflow

Follow the doctrine in `gn help plan`:
1. Before implementing, run `gn search <keywords>` and surface any
   conflicting recorded decisions.
2. Record decisions as they happen with `gn write`.
3. After finishing, run `gn help review`.
