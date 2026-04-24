# AGENTS.md

Orientation notes for agents working in this repo.

## What this is

`gnosis` (binary: `gn`) is a small Go CLI that stores a per-repo knowledge
base of "why" entries — decisions, rejected alternatives, constraints.
Entries are JSONL on disk with a SQLite FTS5 search index.

## Layout

- `cmd/gn/main.go` — thin CLI dispatcher.
- `internal/commands/` — one file per subcommand (`write`, `search`,
  `show`, `topics`, `edit`, `rm`, `reindex`) plus `resolve.go` for ID
  prefix resolution. Each command takes a `*storage.Store` and an
  `io.Writer` so tests can capture output.
- `internal/storage/` — JSONL store, entry type, ID generation, topic
  normalization, repo-root discovery.
- `internal/index/` — FTS5 index wrapper.
- `internal/doctrine/` — embedded help text (`gn help`, `gn help review`).

## Conventions

- IDs: 6-char lowercase, alphabet excludes ambiguous letters (see
  `idAlphabetSet` in `resolve.go`).
- Topics: stored normalized (lowercase, dashes). `NormalizeTopic` is
  the single source of truth.
- Dispatch rule in `show`: target ≤ 6 chars is an ID prefix, otherwise
  a topic name.
- Output currently plain text; no color library in use.
- Tests live next to their code; `go test ./...` runs them.

## Commands

- Build: `make build` → `./gn`.
- Test: `make test` (or `go test ./...`).
- Install: `make install`.

## Agent workflow

Follow the doctrine in `gn help`:
1. Before implementing, run `gn search <keywords>` and surface any
   conflicting recorded decisions.
2. Record decisions as they happen with `gn write`.
3. After finishing, run `gn help review`.
