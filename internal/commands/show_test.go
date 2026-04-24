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
	// The topic name is longer than 6 characters so it is routed to topic
	// lookup rather than ID-prefix lookup.
	entry2 := storage.Entry{
		ID:        "ghjkmn",
		Topics:    []string{"go-language"},
		Text:      "Second chronologically.",
		Related:   []string{},
		CreatedAt: baseTime.Add(time.Hour),
		UpdatedAt: baseTime.Add(time.Hour),
	}
	entry1 := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"go-language"},
		Text:      "First chronologically.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry2, entry1)

	var output strings.Builder
	if err := commands.Show(store, "GoLanguage", &output); err != nil {
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

// --- Dispatch by query length ---

// A query of 6 characters or fewer is always routed to ID-prefix lookup, even
// when the same string could also be a valid topic name.
func TestShow_short_query_routes_to_id_lookup(t *testing.T) {
	store := newTestStore(t)

	// "abcdef" is both a valid entry ID and a valid topic name. Because the
	// query is ≤6 characters, it must be resolved as an ID prefix.
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

// --- No ID-to-topic fallback ---

// A 6-character string that matches no entry ID must return an error. There is
// no fallback to topic lookup for short queries; the split is purely by length.
func TestShow_short_query_no_fallback_to_topic(t *testing.T) {
	store := newTestStore(t)

	// "update" is 6 letters from the allowed alphabet. Even though an entry
	// with that topic exists, Show must not fall back to topic lookup.
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
	err := commands.Show(store, "update", &output)
	if err == nil {
		t.Fatal("expected error for 6-char query with no matching ID, got nil")
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

// Strings containing excluded letters (i, l, o) that are longer than 6
// characters are routed to topic lookup. Short strings (≤6 chars) always go
// to ID-prefix lookup regardless of their character content.
func TestShow_excluded_letters_routed_to_topic_when_long(t *testing.T) {
	store := newTestStore(t)

	// "golifetime" is longer than 6 characters and contains excluded letters,
	// so it is routed to topic lookup.
	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"golifetime"},
		Text:      "Topic with excluded letters.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Show(store, "golifetime", &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	if !strings.Contains(output.String(), "Topic with excluded letters.") {
		t.Errorf("output missing expected text: %q", output.String())
	}
}
