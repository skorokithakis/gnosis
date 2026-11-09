package index

import (
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/go-ego/gse"
)

// cjkRunPattern matches one run of CJK characters (Han, Hiragana, Katakana,
// Hangul). Only these runs need segmentation: Latin text is already
// whitespace-delimited and passes through untouched.
var cjkRunPattern = regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}]+`)

// latinThenCJKPattern and cjkThenLatinPattern match a digit-or-letter glued
// directly to a CJK character (either direction). gse does not split these
// boundaries — a token like "doing状态" would otherwise be indexed whole,
// making the Chinese part unsearchable — so the boundary is split before
// segmentation. Applied as two passes; they converge on mixed strings like
// "重试5次" and "abc中def".
var (
	latinThenCJKPattern = regexp.MustCompile(`([A-Za-z0-9])([\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}])`)
	cjkThenLatinPattern = regexp.MustCompile(`([\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}])([A-Za-z0-9])`)
)

var (
	segmenterOnce sync.Once
	segmenter     *gse.Segmenter
)

// getSegmenter lazily initialises the gse (jieba) segmenter with the embedded
// Chinese dictionary. Loading the dictionary costs noticeable time, so it is
// deferred until the first CJK run is actually seen; pure-Latin commands pay
// nothing. If the dictionary cannot be loaded, segmenter stays nil and
// SegmentCJK degrades to the upstream behaviour (unsegmented CJK).
func getSegmenter() *gse.Segmenter {
	segmenterOnce.Do(func() {
		// gse logs dictionary-loading details through the standard log
		// package on every load; gn is a short-lived CLI, so those lines
		// would precede every CJK-touching command's output. Silence them
		// for the duration of the load.
		log.SetOutput(io.Discard)
		seg, err := gse.New("zh")
		log.SetOutput(os.Stderr)
		if err == nil {
			segmenter = &seg
		}
	})
	return segmenter
}

// SegmentCJK segments every CJK run in s into space-separated words using the
// jieba algorithm (via gse), leaving all other text byte-for-byte unchanged.
// Applied to entry text before it enters the FTS5 index and to search terms
// before they are matched, so that unicode61 — which treats a whole run of
// CJK characters as a single token — can index and match Chinese words.
//
// The transformation only ever inserts spaces between two CJK characters, so
// snippet() output can be restored to its original spacing by stripping
// spaces between CJK characters (see RestoreSnippetSpaces).
func SegmentCJK(s string) string {
	if !cjkRunPattern.MatchString(s) {
		return s
	}
	seg := getSegmenter()
	if seg == nil {
		return s
	}
	s = latinThenCJKPattern.ReplaceAllString(s, "$1 $2")
	s = cjkThenLatinPattern.ReplaceAllString(s, "$1 $2")
	return cjkRunPattern.ReplaceAllStringFunc(s, func(run string) string {
		tokens := seg.Cut(run, true)
		if len(tokens) == 0 {
			return run
		}
		return strings.Join(tokens, " ")
	})
}

// HasCJK reports whether s contains any CJK characters.
func HasCJK(s string) bool {
	return cjkRunPattern.MatchString(s)
}

// cjkSnippetSpacePattern matches a whitespace run between two CJK characters.
// SegmentCJK only ever inserts spaces at such positions, so stripping them
// restores the original text spacing inside snippets.
var cjkSnippetSpacePattern = regexp.MustCompile(`([\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}])\s+([\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}])`)

// RestoreSnippetSpaces reverses the spacing SegmentCJK injected into the
// text stored in the FTS table, so snippets read like the original entry
// text. Applied repeatedly to bridge tokens of two or more characters.
func RestoreSnippetSpaces(s string) string {
	prev := s
	for {
		s = cjkSnippetSpacePattern.ReplaceAllString(s, "$1$2")
		if s == prev {
			return s
		}
		prev = s
	}
}
