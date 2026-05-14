package selector

import (
	"context"
	"database/sql"
	"math"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestScore_Formula(t *testing.T) {
	cases := []struct {
		latency int
		speed   float64
		want    float64
	}{
		// hand-computed: 0.6*(50/100) - 0.4*(100/1000) = 0.30 - 0.04 = 0.26
		{100, 50.0, 0.26},
		// 0.6*(0/100) - 0.4*(500/1000) = 0 - 0.20 = -0.20
		{500, 0, -0.20},
		// 0.6*(100/100) - 0.4*(0/1000) = 0.6 - 0 = 0.6 (max)
		{0, 100, 0.6},
		// 0.6*(0) - 0.4*(0) = 0
		{0, 0, 0},
	}
	for _, c := range cases {
		got := Score(Result{LatencyMs: c.latency, SpeedMbps: c.speed})
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("Score(lat=%d, speed=%v) = %v, want %v", c.latency, c.speed, got, c.want)
		}
	}
}

func TestScore_CapsAtMaxima(t *testing.T) {
	// Latency well past cap: norm_latency caps at 1.0, so penalty is exactly -0.4
	got := Score(Result{LatencyMs: 5000, SpeedMbps: 0})
	if math.Abs(got-(-0.4)) > 1e-9 {
		t.Errorf("over-cap latency: Score = %v, want -0.4", got)
	}
	// Speed well past cap: norm_speed caps at 1.0, contribution 0.6
	got = Score(Result{LatencyMs: 0, SpeedMbps: 500})
	if math.Abs(got-0.6) > 1e-9 {
		t.Errorf("over-cap speed: Score = %v, want 0.6", got)
	}
}

func TestTop3PerCountry_BasicGrouping(t *testing.T) {
	in := []Result{
		{Link: "us-1", LatencyMs: 50, SpeedMbps: 100, Country: "US"},
		{Link: "us-2", LatencyMs: 100, SpeedMbps: 80, Country: "US"},
		{Link: "us-3", LatencyMs: 200, SpeedMbps: 60, Country: "US"},
		{Link: "us-4", LatencyMs: 300, SpeedMbps: 40, Country: "US"}, // out (4th)
		{Link: "de-1", LatencyMs: 30, SpeedMbps: 90, Country: "DE"},
		{Link: "de-2", LatencyMs: 60, SpeedMbps: 80, Country: "DE"},
		{Link: "de-3", LatencyMs: 90, SpeedMbps: 70, Country: "DE"},
	}
	got := Top3PerCountry(in)
	if len(got) != 2 {
		t.Fatalf("countries len = %d, want 2", len(got))
	}
	if got[0].Country != "DE" || got[1].Country != "US" {
		t.Errorf("alpha sort failed: %v", got)
	}
	for _, cs := range got {
		if len(cs.Top) != 3 {
			t.Errorf("country %s top len = %d, want 3", cs.Country, len(cs.Top))
		}
	}
	// US: us-1 has highest score (lowest latency, highest speed), should be first
	if got[1].Top[0].Link != "us-1" {
		t.Errorf("US top should be us-1, got %s", got[1].Top[0].Link)
	}
	// us-4 should NOT appear (only top 3)
	for _, r := range got[1].Top {
		if r.Link == "us-4" {
			t.Error("us-4 should not be in top 3")
		}
	}
}

func TestTop3PerCountry_Partial(t *testing.T) {
	in := []Result{
		{Link: "fr-1", LatencyMs: 50, SpeedMbps: 100, Country: "FR"},
		{Link: "fr-2", LatencyMs: 80, SpeedMbps: 80, Country: "FR"},
	}
	got := Top3PerCountry(in)
	if len(got) != 1 {
		t.Fatalf("countries = %d, want 1", len(got))
	}
	if len(got[0].Top) != 2 {
		t.Errorf("partial output len = %d, want 2", len(got[0].Top))
	}
}

func TestTop3PerCountry_OmitsEmptyCountry(t *testing.T) {
	in := []Result{
		{Link: "x-1", LatencyMs: 50, SpeedMbps: 100, Country: ""},
		{Link: "us-1", LatencyMs: 50, SpeedMbps: 100, Country: "US"},
	}
	got := Top3PerCountry(in)
	if len(got) != 1 {
		t.Fatalf("countries = %d, want 1 (empty-country row should be dropped)", len(got))
	}
	if got[0].Country != "US" {
		t.Errorf("kept country = %s, want US", got[0].Country)
	}
}

