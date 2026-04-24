// Package termcolor provides semantic color helpers for gnosis terminal output.
// It wraps github.com/fatih/color, which automatically disables color when
// stdout is not a TTY and respects the NO_COLOR environment variable, so
// callers need no TTY-detection logic of their own.
package termcolor

import (
	"strings"

	"github.com/fatih/color"
)

var (
	boldMagenta = color.New(color.FgMagenta, color.Bold)
	faint       = color.New(color.Faint)
	cyan        = color.New(color.FgCyan)
	yellow      = color.New(color.FgYellow)
	bold        = color.New(color.Bold)
	boldRed     = color.New(color.FgRed, color.Bold)
)

// HighlightMatches replaces every (start, end) sentinel pair in s with a
// rendering of the enclosed text. When color output is enabled (TTY, no
// NO_COLOR), matches are rendered in bold red — grep's convention. When color
// is disabled, sentinels fall back to literal '[' and ']' brackets so the
// match boundaries remain visible in pipes, files, and dumb terminals.
//
// FTS5 emits balanced start/end pairs, so independently replacing each
// sentinel with its corresponding ANSI sequence yields correct output without
// needing to parse spans.
func HighlightMatches(s, start, end string) string {
	if color.NoColor {
		s = strings.ReplaceAll(s, start, "[")
		s = strings.ReplaceAll(s, end, "]")
		return s
	}
	// Derive the open/close ANSI sequences from boldRed by sprinting a known
	// sentinel character and splitting on it. This keeps the actual escape
	// codes owned by fatih/color rather than hardcoding them here.
	wrapped := boldRed.Sprint("\x01")
	parts := strings.SplitN(wrapped, "\x01", 2)
	if len(parts) != 2 {
		// boldRed produced no escape codes (color unexpectedly disabled);
		// fall back to brackets so output is still readable.
		s = strings.ReplaceAll(s, start, "[")
		s = strings.ReplaceAll(s, end, "]")
		return s
	}
	openSeq, closeSeq := parts[0], parts[1]
	s = strings.ReplaceAll(s, start, openSeq)
	s = strings.ReplaceAll(s, end, closeSeq)
	return s
}

// SplitUniquePrefix returns the shortest prefix of id that does not match the
// start of any other ID in allIDs, plus the remaining characters of id.
//
// The uniqueness rule: a prefix p is unique if no element of allIDs other than
// id itself starts with p. id is always treated as present in allIDs regardless
// of whether it actually appears there, so a single-element list (or a list
// that omits id) still yields a prefix of length 1.
//
// If every other ID in allIDs shares a prefix of length len(id)-1 with id (an
// extreme collision), the full id is returned as the prefix and rest is empty.
func SplitUniquePrefix(id string, allIDs []string) (prefix, rest string) {
	// Build the set of other IDs so we can test prefix uniqueness without
	// repeatedly scanning the full slice.
	others := make([]string, 0, len(allIDs))
	for _, candidate := range allIDs {
		if candidate != id {
			others = append(others, candidate)
		}
	}

	// Walk from length 1 up to len(id), stopping at the first length where no
	// other ID starts with that prefix. We always try at least length 1 so that
	// a single-entry list still produces a non-empty prefix.
	for length := 1; length <= len(id); length++ {
		p := id[:length]
		unique := true
		for _, other := range others {
			if strings.HasPrefix(other, p) {
				unique = false
				break
			}
		}
		if unique {
			return p, id[length:]
		}
	}

	// All prefixes up to the full ID are shared with at least one other entry.
	// Return the full ID as the prefix with an empty remainder.
	return id, ""
}

// UniqueID returns id with its unique prefix rendered in bold magenta and the
// remainder rendered dim/faint. allIDs is the full set of IDs in the store and
// is forwarded to SplitUniquePrefix to determine the boundary.
func UniqueID(id string, allIDs []string) string {
	prefix, rest := SplitUniquePrefix(id, allIDs)
	return boldMagenta.Sprint(prefix) + faint.Sprint(rest)
}

// Topic returns s rendered in cyan, the conventional color for topic names.
func Topic(s string) string {
	return cyan.Sprint(s)
}

// Date returns s rendered in yellow, the conventional color for dates.
func Date(s string) string {
	return yellow.Sprint(s)
}

// Dim returns s rendered faint/dim, used for secondary labels such as
// "Related:" or "created".
func Dim(s string) string {
	return faint.Sprint(s)
}

// Bold returns s rendered bold in the default foreground color, used for
// counts and other emphasis.
func Bold(s string) string {
	return bold.Sprint(s)
}
