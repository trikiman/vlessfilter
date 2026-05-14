// Command vlessfilter discovers, tests, and publishes the top 3 fastest VLESS
// proxy keys per country. See README.md and .planning/PROJECT.md for design.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/trikiman/vlessfilter/internal/pipeline"
	"github.com/trikiman/vlessfilter/internal/xrayknife"
)

const usage = `vlessfilter — top-3 VLESS keys per country, auto-curated.

Usage:
  vlessfilter [global flags] <command> [command flags]

Commands:
  run        Run the full pipeline (or a single stage with --stage)

Global flags:
  --verbose  Enable debug logging

Run flags:
  --sources <path>      Path to sources.yaml (default: ./sources.yaml)
  --out <dir>           Output directory for subs/, README.md, all-results.csv (default: .)
  --stage <name>        Run only one stage: fetch | test | select (default: all stages)
  --threads1 <n>        Stage 1 (handshake) concurrency (default: 500)
  --threads2 <n>        Stage 2 (speedtest) concurrency (default: 20; cap per D-06)
  --limit <n>           Cap number of keys tested in stage 2 (default: 0 = no cap)
  --budget-min <n>      Hard wall-clock budget for the run, minutes (default: 55; <=0 disables)
  --checkpoint-min <n>  Checkpoint cadence: write outputs every N minutes (default: 2; <=0 disables)
  --git-push            Commit + push results after each checkpoint and at end (default: false)
  --git-repo <dir>      Git repo dir to commit in (default: .)
  --git-branch <name>   Branch to push to (default: main)

Environment:
  GH_TOKEN              GitHub PAT used for --git-push (no leak: never written to ~/.gitconfig)

Examples:
  vlessfilter run                                # full pipeline locally
  vlessfilter run --stage fetch                  # only ingest sources
  vlessfilter run --stage select --out results   # regenerate output files only
  GH_TOKEN=ghp_xxx vlessfilter run --git-push    # ephemeral-VPS-style: commit results

Project: https://github.com/trikiman/vlessfilter
`

const (
	exitOK      = 0
	exitUserErr = 1
	exitRuntime = 2
)

func main() {
	os.Exit(run())
}

func run() int {
	root := flag.NewFlagSet("vlessfilter", flag.ContinueOnError)
	verbose := root.Bool("verbose", false, "Enable debug logging")
	root.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	if err := root.Parse(splitGlobal(os.Args[1:])); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUserErr
	}

	setupLogging(*verbose)

	args := afterGlobal(os.Args[1:])
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return exitUserErr
	}

	switch args[0] {
	case "run":
		return runCmd(args[1:])
	case "help", "-h", "--help":
		fmt.Fprint(os.Stderr, usage)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n%s", args[0], usage)
		return exitUserErr
	}
}

func runCmd(args []string) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	sources := fs.String("sources", "./sources.yaml", "Path to sources.yaml")
	outDir := fs.String("out", ".", "Output directory")
	stage := fs.String("stage", "", "Run only one stage: fetch | test | select")
	threads1 := fs.Int("threads1", 500, "Stage 1 concurrency")
	threads2 := fs.Int("threads2", 20, "Stage 2 concurrency")
	limit := fs.Int("limit", 0, "Cap stage-2 key count (0 = no cap)")
	budgetMin := fs.Int("budget-min", 55, "Wall-clock budget in minutes (<=0 disables)")
	checkpointMin := fs.Int("checkpoint-min", 2, "Checkpoint cadence in minutes (<=0 disables)")
	gitPush := fs.Bool("git-push", false, "Commit + push results to git")
	gitRepo := fs.String("git-repo", ".", "Git repo dir for --git-push")
	gitBranch := fs.String("git-branch", "main", "Branch to push to")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUserErr
	}

	switch *stage {
	case "", "fetch", "test", "select":
	default:
		fmt.Fprintf(os.Stderr, "invalid --stage value %q (must be empty, fetch, test, or select)\n", *stage)
		return exitUserErr
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	runner := xrayknife.NewRealRunner()
	if err := xrayknife.EnsureInstalled(ctx, runner); err != nil {
		slog.Error("xray-knife not available", "error", err)
		return exitRuntime
	}

	token := os.Getenv("GH_TOKEN")
	if *gitPush && token == "" {
		slog.Warn("--git-push set but GH_TOKEN env is empty; push will likely fail unless repo permits anonymous push")
	}

	opts := pipeline.Opts{
		SourcesPath:   *sources,
		OutDir:        *outDir,
		Stage:         *stage,
		Threads1:      *threads1,
		Threads2:      *threads2,
		Limit:         *limit,
		BudgetMin:     *budgetMin,
		CheckpointMin: *checkpointMin,
		GitPush:       *gitPush,
		GitRepoDir:    *gitRepo,
		GitBranch:     *gitBranch,
		GitToken:      token,
		Runner:        runner,
		Now:           func() time.Time { return time.Now().UTC() },
	}
	if err := pipeline.Run(ctx, opts); err != nil {
		slog.Error("pipeline failed", "error", err)
		return exitRuntime
	}
	return exitOK
}

func setupLogging(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(h))
}

func splitGlobal(args []string) []string {
	for i, a := range args {
		if !strings.HasPrefix(a, "-") {
			return args[:i]
		}
	}
	return args
}

func afterGlobal(args []string) []string {
	for i, a := range args {
		if !strings.HasPrefix(a, "-") {
			return args[i:]
		}
	}
	return nil
}
