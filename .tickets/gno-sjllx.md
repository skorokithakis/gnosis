---
id: gno-sjllx
status: closed
deps: [gno-qrwtj]
links: []
created: 2026-04-24T08:39:30Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: SQLite FTS5 index

Implement the search index. Source of truth is JSONL; index is disposable and rebuildable.

Location: ~/.cache/gnosis/<repo-hash>/index.db. Repo hash: sha256 of the absolute path to the resolved repo root, hex, first 16 chars. Create parent dirs as needed. XDG_CACHE_HOME should be respected if set.

Schema: an FTS5 virtual table indexing entry text and topics (display form is fine, normalized form also indexed so searches like 'keymaster token' match 'KeymasterTokenAuth'). Store the entry ID as a non-indexed column so search results can look up full entries from JSONL.

Build strategy: on any command that needs the index, check a staleness marker (JSONL mtime vs a stored mtime in the index). If stale or missing, rebuild from JSONL. Rebuild is full, not incremental — simpler, and JSONLs will be small.

API:
- EnsureFresh() error  — rebuild if stale
- Search(query string, limit int) ([]SearchHit, error)  — returns entry IDs + snippets, ranked by FTS5's bm25
- Rebuild() error

SearchHit: EntryID, Snippet, Rank.

Use modernc.org/sqlite (pure Go, FTS5 included). Prefix search enabled on the FTS5 table so partial matches work.

## Acceptance Criteria

Writing entries then searching returns them. Modifying JSONL externally then searching triggers rebuild. Search across topics works: 'token' matches an entry tagged KeymasterTokenAuth.

