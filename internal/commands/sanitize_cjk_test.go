package commands

import "testing"

// CJK queries must survive sanitization (they used to be erased entirely by
// the non-alphanumeric replacement) and must come out segmented, because the
// FTS index stores jieba-segmented CJK (see internal/index.SegmentCJK).
func TestSanitizeQuery_cjk(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bare CJK word run becomes a segmented phrase",
			input:    "状态流转",
			expected: `"状态 流转"`,
		},
		{
			name:     "single-token CJK stays a bareword",
			input:    "失败",
			expected: "失败",
		},
		{
			name:     "CJK survives alongside operators",
			input:    "tab OR 状态流转",
			expected: `tab OR "状态 流转"`,
		},
		{
			name:     "quoted CJK phrase is segmented in place",
			input:    `"状态流转"`,
			expected: `"状态 流转"`,
		},
		{
			name:     "mixed latin and CJK",
			input:    "update_result 失败",
			expected: "update result 失败",
		},
	}
	for _, c := range cases {
		if got := sanitizeQuery(c.input); got != c.expected {
			t.Errorf("sanitizeQuery(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}
