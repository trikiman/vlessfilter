// Command vlessfilter discovers, tests, and publishes the top 3 fastest VLESS proxy
// keys per country. See README.md and .planning/PROJECT.md for the full design.
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
  --sources <path>   Path to sources.yaml (default: ./sources.yaml)
  --out <dir>        Output directory for subs/ and README.md (default: .)
  --stage <name>     Run only one stage: fetch | test | select (default: all stages)
  --threads1 <n>     Stage 1 (handshake) concurrency (default: 200; Phase 2 will raise to 1000)
  --threads2 <n>     Stage 2 (speedtest) concurrency (default: 20; cap per D-06)
  --limit <n>        Cap number of keys tested in stage 2 (default: 0 = no cap)

Examples:
  vlessfilter run                                # full pipeline against ./sources.yaml
  vlessfilter run --stage fetch                  # only ingest sources into xray-knife db
  vlessfilter run --stage select --out results   # regenerate output files only

Project: https://github.com/trikiman/vlessfilter
`

// Exit codes:
//
//	0 — success
//	1 — user/config error (bad flag, missing file, malformed yaml)
//	2 — runtime error (xray-knife failed, network down, db corrupt)
const (
	exitOK      = 0
	exitUserErr = 1
	exitRuntime = 2
)

func main() {
	os.Exit(run())
}

func run() int {
	// Top-level flag set: only --verbose lives here; subcommand flags are parsed below.
	root := flag.NewFlagSet("vlessfilter", flag.ContinueOnError)
	verbose := root.Bool("verbose", false, "Enable debug logging")
	root.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	// Stop the root flag set at the first non-flag arg so the subcommand sees its own flags.
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

// runCmd parses run-specific flags and invokes the pipeline.
func runCmd(args []string) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	sources := fs.String("sources", "./sources.yaml", "Path to sources.yaml")
	outDir := fs.String("out", ".", "Output directory for subs/ and README.md")
	stage := fs.String("stage", "", "Run only one stage: fetch | test | select")
	threads1 := fs.Int("threads1", 200, "Stage 1 (handshake) concurrency")
	threads2 := fs.Int("threads2", 20, "Stage 2 (speedtest) concurrency (cap recommended at 20)")
	limit := fs.Int("limit", 0, "Cap number of keys tested in stage 2 (0 = no cap)")

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

	opts := pipeline.Opts{
		SourcesPath: *sources,
		OutDir:      *outDir,
		Stage:       *stage,
		Threads1:    *threads1,
		Threads2:    *threads2,
		Limit:       *limit,
		Runner:      runner,
		Now:         time.Now().UTC,
	}
	if err := pipeline.Run(ctx, opts); err != nil {
		slog.Error("pipeline failed", "error", err)
		return exitRuntime
	}
	return exitOK
}

// setupLogging configures slog with a text handler at the chosen level.
func setupLogging(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(h))
}

// splitGlobal returns args before the first non-flag positional argument.
// It lets the root flag set parse only --verbose-style global flags and stop
// at the subcommand name, which then parses its own flags.
func splitGlobal(args []string) []string {
	for i, a := range args {
		if !strings.HasPrefix(a, "-") {
			return args[:i]
		}
	}
	return args
}

// afterGlobal returns args from the first non-flag positional argument onward.
func afterGlobal(args []string) []string {
	for i, a := range args {
		if !strings.HasPrefix(a, "-") {
			return args[i:]
		}
	}
	return nil
}
