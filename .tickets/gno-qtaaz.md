---
id: gno-qtaaz
status: closed
deps: []
links: []
created: 2026-04-24T10:49:08Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: fix storage locking for concurrent safety

Both reviewers flagged this as critical. Current behavior silently loses appends during rewrites.

Problem: Append() uses O_APPEND only — never touches .gnosis/.lock. Rewrite() takes exclusive flock during final temp-rename, but the caller (edit/rm) reads entries without the lock. Between ReadAll and Rewrite's rename, concurrent appends go to the old file and get discarded by the rename.

Fix:
1. Append() takes shared flock on .gnosis/.lock for the duration of the append.
2. Add Store.Update(fn func([]Entry) []Entry) error that:
   - Takes exclusive flock.
   - Reads entries under the lock.
   - Calls fn.
   - Rewrites atomically while still holding the lock.
3. Migrate edit and rm to use Update instead of ReadAll+Rewrite.
4. For Rebuild() in the index, take the shared flock around ReadAll + mtime stat, so a concurrent append doesn't leave the index with stale mtime but missing entries. OR re-stat mtime before and after ReadAll and retry if they differ. Either works.
5. Also move ID collision check inside Append/Update — currently write.go does ReadAll+GenerateID+Append with no lock, so two concurrent writers can generate the same ID.

Verify with a concurrent test: N goroutines each appending while another goroutine does rewrites. All appended entries must be present in the final state.

## Acceptance Criteria

The race test from reviewer #2 (five concurrent appends during rewrite, currently 0-1 survive) now shows all 5 surviving. ID collision is atomic — the lock covers ID generation too.

