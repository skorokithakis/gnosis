package commands

import (
	"fmt"
	"os/exec"
	"strings"
)

// gitConfigLookup returns the configured value for a git config key (e.g.
// "user.name"). It is a type rather than a direct call so tests can inject a
// fake lookup and exercise every author-resolution path without spawning git.
type gitConfigLookup func(key string) (string, error)

// defaultGitConfigLookup shells out to "git config <key>" and returns the
// trimmed value. Any failure (git missing, key unset, non-zero exit) yields an
// empty string and the error, which resolveAuthor treats as "no identity".
func defaultGitConfigLookup(key string) (string, error) {
	out, err := exec.Command("git", "config", key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// formatGitIdentity combines a git user.name and user.email into the stored
// author string: "Name <email>" when both are present, just "Name" when the
// email is unset, and "" when there is no name at all (meaning no usable git
// identity exists).
func formatGitIdentity(name, email string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return name
	}
	return fmt.Sprintf("%s <%s>", name, email)
}

// resolveAuthor determines the author for a new entry. Resolution order:
//  1. authorArg, if non-empty (the user was explicit);
//  2. git identity from lookup, formatted as "Name <email>" (or just the name);
//  3. the literal fallback "Unknown <unknown@gnosis>".
//
// The ok result is false only on the fallback path so the caller can print a
// warning. The write is never blocked by missing identity.
func resolveAuthor(authorArg string, lookup gitConfigLookup) (author string, ok bool) {
	if authorArg = strings.TrimSpace(authorArg); authorArg != "" {
		return authorArg, true
	}
	name, _ := lookup("user.name")
	email, _ := lookup("user.email")
	if identity := formatGitIdentity(name, email); identity != "" {
		return identity, true
	}
	return "Unknown <unknown@gnosis>", false
}
