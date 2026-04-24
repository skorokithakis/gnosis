package termcolor_test

import (
	"testing"

	"github.com/skorokithakis/gnosis/internal/termcolor"
)

// SplitUniquePrefix tests cover the four cases called out in the ticket:
//   1. Unique at 1 character.
//   2. Unique only at full length (another ID shares a 5-char prefix).
//   3. Single-entry list (id is the only element).
//   4. id is not present in allIDs at all.

func TestSplitUniquePrefix_unique_at_first_char(t *testing.T) {
	prefix, rest := termcolor.SplitUniquePrefix("abcdef", []string{"abcdef", "ghjkmn", "pqrstu"})
	if prefix != "a" {
		t.Errorf("expected prefix %q, got %q", "a", prefix)
	}
	if rest != "bcdef" {
		t.Errorf("expected rest %q, got %q", "bcdef", rest)
	}
}

func TestSplitUniquePrefix_unique_at_full_length(t *testing.T) {
	// "abcdef" and "abcdez" share the first 5 characters; the prefix is only
	// unique at the full 6-character length.
	prefix, rest := termcolor.SplitUniquePrefix("abcdef", []string{"abcdef", "abcdez"})
	if prefix != "abcdef" {
		t.Errorf("expected prefix %q, got %q", "abcdef", prefix)
	}
	if rest != "" {
		t.Errorf("expected empty rest, got %q", rest)
	}
}

func TestSplitUniquePrefix_single_entry_list(t *testing.T) {
	// With only one ID in the list the prefix should be the first character.
	prefix, rest := termcolor.SplitUniquePrefix("abcdef", []string{"abcdef"})
	if prefix != "a" {
		t.Errorf("expected prefix %q, got %q", "a", prefix)
	}
	if rest != "bcdef" {
		t.Errorf("expected rest %q, got %q", "bcdef", rest)
	}
}

func TestSplitUniquePrefix_id_not_in_list(t *testing.T) {
	// id is treated as present even when absent from allIDs, so the result
	// should be the same as the single-entry case.
	prefix, rest := termcolor.SplitUniquePrefix("abcdef", []string{"ghjkmn"})
	if prefix != "a" {
		t.Errorf("expected prefix %q, got %q", "a", prefix)
	}
	if rest != "bcdef" {
		t.Errorf("expected rest %q, got %q", "bcdef", rest)
	}
}
