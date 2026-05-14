// Package xrayknife wraps the xray-knife binary as a Go subprocess (per D-01).
//
// The Runner interface keeps the wrapper testable: the production
// implementation (RealRunner) shells out to xray-knife; the test double
// (FakeRunner, in xrayknife_test.go) records call args without doing I/O.
//
// Auto-install (per D-02): EnsureInstalled checks for the binary on $PATH and,
// if missing, runs `go install github.com/lilendian0x00/xray-knife/v9@<pin>`.
package xrayknife

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// XrayKnifeVersion pins the engine version. Phase 2 may bump to a specific
// stable tag once we benchmark.
const XrayKnifeVersion = "latest"

// xrayKnifeModule is the Go module path used for `go install`.
const xrayKnifeModule = "github.com/lilendian0x00/xray-knife/v9"

// HTTPOpts are the flags passed to `xray-knife http`.
type HTTPOpts struct {
	Speedtest bool
	Threads   int    // -t flag (0 = let xray-knife pick its default)
	Limit     int    // --limit flag (0 = no limit)
	Protocol  string // --protocol flag (default "vless")
}

// Runner abstracts xray-knife so tests can substitute a fake.
type Runner interface {
	// Available reports whether the xray-knife binary is on PATH and runnable.
	Available(ctx context.Context) (bool, error)
	// Install runs `go install <module>@<XrayKnifeVersion>`.
	Install(ctx context.Context) error
	// SubsAdd registers a subscription URL with the given remark name.
	// Idempotent: returns nil if xray-knife reports the URL already exists.
	SubsAdd(ctx context.Context, url, remark string) error
	// SubsFetch fetches all subscriptions and populates xray-knife's local DB.
	SubsFetch(ctx context.Context) error
	// HTTPTest runs a test pass against configs already in the DB.
	HTTPTest(ctx context.Context, opts HTTPOpts) error
	// SubCount returns the number of subscription rows in the xray-knife DB.
	// Used by the pipeline to refuse stage "test" when no subs have been added.
	SubCount(ctx context.Context) (int, error)
	// DBPath returns the absolute path to ~/.xray-knife/xray-knife.db.
	DBPath() (string, error)
}

// RealRunner shells out to the real xray-knife binary.
type RealRunner struct{}

// NewRealRunner constructs a RealRunner. Exists as a function for test
// symmetry with future runners (e.g., a CSV-output runner).
func NewRealRunner() *RealRunner { return &RealRunner{} }

// Available runs `xray-knife --version` and reports success.
func (r *RealRunner) Available(ctx context.Context) (bool, error) {
	if _, err := exec.LookPath("xray-knife"); err != nil {
		return false, nil //nolint:nilerr // missing-on-PATH is reported as ok=false, not error
	}
	cmd := exec.CommandContext(ctx, "xray-knife", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("xray-knife --version failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	slog.Debug("xray-knife available", "version_output", strings.TrimSpace(string(out)))
	return true, nil
}

// Install runs `go install <module>@<XrayKnifeVersion>`.
func (r *RealRunner) Install(ctx context.Context) error {
	if _, err := exec.LookPath("go"); err != nil {
		return errors.New("go binary not found on PATH; install Go 1.22+ from https://go.dev")
	}
	target := xrayKnifeModule + "@" + XrayKnifeVersion
	slog.Info("installing xray-knife via go install", "target", target)
	cmd := exec.CommandContext(ctx, "go", "install", target)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go install %s: %w", target, err)
	}
	return nil
}

// SubsAdd is idempotent: error messages indicating a duplicate URL are treated as success.
// xray-knife's exact wording varies across versions; we match a few known patterns.
func (r *RealRunner) SubsAdd(ctx context.Context, url, remark string) error {
	cmd := exec.CommandContext(ctx, "xray-knife", "subs", "add", "--url", url, "--remark", remark)
	out, err := cmd.CombinedOutput()
	combined := strings.ToLower(string(out))
	if err != nil {
		// xray-knife v9.12 emits "UNIQUE constraint failed: subscriptions.url"; older versions said "already exists".
		switch {
		case strings.Contains(combined, "already exists"),
			strings.Contains(combined, "duplicate"),
			strings.Contains(combined, "unique constraint failed"):
			slog.Debug("subs add: idempotent (already exists)", "remark", remark)
			return nil
		}
		return fmt.Errorf("xray-knife subs add %q: %w (output: %s)", remark, err, strings.TrimSpace(string(out)))
	}
	slog.Debug("subs add ok", "remark", remark)
	return nil
}

// SubsFetch streams xray-knife's progress to slog.Info; this can take 10-30s.
//
// xray-knife exits 1 if ANY subscription fails. Our pipeline policy is more
// tolerant: as long as the local DB has subscriptions, partial failures are
// logged as warnings, not pipeline-fatal errors. This handles dead URLs
// (e.g., upstream repo moved) without aborting an otherwise successful run.
func (r *RealRunner) SubsFetch(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "xray-knife", "subs", "fetch", "--all")
	out, err := cmd.CombinedOutput()
	// Always show xray-knife's progress to the user.
	if len(out) > 0 {
		os.Stderr.Write(out)
	}
	if err == nil {
		slog.Info("xray-knife subs fetch complete (all subscriptions ok)")
		return nil
	}
	combined := string(out)
	// "All done: N links fetched" appears when at least one subscription
	// succeeded. If we see it, treat exit-1 as a partial-success warning.
	if strings.Contains(combined, "All done:") || strings.Contains(combined, "configs saved") {
		slog.Warn("xray-knife subs fetch: partial failure tolerated", "exit_status", err.Error())
		return nil
	}
	return fmt.Errorf("xray-knife subs fetch: %w", err)
}

