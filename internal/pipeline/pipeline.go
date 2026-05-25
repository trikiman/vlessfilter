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
	"github.com/trikiman/vlessfilter/internal/prepublish"
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

	// Protocols controls which proxy schemes get tested + published.
	// Empty defaults to ["vless"] (legacy v1 behavior). Multi-protocol mode
	// (v2.0+) typically passes ["vless", "vmess", "trojan", "ss"]. Each
	// protocol gets its own test pass + its own subs/<proto>/ output dir.
	Protocols []string

	// SkipPrePublishProbe disables the immediately-before-publish probe
	// (re-tests each top-3 selection, drops dead, aborts publish if
	// drop-rate>75%). Default false (probe runs). Set true for checkpoint
	// runs which fire every 2 min and don't need the full re-validation.
	SkipPrePublishProbe bool
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
			// CRITICAL: use a FRESH ctx with its own timeout for the
			// fallback runSelect. The original ctx is already cancelled,
			// so calling runSelect with it would cancel before the
			// pre-publish probe could finish — leaving CHECKPOINT (and
			// thus unprobed) output as the final state.
			//
			// 5min budget for runSelect: typical probe of 108 keys at
			// 20 concurrency × 5s = ~30s; rest for I/O + diagnostics.
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Minute)
			_ = runSelect(cleanupCtx, opts)
			cleanupCancel()
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

	protocols := opts.Protocols
	if len(protocols) == 0 {
		protocols = []string{"vless"}
	}
	dbPath, err := opts.Runner.DBPath()
	if err != nil {
		return fmt.Errorf("test: db path: %w", err)
	}

	// Per-protocol test pass. Each protocol runs its own stage 1 (handshake)
	// and stage 2 (3x speedtest). xray-knife's --protocol flag filters input
	// configs by scheme, so we can't run all protocols in one invocation.
	//
	// Order matches selector.SupportedProtocols so logs and DB writes are
	// deterministic across runs.
	for _, proto := range protocols {
		slog.Info("test: starting protocol pass", "protocol", proto)
		if err := runTestProtocol(ctx, opts, proto, dbPath, t1, t2); err != nil {
			// Tolerate per-protocol failures: log + continue. A failure of
			// vmess shouldn't abort vless. Budget cancellation propagates
			// via ctx and stops the loop on next iteration.
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return err
			}
			slog.Warn("test: protocol pass failed (continuing other protocols)",
				"protocol", proto, "error", err)
		}
	}
	return nil
}

