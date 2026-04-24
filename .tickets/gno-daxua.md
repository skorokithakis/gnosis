---
id: gno-daxua
status: closed
deps: [gno-qrwtj]
links: []
created: 2026-04-24T08:40:13Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: edit command

Implement 'gnosis edit <id>'. Opens the entry in $EDITOR for modification.

Behavior:
- Load entry by ID (error if missing).
- Write a temp file with the entry in an editable format: a small header (topics, related, as comma-separated lists the user can modify) and the text body below. Use a clear separator between header and body.
- Launch $EDITOR (fallback to 'vi') on the temp file. Wait for it to exit.
- Parse the result. If unchanged, do nothing and report. If changed, validate (topics non-empty, text non-empty, related IDs valid), then rewrite the JSONL with the updated entry replacing the old one.
- Rewrites use write-to-temp-then-rename under the .gnosis/.lock file lock (from the storage layer).

Header format (proposal, confirm in implementation):

  # Topics: KeymasterTokenAuth, SessionManagement
  # Related: defuvw
  # ---
  <text body>

Lines starting with '#' before the '---' separator are the header. After '---' is free-form body.

Non-goals: editing ID or CreatedAt (immutable). Editing multiple entries at once.

## Acceptance Criteria

Editing an entry and saving updates the JSONL. Cancelling (no changes) is a no-op. Invalid edits (empty body, bad related ID) print an error and leave the original entry untouched.

