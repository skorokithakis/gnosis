package textwrap_test

import (
	"testing"

	"github.com/skorokithakis/gnosis/internal/textwrap"
)

// TestWrap_no_op_when_width_zero checks that a zero width returns the input
// unchanged, which is the intended behaviour for non-TTY callers.
func TestWrap_no_op_when_width_zero(t *testing.T) {
	input := "this is a long line that would normally be wrapped"
	result := textwrap.Wrap(input, 0)
	if result != input {
		t.Errorf("expected input unchanged, got %q", result)
	}
}

// TestWrap_no_op_when_width_negative checks that a negative width also returns
// the input unchanged.
func TestWrap_no_op_when_width_negative(t *testing.T) {
	input := "another long line"
	result := textwrap.Wrap(input, -1)
	if result != input {
		t.Errorf("expected input unchanged, got %q", result)
	}
}

// TestWrap_basic_wrap checks that a line longer than width is broken at a word
// boundary and that the resulting lines do not exceed width runes.
func TestWrap_basic_wrap(t *testing.T) {
	result := textwrap.Wrap("one two three four five", 11)
	// "one two" = 7, "three four" = 10, "five" = 4 — all within 11.
	expected := "one two\nthree four\nfive"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestWrap_no_wrap_needed checks that a line already within width is returned
// without modification.
func TestWrap_no_wrap_needed(t *testing.T) {
	input := "short line"
	result := textwrap.Wrap(input, 80)
	if result != input {
		t.Errorf("expected %q, got %q", input, result)
	}
}

// TestWrap_preserves_blank_lines checks that blank lines in the input survive
// wrapping so that paragraph structure is maintained.
func TestWrap_preserves_blank_lines(t *testing.T) {
	input := "first paragraph\n\nsecond paragraph"
	result := textwrap.Wrap(input, 80)
	if result != input {
		t.Errorf("expected %q, got %q", input, result)
	}
}

// TestWrap_preserves_blank_lines_with_wrapping checks that blank lines are
// preserved even when the surrounding paragraphs themselves need wrapping.
func TestWrap_preserves_blank_lines_with_wrapping(t *testing.T) {
	input := "one two three\n\nfour five six"
	result := textwrap.Wrap(input, 7)
	// "one two" fits (7), "three" starts a new line; blank line preserved;
	// "four" fits (4), "five" fits (4+1+4=9 > 7 so new line), "six" fits.
	expected := "one two\nthree\n\nfour\nfive\nsix"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestWrap_oversized_token_overflow checks that a single token longer than
// width is placed on its own line without being broken mid-token.
func TestWrap_oversized_token_overflow(t *testing.T) {
	result := textwrap.Wrap("see https://example.com/very/long/url for details", 20)
	// "see" fits on line 1; the URL (38 runes) overflows onto its own line;
	// "for details" fits on the next line.
	expected := "see\nhttps://example.com/very/long/url\nfor details"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestWrap_multi_paragraph checks that multiple paragraphs separated by blank
// lines are each wrapped independently.
func TestWrap_multi_paragraph(t *testing.T) {
	input := "alpha beta gamma\n\ndelta epsilon zeta"
	result := textwrap.Wrap(input, 11)
	// "alpha beta" = 10, "gamma" = 5; "delta" = 5, "epsilon" = 7, "zeta" = 4.
	expected := "alpha beta\ngamma\n\ndelta\nepsilon\nzeta"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestWrap_unicode_counted_by_runes checks that multi-byte Unicode characters
// each count as one column, not as their byte length.
func TestWrap_unicode_counted_by_runes(t *testing.T) {
	// Each of these words is 3 runes. With width=7 two words fit ("αβγ δεζ" = 7).
	result := textwrap.Wrap("αβγ δεζ ηθι", 7)
	expected := "αβγ δεζ\nηθι"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestWrap_single_word_exactly_at_width checks the boundary condition where a
// single word is exactly as wide as the limit.
func TestWrap_single_word_exactly_at_width(t *testing.T) {
	result := textwrap.Wrap("hello world", 5)
	expected := "hello\nworld"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
