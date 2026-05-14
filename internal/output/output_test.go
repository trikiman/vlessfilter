package output

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/trikiman/vlessfilter/internal/selector"
)

var update = flag.Bool("update", false, "update golden files")

func TestFlagEmoji(t *testing.T) {
	cases := []struct{ in, want string }{
		{"US", "🇺🇸"},
		{"DE", "🇩🇪"},
		{"AR", "🇦🇷"},
		{"USA", "USA"}, // wrong length, fallback
		{"u1", "u1"},   // not letters, fallback
		{"", ""},
	}
	for _, c := range cases {
		got := flagEmoji(c.in)
		if got != c.want {
			t.Errorf("flagEmoji(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMedian(t *testing.T) {
	cases := []struct {
		in   []float64
		want float64
	}{
		{nil, 0},
		{[]float64{}, 0},
		{[]float64{1, 2, 3}, 2.0},
		{[]float64{1, 2, 3, 4}, 2.5},
		{[]float64{50.0, 100.0, 75.0}, 75.0},
		// Tests rounding to 1 decimal: median of {1, 2, 3, 4.7} is (2+3)/2 = 2.5
		{[]float64{1, 2, 3, 4.7}, 2.5},
		// Out-of-order input still works
		{[]float64{100, 1, 50}, 50.0},
	}
	for _, c := range cases {
		got := median(c.in)
		if got != c.want {
			t.Errorf("median(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestWrite_GoldenReadme(t *testing.T) {
	tmp := t.TempDir()

	// Fixed input — must match the golden files byte-for-byte.
	generatedAt := time.Date(2026, 5, 14, 22, 0, 0, 0, time.UTC)
	selections := []selector.CountrySelection{
		{
			Country: "DE",
			Top: []selector.Result{
				{Link: "vless://de-1@example.com:443", LatencyMs: 28, SpeedMbps: 92.1},
				{Link: "vless://de-2@example.com:443", LatencyMs: 35, SpeedMbps: 88.5},
				{Link: "vless://de-3@example.com:443", LatencyMs: 41, SpeedMbps: 75.0},
			},
		},
		{
			Country: "US",
			Top: []selector.Result{
				{Link: "vless://us-1@example.com:443", LatencyMs: 42, SpeedMbps: 78.3},
				{Link: "vless://us-2@example.com:443", LatencyMs: 55, SpeedMbps: 70.0},
				{Link: "vless://us-3@example.com:443", LatencyMs: 60, SpeedMbps: 65.5},
			},
		},
	}

	if err := Write(tmp, selections, generatedAt); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Verify subs/US.txt and subs/DE.txt
	usTxt, err := os.ReadFile(filepath.Join(tmp, "subs", "US.txt"))
	if err != nil {
		t.Fatalf("read US.txt: %v", err)
	}
	deTxt, err := os.ReadFile(filepath.Join(tmp, "subs", "DE.txt"))
	if err != nil {
		t.Fatalf("read DE.txt: %v", err)
	}

	wantUS := "vless://us-1@example.com:443\nvless://us-2@example.com:443\nvless://us-3@example.com:443\n"
	if string(usTxt) != wantUS {
		t.Errorf("US.txt mismatch.\ngot:\n%s\nwant:\n%s", usTxt, wantUS)
	}
	wantDE := "vless://de-1@example.com:443\nvless://de-2@example.com:443\nvless://de-3@example.com:443\n"
	if string(deTxt) != wantDE {
		t.Errorf("DE.txt mismatch.\ngot:\n%s\nwant:\n%s", deTxt, wantDE)
	}

	// Verify README.md against golden file
	gotReadme, err := os.ReadFile(filepath.Join(tmp, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}

	goldenPath := filepath.Join("testdata", "golden-readme.md")
	if *update {
		_ = os.MkdirAll(filepath.Dir(goldenPath), 0o755)
		if err := os.WriteFile(goldenPath, gotReadme, 0o644); err != nil {
			t.Fatalf("update golden: %v", err)
		}
		t.Logf("golden updated: %s", goldenPath)
		return
	}
	wantReadme, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (run with -update first): %v", err)
	}
	if string(gotReadme) != string(wantReadme) {
		t.Errorf("README.md mismatch.\n--- got ---\n%s\n--- want ---\n%s", gotReadme, wantReadme)
	}
}
