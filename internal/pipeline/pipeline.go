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
	"sync"
	"time"

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
	GitName    string
	GitEmail   string
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

	// Stage 1: TLS handshake against every config in the DB. Cheap, fast,
	// filters out the ~99% dead pool.
	//
	// Timeout 5000ms: 1s was too aggressive — many real working proxies
	// take 1-3s to complete handshake from this network. 5s gives reasonable
	// recall without much speed penalty (most dead configs fail at TCP
	// connect in <100ms regardless of timeout).
	slog.Info("test stage 1: handshake/ping", "threads", t1)
	if err := opts.Runner.HTTPTest(ctx, xrayknife.HTTPOpts{
		Speedtest: false,
		Threads:   t1,
		Protocol:  "vless",
		DelayMs:   5000,
	}); err != nil {
		return fmt.Errorf("stage 1 (ping): %w", err)
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
	dbPath, err := opts.Runner.DBPath()
	if err != nil {
		return fmt.Errorf("stage 2: %w", err)
	}
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

	slog.Info("test stage 2: speedtest on stage-1 survivors", "alive", len(aliveLinks), "threads", t2)
	if err := opts.Runner.HTTPTest(ctx, xrayknife.HTTPOpts{
		Speedtest: true,
		Threads:   t2,
		Protocol:  "vless",
		File:      tmp.Name(),
		DelayMs:   5000,
		Retries:   2,
	}); err != nil {
		return fmt.Errorf("stage 2 (speedtest): %w", err)
	}
	return nil
}

// runSelect reads the xray-knife DB, computes selections, and writes ALL
// outputs (subs/, README.md, all-results.csv, raw/dead.txt).
//
// Tolerates an empty DB (logs warn, skips writing) so checkpoint loops can
// be called before stage 2 has any data.
func runSelect(ctx context.Context, opts Opts) error {
	dbPath, err := opts.Runner.DBPath()
	if err != nil {
		return err
	}
	alive, dead, err := selector.LoadAllResults(ctx, dbPath)
	if err != nil {
		return err
	}
	if len(alive) == 0 {
		slog.Debug("select: no alive results yet (skipping output)", "db", dbPath)
		return nil
	}
	selections := selector.Top3PerCountry(alive)
	slog.Info("select complete", "countries", len(selections), "alive", len(alive), "dead", len(dead))
	return output.WriteAll(opts.OutDir, selections, alive, dead, opts.Now())
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
