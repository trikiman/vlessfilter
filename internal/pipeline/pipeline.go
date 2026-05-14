// Package pipeline orchestrates the end-to-end run: load sources →
// xray-knife subs add+fetch → stage 1 ping → stage 2 speedtest → selector →
// output.
//
// Stages can be invoked independently for debugging via Opts.Stage.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/trikiman/vlessfilter/internal/output"
	"github.com/trikiman/vlessfilter/internal/selector"
	"github.com/trikiman/vlessfilter/internal/sources"
	"github.com/trikiman/vlessfilter/internal/xrayknife"
)

// Opts controls a pipeline run. Runner is required; the rest have sensible
// defaults applied in Run.
type Opts struct {
	SourcesPath string
	OutDir      string
	Stage       string // "" | "fetch" | "test" | "select"
	Threads1    int
	Threads2    int
	Limit       int
	Runner      xrayknife.Runner
	Now         func() time.Time // injectable for tests
}

// validStages enumerates accepted Stage values.
var validStages = map[string]bool{
	"":       true,
	"fetch":  true,
	"test":   true,
	"select": true,
}

// Run executes the pipeline according to opts.Stage. Empty stage runs all
// three stages in order.
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

	// Stage "fetch"
	if opts.Stage == "" || opts.Stage == "fetch" {
		if err := runFetch(ctx, opts); err != nil {
			return fmt.Errorf("stage fetch: %w", err)
		}
		if opts.Stage == "fetch" {
			return nil
		}
	}

	// Stage "test" (handshake + speedtest)
	if opts.Stage == "" || opts.Stage == "test" {
		if err := runTest(ctx, opts); err != nil {
			return fmt.Errorf("stage test: %w", err)
		}
		if opts.Stage == "test" {
			return nil
		}
	}

	// Stage "select" (read DB, group, top-3, write outputs)
	if opts.Stage == "" || opts.Stage == "select" {
		if err := runSelect(ctx, opts); err != nil {
			return fmt.Errorf("stage select: %w", err)
		}
	}
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
	count, err := opts.Runner.SubCount(ctx)
	if err != nil {
		slog.Warn("could not query sub count after fetch", "error", err)
	} else {
		slog.Info("fetch complete", "subscriptions_in_db", count)
	}
	return nil
}

func runTest(ctx context.Context, opts Opts) error {
	count, err := opts.Runner.SubCount(ctx)
	if err == nil && count == 0 {
		return errors.New("no subscriptions in xray-knife db; run 'vlessfilter run --stage fetch' first")
	}

	t1 := opts.Threads1
	if t1 <= 0 {
		t1 = 200
	}
	t2 := opts.Threads2
	if t2 <= 0 {
		t2 = 20
	}

	slog.Info("test stage 1: handshake/ping", "threads", t1)
	if err := opts.Runner.HTTPTest(ctx, xrayknife.HTTPOpts{Speedtest: false, Threads: t1, Protocol: "vless"}); err != nil {
		return fmt.Errorf("stage 1 (ping): %w", err)
	}

	slog.Info("test stage 2: speedtest", "threads", t2, "limit", opts.Limit)
	if err := opts.Runner.HTTPTest(ctx, xrayknife.HTTPOpts{Speedtest: true, Threads: t2, Limit: opts.Limit, Protocol: "vless"}); err != nil {
		return fmt.Errorf("stage 2 (speedtest): %w", err)
	}
	return nil
}

func runSelect(ctx context.Context, opts Opts) error {
	dbPath, err := opts.Runner.DBPath()
	if err != nil {
		return err
	}
	results, err := selector.LoadResults(ctx, dbPath)
	if err != nil {
		return err
	}
	selections := selector.Top3PerCountry(results)
	slog.Info("select complete", "countries", len(selections), "total_keys", countKeys(selections))
	if err := output.Write(opts.OutDir, selections, opts.Now()); err != nil {
		return err
	}
	slog.Info("output written", "out_dir", opts.OutDir, "subs_files", len(selections))
	return nil
}

func countKeys(sel []selector.CountrySelection) int {
	total := 0
	for _, cs := range sel {
		total += len(cs.Top)
	}
	return total
}
