package commands_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/skorokithakis/gnosis/internal/commands"
)

func TestWrite_basic(t *testing.T) {
	store := newTestStore(t)

	if err := commands.Write(store, []string{"foo,Bar", "hello world"}, io.Discard); err != nil {
		t.Fatalf("Write: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if len(entry.ID) != 6 {
		t.Errorf("expected 6-char ID, got %q", entry.ID)
	}
	if len(entry.Topics) != 2 {
		t.Errorf("expected 2 topics, got %v", entry.Topics)
	}
	if entry.Topics[0] != "foo" || entry.Topics[1] != "bar" {
		t.Errorf("unexpected topics: %v", entry.Topics)
	}
	if entry.Text != "hello world" {
		t.Errorf("unexpected text: %q", entry.Text)
	}
}

// TestWrite_output_contains_id verifies that Write prints the new entry ID to
// the provided writer.
func TestWrite_output_contains_id(t *testing.T) {
	store := newTestStore(t)

	var buf strings.Builder
	if err := commands.Write(store, []string{"topic", "some text"}, &buf); err != nil {
		t.Fatalf("Write: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if len(output) != 6 {
		t.Errorf("expected 6-char ID in output, got %q", output)
	}
}

// TestWrite_topic_deduplication verifies that topics with the same normalized
// form are collapsed into a single normalized topic.
func TestWrite_topic_deduplication(t *testing.T) {
	store := newTestStore(t)

	// "foo", "Foo", and "FOO" all normalize to "foo"; only one should be stored.
	if err := commands.Write(store, []string{"foo,Foo,FOO", "text"}, io.Discard); err != nil {
		t.Fatalf("Write: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries[0].Topics) != 1 {
		t.Errorf("expected 1 topic after deduplication, got %v", entries[0].Topics)
	}
	if entries[0].Topics[0] != "foo" {
		t.Errorf("expected normalized form %q, got %q", "foo", entries[0].Topics[0])
	}
}

// TestWrite_topic_deduplication_camel verifies that CamelCase and snake_case
// forms of the same topic collapse to the normalized form.
func TestWrite_topic_deduplication_camel(t *testing.T) {
	store := newTestStore(t)

	// "KeymasterTokenAuth" and "keymaster_token_auth" both normalize to
	// "keymaster-token-auth"; only one should be stored.
	if err := commands.Write(store, []string{"KeymasterTokenAuth,keymaster_token_auth", "text"}, io.Discard); err != nil {
		t.Fatalf("Write: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries[0].Topics) != 1 {
		t.Errorf("expected 1 topic after deduplication, got %v", entries[0].Topics)
	}
	if entries[0].Topics[0] != "keymaster-token-auth" {
		t.Errorf("expected normalized form %q, got %q", "keymaster-token-auth", entries[0].Topics[0])
	}
}

// TestWrite_related_validation verifies that --related IDs are checked against
// existing entries and an error is returned for unknown IDs.
func TestWrite_related_validation(t *testing.T) {
	store := newTestStore(t)

	// Write a first entry to get a real ID.
	if err := commands.Write(store, []string{"topic", "first entry"}, io.Discard); err != nil {
		t.Fatalf("Write first entry: %v", err)
	}
	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	existingID := entries[0].ID

	// A valid --related reference should succeed.
	if err := commands.Write(store, []string{"topic", "second entry", "--related", existingID}, io.Discard); err != nil {
		t.Fatalf("Write with valid --related: %v", err)
	}

	// An unknown --related ID must produce an error.
	err = commands.Write(store, []string{"topic", "third entry", "--related", "zzzzzz"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for unknown --related ID, got nil")
	}
	if !strings.Contains(err.Error(), "zzzzzz") {
		t.Errorf("error message should mention the bad ID, got: %v", err)
	}
}

// TestWrite_stdin_fallback verifies that text is read from stdin when it is not
// provided as a command-line argument and stdin is not a TTY. We simulate this
// by replacing os.Stdin with a pipe.
func TestWrite_stdin_fallback(t *testing.T) {
	store := newTestStore(t)

	// Replace os.Stdin with a pipe so isatty returns false.
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	original := os.Stdin
	os.Stdin = reader
	defer func() { os.Stdin = original }()

	if _, err := writer.WriteString("piped text\n"); err != nil {
		t.Fatalf("writing to pipe: %v", err)
	}
	writer.Close()

	// No text argument — should fall back to stdin.
	if err := commands.Write(store, []string{"topic"}, io.Discard); err != nil {
		t.Fatalf("Write with stdin fallback: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if entries[0].Text != "piped text" {
		t.Errorf("expected %q, got %q", "piped text", entries[0].Text)
	}
}

// TestWrite_empty_text_error verifies that an empty text argument is rejected.
func TestWrite_empty_text_error(t *testing.T) {
	store := newTestStore(t)

	err := commands.Write(store, []string{"topic", "   "}, io.Discard)
	if err == nil {
		t.Fatal("expected error for empty text, got nil")
	}
}

// TestWrite_empty_topic_error verifies that an empty topic (after splitting and
// trimming) is rejected.
func TestWrite_empty_topic_error(t *testing.T) {
	store := newTestStore(t)

	err := commands.Write(store, []string{"foo,,bar", "text"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for empty topic, got nil")
	}
}

// TestWrite_no_topics_error verifies that omitting the topics argument is rejected.
func TestWrite_no_topics_error(t *testing.T) {
	store := newTestStore(t)

	err := commands.Write(store, []string{}, io.Discard)
	if err == nil {
		t.Fatal("expected error when no topics provided, got nil")
	}
}

// TestWrite_topic_normalizes_to_empty_error verifies that a topic whose
// normalized form is empty (e.g. "---") is rejected rather than stored as an
// empty string.
func TestWrite_topic_normalizes_to_empty_error(t *testing.T) {
	store := newTestStore(t)

	for _, badTopic := range []string{"---", "-", "___", "--"} {
		err := commands.Write(store, []string{badTopic, "text"}, io.Discard)
		if err == nil {
			t.Errorf("expected error for topic %q that normalizes to empty, got nil", badTopic)
		}
	}

	// Nothing should have been written.
	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after all-rejected writes, got %d", len(entries))
	}
}
