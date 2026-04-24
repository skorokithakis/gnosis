---
id: gno-qrwtj
status: closed
deps: [gno-icdbl]
links: []
created: 2026-04-24T08:39:19Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: storage layer (JSONL + entry model)

Implement the entry model and JSONL-backed storage.

Entry struct fields: ID (6 lowercase letters, no confusables — exclude i/l/o), Topics ([]string, display form preserved), Text (string), Related ([]string, entry IDs), CreatedAt (time.Time). JSON-serialized one per line.

Topic normalization: a helper that takes a display-form topic and returns the storage key. Rules: split CamelCase on case boundaries, replace underscores/spaces with dashes, lowercase, collapse repeated dashes. Example: 'KeymasterTokenAuth', 'keymaster_token_auth', 'keymaster-token-auth' all → 'keymaster-token-auth'. The storage key is what topic lookups compare; the display form (first-written casing) is what we show.

ID generation: 6 chars from alphabet excluding confusables. On write, check existing IDs and retry on collision.

File location: walk up from CWD looking for .git, .jj, or existing .gnosis/. If found, use that repo's .gnosis/entries.jsonl. If none, treat CWD as the root and create .gnosis/ on first write (implicit init).

Concurrency: appends use O_APPEND writes (line-atomic on Linux). Rewrites (for edit/rm) use write-to-temp-then-rename. Add a file lock (flock on .gnosis/.lock) around rewrites to prevent races with concurrent appends.

API surface (internal, used by command packages):
- Append(entry) error
- ReadAll() ([]Entry, error)
- Rewrite(entries []Entry) error  — atomic replace
- GenerateID(existing map[string]bool) string
- NormalizeTopic(display string) string

Non-goals: search, index, any CLI wiring. Pure library code plus unit tests for ID generation, topic normalization, and round-trip JSONL.

## Acceptance Criteria

Round-trip test passes: write N entries, read them back, contents match. Normalization test covers CamelCase, snake_case, kebab-case, mixed. Concurrent-append test (goroutines) produces N well-formed lines.


## Notes

**2026-04-24T09:21:14Z**

Design update: Entry struct must include UpdatedAt (time.Time) alongside CreatedAt. UpdatedAt is set equal to CreatedAt on initial write; updated by the edit command when an entry is modified.
