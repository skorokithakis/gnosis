---
id: gno-ssivu
status: closed
deps: [gno-nsgxu, gno-newbq]
links: []
created: 2026-04-24T10:34:27Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: normalize topics on storage, not just lookup

Current behavior: each entry stores its topics in the display form as originally written. 'topics' command uses first-write-wins for display. Result: same topic shows with different casings depending on which entry you're viewing.

Desired behavior: normalize topics on write. Store only the normalized form (keymaster-token-auth). That's what shows up everywhere — topics, show, search, edit buffer.

Changes needed:
- internal/storage/storage.go: on Append, normalize topics before storing. Entry.Topics always holds normalized forms.
- internal/commands/write.go: stop trying to preserve first-write-wins display — just pass topics through storage, which normalizes.
- internal/commands/topics.go: simplify — no more TopicAggregate with Display field; just count by topic (already normalized).
- internal/commands/show.go: print topics as-is (they're already normalized).
- internal/commands/edit.go: the edit buffer shows normalized forms; users can edit but their input gets re-normalized on save.
- internal/index/index.go: Rebuild can stop indexing 'display form + normalized form' separately — just index the normalized form and the text.
- Update tests across the board.
- Consider: what happens to existing entries in a .gnosis/entries.jsonl that have non-normalized display forms? A migration is overkill for a brand-new tool with no users. Don't handle it — on the next write/edit/reindex their topics get normalized implicitly if we also fix the read path to normalize on load. Actually simpler: add normalization to ReadAll so legacy entries self-heal on next read.

## Acceptance Criteria

After writing 'KeymasterTokenAuth' then 'keymaster_token_auth', both entries store topics as 'keymaster-token-auth'. 'gnosis topics' shows 'keymaster-token-auth 2'. 'gnosis show' on either entry shows '[keymaster-token-auth]'. Lookup still matches case-insensitively — 'gnosis show KeymasterTokenAuth' still works.

