// Package pipeline orchestrates the end-to-end run: load sources →
// xray-knife subs add+fetch → stage 1 ping → stage 2 speedtest → selector →
// output. Stages can be invoked independently for debugging via Opts.Stage.
//
// Phase 2 additions:
//   - Wall-clock budget via context.WithDeadline (Opts.BudgetMin)
//   - Periodic checkpoint that writes outputs and (optionally) git push
//   - kerntune.Apply at the top of stage 1
//   - Diagnostic outputs (all-results.csv, raw/dead.txt) via output.WriteAll
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/trikiman/vlessfilter/internal/accuracy"
	"github.com/trikiman/vlessfilter/internal/git"
	"github.com/trikiman/vlessfilter/internal/kerntune"
	"github.com/trikiman/vlessfilter/internal/output"
	"github.com/trikiman/vlessfilter/internal/selector"
	"github.com/trikiman/vlessfilter/internal/sources"
	"github.com/trikiman/vlessfilter/internal/xrayknife"
)

// Opts controls a pipeline run.
type Opts struct {
	SourcesPath string
	OutDir      string
	Stage       string // "" | "fetch" | "test" | "select"
	Threads1    int
	Threads2    int
	Limit       int
	Runner      xrayknife.Runner
	Now         func() time.Time

	// Budget: max wall-clock minutes the run may use. <=0 disables.
	BudgetMin int

	// Checkpoint: write outputs (and optionally git-push) every N minutes
	// during a full run. <=0 disables.
	CheckpointMin int

	// Git: when GitPush is true and GitToken is non-empty, every checkpoint
	// commits + pushes outputs to GitBranch in GitRepoDir.
	GitPush    bool
	GitRepoDir string
	GitBranch  string
	GitToken   string

	// RunAccuracyProbe: GEO-04 post-publish ground-truth probe. After
	// runSelect writes outputs, samples 5 random keys per country, routes
	// each through xray-knife to ipinfo.io, compares actual exit country
	// to published label, logs report. Set false for fast/dev runs.
	RunAccuracyProbe bool
	GitName          string
	GitEmail         string
}

var validStages = map[string]bool{"": true, "fetch": true, "test": true, "select": true}

// Run executes the pipeline according to opts.Stage. Empty stage runs all
// three stages in order.
//
// On budget expiry: the in-flight stage's xray-knife child is cancelled via
// the context; we then run a final select+commit so any partial progress
// reaches the output dir / repo. Returns nil even on budget expiry — partial
// success is still success for our use case.
func Run(ctx context.Context, opts Opts) error {
	if opts.Runner == nil {
		return errors.New("pipeline: Runner is required")
	}
	if !validStages[opts.Stage] {
		return fmt.Errorf("pipeline: invalid stage %q (must be empty, fetch, test, or select)", opts.Stage)
	}
	if opts.SourcesPath == "" {
		opts.SourcesPath = "./sources.yaml"
	}
	if opts.OutDir == "" {
		opts.OutDir = "."
	}
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	if opts.GitBranch == "" {
		opts.GitBranch = "main"
	}
	if opts.GitRepoDir == "" {
		opts.GitRepoDir = "."
	}
	if opts.GitName == "" {
		opts.GitName = "VlessFilter Bot"
	}
	if opts.GitEmail == "" {
		opts.GitEmail = "vlessfilter-bot@localhost"
	}

	// Budget: derive a deadline.
	if opts.BudgetMin > 0 {
		deadline := opts.Now().Add(time.Duration(opts.BudgetMin) * time.Minute)
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
		slog.Info("pipeline budget set", "minutes", opts.BudgetMin, "deadline", deadline.Format(time.RFC3339))
	}

	// Configure git identity early so checkpoint commits don't fail.
	if opts.GitPush {
		if err := git.Configure(ctx, opts.GitRepoDir, opts.GitName, opts.GitEmail); err != nil {
			slog.Warn("git configure failed; commits may fail", "error", err)
		}
	}

	// Per-stage execution
	switch opts.Stage {
	case "fetch":
		return runFetch(ctx, opts)
	case "test":
		return runTest(ctx, opts)
	case "select":
		return runSelect(ctx, opts)
	}

	// Full pipeline with checkpoint loop running concurrently.
	stopCheckpoint := startCheckpointLoop(ctx, opts)
	defer stopCheckpoint()

	if err := runFetch(ctx, opts); err != nil {
		// Even on fetch failure, attempt a final select to ship any prior outputs.
		_ = runSelect(ctx, opts)
		return fmt.Errorf("stage fetch: %w", err)
	}

	if err := runTest(ctx, opts); err != nil {
		// Likely budget cancellation. Try to ship partial outputs.
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			slog.Warn("budget reached during test stage; shipping partial outputs", "error", err)
			_ = runSelect(ctx, opts)
			finalCommitPush(opts)
			return nil
		}
		return fmt.Errorf("stage test: %w", err)
	}

	if err := runSelect(ctx, opts); err != nil {
		return fmt.Errorf("stage select: %w", err)
	}

	finalCommitPush(opts)
	return nil
}

