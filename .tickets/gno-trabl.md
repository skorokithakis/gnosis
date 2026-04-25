---
id: gno-trabl
status: closed
deps: []
links: []
created: 2026-04-25T00:59:19Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Move flock file from repo into cache dir

Move the flock file currently at <repo>/.gnosis/.lock into the per-repo cache directory (alongside index.db) as <cache-dir>/lock. The lock file is purely runtime coordination state; it has no business living inside the repo.

Scope:
- New package internal/paths exposing CacheDir(repoRoot string) (string, error). Move the existing cacheDir helper out of internal/index/index.go into this package and have index use it. Behaviour must be identical: respect XDG_CACHE_HOME, fall back to ~/.cache, key by sha256(repoRoot)[:16] under gnosis/.
- internal/storage.Store gains a cacheDir field. lockPath() returns filepath.Join(cacheDir, "lock") (no leading dot).
- NewStore computes cacheDir from the resolved repo root via paths.CacheDir.
- NewStoreAt signature becomes NewStoreAt(gnosisDir, cacheDir string). Update all callers (storage_test, commands/search_test, commands/reindex_test, commands/helpers_test, index/index_test). Tests should pass a t.TempDir() for cacheDir (or reuse the one they already create for XDG_CACHE_HOME).
- withSharedLock and withExclusiveLock must no longer call ensureDir on the .gnosis directory. They should ensure the cache dir exists instead (MkdirAll on cacheDir). The .gnosis directory should only be created when something is actually being written to entries.jsonl (Append / Rewrite paths).
- Stale-lock cleanup: opportunistically remove a stale <repo>/.gnosis/.lock if it exists, best-effort, errors ignored. Time-gate this: skip the cleanup entirely once time.Now() is past 2026-07-01 UTC. Implement as a package-level const cutoff and a single guard at the call site, with a comment along the lines of "// TODO: remove this stale-lock cleanup after 2026-07-01; the lock file moved out of the repo on 2026-04-25 and nobody but the author was running gn before then." so the whole block is trivial to delete later.

Non-goals:
- Do not add a migration command, version marker, or warning. The time-gated cleanup above is sufficient.
- Do not change the index.db location or naming.
- Do not rename or restructure anything else in internal/index beyond extracting cacheDir.
- Do not change public CLI behaviour or output.

Caveats:
- Stale-lock cleanup must not run while still holding the flock on the new lock file (the old path is unrelated, but keep the unlink outside any locked critical section to be safe and obvious).
- Update the doc comments in storage.go that currently reference ".gnosis/.lock".

## Acceptance Criteria

go test ./... passes. Fresh run of 'gn search foo' in a repo with no .gnosis directory does not create one. 'gn write ...' still creates .gnosis/entries.jsonl. The cache dir contains both index.db and lock.

