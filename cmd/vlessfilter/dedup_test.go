package main

import (
	"os"
	"path/filepath"
	"testing"
)

// A CRLF subscription file with no #fragment puts the \r immediately after the
// port, so url.Parse sees port "443\r" and the same server spelled CRLF vs LF
// reads as two distinct endpoints — exactly the duplicate this command exists
// to remove. Guard the trim.
func TestDedupEndpoints_CRLF(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    int
	}{
		{"crlf no fragment, same endpoint", "vless://a@8.8.8.8:443\r\nvless://b@8.8.8.8:443\n", 1},
		{"crlf with fragment", "vless://a@1.2.3.4:443#X\r\nvless://b@1.2.3.4:443#X\r\nvless://c@5.6.7.8:443#Y\r\n", 2},
		{"distinct endpoints survive", "vless://a@1.1.1.1:443\nvless://b@2.2.2.2:443\n", 2},
		{"no trailing newline keeps last line", "vless://a@1.1.1.1:443\nvless://b@2.2.2.2:443", 2},
		{"unparseable line is kept, never dropped", "not-a-uri\nvless://a@1.1.1.1:443\n", 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "subs.txt")
			if err := os.WriteFile(path, []byte(c.content), 0o644); err != nil {
				t.Fatal(err)
			}
			if code := dedupEndpointsCmd([]string{path}); code != exitOK {
				t.Fatalf("exit %d, want %d", code, exitOK)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			got := 0
			for _, b := range raw {
				if b == '\n' {
					got++
				}
			}
			if got != c.want {
				t.Errorf("kept %d lines, want %d\ngot:\n%q", got, c.want, raw)
			}
			for _, b := range raw {
				if b == '\r' {
					t.Errorf("stray CR survived into output: %q", raw)
					break
				}
			}
		})
	}
}

// An in-place rewrite must never destroy input it failed to understand.
func TestDedupEndpoints_MissingAndEmptyAreSafe(t *testing.T) {
	dir := t.TempDir()
	// A glob in the workflow may not match anything; that must not fail the step.
	if code := dedupEndpointsCmd([]string{filepath.Join(dir, "absent.txt")}); code != exitOK {
		t.Errorf("missing file: exit %d, want %d", code, exitOK)
	}
	empty := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if code := dedupEndpointsCmd([]string{empty}); code != exitOK {
		t.Errorf("empty file: exit %d, want %d", code, exitOK)
	}
	if raw, _ := os.ReadFile(empty); len(raw) != 0 {
		t.Errorf("empty file gained content: %q", raw)
	}
}