// runTestProtocol runs stage 1 + stage 2 for a single protocol. Extracted so
// runTest can iterate over multiple protocols in v2.0 multi-protocol mode.
func runTestProtocol(ctx context.Context, opts Opts, protocol, dbPath string, t1, t2 int) error {
	// Stage 1: TLS handshake against a manageable batch of NEW (untested)
	// configs PLUS retests of the existing alive set, instead of trying to
	// test the entire 1M+ pool and getting killed by the budget.
	//
	// LIVE-04: --profile dev passes opts.Limit > 0 to cap stage 1 to
	// a small subset for fast iteration.
	untestedBatch := 80000
	if opts.Limit > 0 && opts.Limit < untestedBatch {
		untestedBatch = opts.Limit
	}
	untested, err := selector.LoadUntestedLinks(ctx, dbPath, untestedBatch, protocol)
	if err != nil {
		return fmt.Errorf("stage 1 [%s]: load untested: %w", protocol, err)
	}
	priorAlive, err := selector.LoadAliveLinks(ctx, dbPath, protocol)
	if err != nil {
		return fmt.Errorf("stage 1 [%s]: load alive: %w", protocol, err)
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
		slog.Warn("stage 1: nothing to test (no untested configs and no prior alive)",
			"protocol", protocol)
		// Fall back to full-pool test
		slog.Info("stage 1: full-pool fallback (--from-db)", "protocol", protocol)
		if err := opts.Runner.HTTPTest(ctx, xrayknife.HTTPOpts{
			Speedtest: false,
			Threads:   t1,
			Protocol:  protocol,
			DelayMs:   5000,
		}); err != nil {
			return fmt.Errorf("stage 1 [%s] (ping, full-pool): %w", protocol, err)
		}
	} else {
		stage1Tmp, err := os.CreateTemp("", "vlessfilter-stage1-*.txt")
		if err != nil {
			return fmt.Errorf("stage 1 [%s]: temp file: %w", protocol, err)
		}
		defer os.Remove(stage1Tmp.Name())
		for _, link := range stage1Links {
			if _, err := fmt.Fprintln(stage1Tmp, link); err != nil {
				stage1Tmp.Close()
				return fmt.Errorf("stage 1 [%s]: write temp: %w", protocol, err)
			}
		}
		stage1Tmp.Close()

		slog.Info("test stage 1: handshake/ping",
			"protocol", protocol,
			"threads", t1, "untested", len(untested), "retest_alive", len(priorAlive),
			"total", len(stage1Links))
		if err := opts.Runner.HTTPTest(ctx, xrayknife.HTTPOpts{
			Speedtest: false,
			Threads:   t1,
			Protocol:  protocol,
			File:      stage1Tmp.Name(),
			DelayMs:   5000,
		}); err != nil {
			return fmt.Errorf("stage 1 [%s] (ping): %w", protocol, err)
		}
	}

	// Skip stage 2 if user explicitly disabled it.
	if t2 == 0 {
		slog.Info("stage 2 skipped (Threads2=0)", "protocol", protocol)
		return nil
	}

	// Stage 2: speedtest the stage-1 survivors only.
	aliveLinks, err := selector.LoadAliveLinks(ctx, dbPath, protocol)
	if err != nil {
		return fmt.Errorf("stage 2 [%s]: load alive: %w", protocol, err)
	}
	if len(aliveLinks) == 0 {
		slog.Warn("stage 2 skipped: stage 1 produced 0 alive configs", "protocol", protocol)
		return nil
	}
	tmp, err := os.CreateTemp("", "vlessfilter-alive-*.txt")
	if err != nil {
		return fmt.Errorf("stage 2 [%s]: temp file: %w", protocol, err)
	}
	defer os.Remove(tmp.Name())
	for _, link := range aliveLinks {
		if _, err := fmt.Fprintln(tmp, link); err != nil {
			tmp.Close()
			return fmt.Errorf("stage 2 [%s]: write temp: %w", protocol, err)
		}
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("stage 2 [%s]: close temp: %w", protocol, err)
	}

	// Stage 2: speedtest 3 separate times (LIVE-01).
	const stage2Attempts = 3
	for attempt := 1; attempt <= stage2Attempts; attempt++ {
		slog.Info("test stage 2: speedtest attempt",
			"protocol", protocol,
			"attempt", attempt, "of", stage2Attempts,
			"alive_input", len(aliveLinks), "threads", t2)
		if err := opts.Runner.HTTPTest(ctx, xrayknife.HTTPOpts{
			Speedtest: true,
			Threads:   t2,
			Protocol:  protocol,
			File:      tmp.Name(),
			DelayMs:   5000,
			Retries:   1,
		}); err != nil {
			slog.Warn("stage 2 attempt failed (continuing)",
				"protocol", protocol, "attempt", attempt, "err", err)
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

	protocols := opts.Protocols
	if len(protocols) == 0 {
		protocols = []string{"vless"}
	}

	// Aggregate everything across protocols for cross-protocol diagnostics
	// (all-results.csv + raw/dead.txt). Per-protocol output goes to
	// subs/<proto>/.
	var aggAllTested, aggDead []selector.Result

	// Per-protocol subscription writes. For VLESS specifically, also
	// mirror to subs/<CC>.txt (top-level) for back-compat with v1
	// subscription URLs.
	var perProto []protoOutput

	for _, proto := range protocols {
		stable, rotating, err := selector.LoadStableAndRotating(ctx, dbPath, proto)
		if err != nil {
			return fmt.Errorf("select [%s]: %w", proto, err)
		}
		selections := selector.Top3PerCountry(stable)
		slog.Info("select complete",
			"protocol", proto,
			"countries", len(selections),
			"stable_alive", len(stable),
			"rotating_alive", len(rotating))

		// PRE-PUBLISH PROBE — re-validate each top-3 selection RIGHT NOW
		// (not 50min ago when stage 2 ran). Drops keys that died between
		// stage 2 and publish. Without this, users see "80-90% timeout"
		// in their client because configs churn faster than our test cycle.
		//
		// Skipped for checkpoint runs (only end-of-run validation matters,
		// and probing on every checkpoint would 5x the run time).
		if !opts.SkipPrePublishProbe && len(selections) > 0 {
			probe := &prepublish.Probe{
				Timeout:     5 * time.Second,
				Concurrency: 20,
			}
			res := probe.Filter(ctx, selections)
			prepublish.LogResult(res)
			slog.Info("pre-publish probe summary",
				"protocol", proto,
				"input", res.InputKeys,
				"alive", res.AlivKeys,
				"drop_rate", fmt.Sprintf("%.1f%%", res.DropRate*100))

			// Critical drop rate: if more than 75% of our published
			// keys are dead at publish time, the stage-2 results are
			// stale enough that publishing would mislead users. Skip
			// publish — keep previous run live.
			const maxAllowedDropRate = 0.75
			if res.InputKeys >= 5 && res.DropRate > maxAllowedDropRate {
				slog.Error("pre-publish probe: drop rate exceeds threshold; ABORTING publish for this protocol",
					"protocol", proto,
					"drop_rate", fmt.Sprintf("%.1f%%", res.DropRate*100),
					"threshold", fmt.Sprintf("%.1f%%", maxAllowedDropRate*100))
				// Skip this protocol — leave existing subs/<proto>/ files
				// untouched on disk so previous (presumably better) run
				// stays published.
				continue
			}
			// Use filtered selections for output. May be smaller — that's
			// fine, fewer-but-honest keys is the goal.
			selections = res.Filtered
		}

		perProto = append(perProto, protoOutput{
			proto:      proto,
			selections: selections,
			stable:     stable,
			rotating:   rotating,
		})

		aggAllTested = append(aggAllTested, stable...)
		aggAllTested = append(aggAllTested, rotating...)
	}

	// Dead set is protocol-agnostic — we want to record all configs that
	// failed regardless of which protocol they were tested under.
	_, dead, _ := selector.LoadAllResults(ctx, dbPath)
	aggDead = dead

	totalCountries := 0
	for _, p := range perProto {
		totalCountries += len(p.selections)
	}
	if totalCountries == 0 && len(aggAllTested) == 0 {
		slog.Debug("select: no alive results yet across any protocol (skipping output)", "db", dbPath)
		return nil
	}

	// Write per-protocol subs/<proto>/ output. For vless, also mirror to
	// top-level subs/ for v1 URL back-compat.
	for _, p := range perProto {
		if err := output.WriteProtocol(opts.OutDir, p.proto, p.selections, p.rotating, opts.Now()); err != nil {
			return err
		}
		if p.proto == "vless" {
			// Back-compat: v1 URLs were subs/<CC>.txt, subs/all.txt,
			// subs/rotating.txt — keep them working by mirroring vless
			// output to the top-level subs/ dir.
			if err := output.Write(opts.OutDir, p.selections, p.rotating, opts.Now()); err != nil {
				return err
			}
		}
	}

	// Cross-protocol diagnostics + README at top level.
	if err := output.WriteDiagnostics(opts.OutDir, aggAllTested, aggDead); err != nil {
		return err
	}
	if err := output.WriteMultiProtocolReadme(opts.OutDir, perProtoReadmeData(perProto), opts.Now()); err != nil {
		return err
	}

	// GEO-04: post-publish accuracy probe (VLESS only for now —
	// per-protocol probing arrives in v2.1).
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

// perProtoReadmeData converts pipeline's internal protoOutput slice to the
// shape output.WriteMultiProtocolReadme expects (decoupling the two packages
// without exporting an internal type).
func perProtoReadmeData(perProto []protoOutput) []output.ProtoReadme {
	out := make([]output.ProtoReadme, 0, len(perProto))
	for _, p := range perProto {
		out = append(out, output.ProtoReadme{
			Protocol:   p.proto,
			Selections: p.selections,
			Rotating:   len(p.rotating),
		})
	}
	return out
}

// protoOutput is the per-protocol view collected by runSelect, kept at
// package scope so perProtoReadmeData can reference the type.
type protoOutput struct {
	proto      string
	selections []selector.CountrySelection
	stable     []selector.Result
	rotating   []selector.Result
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
				slog.Info("checkpoint: writing partial output (with pre-publish probe)")
				// CRITICAL: previous behavior was SkipPrePublishProbe=true
				// for checkpoints. That caused a race where, on budget
				// exhaustion, the LATEST CHECKPOINT (unprobed) was the
				// final published output. Users saw dead keys.
				//
				// Now: probe runs on every checkpoint. Probe time is
				// ~30s (108 keys × 5s with 20 concurrency); checkpoint
				// interval is 2min — overhead is acceptable.
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
