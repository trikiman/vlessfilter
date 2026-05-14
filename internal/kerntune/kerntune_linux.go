//go:build linux

package kerntune

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"syscall"
)

// targetNoFile is the file-descriptor ceiling we want for stage 1.
// 100k accommodates 1000 concurrent connections × ~3 fds each plus headroom.
const targetNoFile = 100000

func apply(ctx context.Context) error {
	raiseRLimit()
	trySysctl(ctx, "net.ipv4.tcp_tw_reuse", "1")
	trySysctl(ctx, "net.ipv4.ip_local_port_range", "1024 65535")
	return nil
}

// raiseRLimit best-effort raises RLIMIT_NOFILE for the current process.
// On most modern Linux it can be raised up to the hard limit without root.
func raiseRLimit() {
	var rl syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rl); err != nil {
		slog.Warn("kerntune: getrlimit failed", "error", err)
		return
	}
	if rl.Cur >= targetNoFile {
		slog.Debug("kerntune: RLIMIT_NOFILE already high enough", "cur", rl.Cur, "max", rl.Max)
		return
	}
	want := uint64(targetNoFile)
	if want > rl.Max {
		want = rl.Max
	}
	old := rl.Cur
	rl.Cur = want
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rl); err != nil {
		slog.Warn("kerntune: setrlimit failed (try `ulimit -n 100000` before running)", "wanted", want, "had", old, "error", err)
		return
	}
	slog.Info("kerntune: RLIMIT_NOFILE raised", "from", old, "to", want)
}

// trySysctl runs `sysctl -w key="value"` and logs but never errors.
func trySysctl(ctx context.Context, key, value string) {
	cmd := exec.CommandContext(ctx, "sysctl", "-w", fmt.Sprintf("%s=%s", key, value))
	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("kerntune: sysctl failed (often requires root)", "key", key, "value", value, "error", err, "stderr", strings.TrimSpace(string(out)))
		return
	}
	slog.Info("kerntune: sysctl applied", "key", key, "value", value)
}
