// Command vlessfilter discovers, tests, and publishes the top 3 fastest VLESS
// proxy keys per country. See README.md and .planning/PROJECT.md for design.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/trikiman/vlessfilter/internal/output"
	"github.com/trikiman/vlessfilter/internal/pipeline"
	"github.com/trikiman/vlessfilter/internal/selector"
	"github.com/trikiman/vlessfilter/internal/sources"
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
  --threads1 <n>        Stage 1 (handshake) concurrency (default: 1000)
  --threads2 <n>        Stage 2 (speedtest) concurrency (default: 20; cap per D-06)
  --untested-batch <n>  Cap untested keys per protocol per stage 1 (default: 0 = built-in 80000)
  --min-speed <mbps>    Min Stage-2 speed to publish a stable key (default: 12; 0 disables)
  --stage2-passes <n>   Number of Stage 2 speedtest passes (default: 3)
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
	case "sources-list":
		return sourcesListCmd(args[1:])
	case "regen-readme":
		return regenReadmeCmd(args[1:])
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
	threads1 := fs.Int("threads1", 1000, "Stage 1 concurrency")
	threads2 := fs.Int("threads2", 20, "Stage 2 concurrency")
	limit := fs.Int("limit", 0, "Cap stage-2 key count (0 = no cap)")
	budgetMin := fs.Int("budget-min", 55, "Wall-clock budget in minutes (<=0 disables)")
	checkpointMin := fs.Int("checkpoint-min", 2, "Checkpoint cadence in minutes (<=0 disables)")
	gitPush := fs.Bool("git-push", false, "Commit + push results to git")
	gitRepo := fs.String("git-repo", ".", "Git repo dir for --git-push")
	gitBranch := fs.String("git-branch", "main", "Branch to push to")
	profile := fs.String("profile", "", "Preset for fast iteration: 'dev' = small subset, 2-min budget, dev/ output")
	accuracyProbe := fs.Bool("accuracy-probe", false, "After publish, sample-test keys against ipinfo.io and compare to published country labels (GEO-04)")
	protocols := fs.String("protocols", "vless,vmess,trojan,ss", "Comma-separated proxy protocols to test+publish (default: all 4 supported)")
	untestedBatch := fs.Int("untested-batch", 0, "Cap untested keys per protocol per stage 1 run (0 = use built-in default 80000)")
	minSpeed := fs.Float64("min-speed", 12, "Minimum Stage-2 speed in Mbps; stable keys slower than this are not published (0 = disable). 12 ~ YouTube 1080p at 2x speed with headroom.")
	stage2Passes := fs.Int("stage2-passes", 3, "Number of Stage 2 speedtest passes per key (higher = more stable speed measurement)")

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

	// LIVE-04: --profile dev preset for fast iteration. Caps everything
	// so the full pipeline finishes in <2 minutes wall time on a small
	// subset, writing to dev/ instead of polluting the real subs/ output.
	if *profile == "dev" {
		*outDir = "./dev"
		*budgetMin = 5
		*threads1 = 200
		*threads2 = 10
		*limit = 500
		slog.Info("--profile dev: small-subset run for filter-logic verification",
			"out_dir", *outDir, "budget_min", *budgetMin,
			"threads1", *threads1, "threads2", *threads2, "limit_stage2", *limit)
	}

	// Parse --protocols into a slice. Validate against the supported set.
	var protoList []string
	for _, p := range strings.Split(*protocols, ",") {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}
		switch p {
		case "vless", "vmess", "trojan", "ss":
			protoList = append(protoList, p)
		default:
			fmt.Fprintf(os.Stderr, "invalid --protocols value %q (allowed: vless, vmess, trojan, ss)\n", p)
			return exitUserErr
		}
	}
	if len(protoList) == 0 {
		protoList = []string{"vless"}
	}

	opts := pipeline.Opts{
		SourcesPath:      *sources,
		OutDir:           *outDir,
		Stage:            *stage,
		Threads1:         *threads1,
		Threads2:         *threads2,
		Limit:            *limit,
		BudgetMin:        *budgetMin,
		CheckpointMin:    *checkpointMin,
		GitPush:          *gitPush,
		GitRepoDir:       *gitRepo,
		GitBranch:        *gitBranch,
		GitToken:         token,
		RunAccuracyProbe: *accuracyProbe,
		Protocols:        protoList,
		UntestedBatch:    *untestedBatch,
		MinSpeedMbps:     *minSpeed,
		Stage2Passes:     *stage2Passes,
		Runner:           runner,
		Now:              func() time.Time { return time.Now().UTC() },
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


// sourcesListCmd loads sources.yaml, expands templates, and dumps URLs to stdout.
// Output formats: --format=plain (default; one URL per line) or --format=name-url
// (tab-separated source-name + URL — useful for debugging which source a URL came from).
//
// Used to materialize sources.txt — a human-readable manifest of every
// subscription URL the pipeline pulls from. Helpful for:
//   - diff'ing source sets between commits
//   - manual probing of a specific source
//   - re-creating the source set on a fresh machine without the YAML
func sourcesListCmd(args []string) int {
	fs := flag.NewFlagSet("sources-list", flag.ContinueOnError)
	srcPath := fs.String("sources", "./sources.yaml", "Path to sources.yaml")
	format := fs.String("format", "plain", "Output format: plain (URL per line) | name-url (tab-separated)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUserErr
	}
	cfg, err := sources.Load(*srcPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load sources: %v\n", err)
		return exitRuntime
	}
	expanded := sources.Expand(cfg)
	for _, e := range expanded {
		switch *format {
		case "plain":
			fmt.Println(e.URL)
		case "name-url":
			fmt.Printf("%s\t%s\n", e.Name, e.URL)
		default:
			fmt.Fprintf(os.Stderr, "unknown --format %q\n", *format)
			return exitUserErr
		}
	}
	return exitOK
}

// regenReadmeCmd reads subs/<proto>/_readme-data.json sidecars (written by
// each protocol's matrix job) and regenerates README.md with all 4
// protocols' tables. Used in the merge-and-push job after artifact merge.
//
// Without this, the matrix-merge runner has no xray-knife.db to derive
// fresh selections from, so README would be stale (showing only one
// protocol's data, or worse, the previous run's data).
// protocolIsPublished reports whether subs/<proto>/all.txt exists with keys in
// it, and how many. Used to tell "this protocol legitimately has nothing to
// show" apart from "we lost its data and are about to publish a README that
// hides it".
func protocolIsPublished(outDir, proto string) (bool, int) {
	raw, err := os.ReadFile(filepath.Join(outDir, "subs", proto, "all.txt"))
	if err != nil {
		return false, 0
	}
	n := 0
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.Contains(line, "://") {
			n++
		}
	}
	return n > 0, n
}

