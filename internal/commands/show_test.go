package commands_test

import (
	"strings"
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/commands"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// baseTime is a fixed reference time used across show tests so that date
// formatting is deterministic.
var baseTime = time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)

// --- Show by ID ---

func TestShow_by_id_prints_entry(t *testing.T) {
	store := newTestStore(t)
	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"keymaster-token-auth", "session-management"},
		Text:      "We decided on JWTs for backward compat.",
		Related:   []string{"ghjkmn"},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Show(store, "abcdef", &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	result := output.String()

	if !strings.Contains(result, "abcdef") {
		t.Errorf("output missing ID: %q", result)
	}
	if !strings.Contains(result, "[keymaster-token-auth, session-management]") {
		t.Errorf("output missing topics: %q", result)
	}
	if !strings.Contains(result, "created 2026-04-24") {
		t.Errorf("output missing created date: %q", result)
	}
	if !strings.Contains(result, "updated 2026-04-24") {
		t.Errorf("output missing updated date: %q", result)
	}
	if !strings.Contains(result, "Related: ghjkmn") {
		t.Errorf("output missing related IDs: %q", result)
	}
	if !strings.Contains(result, "We decided on JWTs for backward compat.") {
		t.Errorf("output missing text body: %q", result)
	}
}

func TestShow_by_id_not_found_returns_error(t *testing.T) {
	store := newTestStore(t)

	var output strings.Builder
	err := commands.Show(store, "abcdef", &output)
	if err == nil {
		t.Fatal("expected error for missing ID, got nil")
	}
	if !strings.Contains(err.Error(), "abcdef") {
		t.Errorf("error message should mention the missing ID, got: %v", err)
	}
}

func TestShow_by_id_no_related_omits_related_line(t *testing.T) {
	store := newTestStore(t)
	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"go-lang"},
		Text:      "Some text.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Show(store, "abcdef", &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	if strings.Contains(output.String(), "Related:") {
		t.Errorf("output should not contain Related line when there are no related entries: %q", output.String())
	}
}

// --- Show by topic ---

func TestShow_by_topic_finds_entries(t *testing.T) {
	store := newTestStore(t)

	entry1 := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"keymaster-token-auth", "session-management"},
		Text:      "First entry.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	entry2 := storage.Entry{
		ID:        "ghjkmn",
		Topics:    []string{"keymaster-token-auth"},
		Text:      "Second entry.",
		Related:   []string{},
		CreatedAt: baseTime.Add(time.Hour),
		UpdatedAt: baseTime.Add(time.Hour),
	}
	appendEntries(t, store, entry1, entry2)

	var output strings.Builder
	if err := commands.Show(store, "KeymasterTokenAuth", &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	result := output.String()

	// The header shows the normalized form regardless of how the argument was typed.
	if !strings.Contains(result, "keymaster-token-auth") {
		t.Errorf("output missing normalized topic: %q", result)
	}
	if !strings.Contains(result, "2 entries") {
		t.Errorf("output missing entry count: %q", result)
	}
	if !strings.Contains(result, "First entry.") {
		t.Errorf("output missing first entry text: %q", result)
	}
	if !strings.Contains(result, "Second entry.") {
		t.Errorf("output missing second entry text: %q", result)
	}
}

func TestShow_by_topic_case_insensitive(t *testing.T) {
	store := newTestStore(t)

	// The entry is stored with the normalized form, as it would be after a write.
	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"keymaster-token-auth"},
		Text:      "Some text.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	// All of these normalize to "keymaster-token-auth" and should find the entry.
	for _, target := range []string{"KeymasterTokenAuth", "keymasterTokenAuth", "keymaster_token_auth", "keymaster-token-auth"} {
		var output strings.Builder
		if err := commands.Show(store, target, &output); err != nil {
			t.Errorf("Show(%q): unexpected error: %v", target, err)
			continue
		}
		if !strings.Contains(output.String(), "Some text.") {
			t.Errorf("Show(%q): output missing entry text: %q", target, output.String())
		}
	}
}

