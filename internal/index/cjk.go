package index

import (
	"strings"
	"unicode"
)

// katakanaLongVowelMark is U+30FC, the Katakana-Hiragana prolonged sound mark.
// Unicode classes it as script Common, so unicode.Is(unicode.Katakana) misses
// it, but it occurs inside Katakana words and must be split out like the
// characters around it.
const katakanaLongVowelMark = '\u30FC'

// IsCJK reports whether r belongs to a script written without spaces between
// words: Han, Hiragana, and Katakana, plus the Katakana long-vowel mark. Hangul
// is deliberately excluded because Korean separates words with spaces, so the
// unicode61 tokenizer already produces useful tokens for it.
func IsCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		r == katakanaLongVowelMark
}

// SplitCJK returns text with every CJK character separated from its neighbours
// by a single space, so that FTS5's unicode61 tokenizer emits each CJK
// character as its own token. Text without CJK is returned unchanged, and
// existing whitespace is never doubled.
//
// No space is inserted next to punctuation or symbols, because the tokenizer
// already treats those as separators; that keeps snippets free of stray spaces
// before "。" or "，" in ordinary Chinese text. For the same reason a space is
// only needed when at least one of the two neighbours is CJK.
//
// This is necessary because unicode61 treats an unbroken run of CJK characters
// as a single token, which makes a shorter phrase unmatchable: searching for
// 状态 would never find 非法状态流转.
func SplitCJK(text string) string {
	if !containsCJK(text) {
		return text
	}

	var builder strings.Builder
	builder.Grow(len(text) + len(text)/2)
	// Track the previous rune so each pair of neighbours is considered once:
	// the separator decision depends on both sides of the pair.
	var previous rune
	havePrevious := false
	for _, r := range text {
		if havePrevious && needsCJKSeparator(previous, r) {
			builder.WriteByte(' ')
		}
		builder.WriteRune(r)
		previous = r
		havePrevious = true
	}

	return builder.String()
}

// needsCJKSeparator reports whether SplitCJK should insert a space between two
// neighbouring runes. A pair needs separating only when at least one side is
// CJK and neither side is already a separator on its own.
func needsCJKSeparator(previous, current rune) bool {
	if !IsCJK(previous) && !IsCJK(current) {
		return false
	}
	return !isTokenSeparator(previous) && !isTokenSeparator(current)
}

// isTokenSeparator reports whether r already splits tokens for unicode61.
// Whitespace does so by definition, and punctuation and symbols such as "。",
// "，" or "-" do too.
func isTokenSeparator(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)
}

// containsCJK reports whether text contains any rune for which IsCJK is true.
func containsCJK(text string) bool {
	for _, r := range text {
		if IsCJK(r) {
			return true
		}
	}
	return false
}

// joinCJKSnippet removes the spaces that SplitCJK inserted between CJK
// characters, so that a snippet reads as the original text. snippet() places
// the match markers around the matched tokens, which can put a marker between a
// space and its neighbouring character, so the checks look through the markers
// instead of treating them as boundaries. Spaces between a CJK character and
// non-CJK text stay, because those may be genuine word separators.
func joinCJKSnippet(snippet string) string {
	runes := []rune(snippet)
	var builder strings.Builder
	builder.Grow(len(snippet))
	for i, r := range runes {
		if r == ' ' && spaceJoinsCJK(runes, i) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

// spaceJoinsCJK reports whether the space at position i in runes has a CJK
// character on both sides, ignoring any match markers in between.
func spaceJoinsCJK(runes []rune, i int) bool {
	left := i - 1
	for left >= 0 && isMatchMarker(runes[left]) {
		left--
	}
	if left < 0 || !IsCJK(runes[left]) {
		return false
	}

	right := i + 1
	for right < len(runes) && isMatchMarker(runes[right]) {
		right++
	}
	return right < len(runes) && IsCJK(runes[right])
}

// isMatchMarker reports whether r is one of the MatchStart/MatchEnd sentinels.
func isMatchMarker(r rune) bool {
	return strings.ContainsRune(MatchStart, r) || strings.ContainsRune(MatchEnd, r)
}
