package index_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/index"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// newTestIndex creates a Store backed by a temporary directory, writes the
// given entries into it, and opens an Index pointing at a separate temp
// directory for the cache. It returns the index and a cleanup function.
func newTestIndex(t *testing.T, entries []storage.Entry) (*index.Index, *storage.Store, string) {
	t.Helper()

	storeDir := t.TempDir()
	gnosisDir := filepath.Join(storeDir, ".gnosis")
	store, err := storage.NewStoreAt(gnosisDir)
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}

	for _, entry := range entries {
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append entry %q: %v", entry.ID, err)
		}
	}

	// Use a separate temp directory as XDG_CACHE_HOME so the test never
	// touches the real user cache.
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	// repoRoot is the parent of .gnosis, which is storeDir.
	idx, err := index.Open(storeDir, store)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { idx.Close() })

	return idx, store, storeDir
}

func makeEntry(id string, topics []string, text string) storage.Entry {
	now := time.Now().UTC().Truncate(time.Second)
	return storage.Entry{
		ID:        id,
		Topics:    topics,
		Text:      text,
		Related:   []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// TestSearch_by_text verifies that a full-text search on the entry body returns
// the correct entry ID.
func TestSearch_by_text(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"golang"}, "The quick brown fox jumps over the lazy dog"),
		makeEntry("bbbbbb", []string{"python"}, "Monty Python and the Holy Grail"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	hits, err := idx.Search("fox", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].EntryID != "aaaaaa" {
		t.Errorf("expected entry ID %q, got %q", "aaaaaa", hits[0].EntryID)
	}
}

// TestSearch_by_topic_display verifies that searching for a topic's display
// form returns the entry tagged with that topic.
func TestSearch_by_topic_display(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"KeymasterTokenAuth"}, "Some entry about authentication"),
		makeEntry("bbbbbb", []string{"golang"}, "Unrelated entry"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	// "token" should match "KeymasterTokenAuth" via the normalized form
	// "keymaster-token-auth" which the porter stemmer tokenises into individual
	// words including "token".
	hits, err := idx.Search("token", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected at least 1 hit for 'token', got 0")
	}
	if hits[0].EntryID != "aaaaaa" {
		t.Errorf("expected entry ID %q, got %q", "aaaaaa", hits[0].EntryID)
	}
}

// TestSearch_by_topic_normalized verifies that searching for a word that
// appears only in the normalized form of a CamelCase topic returns the entry.
func TestSearch_by_topic_normalized(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("cccccc", []string{"KeymasterTokenAuth"}, "Entry about keymaster"),
		makeEntry("dddddd", []string{"unrelated"}, "Nothing to see here"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	hits, err := idx.Search("keymaster", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected at least 1 hit for 'keymaster', got 0")
	}
	found := false
	for _, hit := range hits {
		if hit.EntryID == "cccccc" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected entry %q in results, got %v", "cccccc", hits)
	}
}

// TestSearch_prefix_match verifies that a prefix query (e.g. "fox*") matches
// entries containing words that start with that prefix.
func TestSearch_prefix_match(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"animals"}, "The quick brown fox jumps"),
		makeEntry("bbbbbb", []string{"food"}, "Foxglove is a plant"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	hits, err := idx.Search("fox*", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) < 2 {
		t.Errorf("expected at least 2 hits for prefix 'fox*', got %d", len(hits))
	}
}

// TestEnsureFresh_rebuilds_when_stale verifies that EnsureFresh triggers a
// rebuild when the JSONL file is modified after the index was last built.
func TestEnsureFresh_rebuilds_when_stale(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"golang"}, "Initial entry"),
	}

	idx, store, _ := newTestIndex(t, entries)

	// Build the index for the first time.
	if err := idx.Rebuild(); err != nil {
		t.Fatalf("initial Rebuild: %v", err)
	}

	// Confirm the initial entry is searchable.
	hits, err := idx.Search("initial", 10)
	if err != nil {
		t.Fatalf("Search before staleness: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit before staleness, got %d", len(hits))
	}

	// Wait a moment so the mtime will differ, then append a new entry.
	time.Sleep(10 * time.Millisecond)
	newEntry := makeEntry("bbbbbb", []string{"python"}, "Newly added entry about snakes")
	if err := store.Append(newEntry); err != nil {
		t.Fatalf("Append new entry: %v", err)
	}

	// Touch the file to ensure the mtime changes even if the write was fast.
	gnosisDir := store.GnosisDir()
	entriesPath := filepath.Join(gnosisDir, "entries.jsonl")
	now := time.Now()
	if err := os.Chtimes(entriesPath, now, now); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	// EnsureFresh should detect the stale mtime and rebuild.
	if err := idx.EnsureFresh(); err != nil {
		t.Fatalf("EnsureFresh: %v", err)
	}

	// The new entry should now be searchable.
	hits, err = idx.Search("snakes", 10)
	if err != nil {
		t.Fatalf("Search after EnsureFresh: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit after EnsureFresh, got %d", len(hits))
	}
	if hits[0].EntryID != "bbbbbb" {
		t.Errorf("expected entry ID %q, got %q", "bbbbbb", hits[0].EntryID)
	}
}

// TestEnsureFresh_no_rebuild_when_fresh verifies that EnsureFresh does not
// rebuild when the JSONL mtime matches the stored mtime.
func TestEnsureFresh_no_rebuild_when_fresh(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"golang"}, "Only entry"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	// EnsureFresh on an already-fresh index should be a no-op. We verify this
	// indirectly by confirming Search still returns the expected result.
	if err := idx.EnsureFresh(); err != nil {
		t.Fatalf("EnsureFresh on fresh index: %v", err)
	}

	hits, err := idx.Search("only", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 || hits[0].EntryID != "aaaaaa" {
		t.Errorf("unexpected hits after no-op EnsureFresh: %v", hits)
	}
}

// TestSearch_snippet_present verifies that search results include a non-empty
// snippet string.
func TestSearch_snippet_present(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"golang"}, "The quick brown fox jumps over the lazy dog"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	hits, err := idx.Search("fox", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected at least 1 hit")
	}
	if hits[0].Snippet == "" {
		t.Error("expected non-empty snippet")
	}
}

// TestSearch_empty_index verifies that searching an empty index returns no
// results without error.
func TestSearch_empty_index(t *testing.T) {
	idx, _, _ := newTestIndex(t, nil)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild on empty store: %v", err)
	}

	hits, err := idx.Search("anything", 10)
	if err != nil {
		t.Fatalf("Search on empty index: %v", err)
	}
	if len(hits) != 0 {
		t.Errorf("expected 0 hits on empty index, got %d", len(hits))
	}
}
