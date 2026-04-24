package commands_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/commands"
	"github.com/skorokithakis/gnosis/internal/storage"
)

func entryWithTopics(id string, topics []string) storage.Entry {
	now := time.Now().UTC().Truncate(time.Second)
	return storage.Entry{
		ID:        id,
		Topics:    topics,
		Text:      "text",
		Related:   []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// TestAggregateTopics_counts_normalized_topics verifies that topics stored in
// normalized form are counted correctly. Since normalization now happens on
// write, entries in the store already hold the normalized form.
func TestAggregateTopics_counts_normalized_topics(t *testing.T) {
	entries := []storage.Entry{
		entryWithTopics("aaaaaa", []string{"keymaster-token-auth"}),
		entryWithTopics("bbbbbb", []string{"keymaster-token-auth"}),
		entryWithTopics("cccccc", []string{"keymaster-token-auth"}),
	}

	aggregates := commands.AggregateTopics(entries)

	if len(aggregates) != 1 {
		t.Fatalf("expected 1 aggregate, got %d: %+v", len(aggregates), aggregates)
	}
	if aggregates[0].Count != 3 {
		t.Errorf("expected count 3, got %d", aggregates[0].Count)
	}
	if aggregates[0].Topic != "keymaster-token-auth" {
		t.Errorf("expected topic %q, got %q", "keymaster-token-auth", aggregates[0].Topic)
	}
}

// TestAggregateTopics_multiple_topics_per_entry verifies that all topics on a
// single entry are counted independently.
func TestAggregateTopics_multiple_topics_per_entry(t *testing.T) {
	entries := []storage.Entry{
		entryWithTopics("aaaaaa", []string{"billing", "session-management"}),
		entryWithTopics("bbbbbb", []string{"billing"}),
	}

	aggregates := commands.AggregateTopics(entries)

	if len(aggregates) != 2 {
		t.Fatalf("expected 2 aggregates, got %d: %+v", len(aggregates), aggregates)
	}

	counts := map[string]int{}
	for _, aggregate := range aggregates {
		counts[aggregate.Topic] = aggregate.Count
	}

	if counts["billing"] != 2 {
		t.Errorf("billing: expected count 2, got %d", counts["billing"])
	}
	if counts["session-management"] != 1 {
		t.Errorf("session-management: expected count 1, got %d", counts["session-management"])
	}
}

// TestTopics_sort_order verifies that Topics prints entries sorted by count
// descending, with ties broken alphabetically by topic name.
func TestTopics_sort_order(t *testing.T) {
	store := newTestStore(t)

	entries := []storage.Entry{
		entryWithTopics("aaaaaa", []string{"billing"}),
		entryWithTopics("bbbbbb", []string{"keymaster-token-auth", "session-management"}),
		entryWithTopics("cccccc", []string{"keymaster-token-auth", "session-management"}),
		entryWithTopics("dddddd", []string{"keymaster-token-auth"}),
	}
	for _, entry := range entries {
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	var buffer bytes.Buffer
	if err := commands.Topics(store, &buffer); err != nil {
		t.Fatalf("Topics: %v", err)
	}

	lines := strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), lines)
	}

	// keymaster-token-auth appears 3 times, session-management 2 times, billing 1 time.
	if !strings.Contains(lines[0], "keymaster-token-auth") {
		t.Errorf("line 0 should contain keymaster-token-auth, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "session-management") {
		t.Errorf("line 1 should contain session-management, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "billing") {
		t.Errorf("line 2 should contain billing, got %q", lines[2])
	}
}

// TestTopics_tie_broken_alphabetically verifies that when two topics have the
// same count, they are sorted alphabetically by topic name.
func TestTopics_tie_broken_alphabetically(t *testing.T) {
	store := newTestStore(t)

	entries := []storage.Entry{
		entryWithTopics("aaaaaa", []string{"zebra", "apple"}),
	}
	for _, entry := range entries {
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	var buffer bytes.Buffer
	if err := commands.Topics(store, &buffer); err != nil {
		t.Fatalf("Topics: %v", err)
	}

	lines := strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), lines)
	}

	if !strings.Contains(lines[0], "apple") {
		t.Errorf("line 0 should contain apple (alphabetically first), got %q", lines[0])
	}
	if !strings.Contains(lines[1], "zebra") {
		t.Errorf("line 1 should contain zebra, got %q", lines[1])
	}
}

// TestTopics_empty_store verifies that Topics writes nothing when there are no
// entries.
func TestTopics_empty_store(t *testing.T) {
	store := newTestStore(t)

	var buffer bytes.Buffer
	if err := commands.Topics(store, &buffer); err != nil {
		t.Fatalf("Topics: %v", err)
	}

	if buffer.Len() != 0 {
		t.Errorf("expected no output for empty store, got %q", buffer.String())
	}
}
