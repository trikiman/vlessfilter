package kerntune

import (
	"context"
	"testing"
)

// TestApply_NeverErrors documents the best-effort contract: Apply always
// returns nil even when individual tunables fail.
func TestApply_NeverErrors(t *testing.T) {
	if err := Apply(context.Background()); err != nil {
		t.Errorf("Apply must be best-effort (always nil), got: %v", err)
	}
}
