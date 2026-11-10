package index_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/index"
	"github.com/skorokithakis/gnosis/internal/paths"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// newTestIndex creates a Store backed by a temporary directory, writes the
// given entries into it, and opens an Index pointing at a separate temp
// directory for the cache. It returns the index and a cleanup function.
func newTestIndex(t *testing.T, entries []storage.Entry) (*index.Index, *storage.Store, string) {
	t.Helper()

	storeDir := t.TempDir()
	gnosisDir := filepath.Join(storeDir, ".gnosis")
	// Use a separate temp directory as XDG_CACHE_HOME so the test never
	// touches the real user cache. Derive the per-repo cache subdirectory via
	// paths.CacheDir so the store's lock file lands in the same hashed
	// directory that index.Open will use.
	xdgCacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", xdgCacheHome)
	storeCacheDir, err := paths.CacheDir(storeDir)
	if err != nil {
		t.Fatalf("paths.CacheDir: %v", err)
	}
	store, err := storage.NewStoreAt(gnosisDir, storeCacheDir)
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}

	for _, entry := range entries {
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append entry %q: %v", entry.ID, err)
		}
	}

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

// TestIsCJK verifies which scripts are treated as CJK. Hangul is excluded
// because Korean separates words with spaces.
func TestIsCJK(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"Han", '非', true},
		{"Hiragana", 'あ', true},
		{"Katakana", 'ア', true},
		{"Katakana long vowel mark", 'ー', true},
		{"Latin", 'a', false},
		{"Hangul", '한', false},
		{"Digit", '1', false},
		{"CJK punctuation", '。', false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := index.IsCJK(test.r); got != test.want {
				t.Errorf("IsCJK(%q) = %v, want %v", test.r, got, test.want)
			}
		})
	}
}

// TestSplitCJK verifies that CJK characters are separated from their
// neighbours without disturbing existing whitespace or non-CJK text.
func TestSplitCJK(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"CJK run", "非法状态流转", "非 法 状 态 流 转"},
		{"Latin next to CJK", "API状态ok", "API 状 态 ok"},
		{"already separated", "非 法", "非 法"},
		{"newline is whitespace", "非\n法", "非\n法"},
		{"hiragana", "ひらがな", "ひ ら が な"},
		{"katakana long vowel mark", "ラーメン", "ラ ー メ ン"},
		{"CJK next to punctuation", "状态。", "状 态。"},
		{"CJK next to CJK punctuation and text", "状态。因为", "状 态。因 为"},
		{"CJK inside punctuation", "（状态）", "（状 态）"},
		{"CJK next to ASCII punctuation", "状态, ok", "状 态, ok"},
		{"CJK next to symbol", "状态+ok", "状 态+ok"},
		{"no CJK", "The quick brown fox", "The quick brown fox"},
		{"hangul is not split", "한국어 공부", "한국어 공부"},
		{"empty", "", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := index.SplitCJK(test.in); got != test.want {
				t.Errorf("SplitCJK(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}

// TestSearch_CJK_phrase verifies that a CJK phrase can be found even though the
// stored text has no spaces, and that the snippet shows the original text with
// the match sentinels intact.
func TestSearch_CJK_phrase(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"golang"}, "非法状态流转"),
		makeEntry("bbbbbb", []string{"golang"}, "Unrelated English text"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	hits, err := idx.Search(`"状 态 流 转"`, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].EntryID != "aaaaaa" {
		t.Errorf("expected entry ID %q, got %q", "aaaaaa", hits[0].EntryID)
	}

	snippet := hits[0].Snippet
	if !strings.Contains(snippet, "状态流转") {
		t.Errorf("expected snippet to contain %q, got %q", "状态流转", snippet)
	}
	if strings.Contains(snippet, "状 态") {
		t.Errorf("expected snippet without inserted CJK spaces, got %q", snippet)
	}
	if !strings.Contains(snippet, index.MatchStart) || !strings.Contains(snippet, index.MatchEnd) {
		t.Errorf("expected match sentinels in snippet, got %q", snippet)
	}
}

// TestSearch_CJK_next_to_latin verifies that a CJK run embedded in Latin text is
// split into separate tokens and can still be found, while the genuine space
// between the Latin and CJK text is preserved in the snippet.
func TestSearch_CJK_next_to_latin(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"golang"}, "API状态"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	hits, err := idx.Search(`"状 态"`, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}

	snippet := hits[0].Snippet
	if !strings.Contains(snippet, "状态") {
		t.Errorf("expected snippet to contain %q, got %q", "状态", snippet)
	}
	if strings.Contains(snippet, "状 态") {
		t.Errorf("expected snippet without inserted CJK spaces, got %q", snippet)
	}
	if !strings.Contains(snippet, "API") {
		t.Errorf("expected snippet to contain %q, got %q", "API", snippet)
	}
}

// TestSearch_CJK_topic verifies that a CJK topic is split before indexing and
// remains searchable.
func TestSearch_CJK_topic(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"状态流转"}, "English only text"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	hits, err := idx.Search(`"状 态"`, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].EntryID != "aaaaaa" {
		t.Errorf("expected entry ID %q, got %q", "aaaaaa", hits[0].EntryID)
	}
	if strings.Contains(hits[0].Snippet, "状 态") {
		t.Errorf("expected snippet without inserted CJK spaces, got %q", hits[0].Snippet)
	}
}

