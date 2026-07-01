---
id: gno-cyvqa
status: closed
deps: []
links: []
created: 2026-07-01T21:22:17Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Cleanup pass: dead code, duplication, convention fixes

Ready for implementation.

Objective: hygiene pass across the codebase, net deletion expected. Seven items:

1. Remove the expired stale-lock migration shim in internal/storage/storage.go: delete staleLockCutoff, removeStaleRepoLock, and its call in NewStore. The TODO says to remove after 2026-07-01, which has passed.
2. Delete the unused notImplemented function in cmd/gn/main.go.
3. In cmd/gn/main.go, hoist the repeated storage.NewStore() + error-print boilerplate so it happens once after the help case is dispatched, instead of once per subcommand case. Keep behavior identical (help must not require a store).
4. Deduplicate topic validation: make Write in internal/commands/write.go use normalizeAndDeduplicateTopics from edit.go (move/rename the helper if a better home is warranted). Preserve Write's current stricter behavior of erroring on an empty topic segment (e.g. trailing comma) rather than silently dropping it — reconcile the helper accordingly and keep edit's parse behavior working.
5. Remove the duplicated idAlphabetSet in internal/commands/resolve.go: export the alphabet from internal/storage (same pattern as storage.IDLength) and use it.
6. Make Edit consistent with the io.Writer convention: the 'no changes' message currently goes through fmt.Println directly; route it through an injected writer like every other command. Update cmd/gn/main.go call site.
7. Fix the inaccurate comment on Store.Append: PIPE_BUF atomicity applies to pipes/FIFOs, not regular files. Reword to reflect the real protection (flock + single-user CLI usage). Also note in its doc comment that production code uses AppendNew and Append exists for callers that already have a unique ID (currently tests).

Non-goals: no behavior changes beyond item 6's plumbing, no output format changes, no refactor of search/latest row formatting, no flag-parsing rework.

Constraints: go test ./... must pass; keep the existing comment style (why-focused, full sentences).

