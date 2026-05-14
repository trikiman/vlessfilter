package pipeline

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/trikiman/vlessfilter/internal/xrayknife"
)

// fakeRunner is a minimal Runner for pipeline tests; mirrors xrayknife.FakeRunner
// but kept local to avoid depending on an internal _test.go type from another package.
type fakeRunner struct {
	IsAvailable bool
	Calls       []string
	SubCountVal int
	DBPathVal   string
	SubsAddErr  error
	HTTPErr     error
}

func (f *fakeRunner) Available(ctx context.Context) (bool, error) { return true, nil }
func (f *fakeRunner) Install(ctx context.Context) error           { return nil }
func (f *fakeRunner) SubsAdd(ctx context.Context, url, remark string) error {
	f.Calls = append(f.Calls, "SubsAdd:"+remark)
	return f.SubsAddErr
}
func (f *fakeRunner) SubsFetch(ctx context.Context) error {
	f.Calls = append(f.Calls, "SubsFetch")
	return nil
}
func (f *fakeRunner) HTTPTest(ctx context.Context, opts xrayknife.HTTPOpts) error {
	tag := "HTTPTest:ping"
	if opts.Speedtest {
		tag = "HTTPTest:speed"
	}
	f.Calls = append(f.Calls, tag)
	return f.HTTPErr
}
func (f *fakeRunner) SubCount(ctx context.Context) (int, error) { return f.SubCountVal, nil }
func (f *fakeRunner) DBPath() (string, error) {
	return f.DBPathVal, nil
}

// writeMinimalSourcesYaml creates a minimal valid sources.yaml in tmp and returns its path.
func writeMinimalSourcesYaml(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "sources.yaml")
	body := `
sources:
  - name: test-source
    url: https://example.com/vless.txt
    kind: plain
    enabled: true
`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// makeFixtureDB creates a SQLite db with valid xray-knife-shaped results so
// the select stage can exercise the full path.
func makeFixtureDB(t *testing.T) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "fixture.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`
CREATE TABLE test_results (
  id INTEGER PRIMARY KEY,
  link TEXT,
  delay INTEGER,
  download_speed REAL,
  location TEXT
);
INSERT INTO test_results (link, delay, download_speed, location) VALUES
  ('vless://us-1@example.com:443', 50, 80.0, 'US'),
  ('vless://us-2@example.com:443', 70, 60.0, 'US'),
  ('vless://us-3@example.com:443', 90, 50.0, 'US');
`); err != nil {
		t.Fatal(err)
	}
	return dbPath
}

func TestRun_FullPipeline_WithFakeRunner(t *testing.T) {
	yamlPath := writeMinimalSourcesYaml(t)
	dbPath := makeFixtureDB(t)
	outDir := t.TempDir()

	r := &fakeRunner{SubCountVal: 1, DBPathVal: dbPath}
	err := Run(context.Background(), Opts{
		SourcesPath: yamlPath,
		OutDir:      outDir,
		Runner:      r,
		Now:         func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Calls should include SubsAdd (per source), SubsFetch, HTTPTest:ping, HTTPTest:speed.
	want := []string{"SubsAdd", "SubsFetch", "HTTPTest:ping", "HTTPTest:speed"}
	for _, w := range want {
		found := false
		for _, c := range r.Calls {
			if strings.HasPrefix(c, w) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Calls missing %q; got %v", w, r.Calls)
		}
	}

	// Output files should exist.
	if _, err := os.Stat(filepath.Join(outDir, "README.md")); err != nil {
		t.Errorf("README.md not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "subs", "US.txt")); err != nil {
		t.Errorf("subs/US.txt not written: %v", err)
	}
}

func TestRun_StageFetchOnly(t *testing.T) {
	yamlPath := writeMinimalSourcesYaml(t)
	r := &fakeRunner{SubCountVal: 1, DBPathVal: "/tmp/unused.db"}

	err := Run(context.Background(), Opts{
		SourcesPath: yamlPath,
		OutDir:      t.TempDir(),
		Stage:       "fetch",
		Runner:      r,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, c := range r.Calls {
		if strings.HasPrefix(c, "HTTPTest") {
			t.Errorf("HTTPTest should not be called in fetch-only stage; calls: %v", r.Calls)
		}
	}
}

func TestRun_RejectsBadStage(t *testing.T) {
	r := &fakeRunner{}
	err := Run(context.Background(), Opts{Stage: "bogus", Runner: r})
	if err == nil {
		t.Fatal("expected error for invalid stage")
	}
}

func TestRun_RejectsNilRunner(t *testing.T) {
	err := Run(context.Background(), Opts{})
	if err == nil {
		t.Fatal("expected error for nil Runner")
	}
}

func TestRun_StageTestRequiresFetched(t *testing.T) {
	r := &fakeRunner{SubCountVal: 0}
	err := Run(context.Background(), Opts{Stage: "test", Runner: r})
	if err == nil || !strings.Contains(err.Error(), "no subscriptions") {
		t.Fatalf("expected 'no subscriptions' error; got: %v", err)
	}
}

// Sentinel test ensuring SubsAdd errors don't crash the pipeline (logged + skipped).
func TestRun_SubsAddErrorIsSkipped(t *testing.T) {
	yamlPath := writeMinimalSourcesYaml(t)
	r := &fakeRunner{
		SubCountVal: 1,
		DBPathVal:   makeFixtureDB(t),
		SubsAddErr:  errors.New("simulated"),
	}
	err := Run(context.Background(), Opts{
		SourcesPath: yamlPath,
		OutDir:      t.TempDir(),
		Stage:       "fetch",
		Runner:      r,
	})
	if err != nil {
		t.Fatalf("fetch should swallow per-source errors; got %v", err)
	}
}