func regenReadmeCmd(args []string) int {
	fs := flag.NewFlagSet("regen-readme", flag.ContinueOnError)
	outDir := fs.String("out", ".", "Output directory (must contain subs/)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUserErr
	}

	type sidecar struct {
		Protocol    string                      `json:"protocol"`
		Selections  []selector.CountrySelection `json:"selections"`
		Rotating    int                         `json:"rotating"`
		GeneratedAt time.Time                   `json:"generated_at"`
	}

	// Iterate canonical protocol order so the README has stable section order.
	var protos []output.ProtoReadme
	var latest time.Time
	for _, proto := range []string{"vless", "vmess", "trojan", "ss"} {
		path := filepath.Join(*outDir, ".readme-data", proto+".json")
		raw, err := os.ReadFile(path)
		if err != nil {
			// A slog warning scrolls past in CI. If this protocol still has
			// published keys, the README is about to drop its section and its
			// subscription URLs while those keys remain live — readers simply
			// never learn the protocol exists. Annotate so it surfaces in the
			// Actions summary.
			if published, n := protocolIsPublished(*outDir, proto); published {
				fmt.Printf("::warning::README will omit %s despite %d published keys — sidecar missing at %s\n",
					proto, n, path)
			}
			slog.Warn("regen-readme: sidecar missing, protocol omitted from README",
				"protocol", proto, "path", path)
			continue
		}
		var sc sidecar
		if err := json.Unmarshal(raw, &sc); err != nil {
			slog.Warn("regen-readme: sidecar parse failed, omitting protocol",
				"protocol", proto, "error", err)
			continue
		}
		protos = append(protos, output.ProtoReadme{
			Protocol:   sc.Protocol,
			Selections: sc.Selections,
			Rotating:   sc.Rotating,
		})
		if sc.GeneratedAt.After(latest) {
			latest = sc.GeneratedAt
		}
	}
	if len(protos) == 0 {
		fmt.Fprintln(os.Stderr, "regen-readme: no _readme-data.json sidecars found in subs/")
		return exitRuntime
	}
	if latest.IsZero() {
		latest = time.Now().UTC()
	}
	if err := output.WriteMultiProtocolReadme(*outDir, protos, latest); err != nil {
		fmt.Fprintf(os.Stderr, "regen-readme: write failed: %v\n", err)
		return exitRuntime
	}
	slog.Info("regen-readme ok", "protocols", len(protos), "generated_at", latest.Format(time.RFC3339))
	return exitOK
}
