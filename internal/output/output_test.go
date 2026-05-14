package output

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
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
		{"USA", "USA"},
		{"u1", "u1"},
		{"", ""},
	}
	for _, c := range cases {
		if got := flagEmoji(c.in); got != c.want {
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
		{[]float64{100, 1, 50}, 50.0},
	}
	for _, c := range cases {
		if got := median(c.in); got != c.want {
			t.Errorf("median(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestWrite_GoldenReadme(t *testing.T) {
	tmp := t.TempDir()
	generatedAt := time.Date(2026, 5, 14, 22, 0, 0, 0, time.UTC)
	selections := goldenSelections()

	if err := Write(tmp, selections, generatedAt); err != nil {
		t.Fatalf("Write: %v", err)
	}

	usTxt, err := os.ReadFile(filepath.Join(tmp, "subs", "US.txt"))
	if err != nil {
		t.Fatalf("read US.txt: %v", err)
	}
	wantUS := "vless://us-1@example.com:443\nvless://us-2@example.com:443\nvless://us-3@example.com:443\n"
	if string(usTxt) != wantUS {
		t.Errorf("US.txt mismatch:\n%s", usTxt)
	}

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
		t.Logf("updated %s", goldenPath)
		return
	}
	wantReadme, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (run -update first): %v", err)
	}
	if string(gotReadme) != string(wantReadme) {
		t.Errorf("README.md mismatch.\n--- got ---\n%s\n--- want ---\n%s", gotReadme, wantReadme)
	}
}

// TestWrite_TimestampInComment ensures only the bottom HTML comment varies
// between runs with identical inputs (D-17 determinism).
func TestWrite_TimestampInComment(t *testing.T) {
	tmp := t.TempDir()
	selections := goldenSelections()

	if err := Write(tmp, selections, time.Date(2026, 5, 14, 22, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("Write 1: %v", err)
	}
	first, _ := os.ReadFile(filepath.Join(tmp, "README.md"))

	if err := Write(tmp, selections, time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("Write 2: %v", err)
	}
	second, _ := os.ReadFile(filepath.Join(tmp, "README.md"))

	// They MUST differ (timestamp changed)
	if string(first) == string(second) {
		t.Fatal("expected README to differ when timestamp changed")
	}
	// They must differ ONLY in the last-tested comment line. Compare line-by-line.
	firstLines := strings.Split(string(first), "\n")
	secondLines := strings.Split(string(second), "\n")
	if len(firstLines) != len(secondLines) {
		t.Fatalf("line count differs: %d vs %d", len(firstLines), len(secondLines))
	}
	diffs := 0
	for i := range firstLines {
		if firstLines[i] != secondLines[i] {
			diffs++
			if !strings.Contains(firstLines[i], "last-tested") {
				t.Errorf("non-timestamp line differs at %d: %q vs %q", i, firstLines[i], secondLines[i])
			}
		}
	}
	if diffs != 1 {
		t.Errorf("expected exactly 1 differing line (the timestamp comment); got %d", diffs)
	}
}

func TestWriteDiagnostics_Deterministic(t *testing.T) {
	tmp := t.TempDir()
	all := []selector.Result{
		{Link: "vless://us-2", LatencyMs: 80, SpeedMbps: 60, Country: "US"},
		{Link: "vless://us-1", LatencyMs: 50, SpeedMbps: 100, Country: "US"},
		{Link: "vless://de-1", LatencyMs: 30, SpeedMbps: 90, Country: "DE"},
	}
	dead := []selector.Result{
		{Link: "vless://dead-2"},
		{Link: "vless://dead-1"},
	}
	if err := WriteDiagnostics(tmp, all, dead); err != nil {
		t.Fatalf("WriteDiagnostics: %v", err)
	}

	csv, _ := os.ReadFile(filepath.Join(tmp, "all-results.csv"))
	csvStr := string(csv)
	if !strings.HasPrefix(csvStr, "Link,LatencyMs,SpeedMbps,Country,Score\n") {
		t.Errorf("csv header wrong: %s", csvStr)
	}
	// First data row must be DE (alphabetical first), second & third US in score order
	lines := strings.Split(strings.TrimSpace(csvStr), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected header + 3 rows, got %d lines", len(lines))
	}
	if !strings.HasPrefix(lines[1], "vless://de-1,") {
		t.Errorf("expected DE row first; got %q", lines[1])
	}
	// US rows: us-1 has higher score than us-2 (lower latency, higher speed)
	if !strings.HasPrefix(lines[2], "vless://us-1,") {
		t.Errorf("expected us-1 second; got %q", lines[2])
	}
	if !strings.HasPrefix(lines[3], "vless://us-2,") {
		t.Errorf("expected us-2 third; got %q", lines[3])
	}

	dead0, _ := os.ReadFile(filepath.Join(tmp, "raw", "dead.txt"))
	wantDead := "vless://dead-1\nvless://dead-2\n"
	if string(dead0) != wantDead {
		t.Errorf("dead.txt mismatch:\n%s", dead0)
	}
}

func goldenSelections() []selector.CountrySelection {
	return []selector.CountrySelection{
		{Country: "DE", Top: []selector.Result{
			{Link: "vless://de-1@example.com:443", LatencyMs: 28, SpeedMbps: 92.1},
			{Link: "vless://de-2@example.com:443", LatencyMs: 35, SpeedMbps: 88.5},
			{Link: "vless://de-3@example.com:443", LatencyMs: 41, SpeedMbps: 75.0},
		}},
		{Country: "US", Top: []selector.Result{
			{Link: "vless://us-1@example.com:443", LatencyMs: 42, SpeedMbps: 78.3},
			{Link: "vless://us-2@example.com:443", LatencyMs: 55, SpeedMbps: 70.0},
			{Link: "vless://us-3@example.com:443", LatencyMs: 60, SpeedMbps: 65.5},
		}},
	}
}
