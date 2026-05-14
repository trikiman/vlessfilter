//go:build !linux

package kerntune

import (
	"context"
	"log/slog"
	"runtime"
)

func apply(_ context.Context) error {
	slog.Debug("kerntune: no-op on this platform", "goos", runtime.GOOS)
	return nil
}