// HTTPTest runs `xray-knife http --from-db --protocol <p> [--speedtest] [-t N] [--limit N]`.
func (r *RealRunner) HTTPTest(ctx context.Context, opts HTTPOpts) error {
	args := []string{"http", "--from-db"}
	proto := opts.Protocol
	if proto == "" {
		proto = "vless"
	}
	args = append(args, "--protocol", proto)
	if opts.Speedtest {
		args = append(args, "--speedtest")
	}
	if opts.Threads > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", opts.Threads))
	}
	if opts.Limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", opts.Limit))
	}
	slog.Info("xray-knife http", "args", args, "speedtest", opts.Speedtest)
	cmd := exec.CommandContext(ctx, "xray-knife", args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xray-knife http %v: %w", args, err)
	}
	return nil
}

// SubCount queries `xray-knife subs show` and counts subscription rows.
// The exact output format is fragile; we count any line that contains "http"
// (a URL signature). Plan 02-02 may improve this once we read the DB directly.
func (r *RealRunner) SubCount(ctx context.Context) (int, error) {
	cmd := exec.CommandContext(ctx, "xray-knife", "subs", "show")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("xray-knife subs show: %w", err)
	}
	count := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(strings.ToLower(line), "http") {
			count++
		}
	}
	return count, nil
}

// DBPath returns the standard xray-knife DB location.
//
// On Linux/macOS: $HOME/.xray-knife/xray-knife.db
// Windows: %USERPROFILE%/.xray-knife/xray-knife.db (Phase 1 doesn't deploy on Windows;
// Phase 2 ephemeral VPS is Linux only. This is a TODO for portability.)
func (r *RealRunner) DBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}
	return filepath.Join(home, ".xray-knife", "xray-knife.db"), nil
}

// EnsureInstalled returns nil if xray-knife is already on PATH, or installs it.
//
// The contract: after EnsureInstalled returns nil, runner.Available must be
// true. If install succeeds but the binary still isn't on PATH (typical when
// $GOPATH/bin or $GOBIN isn't on PATH), we return a clear error pointing the
// user at the fix.
func EnsureInstalled(ctx context.Context, r Runner) error {
	ok, err := r.Available(ctx)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	slog.Info("xray-knife not found — auto-installing")
	if err := r.Install(ctx); err != nil {
		return err
	}
	ok, err = r.Available(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("xray-knife installation reported success but binary not found on PATH; ensure $GOPATH/bin or $GOBIN is on PATH (typically $HOME/go/bin)")
	}
	return nil
}
