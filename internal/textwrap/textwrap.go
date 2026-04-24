// Package textwrap provides a word-wrap helper for terminal output.
package textwrap

import (
	"strings"
	"unicode/utf8"
)

// Wrap word-wraps s so that no line exceeds width runes. When width is zero or
// negative the string is returned unchanged, which is the correct behaviour for
// non-TTY output where the consumer controls line length.
//
// Each newline-delimited paragraph in s is wrapped independently so that blank
// separator lines and intentional hard breaks are preserved. Within a paragraph
// words are placed greedily: as many words as fit within width are placed on
// one line, then a new line is started. A single token that is longer than
// width is placed on its own line without being broken mid-token, because
// mid-token breaks would corrupt URLs, identifiers, and similar content.
//
// Width is measured in runes, not bytes, so multi-byte Unicode characters each
// count as one column.
func Wrap(s string, width int) string {
	if width <= 0 {
		return s
	}

	// Split on existing newlines and wrap each line independently so that
	// paragraph structure and blank lines survive the transformation.
	lines := strings.Split(s, "\n")
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, wrapLine(line, width))
	}
	return strings.Join(wrapped, "\n")
}

// wrapLine wraps a single line (containing no newlines) to width runes.
func wrapLine(line string, width int) string {
	words := strings.Fields(line)
	if len(words) == 0 {
		// Preserve blank lines and lines that are only whitespace.
		return line
	}

	var builder strings.Builder
	// currentWidth tracks the rune count of the current output line.
	currentWidth := 0

	for index, word := range words {
		wordWidth := utf8.RuneCountInString(word)

		if index == 0 {
			// First word always starts the line without a leading space.
			builder.WriteString(word)
			currentWidth = wordWidth
			continue
		}

		// A word fits on the current line when adding a space and the word
		// keeps the total at or below width.
		if currentWidth+1+wordWidth <= width {
			builder.WriteByte(' ')
			builder.WriteString(word)
			currentWidth += 1 + wordWidth
		} else {
			// The word does not fit; start a new line. Oversized tokens
			// (wordWidth > width) are placed on their own line rather than
			// being broken mid-token.
			builder.WriteByte('\n')
			builder.WriteString(word)
			currentWidth = wordWidth
		}
	}

	return builder.String()
}
