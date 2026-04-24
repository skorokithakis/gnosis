package commands_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/commands"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// newSearchTestStore creates a Store backed by a temporary directory and sets
// XDG_CACHE_HOME to a separate temp directory so the index never touches the
// real user cache.
func newSearchTestStore(t *testing.T) *storage.Store {
	t.Helper()
	storeDir := t.TempDir()
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	// FindRepoRoot walks up from the working directory looking for .gnosis, so
	// we place the .gnosis directory inside storeDir and change the working
	// directory to storeDir so that FindRepoRoot resolves to it.
	gnosisDir := filepath.Join(storeDir, ".gnosis")
	if err := os.MkdirAll(gnosisDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(storeDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(original) }) //nolint:errcheck

	store, err := storage.NewStoreAt(gnosisDir)
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	return store
}

func makeSearchEntry(id string, topics []string, text string) storage.Entry {
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

// TestSearch_finds_entry_by_text verifies that writing an entry and then
// searching for a word from its body returns that entry.
func TestSearch_finds_entry_by_text(t *testing.T) {
	store := newSearchTestStore(t)
	entry := makeSearchEntry("aaaaaa", []string{"GoLang"}, "hello world from gnosis")
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	var output strings.Builder
	if err := commands.Search(store, []string{"hello"}, &output); err != nil {
		t.Fatalf("Search: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "aaaaaa") {
		t.Errorf("output missing entry ID: %q", result)
	}
	// Topics are stored normalized, so the output shows the normalized form.
	if !strings.Contains(result, "go-lang") {
		t.Errorf("output missing normalized topic: %q", result)
	}
}

// TestSearch_finds_entry_by_topic verifies that topic names are searchable.
func TestSearch_finds_entry_by_topic(t *testing.T) {
	store := newSearchTestStore(t)
	entry := makeSearchEntry("bbbbbb", []string{"KeymasterTokenAuth"}, "some unrelated body text")
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	var output strings.Builder
	if err := commands.Search(store, []string{"keymaster"}, &output); err != nil {
		t.Fatalf("Search: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "bbbbbb") {
		t.Errorf("output missing entry ID: %q", result)
	}
	// Topics are stored normalized, so the output shows the normalized form.
	if !strings.Contains(result, "keymaster-token-auth") {
		t.Errorf("output missing normalized topic: %q", result)
	}
}

// TestSearch_limit_truncates verifies that --limit restricts the number of
// results returned.
func TestSearch_limit_truncates(t *testing.T) {
	store := newSearchTestStore(t)

	// Write five entries that all match the query.
	for i, id := range []string{"aaaaaa", "bbbbbb", "cccccc", "dddddd", "eeeeee"} {
		entry := makeSearchEntry(id, []string{"GoLang"}, strings.Repeat("searchable ", i+1))
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append %s: %v", id, err)
		}
	}

	var output strings.Builder
	if err := commands.Search(store, []string{"searchable", "--limit", "2"}, &output); err != nil {
		t.Fatalf("Search: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 result lines with --limit 2, got %d: %q", len(lines), output.String())
	}
}

// TestSearch_no_results_exits_cleanly verifies that a query with no matches
// produces no output and no error.
func TestSearch_no_results_exits_cleanly(t *testing.T) {
	store := newSearchTestStore(t)
	entry := makeSearchEntry("aaaaaa", []string{"GoLang"}, "hello world")
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	var output strings.Builder
	if err := commands.Search(store, []string{"zzznomatch"}, &output); err != nil {
		t.Fatalf("Search: %v", err)
	}

	if output.String() != "" {
		t.Errorf("expected empty output for no results, got: %q", output.String())
	}
}

// TestSearch_multiword_query verifies that a multi-word query (passed as
// separate argv elements) matches entries containing all those words.
func TestSearch_multiword_query(t *testing.T) {
	store := newSearchTestStore(t)
	entry := makeSearchEntry("aaaaaa", []string{"GoLang"}, "the quick brown fox jumps")
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	var output strings.Builder
	// FTS5 treats space-separated terms as implicit AND by default.
	if err := commands.Search(store, []string{"quick brown"}, &output); err != nil {
		t.Fatalf("Search: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "aaaaaa") {
		t.Errorf("multi-word query did not find expected entry: %q", result)
	}
}

// TestSearch_missing_query_returns_error verifies that calling Search with no
// query arguments returns an error.
func TestSearch_missing_query_returns_error(t *testing.T) {
	store := newSearchTestStore(t)

	var output strings.Builder
	err := commands.Search(store, []string{}, &output)
	if err == nil {
		t.Fatal("expected error for missing query, got nil")
	}
}

// TestSearch_invalid_limit_returns_error verifies that a non-integer --limit
// value returns an error.
func TestSearch_invalid_limit_returns_error(t *testing.T) {
	store := newSearchTestStore(t)

	var output strings.Builder
	err := commands.Search(store, []string{"hello", "--limit", "notanumber"}, &output)
	if err == nil {
		t.Fatal("expected error for invalid --limit, got nil")
	}
}

// TestSearch_hyphenated_query_finds_entry verifies that a hyphenated query
// like "keymaster-token-auth" (the form shown by "topics" and "show") is
// sanitized before being passed to FTS5 so that it does not trigger FTS5's
// column-qualifier parser and instead matches entries tagged with the
// corresponding topic.
func TestSearch_hyphenated_query_finds_entry(t *testing.T) {
	store := newSearchTestStore(t)
	entry := makeSearchEntry("cccccc", []string{"KeymasterTokenAuth"}, "some unrelated body text")
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	var output strings.Builder
	if err := commands.Search(store, []string{"keymaster-token-auth"}, &output); err != nil {
		t.Fatalf("Search returned unexpected error: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "cccccc") {
		t.Errorf("hyphenated query did not find expected entry: %q", result)
	}
}

// TestSearch_explicit_operator_passthrough verifies that a query containing
// explicit FTS5 boolean operators (AND, OR, NOT) is passed through to FTS5
// unchanged so that users who know the syntax can use it.
func TestSearch_explicit_operator_passthrough(t *testing.T) {
	store := newSearchTestStore(t)
	entry := makeSearchEntry("dddddd", []string{"GoLang"}, "foo and bar together")
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	var output strings.Builder
	// "foo AND bar" uses explicit FTS5 AND; both terms must appear.
	if err := commands.Search(store, []string{"foo AND bar"}, &output); err != nil {
		t.Fatalf("Search returned unexpected error: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "dddddd") {
		t.Errorf("explicit AND query did not find expected entry: %q", result)
	}
}

// TestSearch_snippet_has_no_newlines verifies that when an entry's text
// contains embedded newlines, the snippet printed for a matching search result
// contains no newline characters — preserving the one-entry-per-line output
// contract.
func TestSearch_snippet_has_no_newlines(t *testing.T) {
	store := newSearchTestStore(t)
	// Embed newlines in the text so that FTS5's snippet() is likely to include
	// them in the excerpt it selects.
	multilineText := "first paragraph with the term uniqueword\n\nsecond paragraph\n\nthird paragraph"
	entry := makeSearchEntry("eeeeee", []string{"GoLang"}, multilineText)
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	var output strings.Builder
	if err := commands.Search(store, []string{"uniqueword"}, &output); err != nil {
		t.Fatalf("Search returned unexpected error: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "eeeeee") {
		t.Fatalf("search did not find expected entry: %q", result)
	}

	// Each line in the output corresponds to one result. If any line (after
	// splitting on \n) contains the entry ID, that is the result line — it
	// must not itself contain a newline, which is guaranteed by the split, but
	// we also verify that the raw output has exactly one trailing newline for
	// the single result (i.e. no embedded newlines in the snippet).
	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected exactly 1 output line for 1 result, got %d lines: %q", len(lines), result)
	}
}