func runFetch(ctx context.Context, opts Opts) error {
	cfg, err := sources.Load(opts.SourcesPath)
	if err != nil {
		return err
	}
	expanded := sources.Expand(cfg)
	slog.Info("fetch: ingesting subscriptions", "count", len(expanded))
	for _, e := range expanded {
		if err := opts.Runner.SubsAdd(ctx, e.URL, e.Name); err != nil {
			slog.Warn("subs add failed; skipping", "name", e.Name, "url", e.URL, "error", err)
			continue
		}
	}
	if err := opts.Runner.SubsFetch(ctx); err != nil {
		return err
	}
	if count, err := opts.Runner.SubCount(ctx); err == nil {
		slog.Info("fetch complete", "subscriptions_in_db", count)
	}
	return nil
}

func runTest(ctx context.Context, opts Opts) error {
	count, err := opts.Runner.SubCount(ctx)
	if err == nil && count == 0 {
		return errors.New("no subscriptions in xray-knife db; run 'vlessfilter run --stage fetch' first")
	}

	// Apply kernel tunables before the high-concurrency stage.
	_ = kerntune.Apply(ctx)

	t1 := opts.Threads1
	if t1 <= 0 {
		t1 = 200
	}
	t2 := opts.Threads2
	if t2 <= 0 {
		t2 = 20
	}

	// Stage 1: TLS handshake against a manageable batch of NEW (untested)
	// configs PLUS retests of the existing alive set, instead of trying to
	// test the entire 700k+ pool and getting killed by the budget.
	//
	// Why this matters: previously stage 1 used --from-db which iterates
	// the full pool. With 700k+ configs at ~30/s sustained, a 60-min run
	// only covered ~108k = 14% before getting killed. Worse, xray-knife
	// didn't flush partial results when killed, so several full-pool
	// scheduled runs produced 0 saved alive entries.
	//
	// New design: each run tests up to untestedBatch new configs +
	// re-validates currently-alive ones. Over ~5-7 runs the pool gets
	// fully covered, and once-alive configs that died get marked as such.
	//
	// LIVE-04: --profile dev passes opts.Limit > 0 to cap stage 1 to
	// a small subset for fast iteration.
	untestedBatch := 80000
	if opts.Limit > 0 && opts.Limit < untestedBatch {
		untestedBatch = opts.Limit
	}
	dbPath, err := opts.Runner.DBPath()
	if err != nil {
		return fmt.Errorf("stage 1: db path: %w", err)
	}
	untested, err := selector.LoadUntestedLinks(ctx, dbPath, untestedBatch)
	if err != nil {
		return fmt.Errorf("stage 1: load untested: %w", err)
	}
	priorAlive, err := selector.LoadAliveLinks(ctx, dbPath)
	if err != nil {
		return fmt.Errorf("stage 1: load alive: %w", err)
	}

	// Combine the two sets for stage 1. Dedup in case a once-alive config
	// somehow shows up in untested too (shouldn't, but defensive).
	seen := make(map[string]bool, len(untested)+len(priorAlive))
	stage1Links := make([]string, 0, len(untested)+len(priorAlive))
	for _, l := range append(untested, priorAlive...) {
		if !seen[l] {
			seen[l] = true
			stage1Links = append(stage1Links, l)
		}
	}

	if len(stage1Links) == 0 {
		slog.Warn("stage 1: nothing to test (no untested configs and no prior alive)")
		// Fall back to full-pool test
		slog.Info("stage 1: full-pool fallback (--from-db)")
		if err := opts.Runner.HTTPTest(ctx, xrayknife.HTTPOpts{
			Speedtest: false,
			Threads:   t1,
			Protocol:  "vless",
			DelayMs:   5000,
		}); err != nil {
			return fmt.Errorf("stage 1 (ping, full-pool): %w", err)
		}
	} else {
		stage1Tmp, err := os.CreateTemp("", "vlessfilter-stage1-*.txt")
		if err != nil {
			return fmt.Errorf("stage 1: temp file: %w", err)
		}
		defer os.Remove(stage1Tmp.Name())
		for _, link := range stage1Links {
			if _, err := fmt.Fprintln(stage1Tmp, link); err != nil {
				stage1Tmp.Close()
				return fmt.Errorf("stage 1: write temp: %w", err)
			}
		}
		stage1Tmp.Close()

		slog.Info("test stage 1: handshake/ping",
			"threads", t1, "untested", len(untested), "retest_alive", len(priorAlive),
			"total", len(stage1Links))
		if err := opts.Runner.HTTPTest(ctx, xrayknife.HTTPOpts{
			Speedtest: false,
			Threads:   t1,
			Protocol:  "vless",
			File:      stage1Tmp.Name(),
			DelayMs:   5000,
		}); err != nil {
			return fmt.Errorf("stage 1 (ping): %w", err)
		}
	}

	// Skip stage 2 if user explicitly disabled it.
	if t2 == 0 {
		slog.Info("stage 2 skipped (Threads2=0)")
		return nil
	}

	// Stage 2: speedtest the stage-1 survivors only. We need real HTTP
	// traffic flowing through each proxy, not just a TLS handshake — many
	// VLESS endpoints accept handshakes but refuse to forward traffic.
	//
	// xray-knife's --from-db doesn't filter to "previous run passed", so we
	// extract the alive set from the DB ourselves and feed it back via -f.
	// (dbPath already obtained above for stage 1.)
	aliveLinks, err := selector.LoadAliveLinks(ctx, dbPath)
	if err != nil {
		return fmt.Errorf("stage 2: load alive: %w", err)
	}
	if len(aliveLinks) == 0 {
		slog.Warn("stage 2 skipped: stage 1 produced 0 alive configs")
		return nil
	}
	tmp, err := os.CreateTemp("", "vlessfilter-alive-*.txt")
	if err != nil {
		return fmt.Errorf("stage 2: temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	for _, link := range aliveLinks {
		if _, err := fmt.Fprintln(tmp, link); err != nil {
			tmp.Close()
			return fmt.Errorf("stage 2: write temp: %w", err)
		}
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("stage 2: close temp: %w", err)
	}

	// Stage 2: speedtest 3 separate times (LIVE-01).
	//
	// Each invocation writes a fresh test_run row to the DB. The selector
	// then groups by config_link and applies "passes >= 2" to decide
	// alive vs flaky. This catches handshake-passes-but-no-real-traffic
	// false positives and flaky proxies that succeed once but fail next
	// time. xray-knife's internal --retries=2 fakes "1 pass on retry" as
	// a success; we want 3 INDEPENDENT runs to count distinct successes.
	const stage2Attempts = 3
	for attempt := 1; attempt <= stage2Attempts; attempt++ {
		slog.Info("test stage 2: speedtest attempt",
			"attempt", attempt, "of", stage2Attempts,
			"alive_input", len(aliveLinks), "threads", t2)
		if err := opts.Runner.HTTPTest(ctx, xrayknife.HTTPOpts{
			Speedtest: true,
			Threads:   t2,
			Protocol:  "vless",
			File:      tmp.Name(),
			DelayMs:   5000,
			Retries:   1, // small per-attempt retry; cross-attempt check via 3 runs below
		}); err != nil {
			// Tolerate intermittent failures of an individual attempt —
			// other attempts still produce useful evidence. Fail only if
			// none of the attempts produced any DB writes (caller can
			// detect via empty selector results).
			slog.Warn("stage 2 attempt failed (continuing)",
				"attempt", attempt, "err", err)
		}
	}
	return nil
}

// runSelect reads the xray-knife DB, computes selections, and writes ALL
// outputs (subs/, README.md, all-results.csv, raw/dead.txt).
//
// Tolerates an empty DB (logs warn, skips writing) so checkpoint loops can
// be called before stage 2 has any data.
//
// Country resolution: instead of trusting the latest test's reported country
// (which can rotate per-connection on proxy chains, load balancers, and
// Cloudflare Workers), we look at the FULL history of passes per config:
//   - Same country across all passes → "stable" → goes into subs/<CC>.txt
//   - Different countries across passes → "rotating" → goes into
//     subs/rotating.txt (still useful, just no country guarantee)
//   - Cloudflare Worker / Pages → always rotating regardless of history
func runSelect(ctx context.Context, opts Opts) error {
	dbPath, err := opts.Runner.DBPath()
	if err != nil {
		return err
	}
	stable, rotating, err := selector.LoadStableAndRotating(ctx, dbPath)
	if err != nil {
		return err
	}
	// We still need dead set for diagnostics; fetch them via the legacy
	// LoadAllResults (it returns alive+dead but we only use dead here).
	_, dead, _ := selector.LoadAllResults(ctx, dbPath)

	if len(stable) == 0 && len(rotating) == 0 {
		slog.Debug("select: no alive results yet (skipping output)", "db", dbPath)
		return nil
	}
	selections := selector.Top3PerCountry(stable)
	slog.Info("select complete",
		"countries", len(selections),
		"stable_alive", len(stable),
		"rotating_alive", len(rotating),
		"dead", len(dead))
	if err := output.WriteAll(opts.OutDir, selections, append(stable, rotating...), dead, rotating, opts.Now()); err != nil {
		return err
	}

	// GEO-04: post-publish accuracy probe. Sample N random keys per
	// country, route HTTP through them to ipinfo.io, compare actual
	// exit country to our published label. Logs per-country accuracy.
	//
	// Skipped for --profile dev (small subsets aren't worth probing).
	// Skipped on checkpoint runs (only end-of-run validation matters).
	if opts.RunAccuracyProbe {
		probe := &accuracy.Probe{
			SubsDir:       filepath.Join(opts.OutDir, "subs"),
			MaxPerCountry: 5,
			Threshold:     0.80,
			Timeout:       15 * time.Second,
		}
		if report, err := probe.Run(ctx); err != nil {
			slog.Warn("accuracy probe failed to run (publish proceeds)", "err", err)
		} else {
			accuracy.LogReport(report)
			if !report.Passed {
				slog.Error("accuracy below threshold; consider rolling back this run",
					"overall", fmt.Sprintf("%.1f%%", report.OverallPercent*100),
					"threshold", fmt.Sprintf("%.1f%%", report.Threshold*100))
			}
		}
	}
	return nil
}

// startCheckpointLoop launches a goroutine that, every CheckpointMin
// minutes, runs runSelect and (when GitPush is true) commits + pushes.
//
// Returns a stop function that the caller defers.
func startCheckpointLoop(ctx context.Context, opts Opts) func() {
	if opts.CheckpointMin <= 0 {
		return func() {}
	}
	interval := time.Duration(opts.CheckpointMin) * time.Minute
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-stopCh:
				return
			case <-t.C:
				slog.Info("checkpoint: writing partial output")
				if err := runSelect(ctx, opts); err != nil {
					slog.Warn("checkpoint runSelect failed", "error", err)
					continue
				}
				if opts.GitPush {
					commitPush(ctx, opts, fmt.Sprintf("checkpoint: results @ %s", opts.Now().Format(time.RFC3339)))
				}
			}
		}
	}()
	return func() {
		close(stopCh)
		wg.Wait()
	}
}

// finalCommitPush is the end-of-run commit. Best-effort: failures only logged.
func finalCommitPush(opts Opts) {
	if !opts.GitPush {
		return
	}
	commitPush(context.Background(), opts, fmt.Sprintf("vlessfilter run: results @ %s", opts.Now().Format(time.RFC3339)))
}

func commitPush(ctx context.Context, opts Opts, msg string) {
	committed, err := git.CommitAll(ctx, opts.GitRepoDir, msg)
	if err != nil {
		slog.Warn("git commit failed", "error", err)
		return
	}
	if !committed {
		slog.Debug("git: nothing to commit at this checkpoint")
		return
	}
	if err := git.Push(ctx, opts.GitRepoDir, opts.GitBranch, opts.GitToken); err != nil {
		slog.Warn("git push failed", "error", err)
		return
	}
	slog.Info("git push ok", "branch", opts.GitBranch, "msg", msg)
}
