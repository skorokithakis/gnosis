package commands

import (
	"testing"
)

func TestSanitizeQuery(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "OR operator with hyphenated terms",
			input:    "install warning OR latest-release OR update track",
			expected: "install warning OR latest release OR update track",
		},
		{
			name:     "quoted phrase with hyphen and OR",
			input:    `"foo-bar" OR baz-qux`,
			expected: `"foo-bar" OR baz qux`,
		},
		{
			name:     "bare hyphenated query still works (regression)",
			input:    "keymaster-token-auth",
			expected: "keymaster token auth",
		},
		{
			name:     "clean operator query unchanged",
			input:    "foo OR bar",
			expected: "foo OR bar",
		},
		{
			name:     "unmatched quote — consumed verbatim to end-of-string",
			input:    `foo "bar-baz`,
			expected: `foo "bar-baz`,
		},
		{
			name:     "lowercase 'or' is a bareword not an operator",
			input:    "foo or bar-baz",
			expected: "foo or bar baz",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeQuery(tc.input)
			if got != tc.expected {
				t.Errorf("sanitizeQuery(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
