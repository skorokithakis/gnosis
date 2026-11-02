package commands_test

import (
	"io"
	"strings"
	"testing"

	"github.com/skorokithakis/gnosis/internal/commands"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// --- ResolveIDPrefix unit tests ---

func TestResolveIDPrefix_full_id_matches(t *testing.T) {
	entries := []storage.Entry{
		{ID: "abcdef"},
		{ID: "ghjkmn"},
	}
	got, err := commands.ResolveIDPrefix(entries, "abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abcdef" {
		t.Errorf("expected %q, got %q", "abcdef", got)
	}
}

func TestResolveIDPrefix_short_prefix_matches_unique(t *testing.T) {
	entries := []storage.Entry{
		{ID: "abcdef"},
		{ID: "ghjkmn"},
	}
	got, err := commands.ResolveIDPrefix(entries, "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abcdef" {
		t.Errorf("expected %q, got %q", "abcdef", got)
	}
}

func TestResolveIDPrefix_single_char_prefix(t *testing.T) {
	entries := []storage.Entry{
		{ID: "abcdef"},
		{ID: "ghjkmn"},
	}
	got, err := commands.ResolveIDPrefix(entries, "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abcdef" {
		t.Errorf("expected %q, got %q", "abcdef", got)
	}
}

func TestResolveIDPrefix_not_found(t *testing.T) {
	entries := []storage.Entry{
		{ID: "abcdef"},
	}
	_, err := commands.ResolveIDPrefix(entries, "xyz")
	if err == nil {
		t.Fatal("expected error for non-matching prefix, got nil")
	}
	if !strings.Contains(err.Error(), "xyz") {
		t.Errorf("error should mention the prefix, got: %v", err)
	}
}

func TestResolveIDPrefix_ambiguous_returns_candidates(t *testing.T) {
	entries := []storage.Entry{
		{ID: "abcdef"},
		{ID: "abcxyz"},
	}
	_, err := commands.ResolveIDPrefix(entries, "abc")
	if err == nil {
		t.Fatal("expected error for ambiguous prefix, got nil")
	}
	if !strings.Contains(err.Error(), "abcdef") || !strings.Contains(err.Error(), "abcxyz") {
		t.Errorf("error should list candidate IDs, got: %v", err)
	}
}

func TestResolveIDPrefix_excluded_letter_not_found(t *testing.T) {
	// 'i', 'l', 'o' are excluded from the ID alphabet, so any prefix
	// containing them can never match an entry ID.
	entries := []storage.Entry{
		{ID: "abcdef"},
	}
	for _, prefix := range []string{"oil", "life", "lion"} {
		_, err := commands.ResolveIDPrefix(entries, prefix)
		if err == nil {
			t.Errorf("expected error for prefix %q with excluded letters, got nil", prefix)
		}
	}
}

func TestResolveIDPrefix_empty_store(t *testing.T) {
	_, err := commands.ResolveIDPrefix([]storage.Entry{}, "abc")
	if err == nil {
		t.Fatal("expected error for empty store, got nil")
	}
}

func TestResolveIDPrefix_empty_prefix_rejected(t *testing.T) {
	// An empty prefix would match every entry. It must be rejected regardless
	// of how many entries are in the store, including the single-entry case
	// where it would otherwise silently resolve.
	entries := []storage.Entry{
		{ID: "abcdef"},
	}
	_, err := commands.ResolveIDPrefix(entries, "")
	if err == nil {
		t.Fatal("expected error for empty prefix, got nil")
	}
}

// --- Show with prefix ---

func TestShow_prefix_resolves_to_entry(t *testing.T) {
	store := newTestStore(t)
	entry := storage.Entry{
		ID:        "abcdef",
		Topics:    []string{"keymaster-token-auth"},
		Text:      "Found by prefix.",
		Related:   []string{},
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Show(store, "abc", 0, &output); err != nil {
		t.Fatalf("Show with prefix: %v", err)
	}
	if !strings.Contains(output.String(), "Found by prefix.") {
		t.Errorf("output missing expected text: %q", output.String())
	}
}

func TestShow_ambiguous_prefix_returns_error(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store,
		sampleEntry("abcdef"),
		sampleEntry("abcxyz"),
	)

	var output strings.Builder
	err := commands.Show(store, "abc", 0, &output)
	if err == nil {
		t.Fatal("expected error for ambiguous prefix, got nil")
	}
	if !strings.Contains(err.Error(), "abcdef") || !strings.Contains(err.Error(), "abcxyz") {
		t.Errorf("error should list candidates, got: %v", err)
	}
}

// --- Remove with prefix ---

func TestRemove_prefix_resolves_and_removes(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store,
		sampleEntry("aaaaaa"),
		sampleEntry("bbbbbb"),
	)

	var stdout, stderr strings.Builder
	if err := commands.Remove(store, []string{"aaa"}, &stdout, &stderr); err != nil {
		t.Fatalf("Remove with prefix: %v", err)
	}

	remaining, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != "bbbbbb" {
		t.Errorf("expected only bbbbbb to survive, got %v", remaining)
	}
	// stdout should contain the full resolved ID, not the prefix.
	if !strings.Contains(stdout.String(), "aaaaaa") {
		t.Errorf("stdout should contain full resolved ID, got %q", stdout.String())
	}
}

func TestRemove_ambiguous_prefix_refuses_all(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store,
		sampleEntry("aaaaaa"),
		sampleEntry("aabbbb"),
	)

	var stdout, stderr strings.Builder
	err := commands.Remove(store, []string{"aa"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for ambiguous prefix, got nil")
	}

	// Neither entry should have been removed.
	remaining, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(remaining) != 2 {
		t.Errorf("expected 2 entries after failed remove, got %d", len(remaining))
	}
}

// --- Write --related with prefix ---

func TestWrite_related_prefix_resolves_to_full_id(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store, sampleEntry("aaaaaa"))

	var output strings.Builder
	// Use a prefix "aaa" that uniquely resolves to "aaaaaa".
	if err := commands.Write(store, []string{"go-language", "new entry", "--related", "aaa"}, &output); err != nil {
		t.Fatalf("Write with prefix --related: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	// Find the newly written entry (not "aaaaaa").
	var newEntry storage.Entry
	for _, entry := range entries {
		if entry.ID != "aaaaaa" {
			newEntry = entry
		}
	}
	if len(newEntry.Related) != 1 || newEntry.Related[0] != "aaaaaa" {
		t.Errorf("expected Related to contain full ID %q, got %v", "aaaaaa", newEntry.Related)
	}
}

func TestWrite_related_ambiguous_prefix_errors(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store,
		sampleEntry("aaaaaa"),
		sampleEntry("aabbbb"),
	)

	err := commands.Write(store, []string{"go-language", "new entry", "--related", "aa"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for ambiguous --related prefix, got nil")
	}
}
