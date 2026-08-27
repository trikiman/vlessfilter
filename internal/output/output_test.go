package output

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/trikiman/vlessfilter/internal/selector"
)

var update = flag.Bool("update", false, "update golden files")

func TestRewriteRemark(t *testing.T) {
	cases := []struct {
		name, link, cc string
		speed          float64
		want           string
	}{
		{
			name:  "flag first, then icon + [mb] + CC + name (video tier)",
			link:  "vless://abc@example.com:443",
			cc:    "DE",
			speed: 15,
			// 🇩🇪 + 📺(%F0%9F%93%BA) + [15.0 Mbps] + DE Germany
			want: "vless://abc@example.com:443#%F0%9F%87%A9%F0%9F%87%AA%20%F0%9F%93%BA%20%5B15.0%20Mbps%5D%20DE%20Germany",
		},
		{
			name:  "existing fragment replaced (25+ = clapper)",
			link:  "vless://abc@example.com:443?security=tls#OldName",
			cc:    "PL",
			speed: 30,
			// 🇵🇱 + 🎬(%F0%9F%8E%AC) + [30.0 Mbps] + PL Poland
			want: "vless://abc@example.com:443?security=tls#%F0%9F%87%B5%F0%9F%87%B1%20%F0%9F%8E%AC%20%5B30.0%20Mbps%5D%20PL%20Poland",
		},
		{
			name:  "zero speed: no icon, no bracket",
			link:  "vless://x@y.com:443",
			cc:    "jp",
			speed: 0,
			want:  "vless://x@y.com:443#%F0%9F%87%AF%F0%9F%87%B5%20JP%20Japan",
		},
		{
			name:  "unknown CC: no name duplication (60+ = bolt)",
			link:  "vless://q@z.com:443",
			cc:    "ZZ",
			speed: 70,
			// 🇿🇿 + ⚡(%E2%9A%A1) + [70.0 Mbps] + ZZ
			want: "vless://q@z.com:443#%F0%9F%87%BF%F0%9F%87%BF%20%E2%9A%A1%20%5B70.0%20Mbps%5D%20ZZ",
		},
		{
			name:  "unparseable input returned as-is",
			link:  "not a url at all !!! ::: ???",
			cc:    "US",
			speed: 50,
			want:  "not a url at all !!! ::: ???",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := rewriteRemark(c.link, c.cc, c.speed); got != c.want {
				t.Errorf("rewriteRemark(%q, %q, %v)\ngot:  %q\nwant: %q", c.link, c.cc, c.speed, got, c.want)
			}
		})
	}
}

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

	if err := Write(tmp, selections, nil, generatedAt); err != nil {
		t.Fatalf("Write: %v", err)
	}

	usTxt, err := os.ReadFile(filepath.Join(tmp, "subs", "US.txt"))
	if err != nil {
		t.Fatalf("read US.txt: %v", err)
	}
	// Expect "<flag> United States" (full country name from map).
	// 🇺🇸 is U+1F1FA U+1F1F8 = %F0%9F%87%BA%F0%9F%87%B8.
	wantUS := "vless://us-1@example.com:443#%F0%9F%87%BA%F0%9F%87%B8%20%E2%9A%A1%20%5B78.3%20Mbps%5D%20US%20United%20States\n" +
		"vless://us-2@example.com:443#%F0%9F%87%BA%F0%9F%87%B8%20%E2%9A%A1%20%5B70.0%20Mbps%5D%20US%20United%20States\n" +
		"vless://us-3@example.com:443#%F0%9F%87%BA%F0%9F%87%B8%20%E2%9A%A1%20%5B65.5%20Mbps%5D%20US%20United%20States\n"
	if string(usTxt) != wantUS {
		t.Errorf("US.txt mismatch:\ngot:\n%s\nwant:\n%s", usTxt, wantUS)
	}

	// subs/all.txt must contain DE entries first (alphabetical), then US.
	allTxt, err := os.ReadFile(filepath.Join(tmp, "subs", "all.txt"))
	if err != nil {
		t.Fatalf("read all.txt: %v", err)
	}
	allStr := string(allTxt)
	if !strings.Contains(allStr, "vless://de-1@example.com:443#%F0%9F%87%A9%F0%9F%87%AA%20%E2%9A%A1%20%5B92.1%20Mbps%5D%20DE%20Germany") {
		t.Errorf("all.txt missing DE-flagged entry; got:\n%s", allStr)
	}
	if !strings.Contains(allStr, "vless://us-1@example.com:443#%F0%9F%87%BA%F0%9F%87%B8%20%E2%9A%A1%20%5B78.3%20Mbps%5D%20US%20United%20States") {
		t.Errorf("all.txt missing US-flagged entry; got:\n%s", allStr)
	}
	// 6 lines (3 DE + 3 US) + trailing newline.
	if lines := strings.Count(allStr, "\n"); lines != 6 {
		t.Errorf("all.txt expected 6 lines, got %d:\n%s", lines, allStr)
	}
	// DE rows must come before US rows (alphabetical).
	dePos := strings.Index(allStr, "de-1")
	usPos := strings.Index(allStr, "us-1")
	if dePos < 0 || usPos < 0 || dePos > usPos {
		t.Errorf("all.txt order wrong: DE at %d, US at %d", dePos, usPos)
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

	if err := Write(tmp, selections, nil, time.Date(2026, 5, 14, 22, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("Write 1: %v", err)
	}
	first, _ := os.ReadFile(filepath.Join(tmp, "README.md"))

	if err := Write(tmp, selections, nil, time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
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

// all-results.csv has no natural bound: it is the union of every tested config
// across four protocols, and a single vless sweep alone can reach 240k rows.
// Unbounded, it hit 131 MB and GitHub rejected the entire push at the
// pre-receive hook (GH001, 100 MB hard limit) — six refresh runs tested for
// ~55 minutes and then threw the results away. Cap it, and keep the best rows.
func TestWriteAllResultsCSV_CapsRowsUnderBlobLimit(t *testing.T) {
	dir := t.TempDir()

	// One more row than the cap, with score ascending so the highest scores
	// are at the END of the input — a prefix truncation would keep the worst.
	n := MaxCSVRows + 1000
	rows := make([]selector.Result, 0, n)
	for i := 0; i < n; i++ {
		rows = append(rows, selector.Result{
			Link:      fmt.Sprintf("vless://u%d@10.0.0.1:443", i),
			LatencyMs: 1000 - (i % 1000),
			SpeedMbps: float64(i%100) / 10,
			Country:   "US",
		})
	}
	if err := WriteDiagnostics(dir, rows, nil); err != nil {
		t.Fatalf("WriteDiagnostics: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "all-results.csv"))
	if err != nil {
		t.Fatal(err)
	}
	data := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	got := len(data) - 1 // minus header
	if got != MaxCSVRows {
		t.Errorf("wrote %d rows, want the %d cap", got, MaxCSVRows)
	}
	if len(raw) > 100*1024*1024 {
		t.Errorf("file is %d bytes — over GitHub's 100 MB hard limit", len(raw))
	}

	// The retained set must be the best-scoring, so truncation happens after
	// the sort. The top row's score must beat the worst possible input score.
	if got > 0 {
		first := data[1]
		if strings.HasSuffix(first, ",0.0000") || strings.Contains(first, ",-0.4") {
			t.Errorf("kept a worst-scoring row first (%q) — truncated before sorting?", first)
		}
	}
}

// The README users actually read comes from buildMultiProtocolReadme, which had
// NO test coverage: golden-readme.md pins the deprecated single-protocol Write
// path ("top 3 fastest VLESS keys"), so it never exercised this function. That
// gap let a real defect ship — the published README documented VMESS and SS
// only, while 43 vless and 51 trojan keys were served with no section and no
// subscription URL anywhere in the file.
func TestBuildMultiProtocolReadme_SectionPerProtocol(t *testing.T) {
	mk := func(proto, cc string) ProtoReadme {
		return ProtoReadme{
			Protocol: proto,
			Selections: []selector.CountrySelection{{
				Country: cc,
				Top: []selector.Result{{
					Link: proto + "://u@1.2.3.4:443", LatencyMs: 100, SpeedMbps: 20, Country: cc,
				}},
			}},
		}
	}
	protos := []ProtoReadme{mk("vless", "DE"), mk("vmess", "US"), mk("trojan", "NL"), mk("ss", "FI")}
	got := buildMultiProtocolReadme(protos, time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC))

	for _, p := range protos {
		up := strings.ToUpper(p.Protocol)
		// A usable section needs both a heading and a subscription URL —
		// a heading alone still leaves the user unable to subscribe.
		if !strings.Contains(got, up) {
			t.Errorf("README has no mention of %s despite published keys", up)
		}
		url := "subs/" + p.Protocol + "/all.txt"
		if !strings.Contains(got, url) {
			t.Errorf("README omits the %s subscription URL (%s)", up, url)
		}
	}

	// The intro advertises all four protocols; it must not do that while a
	// section is missing. Guard the inverse case too: one protocol only.
	single := buildMultiProtocolReadme([]ProtoReadme{mk("ss", "FI")}, time.Now())
	if !strings.Contains(single, "subs/ss/all.txt") {
		t.Error("single-protocol README omits its own subscription URL")
	}
	if strings.Contains(single, "subs/trojan/all.txt") {
		t.Error("README advertises a trojan URL when no trojan keys were passed")
	}
}