func TestTop3PerCountry_TieBreakByLatency(t *testing.T) {
	// Both have latency=100 speed=50 → identical score. Lower latency tied;
	// distinguish via Link to verify which row "won". Set distinct latency
	// pair that yields identical score: (lat=200,sp=80) vs (lat=100,sp=70):
	//   r1: 0.6*(80/100) - 0.4*(200/1000) = 0.48 - 0.08 = 0.40
	//   r2: 0.6*(70/100) - 0.4*(100/1000) = 0.42 - 0.04 = 0.38  -- not equal
	// Force a tie with identical scores:
	in := []Result{
		{Link: "slow", LatencyMs: 200, SpeedMbps: 50, Country: "GB"}, // 0.6*0.5 - 0.4*0.2 = 0.22
		{Link: "fast", LatencyMs: 100, SpeedMbps: 50, Country: "GB"}, // 0.6*0.5 - 0.4*0.1 = 0.26  (different)
	}
	// To get a tie, hand-craft scores:
	in = []Result{
		{Link: "slow", LatencyMs: 400, SpeedMbps: 100, Country: "GB"}, // 0.6 - 0.16 = 0.44
		{Link: "fast", LatencyMs: 100, SpeedMbps: 70, Country: "GB"},  // 0.42 - 0.04 = 0.38  -- still not equal
	}
	// Easier: use identical inputs differing only in latency
	in = []Result{
		{Link: "slow", LatencyMs: 200, SpeedMbps: 100, Country: "GB"},
		{Link: "fast", LatencyMs: 100, SpeedMbps: 100, Country: "GB"},
	}
	// Scores: slow=0.52, fast=0.56 -- "fast" naturally wins on score, not just tie-break.
	// To genuinely tie, set speed differently to compensate:
	//   want: 0.6*sA - 0.4*lA == 0.6*sB - 0.4*lB
	//   pick sA=100, lA=200 → 0.6 - 0.08 = 0.52
	//   pick lB=100 → need 0.6*sB - 0.04 = 0.52 → sB = 0.56/0.6 → sB ≈ 0.9333 → speed = 93.33
	in = []Result{
		{Link: "slow", LatencyMs: 200, SpeedMbps: 100, Country: "GB"},
		{Link: "fast", LatencyMs: 100, SpeedMbps: 93.33333333, Country: "GB"},
	}
	// Both should compute to ~0.52. Lower latency ("fast") wins tie-break.
	got := Top3PerCountry(in)
	if len(got) != 1 || len(got[0].Top) != 2 {
		t.Fatalf("want 1 country with 2 entries, got %v", got)
	}
	if got[0].Top[0].Link != "fast" {
		t.Errorf("tie-break should pick lower latency (fast); got %s first", got[0].Top[0].Link)
	}
}

func TestTop3PerCountry_AlphaSorted(t *testing.T) {
	in := []Result{
		{Link: "z-1", LatencyMs: 50, SpeedMbps: 100, Country: "ZW"},
		{Link: "a-1", LatencyMs: 50, SpeedMbps: 100, Country: "AR"},
		{Link: "m-1", LatencyMs: 50, SpeedMbps: 100, Country: "MX"},
	}
	got := Top3PerCountry(in)
	want := []string{"AR", "MX", "ZW"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, cc := range want {
		if got[i].Country != cc {
			t.Errorf("[%d] = %s, want %s", i, got[i].Country, cc)
		}
	}
}

// TestLoadResults_NoDB ensures missing DB file is a clean error, not a panic.
func TestLoadResults_NoDB(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "definitely-not-here.db")
	_, err := LoadResults(context.Background(), missing)
	if err == nil {
		t.Fatal("expected error for missing db")
	}
}

// TestLoadResults_FromFixtureDB writes a hand-crafted SQLite file matching the
// expected xray-knife schema and verifies LoadResults parses it correctly.
//
// Semantics: latest result per config_link across all runs is returned, dead
// rows (latency<=0 or >10000) are dropped. A link only present in run 1 is
// still our latest knowledge of it and is NOT filtered out by run-id.
func TestLoadResults_FromFixtureDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "fixture.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// Create a table with column names matching what mapColumns picks up.
	if _, err := db.Exec(`
CREATE TABLE test_results (
  id INTEGER PRIMARY KEY,
  link TEXT,
  delay INTEGER,
  download_speed REAL,
  location TEXT,
  run_id INTEGER
);
INSERT INTO test_results (link, delay, download_speed, location, run_id) VALUES
  ('vless://us-1@e.com:443', 50, 100.0, 'US', 2),
  ('vless://us-2@e.com:443', 80,  90.0, 'US', 2),
  ('vless://de-1@e.com:443', 30,  85.0, 'DE', 2),
  ('vless://us-1@e.com:443', 999, 50.0, 'US', 1), -- older run for us-1, should be overridden by run 2
  ('vless://only-old@e.com:443', 20, 999.0, 'US', 1), -- only present in old run, kept
  ('vless://dead@e.com:443',  0,   0.0, 'US', 2); -- dead, latency=0, dropped
`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	db.Close()

	results, err := LoadResults(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	// Want 4: us-1 (run 2 wins over run 1), us-2, de-1, only-old (no newer run)
	if len(results) != 4 {
		t.Errorf("got %d results, want 4 (dead dropped, older us-1 superseded by newer)", len(results))
	}
	for _, r := range results {
		if r.LatencyMs == 0 {
			t.Errorf("dead row leaked through: %v", r)
		}
		if r.Country != "US" && r.Country != "DE" {
			t.Errorf("unexpected country: %v", r)
		}
		// Verify us-1 has the run-2 latency (50), not run-1 (999).
		if r.Link == "vless://us-1@e.com:443" && r.LatencyMs != 50 {
			t.Errorf("us-1 should have latest result (50ms), got %dms", r.LatencyMs)
		}
	}
}
