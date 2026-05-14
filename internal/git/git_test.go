package git

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitize_RedactsAuthHeader(t *testing.T) {
	in := "remote: error: foo\nfatal: bar Authorization: Basic SECRETTOKEN==\nbaz\n"
	got := sanitize(in)
	if strings.Contains(got, "SECRETTOKEN") {
		t.Errorf("sanitize leaked secret: %s", got)
	}
	if !strings.Contains(got, "[REDACTED-AUTH]") {
		t.Errorf("expected [REDACTED-AUTH] marker in output, got: %s", got)
	}
}

func TestSanitize_NoChangesWhenNoMatch(t *testing.T) {
	in := "nothing sensitive here"
	if got := sanitize(in); got != in {
		t.Errorf("sanitize altered safe input: %q", got)
	}
}

func TestEncodeAuthHeader(t *testing.T) {
	// The pattern we use: "Authorization: Basic " + base64("oauth2:TOKEN")
	token := "ghp_TESTTOKEN123"
	header := "Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte("oauth2:"+token))
	// Decode and confirm round-trip
	prefix := "Authorization: Basic "
	encoded := strings.TrimPrefix(header, prefix)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(decoded) != "oauth2:"+token {
		t.Errorf("round-trip mismatch: got %q", decoded)
	}
}

func TestCommitAll_NoChanges(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := Configure(ctx, dir, "Test User", "test@example.com"); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	// Initial commit so the branch exists.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	committed, err := CommitAll(ctx, dir, "init")
	if err != nil {
		t.Fatalf("first CommitAll: %v", err)
	}
	if !committed {
		t.Fatal("expected first commit to succeed")
	}
	// Second commit with no changes should report committed=false, no error.
	committed, err = CommitAll(ctx, dir, "no-op")
	if err != nil {
		t.Fatalf("no-op CommitAll: %v", err)
	}
	if committed {
		t.Error("expected no commit on no-changes; CommitAll reported committed=true")
	}
}

func TestConfigure_Idempotent(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	if _, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := Configure(ctx, dir, "A", "a@b.c"); err != nil {
			t.Fatalf("Configure pass %d: %v", i, err)
		}
	}
}