func TestShow_by_topic_sorted_chronologically(t *testing.T) {
	store := newTestStore(t)

	// Append in reverse chronological order to confirm sorting is applied.
	entry2 := storage.Entry{
		ID:        "ghjkmn",
		Topics:    []string{"go-lang"},
		Text:      "Second chronologically.",
		Related:   []string{},
		CreatedAt: baseTime.Add(time.Hour),
		UpdatedAt: baseTime.Add(time.Hour),
	}
	entry1 := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"go-lang"},
		Text:      "First chronologically.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry2, entry1)

	var output strings.Builder
	if err := commands.Show(store, "GoLang", &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	result := output.String()
	firstPosition := strings.Index(result, "First chronologically.")
	secondPosition := strings.Index(result, "Second chronologically.")

	if firstPosition == -1 || secondPosition == -1 {
		t.Fatalf("output missing expected text: %q", result)
	}
	if firstPosition > secondPosition {
		t.Errorf("entries not sorted chronologically: first appears after second in output")
	}
}

func TestShow_by_topic_not_found_returns_error(t *testing.T) {
	store := newTestStore(t)

	var output strings.Builder
	err := commands.Show(store, "NonExistentTopic", &output)
	if err == nil {
		t.Fatal("expected error for missing topic, got nil")
	}
	if !strings.Contains(err.Error(), "NonExistentTopic") {
		t.Errorf("error message should mention the missing topic, got: %v", err)
	}
}

// --- ID pattern detection ---

func TestShow_id_pattern_takes_priority_over_topic(t *testing.T) {
	store := newTestStore(t)

	// Create an entry whose ID happens to also be a valid topic name.
	// This is contrived but exercises the "prefer ID" rule. "abcdef" is already
	// normalized (all lowercase, no separators needed).
	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"abcdef"},
		Text:      "Entry found by ID.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	// "abcdef" matches the ID pattern, so it should be treated as an ID lookup.
	if err := commands.Show(store, "abcdef", &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	result := output.String()
	// ID-mode output starts with the ID on the first line, not a "Topic:" header.
	if strings.HasPrefix(result, "Topic:") {
		t.Errorf("expected ID-mode output, got topic-mode output: %q", result)
	}
	if !strings.Contains(result, "abcdef") {
		t.Errorf("output missing ID: %q", result)
	}
}

// --- ID-pattern fallback to topic ---

// "update" is 6 letters from the allowed alphabet, so it matches the ID
// pattern. When no entry with that ID exists, Show must fall back to topic
// lookup and return the matching entries.
func TestShow_id_pattern_falls_back_to_topic_on_miss(t *testing.T) {
	store := newTestStore(t)

	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"update"},
		Text:      "Update strategy entry.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Show(store, "update", &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Update strategy entry.") {
		t.Errorf("output missing expected text: %q", result)
	}
}

// When an entry exists with the exact ID, it must be returned even though the
// same string could also be a topic name.
func TestShow_id_pattern_prefers_id_when_entry_exists(t *testing.T) {
	store := newTestStore(t)

	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"some-topic"},
		Text:      "Entry found by ID.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Show(store, "abcdef", &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	result := output.String()
	// ID-mode output starts with the ID, not a "Topic:" header.
	if strings.HasPrefix(result, "Topic:") {
		t.Errorf("expected ID-mode output, got topic-mode output: %q", result)
	}
	if !strings.Contains(result, "Entry found by ID.") {
		t.Errorf("output missing expected text: %q", result)
	}
}

// A target that matches neither an ID nor a topic must return an error.
func TestShow_not_found_anywhere_returns_error(t *testing.T) {
	store := newTestStore(t)

	var output strings.Builder
	err := commands.Show(store, "nonsense-that-isnt-in-the-kb", &output)
	if err == nil {
		t.Fatal("expected error for unknown target, got nil")
	}
}

// Strings containing excluded letters (i, l, o) must not be treated as IDs
// even if they are 6 characters long.
func TestShow_excluded_letters_not_treated_as_id(t *testing.T) {
	store := newTestStore(t)

	// "golife" is already normalized (lowercase, no separators needed).
	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"golife"},
		Text:      "Topic with excluded letters.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	// "golife" contains 'l', 'i' — excluded from the ID alphabet — so it must
	// be treated as a topic, not an ID.
	if err := commands.Show(store, "golife", &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	if !strings.Contains(output.String(), "Topic with excluded letters.") {
		t.Errorf("output missing expected text: %q", output.String())
	}
}
