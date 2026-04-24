package commands_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/commands"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// --- ParseEditBuffer ---

func TestParseEditBuffer_basic(t *testing.T) {
	input := "# Topics: GoLang, Testing\n# Related: aaaaaa\n# ---\nsome body text"
	parsed, err := commands.ParseEditBuffer(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.Topics) != 2 || parsed.Topics[0] != "go-lang" || parsed.Topics[1] != "testing" {
		t.Errorf("unexpected topics: %v", parsed.Topics)
	}
	if len(parsed.Related) != 1 || parsed.Related[0] != "aaaaaa" {
		t.Errorf("unexpected related: %v", parsed.Related)
	}
	if parsed.Text != "some body text" {
		t.Errorf("unexpected text: %q", parsed.Text)
	}
}

func TestParseEditBuffer_empty_related(t *testing.T) {
	input := "# Topics: GoLang\n# Related: \n# ---\nbody"
	parsed, err := commands.ParseEditBuffer(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.Related) != 0 {
		t.Errorf("expected empty related, got %v", parsed.Related)
	}
}

func TestParseEditBuffer_multiline_body(t *testing.T) {
	input := "# Topics: GoLang\n# Related: \n# ---\nline one\nline two\nline three"
	parsed, err := commands.ParseEditBuffer(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Text != "line one\nline two\nline three" {
		t.Errorf("unexpected text: %q", parsed.Text)
	}
}

func TestParseEditBuffer_body_with_hash(t *testing.T) {
	// Hash characters in the body should not be treated as header lines.
	input := "# Topics: GoLang\n# Related: \n# ---\n# this is not a header\nbody"
	parsed, err := commands.ParseEditBuffer(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Text != "# this is not a header\nbody" {
		t.Errorf("unexpected text: %q", parsed.Text)
	}
}

func TestParseEditBuffer_missing_separator(t *testing.T) {
	input := "# Topics: GoLang\n# Related: \nbody without separator"
	_, err := commands.ParseEditBuffer(input)
	if err == nil {
		t.Fatal("expected error for missing separator, got nil")
	}
}

func TestParseEditBuffer_trims_body_whitespace(t *testing.T) {
	input := "# Topics: GoLang\n# Related: \n# ---\n\n  body with leading blank  \n\n"
	parsed, err := commands.ParseEditBuffer(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Text != "body with leading blank" {
		t.Errorf("unexpected text: %q", parsed.Text)
	}
}

// --- FormatEditBuffer / round-trip ---

func TestFormatEditBuffer_round_trip(t *testing.T) {
	entry := sampleEntry("aaaaaa")
	entry.Topics = []string{"GoLang", "Testing"}
	entry.Related = []string{"bbbbbb"}
	entry.Text = "hello world"

	buf := commands.FormatEditBuffer(entry)
	parsed, err := commands.ParseEditBuffer(buf)
	if err != nil {
		t.Fatalf("ParseEditBuffer: %v", err)
	}

	if len(parsed.Topics) != 2 || parsed.Topics[0] != "go-lang" || parsed.Topics[1] != "testing" {
		t.Errorf("topics mismatch: %v", parsed.Topics)
	}
	if len(parsed.Related) != 1 || parsed.Related[0] != "bbbbbb" {
		t.Errorf("related mismatch: %v", parsed.Related)
	}
	if parsed.Text != "hello world" {
		t.Errorf("text mismatch: %q", parsed.Text)
	}
}

func TestFormatEditBuffer_empty_related_line_present(t *testing.T) {
	entry := sampleEntry("aaaaaa")
	entry.Related = []string{}
	buf := commands.FormatEditBuffer(entry)

	// The Related line must be present even when empty so the user can add IDs.
	if !containsPrefix(buf, "# Related:") {
		t.Errorf("Related line missing from buffer:\n%s", buf)
	}
}

func containsPrefix(content, prefix string) bool {
	for _, line := range splitLines(content) {
		if len(line) >= len(prefix) && line[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for index, character := range s {
		if character == '\n' {
			lines = append(lines, s[start:index])
			start = index + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

// --- Edit integration (editor stubbed via $EDITOR) ---

// writeScript writes a small shell script to a temp file and makes it
// executable. It returns the path to the script.
func writeScript(t *testing.T, content string) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "editor-*.sh")
	if err != nil {
		t.Fatalf("creating script: %v", err)
	}
	if _, err := file.WriteString("#!/bin/sh\n" + content + "\n"); err != nil {
		t.Fatalf("writing script: %v", err)
	}
	if err := file.Chmod(0o755); err != nil {
		t.Fatalf("chmod script: %v", err)
	}
	file.Close()
	return file.Name()
}

func TestEdit_no_changes(t *testing.T) {
	store := newTestStore(t)
	entry := sampleEntry("aaaaaa")
	entry.Topics = []string{"GoLang"}
	entry.Text = "original text"
	entry.Related = []string{}
	appendEntries(t, store, entry)

	// Editor that does nothing (leaves the file unchanged).
	t.Setenv("EDITOR", "true")

	if err := commands.Edit(store, "aaaaaa"); err != nil {
		t.Fatalf("Edit returned error: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// UpdatedAt must not have changed.
	if !entries[0].UpdatedAt.Equal(entry.UpdatedAt) {
		t.Errorf("UpdatedAt changed unexpectedly: got %v, want %v", entries[0].UpdatedAt, entry.UpdatedAt)
	}
}

func TestEdit_updates_entry(t *testing.T) {
	store := newTestStore(t)
	entry := sampleEntry("aaaaaa")
	entry.Topics = []string{"GoLang"}
	entry.Text = "original text"
	entry.Related = []string{}
	appendEntries(t, store, entry)

	// Editor that replaces the file content with a modified version.
	script := writeScript(t, fmt.Sprintf(`cat > "$1" <<'EOF'
# Topics: GoLang, NewTopic
# Related: 
# ---
updated text
EOF`))
	t.Setenv("EDITOR", script)

	before := time.Now().UTC()
	if err := commands.Edit(store, "aaaaaa"); err != nil {
		t.Fatalf("Edit returned error: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	got := entries[0]
	if got.ID != "aaaaaa" {
		t.Errorf("ID changed: got %q", got.ID)
	}
	if !got.CreatedAt.Equal(entry.CreatedAt) {
		t.Errorf("CreatedAt changed: got %v, want %v", got.CreatedAt, entry.CreatedAt)
	}
	if got.UpdatedAt.Before(before) {
		t.Errorf("UpdatedAt not updated: got %v", got.UpdatedAt)
	}
	if got.Text != "updated text" {
		t.Errorf("Text not updated: got %q", got.Text)
	}
	if len(got.Topics) != 2 || got.Topics[1] != "new-topic" {
		t.Errorf("Topics not updated: got %v", got.Topics)
	}
}

func TestEdit_missing_entry(t *testing.T) {
	store := newTestStore(t)
	t.Setenv("EDITOR", "true")

	err := commands.Edit(store, "zzzzzz")
	if err == nil {
		t.Fatal("expected error for missing entry, got nil")
	}
}

func TestEdit_empty_topics_rejected(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store, sampleEntry("aaaaaa"))

	script := writeScript(t, `cat > "$1" <<'EOF'
# Topics: 
# Related: 
# ---
some text
EOF`)
	t.Setenv("EDITOR", script)

	err := commands.Edit(store, "aaaaaa")
	if err == nil {
		t.Fatal("expected error for empty topics, got nil")
	}
}

func TestEdit_empty_body_rejected(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store, sampleEntry("aaaaaa"))

	script := writeScript(t, `cat > "$1" <<'EOF'
# Topics: GoLang
# Related: 
# ---
EOF`)
	t.Setenv("EDITOR", script)

	err := commands.Edit(store, "aaaaaa")
	if err == nil {
		t.Fatal("expected error for empty body, got nil")
	}
}

func TestEdit_invalid_related_id_rejected(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store, sampleEntry("aaaaaa"))

	script := writeScript(t, `cat > "$1" <<'EOF'
# Topics: GoLang
# Related: zzzzzz
# ---
some text
EOF`)
	t.Setenv("EDITOR", script)

	err := commands.Edit(store, "aaaaaa")
	if err == nil {
		t.Fatal("expected error for invalid related ID, got nil")
	}
}

func TestEdit_valid_related_id_accepted(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store, sampleEntry("aaaaaa"), sampleEntry("bbbbbb"))

	script := writeScript(t, `cat > "$1" <<'EOF'
# Topics: GoLang
# Related: bbbbbb
# ---
updated text
EOF`)
	t.Setenv("EDITOR", script)

	if err := commands.Edit(store, "aaaaaa"); err != nil {
		t.Fatalf("Edit returned unexpected error: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	var updated storage.Entry
	for _, entry := range entries {
		if entry.ID == "aaaaaa" {
			updated = entry
		}
	}
	if len(updated.Related) != 1 || updated.Related[0] != "bbbbbb" {
		t.Errorf("Related not updated: got %v", updated.Related)
	}
}

// TestEdit_topic_casing_change_is_no_change verifies that changing only the
// casing of a topic (e.g. "go-lang" → "GoLang") is not treated as a change,
// because both forms normalize to the same string.
func TestEdit_topic_casing_change_is_no_change(t *testing.T) {
	store := newTestStore(t)
	entry := sampleEntry("aaaaaa")
	// Store normalizes "go-lang" on write; original.Topics will be ["go-lang"].
	entry.Topics = []string{"go-lang"}
	entry.Text = "original text"
	entry.Related = []string{}
	appendEntries(t, store, entry)

	// Editor writes "GoLang" — which normalizes to "go-lang", the same as stored.
	script := writeScript(t, `cat > "$1" <<'EOF'
# Topics: GoLang
# Related: 
# ---
original text
EOF`)
	t.Setenv("EDITOR", script)

	if err := commands.Edit(store, "aaaaaa"); err != nil {
		t.Fatalf("Edit returned error: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	// UpdatedAt must not have changed because there were no real changes.
	if !entries[0].UpdatedAt.Equal(entry.UpdatedAt) {
		t.Errorf("UpdatedAt changed despite only a casing difference: got %v, want %v", entries[0].UpdatedAt, entry.UpdatedAt)
	}
}

// TestEdit_duplicate_normalized_topics_in_buffer_saves_one verifies that
// writing "FooTopic, foo-topic" in the topics line saves only one topic, not two.
func TestEdit_duplicate_normalized_topics_in_buffer_saves_one(t *testing.T) {
	store := newTestStore(t)
	entry := sampleEntry("aaaaaa")
	entry.Topics = []string{"foo-topic"}
	entry.Text = "original text"
	entry.Related = []string{}
	appendEntries(t, store, entry)

	// Editor writes "FooTopic, foo-topic" — both normalize to "foo-topic"; only one should be saved.
	script := writeScript(t, `cat > "$1" <<'EOF'
# Topics: FooTopic, foo-topic
# Related: 
# ---
updated text
EOF`)
	t.Setenv("EDITOR", script)

	if err := commands.Edit(store, "aaaaaa"); err != nil {
		t.Fatalf("Edit returned error: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries[0].Topics) != 1 {
		t.Errorf("expected 1 topic after deduplication, got %v", entries[0].Topics)
	}
	if entries[0].Topics[0] != "foo-topic" {
		t.Errorf("expected topic %q, got %q", "foo-topic", entries[0].Topics[0])
	}
}

// TestEdit_editor_with_arguments verifies that $EDITOR values containing flags
// (e.g. "sh -c true") are split correctly so that the first field is used as
// the program and the remaining fields are passed as arguments before the file
// path. Without the strings.Fields fix, exec.Command would try to look up the
// entire string "sh -c true" as a binary name and fail with "no such file".
func TestEdit_editor_with_arguments(t *testing.T) {
	store := newTestStore(t)
	entry := sampleEntry("aaaaaa")
	entry.Topics = []string{"GoLang"}
	entry.Text = "original text"
	entry.Related = []string{}
	appendEntries(t, store, entry)

	// "sh -c true" is a no-op editor that exits 0 without modifying the file,
	// but it exercises the multi-field EDITOR splitting path.
	t.Setenv("EDITOR", "sh -c true")

	if err := commands.Edit(store, "aaaaaa"); err != nil {
		t.Fatalf("Edit with multi-word EDITOR returned error: %v", err)
	}
}

// TestEdit_short_topic_rejected verifies that editing an entry to use a topic
// whose normalized form is fewer than 7 characters is rejected.
func TestEdit_short_topic_rejected(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store, sampleEntry("aaaaaa"))

	script := writeScript(t, `cat > "$1" <<'EOF'
# Topics: foo
# Related: 
# ---
some text
EOF`)
	t.Setenv("EDITOR", script)

	err := commands.Edit(store, "aaaaaa")
	if err == nil {
		t.Fatal("expected error for short topic, got nil")
	}
	if !strings.Contains(err.Error(), "foo") {
		t.Errorf("error should mention the offending topic, got: %v", err)
	}
}

// TestEdit_related_prefix_resolves_to_full_id verifies that a prefix typed in
// the Related line of the edit buffer is resolved to the full ID before being
// stored, matching the behaviour of gn write --related.
func TestEdit_related_prefix_resolves_to_full_id(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store, sampleEntry("aaaaaa"), sampleEntry("bbbbbb"))

	// Editor writes a prefix "bbb" in the Related line instead of the full ID.
	script := writeScript(t, `cat > "$1" <<'EOF'
# Topics: GoLang
# Related: bbb
# ---
updated text
EOF`)
	t.Setenv("EDITOR", script)

	if err := commands.Edit(store, "aaaaaa"); err != nil {
		t.Fatalf("Edit returned unexpected error: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	var updated storage.Entry
	for _, entry := range entries {
		if entry.ID == "aaaaaa" {
			updated = entry
		}
	}
	// The stored Related list must contain the full ID, not the prefix.
	if len(updated.Related) != 1 || updated.Related[0] != "bbbbbb" {
		t.Errorf("expected Related to contain full ID %q, got %v", "bbbbbb", updated.Related)
	}
}

// TestEdit_related_ambiguous_prefix_rejected verifies that an ambiguous prefix
// in the Related line is rejected with an error rather than silently stored.
func TestEdit_related_ambiguous_prefix_rejected(t *testing.T) {
	store := newTestStore(t)
	// Three entries: "cccccc" is the one being edited; "aaaaaa" and "aabbbb"
	// are the candidates. Prefix "aa" is ambiguous among the non-edited entries.
	appendEntries(t, store, sampleEntry("cccccc"), sampleEntry("aaaaaa"), sampleEntry("aabbbb"))

	// Editor writes prefix "aa" which matches both "aaaaaa" and "aabbbb".
	script := writeScript(t, `cat > "$1" <<'EOF'
# Topics: GoLang
# Related: aa
# ---
updated text
EOF`)
	t.Setenv("EDITOR", script)

	err := commands.Edit(store, "cccccc")
	if err == nil {
		t.Fatal("expected error for ambiguous related prefix, got nil")
	}
}

func TestEdit_editor_nonzero_exit_aborts(t *testing.T) {
	store := newTestStore(t)
	entry := sampleEntry("aaaaaa")
	entry.Text = "original"
	appendEntries(t, store, entry)

	// Editor that exits non-zero.
	t.Setenv("EDITOR", "false")

	err := commands.Edit(store, "aaaaaa")
	if err == nil {
		t.Fatal("expected error when editor exits non-zero, got nil")
	}

	// The original entry must be untouched.
	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if entries[0].Text != "original" {
		t.Errorf("entry was modified despite editor failure: %q", entries[0].Text)
	}
}
