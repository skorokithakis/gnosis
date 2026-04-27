package commands_test

import (
	"strings"
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/commands"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// makeLatestEntry builds an Entry with an explicit CreatedAt so tests can
// control ordering without relying on wall-clock timing.
func makeLatestEntry(id string, topics []string, text string, createdAt time.Time) storage.Entry {
	return storage.Entry{
		ID:        id,
		Topics:    topics,
		Text:      text,
		Related:   []string{},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
}

// TestLatest_empty_store_exits_cleanly verifies that an empty store produces no
// output and no error.
func TestLatest_empty_store_exits_cleanly(t *testing.T) {
	store := newTestStore(t)

	var output strings.Builder
	if err := commands.Latest(store, []string{}, &output); err != nil {
		t.Fatalf("Latest: %v", err)
	}

	if output.String() != "" {
		t.Errorf("expected empty output for empty store, got: %q", output.String())
	}
}

// TestLatest_newest_first verifies that entries are printed newest-first.
func TestLatest_newest_first(t *testing.T) {
	store := newTestStore(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	older := makeLatestEntry("aaaaaa", []string{"golang"}, "older entry", base)
	newer := makeLatestEntry("bbbbbb", []string{"golang"}, "newer entry", base.Add(time.Hour))

	appendEntries(t, store, older, newer)

	var output strings.Builder
	if err := commands.Latest(store, []string{}, &output); err != nil {
		t.Fatalf("Latest: %v", err)
	}

	result := output.String()
	olderPos := strings.Index(result, "aaaaaa")
	newerPos := strings.Index(result, "bbbbbb")

	if olderPos == -1 || newerPos == -1 {
		t.Fatalf("expected both entry IDs in output, got: %q", result)
	}
	if newerPos > olderPos {
		t.Errorf("expected newer entry (bbbbbb) before older entry (aaaaaa), got:\n%s", result)
	}
}

// TestLatest_limit_truncates verifies that --limit N restricts output to N lines.
func TestLatest_limit_truncates(t *testing.T) {
	store := newTestStore(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, id := range []string{"aaaaaa", "bbbbbb", "cccccc", "dddddd", "eeeeee"} {
		entry := makeLatestEntry(id, []string{"golang"}, "entry text", base.Add(time.Duration(i)*time.Hour))
		appendEntries(t, store, entry)
	}

	var output strings.Builder
	if err := commands.Latest(store, []string{"--limit", "3"}, &output); err != nil {
		t.Fatalf("Latest: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines with --limit 3, got %d: %q", len(lines), output.String())
	}
}

// TestLatest_default_limit_is_20 verifies that without --limit the output is
// capped at 20 entries even when the store holds more.
func TestLatest_default_limit_is_20(t *testing.T) {
	store := newTestStore(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	// alphabet used by gnosis IDs excludes i, l, o; build 25 distinct IDs
	// manually using only safe characters.
	ids := []string{
		"aaaaaa", "aaaaab", "aaaaac", "aaaaad", "aaaaae",
		"aaaaaf", "aaaaag", "aaaaah", "aaaaaj", "aaaaak",
		"aaaaam", "aaaaan", "aaaaap", "aaaaaq", "aaaaar",
		"aaaaat", "aaaaau", "aaaaav", "aaaaaw", "aaaaax",
		"aaaaay", "aaaaaz", "aaaaba", "aaaabb", "aaaabc",
	}
	for index, id := range ids {
		entry := makeLatestEntry(id, []string{"golang"}, "entry text", base.Add(time.Duration(index)*time.Hour))
		appendEntries(t, store, entry)
	}

	var output strings.Builder
	if err := commands.Latest(store, []string{}, &output); err != nil {
		t.Fatalf("Latest: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 20 {
		t.Errorf("expected 20 lines (default limit), got %d", len(lines))
	}
}

// TestLatest_shows_id_and_topic verifies that each output line contains the
// entry ID and its primary topic.
func TestLatest_shows_id_and_topic(t *testing.T) {
	store := newTestStore(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	entry := makeLatestEntry("aaaaaa", []string{"MyTopic"}, "some text", base)
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Latest(store, []string{}, &output); err != nil {
		t.Fatalf("Latest: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "aaaaaa") {
		t.Errorf("output missing entry ID: %q", result)
	}
	// Topics are stored normalized.
	if !strings.Contains(result, "my-topic") {
		t.Errorf("output missing normalized topic: %q", result)
	}
}

// TestLatest_snippet_has_no_newlines verifies that embedded newlines in entry
// text are collapsed so each result occupies exactly one output line.
func TestLatest_snippet_has_no_newlines(t *testing.T) {
	store := newTestStore(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	entry := makeLatestEntry("aaaaaa", []string{"golang"}, "line one\nline two\nline three", base)
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Latest(store, []string{}, &output); err != nil {
		t.Fatalf("Latest: %v", err)
	}

	lines := strings.Split(strings.TrimRight(output.String(), "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected exactly 1 output line for 1 entry, got %d: %q", len(lines), output.String())
	}
}

// TestLatest_invalid_limit_returns_error verifies that a non-integer --limit
// value returns an error.
func TestLatest_invalid_limit_returns_error(t *testing.T) {
	store := newTestStore(t)

	var output strings.Builder
	err := commands.Latest(store, []string{"--limit", "notanumber"}, &output)
	if err == nil {
		t.Fatal("expected error for invalid --limit, got nil")
	}
}

// TestLatest_missing_limit_value_returns_error verifies that --limit with no
// following value returns an error.
func TestLatest_missing_limit_value_returns_error(t *testing.T) {
	store := newTestStore(t)

	var output strings.Builder
	err := commands.Latest(store, []string{"--limit"}, &output)
	if err == nil {
		t.Fatal("expected error for missing --limit value, got nil")
	}
}

// TestLatest_snippet_truncated_at_200_runes verifies that entry text longer
// than 200 runes is cut at exactly 200 runes and followed by "…". Text at or
// below the limit must appear in full without the ellipsis.
func TestLatest_snippet_truncated_at_200_runes(t *testing.T) {
	store := newTestStore(t)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Build a 201-rune string using a multi-byte rune (é, U+00E9) to confirm
	// the boundary is measured in runes, not bytes.
	long := strings.Repeat("é", 201)
	entry := makeLatestEntry("aaaaaa", []string{"golang"}, long, base)
	appendEntries(t, store, entry)

	var output strings.Builder
	if err := commands.Latest(store, []string{}, &output); err != nil {
		t.Fatalf("Latest: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "…") {
		t.Errorf("expected ellipsis in truncated snippet, got: %q", result)
	}

	// Extract the snippet portion (everything after the two-space gap following
	// the topic column) and verify it is exactly 200 runes + the ellipsis.
	// The output line format is: "<id>  <topic>  <snippet>\n"
	// Split on the double-space that precedes the snippet.
	parts := strings.SplitN(strings.TrimRight(result, "\n"), "  ", 3)
	if len(parts) != 3 {
		t.Fatalf("unexpected output format: %q", result)
	}
	snippetRunes := []rune(parts[2])
	// Last rune must be the ellipsis; the preceding 200 runes are the text.
	if snippetRunes[len(snippetRunes)-1] != '…' {
		t.Errorf("last rune of snippet should be ellipsis, got %q", string(snippetRunes[len(snippetRunes)-1]))
	}
	if len(snippetRunes) != 201 { // 200 text runes + 1 ellipsis rune
		t.Errorf("expected 201 runes in truncated snippet (200 text + ellipsis), got %d", len(snippetRunes))
	}

	// A 200-rune entry must not be truncated.
	store2 := newTestStore(t)
	exact := strings.Repeat("é", 200)
	entry2 := makeLatestEntry("bbbbbb", []string{"golang"}, exact, base)
	appendEntries(t, store2, entry2)

	var output2 strings.Builder
	if err := commands.Latest(store2, []string{}, &output2); err != nil {
		t.Fatalf("Latest (exact): %v", err)
	}
	if strings.Contains(output2.String(), "…") {
		t.Errorf("200-rune entry should not be truncated, got: %q", output2.String())
	}
}

// TestLatest_unknown_flag_returns_error verifies that unrecognised arguments
// return an error rather than silently being ignored.
func TestLatest_unknown_flag_returns_error(t *testing.T) {
	store := newTestStore(t)

	var output strings.Builder
	err := commands.Latest(store, []string{"--unknown"}, &output)
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}
