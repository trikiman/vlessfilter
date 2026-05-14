// Package kerntune applies Linux kernel tunables required for the high-
// concurrency stage 1 ping (1000+ outbound TCP/TLS handshakes in flight).
//
// Apply is best-effort: failures are logged at WARN and never returned.
// This keeps the pipeline running on misconfigured boxes (e.g., non-root
// user, locked-down sysctl) — degraded throughput is preferable to a fatal
// abort 30 seconds before the VPS auto-deletes.
//
// Build-tagged: real implementation only on Linux. Other GOOSes get a no-op.
package kerntune

import "context"

// Apply runs the platform-specific tuning. See *_linux.go and *_other.go.
// Always returns nil; signature kept error-returning so future fatal cases
// (e.g., refuse to run if some critical sysctl is impossible) can be added
// without rewiring callers.
func Apply(ctx context.Context) error {
	return apply(ctx)
}
