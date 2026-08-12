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

// regen-readme builds each protocol's section purely from its
// .readme-data/<proto>.json sidecar. A missing sidecar used to just `continue`,
// which silently deleted that protocol's section AND its subscription URLs from
// the README while its keys stayed published — vless (43 keys) and trojan (51)
// were undocumented that way. It must now refuse rather than publish a README
// that hides live output.
func TestRegenReadme_RefusesToDropPublishedProtocol(t *testing.T) {
	newRepo := func(t *testing.T, publishedProtos []string, sidecarProtos []string) string {
		t.Helper()
		dir := t.TempDir()
		for _, p := range publishedProtos {
			d := filepath.Join(dir, "subs", p)
			if err := os.MkdirAll(d, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(d, "all.txt"),
				[]byte("vless://u@1.2.3.4:443#X\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		sd := filepath.Join(dir, ".readme-data")
		if err := os.MkdirAll(sd, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, p := range sidecarProtos {
			body := `{"protocol":"` + p + `","selections":[],"rotating":0,` +
				`"generated_at":"2026-08-12T00:00:00Z"}`
			if err := os.WriteFile(filepath.Join(sd, p+".json"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}

	t.Run("published protocol missing its sidecar aborts", func(t *testing.T) {
		// trojan has keys but no sidecar — the exact production failure.
		dir := newRepo(t, []string{"vless", "trojan"}, []string{"vless"})
		if code := regenReadmeCmd([]string{"--out", dir}); code == exitOK {
			t.Error("regen-readme succeeded while dropping trojan; want non-zero exit")
		}
	})

	t.Run("unpublished protocol missing its sidecar is fine", func(t *testing.T) {
		// Only vless is published, and it has its sidecar. The other three
		// protocols have neither keys nor sidecars — nothing is being hidden.
		dir := newRepo(t, []string{"vless"}, []string{"vless"})
		if code := regenReadmeCmd([]string{"--out", dir}); code != exitOK {
			t.Errorf("exit %d, want %d: nothing published was omitted", code, exitOK)
		}
	})

	t.Run("corrupt sidecar for a published protocol aborts", func(t *testing.T) {
		dir := newRepo(t, []string{"vless"}, nil)
		if err := os.WriteFile(filepath.Join(dir, ".readme-data", "vless.json"),
			[]byte("{not json"), 0o644); err != nil {
			t.Fatal(err)
		}
		if code := regenReadmeCmd([]string{"--out", dir}); code == exitOK {
			t.Error("regen-readme accepted a corrupt sidecar for a published protocol")
		}
	})
}
