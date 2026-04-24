---
id: gno-sytgy
status: closed
deps: []
links: []
created: 2026-04-24T10:49:16Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: topic normalization edge cases + empty-topic rejection

Bugs flagged by code review:

1. Empty-after-normalization topics are silently stored. NormalizeTopic('---'), NormalizeTopic('-'), NormalizeTopic('___') all return ''. Current write.go rejects empty-trimmed input but accepts input that normalizes to empty. Result: Entry.Topics contains ''.
   Fix: after normalization, validate that each resulting topic is non-empty. Reject the write if any would-be topic is empty post-normalization.
   Apply the same fix to edit's validation (internal/commands/edit.go's validateEditedEntry).

2. Edit doesn't normalize parsed topics before the change-detection comparison. So editing 'go-lang' to 'GoLang' produces a false-positive 'change' (they normalize to the same thing). Also, edit doesn't dedupe topics by normalized form, so 'Foo,foo' saves two identical topics.
   Fix: in ParseEditBuffer or before comparison in Edit, normalize + dedupe parsed topics the same way write.go does.

## Acceptance Criteria

Writing with topic '---' errors with a useful message. Editing an entry and changing only the casing of a topic is detected as 'no changes'. Editing to include 'Foo,foo' saves one topic, not two.

