package commands

import (
	"errors"
	"testing"
)

// TestFormatGitIdentity covers the formatting rules: "Name <email>" when both
// are set, just the name when email is missing, and "" when there is no name.
func TestFormatGitIdentity(t *testing.T) {
	cases := []struct {
		name, email, want string
	}{
		{"Alice", "alice@example.com", "Alice <alice@example.com>"},
		{"Alice", "", "Alice"},
		{"Alice", "   ", "Alice"},                           // whitespace-only email treated as unset
		{"  Bob  ", "b@example.com", "Bob <b@example.com>"}, // name trimmed
		{"", "someone@example.com", ""},                     // no name -> no identity
		{"", "", ""},
		{"   ", "", ""}, // whitespace-only name -> no identity
	}
	for _, tc := range cases {
		if got := formatGitIdentity(tc.name, tc.email); got != tc.want {
			t.Errorf("formatGitIdentity(%q, %q) = %q, want %q", tc.name, tc.email, got, tc.want)
		}
	}
}

// TestResolveAuthor exercises all three resolution paths using an injected
// lookup so no real git process is spawned.
func TestResolveAuthor(t *testing.T) {
	// A fake lookup backed by a map of key -> value.
	lookupFrom := func(values map[string]string) gitConfigLookup {
		return func(key string) (string, error) {
			if v, ok := values[key]; ok {
				return v, nil
			}
			return "", errors.New("key not set")
		}
	}

	cases := []struct {
		name      string
		authorArg string
		lookup    gitConfigLookup
		want      string
		wantOK    bool
	}{
		{
			name:      "explicit author wins",
			authorArg: "Carol <carol@example.com>",
			lookup:    lookupFrom(map[string]string{"user.name": "Ignored", "user.email": "ignored@example.com"}),
			want:      "Carol <carol@example.com>",
			wantOK:    true,
		},
		{
			name:      "git identity with name and email",
			authorArg: "",
			lookup:    lookupFrom(map[string]string{"user.name": "Dave", "user.email": "dave@example.com"}),
			want:      "Dave <dave@example.com>",
			wantOK:    true,
		},
		{
			name:      "git identity with name only",
			authorArg: "",
			lookup:    lookupFrom(map[string]string{"user.name": "Eve"}),
			want:      "Eve",
			wantOK:    true,
		},
		{
			name:      "fallback when no git identity",
			authorArg: "",
			lookup:    lookupFrom(map[string]string{}),
			want:      "Unknown <unknown@gnosis>",
			wantOK:    false,
		},
		{
			name:      "fallback when lookup errors",
			authorArg: "",
			lookup:    func(key string) (string, error) { return "", errors.New("git not installed") },
			want:      "Unknown <unknown@gnosis>",
			wantOK:    false,
		},
		{
			name:      "explicit author beats broken git",
			authorArg: "Frank",
			lookup:    func(key string) (string, error) { return "", errors.New("boom") },
			want:      "Frank",
			wantOK:    true,
		},
		{
			name:      "whitespace-only author arg falls through to git",
			authorArg: "   ",
			lookup:    lookupFrom(map[string]string{"user.name": "Grace", "user.email": "g@example.com"}),
			want:      "Grace <g@example.com>",
			wantOK:    true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := resolveAuthor(tc.authorArg, tc.lookup)
			if got != tc.want {
				t.Errorf("resolveAuthor author = %q, want %q", got, tc.want)
			}
			if ok != tc.wantOK {
				t.Errorf("resolveAuthor ok = %v, want %v", ok, tc.wantOK)
			}
		})
	}
}
