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
	if err := commands.Show(store, "abcdef", 0, &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	result := output.String()

	if !strings.Contains(result, "abcdef") {
		t.Errorf("output missing ID: %q", result)
	}
	if !strings.Contains(result, "topics:") {
		t.Errorf("output missing topics label: %q", result)
	}
	if !strings.Contains(result, "keymaster-token-auth") {
		t.Errorf("output missing topic keymaster-token-auth: %q", result)
	}
	if !strings.Contains(result, "session-management") {
		t.Errorf("output missing topic session-management: %q", result)
	}
	if !strings.Contains(result, "created:") {
		t.Errorf("output missing created label: %q", result)
	}
	if !strings.Contains(result, "2026-04-24") {
		t.Errorf("output missing date: %q", result)
	}
	if !strings.Contains(result, "updated:") {
		t.Errorf("output missing updated label: %q", result)
	}
	if !strings.Contains(result, "related:") {
		t.Errorf("output missing related label: %q", result)
	}
	if !strings.Contains(result, "ghjkmn") {
		t.Errorf("output missing related ID: %q", result)
	}
	if !strings.Contains(result, "We decided on JWTs for backward compat.") {
		t.Errorf("output missing text body: %q", result)
	}
}

func TestShow_by_id_not_found_returns_error(t *testing.T) {
	store := newTestStore(t)

	var output strings.Builder
	err := commands.Show(store, "abcdef", 0, &output)
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
	if err := commands.Show(store, "abcdef", 0, &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	if strings.Contains(output.String(), "related:") {
		t.Errorf("output should not contain related line when there are no related entries: %q", output.String())
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
	if err := commands.Show(store, "KeymasterTokenAuth", 0, &output); err != nil {
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
		if err := commands.Show(store, target, 0, &output); err != nil {
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
	if err := commands.Show(store, "GoLanguage", 0, &output); err != nil {
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
	err := commands.Show(store, "NonExistentTopic", 0, &output)
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
	if err := commands.Show(store, "abcdef", 0, &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	result := output.String()
	// ID-mode output starts with the id: label, not a "topic:" banner.
	if strings.HasPrefix(result, "topic:") {
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
	err := commands.Show(store, "update", 0, &output)
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
	if err := commands.Show(store, "abcdef", 0, &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	result := output.String()
	// ID-mode output starts with the id: label, not a "topic:" banner.
	if strings.HasPrefix(result, "topic:") {
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
	err := commands.Show(store, "nonsense-that-isnt-in-the-kb", 0, &output)
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
	if err := commands.Show(store, "golifetime", 0, &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	if !strings.Contains(output.String(), "Topic with excluded letters.") {
		t.Errorf("output missing expected text: %q", output.String())
	}
}

// --- Body wrapping ---

// When wrapWidth > 0, long lines in the body are wrapped at that width.
func TestShow_body_wraps_when_wrapWidth_positive(t *testing.T) {
	store := newTestStore(t)
	// Construct a body that is a single long line well over 20 characters.
	longBody := "one two three four five six seven eight nine ten"
	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"go-lang"},
		Text:      longBody,
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Show(store, "abcdef", 20, &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	result := output.String()
	// textwrap.Wrap with width=20 must produce exactly these line breaks for
	// the 48-character input. Comparing against the exact wrapped string rules
	// out the false-positive where the trailing newline from fmt.Fprintln
	// satisfies a bare "contains newline" check even when wrapping is broken.
	wantBody := "one two three four\nfive six seven eight\nnine ten"
	if !strings.Contains(result, wantBody) {
		t.Errorf("expected wrapped body %q in output, got: %q", wantBody, result)
	}
}

// When wrapWidth is 0, the body is returned unchanged regardless of line length.
func TestShow_body_not_wrapped_when_wrapWidth_zero(t *testing.T) {
	store := newTestStore(t)
	longBody := "one two three four five six seven eight nine ten"
	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"go-lang"},
		Text:      longBody,
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Show(store, "abcdef", 0, &output); err != nil {
		t.Fatalf("Show: %v", err)
	}

	// The body must appear verbatim as a single line (no wrapping newlines).
	if !strings.Contains(output.String(), longBody) {
		t.Errorf("expected unwrapped body to appear verbatim, got: %q", output.String())
	}
}
