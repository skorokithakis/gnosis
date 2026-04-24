package commands_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/skorokithakis/gnosis/internal/commands"
)

func TestRemove_removes_existing_entries(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store,
		sampleEntry("aaaaaa"),
		sampleEntry("bbbbbb"),
		sampleEntry("cccccc"),
	)

	var stdout, stderr bytes.Buffer
	if err := commands.Remove(store, []string{"aaaaaa", "cccccc"}, &stdout, &stderr); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	remaining, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != "bbbbbb" {
		t.Errorf("expected only bbbbbb to survive, got %v", remaining)
	}

	// Both removed IDs should appear on stdout.
	output := stdout.String()
	if !strings.Contains(output, "aaaaaa") || !strings.Contains(output, "cccccc") {
		t.Errorf("expected removed IDs on stdout, got %q", output)
	}

	if stderr.Len() != 0 {
		t.Errorf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRemove_nonexistent_id_errors_without_modifying_file(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store, sampleEntry("aaaaaa"), sampleEntry("bbbbbb"))

	var stdout, stderr bytes.Buffer
	err := commands.Remove(store, []string{"aaaaaa", "zzzzzz"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
	if !strings.Contains(err.Error(), "zzzzzz") {
		t.Errorf("error should mention the missing ID, got %q", err.Error())
	}

	// The file must be untouched: both original entries must still be present.
	remaining, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(remaining) != 2 {
		t.Errorf("expected 2 entries after failed remove, got %d", len(remaining))
	}

	if stdout.Len() != 0 {
		t.Errorf("expected no stdout on error, got %q", stdout.String())
	}
}

func TestRemove_dangling_related_refs_produce_warnings(t *testing.T) {
	store := newTestStore(t)

	entryA := sampleEntry("aaaaaa")
	entryB := sampleEntry("bbbbbb")
	// bbbbbb references aaaaaa in its Related list.
	entryB.Related = []string{"aaaaaa"}
	appendEntries(t, store, entryA, entryB)

	var stdout, stderr bytes.Buffer
	if err := commands.Remove(store, []string{"aaaaaa"}, &stdout, &stderr); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	// The surviving entry (bbbbbb) must not be modified.
	remaining, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != "bbbbbb" {
		t.Errorf("expected only bbbbbb to survive, got %v", remaining)
	}
	if len(remaining[0].Related) != 1 || remaining[0].Related[0] != "aaaaaa" {
		t.Errorf("surviving entry Related should be unchanged, got %v", remaining[0].Related)
	}

	// A warning about the dangling reference must appear on stderr.
	warning := stderr.String()
	if !strings.Contains(warning, "bbbbbb") || !strings.Contains(warning, "aaaaaa") {
		t.Errorf("expected dangling-ref warning on stderr, got %q", warning)
	}
}

func TestRemove_no_args_returns_usage_error(t *testing.T) {
	store := newTestStore(t)
	var stdout, stderr bytes.Buffer
	err := commands.Remove(store, []string{}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing arguments, got nil")
	}
}

func TestRemove_all_entries(t *testing.T) {
	store := newTestStore(t)
	appendEntries(t, store, sampleEntry("aaaaaa"), sampleEntry("bbbbbb"))

	var stdout, stderr bytes.Buffer
	if err := commands.Remove(store, []string{"aaaaaa", "bbbbbb"}, &stdout, &stderr); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	remaining, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("expected empty store after removing all entries, got %v", remaining)
	}
}
