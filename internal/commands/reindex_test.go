package commands_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/skorokithakis/gnosis/internal/commands"
	"github.com/skorokithakis/gnosis/internal/index"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// newTestStoreWithRoot creates a Store backed by a temporary directory and
// returns both the store and the repo root (the parent of .gnosis). The index
// package expects the repo root, not the .gnosis directory itself.
func newTestStoreWithRoot(t *testing.T) (*storage.Store, string) {
	t.Helper()
	repoRoot := t.TempDir()
	store, err := storage.NewStoreAt(filepath.Join(repoRoot, ".gnosis"))
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	return store, repoRoot
}

// TestReindex_reports_count_and_path verifies that Reindex prints the number of
// indexed entries and the path to the database file.
func TestReindex_reports_count_and_path(t *testing.T) {
	store, repoRoot := newTestStoreWithRoot(t)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	appendEntries(t, store,
		sampleEntry("aaaaaa"),
		sampleEntry("bbbbbb"),
	)

	var output strings.Builder
	if err := commands.Reindex(repoRoot, store, &output); err != nil {
		t.Fatalf("Reindex: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "2 entries") {
		t.Errorf("output should report 2 entries, got: %q", result)
	}
	if !strings.Contains(result, "index.db") {
		t.Errorf("output should include path to index.db, got: %q", result)
	}
}

// TestReindex_empty_store reports zero entries without error.
func TestReindex_empty_store(t *testing.T) {
	store, repoRoot := newTestStoreWithRoot(t)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	var output strings.Builder
	if err := commands.Reindex(repoRoot, store, &output); err != nil {
		t.Fatalf("Reindex on empty store: %v", err)
	}

	if !strings.Contains(output.String(), "0 entries") {
		t.Errorf("expected 0 entries in output, got: %q", output.String())
	}
}

// TestReindex_single_entry_uses_singular verifies that the output uses "entry"
// (not "entries") when exactly one entry is indexed.
func TestReindex_single_entry_uses_singular(t *testing.T) {
	store, repoRoot := newTestStoreWithRoot(t)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	appendEntries(t, store, sampleEntry("ffffff"))

	var output strings.Builder
	if err := commands.Reindex(repoRoot, store, &output); err != nil {
		t.Fatalf("Reindex: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "1 entry") {
		t.Errorf("output should say '1 entry', got: %q", result)
	}
	if strings.Contains(result, "1 entries") {
		t.Errorf("output must not say '1 entries', got: %q", result)
	}
}

// TestReindex_subsequent_search_returns_results verifies the core acceptance
// criterion: after Reindex, a search on the index returns the expected entries.
func TestReindex_subsequent_search_returns_results(t *testing.T) {
	store, repoRoot := newTestStoreWithRoot(t)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	entry := storage.Entry{
		ID:      "cccccc",
		Topics:  []string{"GoLang"},
		Text:    "goroutines are lightweight threads managed by the Go runtime",
		Related: []string{},
	}
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Reindex(repoRoot, store, &output); err != nil {
		t.Fatalf("Reindex: %v", err)
	}

	// Open the same index (same XDG_CACHE_HOME, same repoRoot) and confirm the
	// entry is searchable, proving Rebuild actually populated the FTS table.
	idx, err := index.Open(repoRoot, store)
	if err != nil {
		t.Fatalf("index.Open after Reindex: %v", err)
	}
	defer idx.Close()

	hits, err := idx.Search("goroutine", 10)
	if err != nil {
		t.Fatalf("Search after Reindex: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit after Reindex, got %d", len(hits))
	}
	if hits[0].EntryID != "cccccc" {
		t.Errorf("expected entry ID %q, got %q", "cccccc", hits[0].EntryID)
	}
}

// TestReindex_overwrites_stale_index verifies that Reindex replaces a stale
// index: entries added after the initial build become searchable after reindex.
func TestReindex_overwrites_stale_index(t *testing.T) {
	store, repoRoot := newTestStoreWithRoot(t)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	// Populate the store with one entry and build the index.
	appendEntries(t, store, storage.Entry{
		ID:      "dddddd",
		Topics:  []string{"initial"},
		Text:    "first entry text",
		Related: []string{},
	})

	var output strings.Builder
	if err := commands.Reindex(repoRoot, store, &output); err != nil {
		t.Fatalf("first Reindex: %v", err)
	}

	// Add a second entry without rebuilding — the index is now stale.
	appendEntries(t, store, storage.Entry{
		ID:      "eeeeee",
		Topics:  []string{"added"},
		Text:    "second entry with unique word: xylophone",
		Related: []string{},
	})

	// Reindex again — this must pick up the new entry.
	output.Reset()
	if err := commands.Reindex(repoRoot, store, &output); err != nil {
		t.Fatalf("second Reindex: %v", err)
	}

	if !strings.Contains(output.String(), "2 entries") {
		t.Errorf("expected 2 entries after second Reindex, got: %q", output.String())
	}

	idx, err := index.Open(repoRoot, store)
	if err != nil {
		t.Fatalf("index.Open after second Reindex: %v", err)
	}
	defer idx.Close()

	hits, err := idx.Search("xylophone", 10)
	if err != nil {
		t.Fatalf("Search for new entry: %v", err)
	}
	if len(hits) != 1 || hits[0].EntryID != "eeeeee" {
		t.Errorf("expected entry %q after reindex, got %v", "eeeeee", hits)
	}
}
