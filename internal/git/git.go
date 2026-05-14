// Package git is a thin os/exec wrapper around the git CLI for the use cases
// VlessFilter cares about: configure local identity, commit-all, push using
// a PAT injected per-call via http.extraheader so secrets never land in
// .git/config or ~/.gitconfig.
package git

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"
)

// Configure sets local user.name and user.email if they're not already set.
// Idempotent. Required for unattended commits on fresh ephemeral VPSes.
func Configure(ctx context.Context, repoDir, name, email string) error {
	if err := setIfMissing(ctx, repoDir, "user.name", name); err != nil {
		return err
	}
	return setIfMissing(ctx, repoDir, "user.email", email)
}

func setIfMissing(ctx context.Context, repoDir, key, value string) error {
	out, err := runIn(ctx, repoDir, "config", "--get", key)
	if err == nil && strings.TrimSpace(out) != "" {
		return nil
	}
	if _, err := runIn(ctx, repoDir, "config", key, value); err != nil {
		return fmt.Errorf("git config %s: %w", key, err)
	}
	return nil
}

// CommitAll stages all changes and commits with msg.
// Returns committed=false (without error) when there's nothing to commit.
func CommitAll(ctx context.Context, repoDir, msg string) (committed bool, err error) {
	if _, err := runIn(ctx, repoDir, "add", "-A"); err != nil {
		return false, fmt.Errorf("git add: %w", err)
	}
	// `git diff --cached --quiet` exits 0 when there are no staged changes.
	if _, qerr := runIn(ctx, repoDir, "diff", "--cached", "--quiet"); qerr == nil {
		return false, nil
	}
	if _, err := runIn(ctx, repoDir, "commit", "-m", msg); err != nil {
		return false, fmt.Errorf("git commit: %w", err)
	}
	return true, nil
}

// Push pushes to origin/branch using `git -c http.extraheader=...` so the PAT
// never persists in .git/config or ~/.gitconfig and isn't visible to ps for
// the git child process (it's still in the parent's argv, so callers should
// not log it).
//
// Empty token = unauthenticated push (will only succeed for public repos
// with anonymous push, which is essentially nothing — kept for ergonomics).
func Push(ctx context.Context, repoDir, branch, token string) error {
	if branch == "" {
		branch = "main"
	}
	args := []string{}
	if token != "" {
		header := "Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte("oauth2:"+token))
		args = append(args, "-c", "http.extraheader="+header)
	}
	args = append(args, "push", "origin", branch)
	out, err := runIn(ctx, repoDir, args...)
	if err != nil {
		// Sanitize: never leak the auth header in error logs.
		sanitized := sanitize(out)
		return fmt.Errorf("git push: %w (output: %s)", err, sanitized)
	}
	slog.Debug("git push ok", "branch", branch)
	return nil
}

// runIn runs `git -C repoDir <args...>` and returns combined output.
func runIn(ctx context.Context, repoDir string, args ...string) (string, error) {
	if repoDir == "" {
		return "", errors.New("git: repoDir is required")
	}
	full := append([]string{"-C", repoDir}, args...)
	cmd := exec.CommandContext(ctx, "git", full...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// sanitize strips any "Authorization: Basic ..." substrings that might leak
// into a child-process error message.
//
// Implementation note: previous string-based loop was O(N²) and incorrectly
// re-matched the replacement when the replacement string itself contained
// the marker prefix. Regexp is correct and robust.
func sanitize(s string) string {
	return authRE.ReplaceAllString(s, "[REDACTED-AUTH]")
}

// authRE matches an Authorization header until the next whitespace/newline.
var authRE = regexp.MustCompile(`Authorization: Basic [^\s\n]+`)