// TestEnsureFresh_rebuilds_on_version_mismatch verifies that EnsureFresh
// rebuilds an index written by an older format version even though the JSONL
// mtime is unchanged.
func TestEnsureFresh_rebuilds_on_version_mismatch(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"golang"}, "非法状态流转"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	// Simulate a cache written by a binary that predates format versioning:
	// no version key, no rows, but an mtime that still matches the JSONL file.
	withIndexDB(t, idx, func(db *sql.DB) error {
		if _, err := db.Exec(`DELETE FROM entries_fts`); err != nil {
			return err
		}
		_, err := db.Exec(`DELETE FROM metadata WHERE key = 'format_version'`)
		return err
	})

	if err := idx.EnsureFresh(); err != nil {
		t.Fatalf("EnsureFresh: %v", err)
	}

	hits, err := idx.Search(`"状 态 流 转"`, 10)
	if err != nil {
		t.Fatalf("Search after EnsureFresh: %v", err)
	}
	if len(hits) != 1 || hits[0].EntryID != "aaaaaa" {
		t.Fatalf("expected rebuilt index to return entry %q, got %v", "aaaaaa", hits)
	}

	var version string
	withIndexDB(t, idx, func(db *sql.DB) error {
		return db.QueryRow(`SELECT value FROM metadata WHERE key = 'format_version'`).Scan(&version)
	})
	if version == "" {
		t.Error("expected Rebuild to write a format version")
	}
}

// TestSearch_CJK_punctuation_snippet verifies that CJK text containing ordinary
// clause punctuation round-trips through indexing and snippet cleanup without
// stray spaces around the punctuation.
func TestSearch_CJK_punctuation_snippet(t *testing.T) {
	entries := []storage.Entry{
		makeEntry("aaaaaa", []string{"golang"}, "状态。因为，所以。"),
	}

	idx, _, _ := newTestIndex(t, entries)

	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	hits, err := idx.Search(`"状 态"`, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}

	snippet := hits[0].Snippet
	if !strings.Contains(snippet, "状态") || !strings.Contains(snippet, "。因为，所以。") {
		t.Errorf("expected snippet to join CJK around punctuation, got %q", snippet)
	}
	for _, stray := range []string{" 。", "。 ", " ，", "， "} {
		if strings.Contains(snippet, stray) {
			t.Errorf("expected no stray space %q in snippet, got %q", stray, snippet)
		}
	}
}

// withIndexDB opens the index's SQLite file directly and runs fn against it.
// Tests use it to put the cache into states that the public API cannot produce,
// such as a cache written by an older format version.
func withIndexDB(t *testing.T, idx *index.Index, fn func(*sql.DB) error) {
	t.Helper()

	db, err := sql.Open("sqlite", idx.DBPath())
	if err != nil {
		t.Fatalf("open index database: %v", err)
	}
	defer db.Close()

	if err := fn(db); err != nil {
		t.Fatalf("using index database: %v", err)
	}
}
