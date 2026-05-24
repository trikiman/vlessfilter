package prepublish

import (
	"context"
	"testing"
	"time"

	"github.com/trikiman/vlessfilter/internal/selector"
)

// TestFilter_EmptyInput verifies graceful handling of empty selections.
func TestFilter_EmptyInput(t *testing.T) {
	p := &Probe{Timeout: 100 * time.Millisecond}
	res := p.Filter(context.Background(), nil)
	if res.InputKeys != 0 {
		t.Errorf("InputKeys=%d, want 0", res.InputKeys)
	}
	if len(res.Filtered) != 0 {
		t.Errorf("Filtered len=%d, want 0", len(res.Filtered))
	}
}

// TestFilter_AllFailDropsCountries verifies that countries with all-dead
// keys get removed from the filtered output entirely.
//
// Uses a probe binary that always fails (we set XrayKnifeBin to a name
// that doesn't exist; exec returns error; probeOne returns false).
func TestFilter_AllFailDropsCountries(t *testing.T) {
	selections := []selector.CountrySelection{
		{
			Country: "US",
			Top: []selector.Result{
				{Link: "vless://aaa@example.com:443"},
				{Link: "vless://bbb@example.com:443"},
			},
		},
		{
			Country: "DE",
			Top: []selector.Result{
				{Link: "vless://ccc@example.com:443"},
			},
		},
	}
	p := &Probe{
		XrayKnifeBin: "definitely-not-a-real-binary-12345",
		Timeout:      100 * time.Millisecond,
		Concurrency:  4,
	}
	res := p.Filter(context.Background(), selections)
	if res.InputKeys != 3 {
		t.Errorf("InputKeys=%d, want 3", res.InputKeys)
	}
	if res.AlivKeys != 0 {
		t.Errorf("AlivKeys=%d, want 0", res.AlivKeys)
	}
	if res.DropRate != 1.0 {
		t.Errorf("DropRate=%.2f, want 1.0", res.DropRate)
	}
	if len(res.Filtered) != 0 {
		t.Errorf("Filtered len=%d, want 0 (all keys dead)", len(res.Filtered))
	}
}

// TestFilter_RespectsContextCancel verifies the probe doesn't deadlock when
// the parent context is cancelled.
func TestFilter_RespectsContextCancel(t *testing.T) {
	selections := []selector.CountrySelection{
		{
			Country: "US",
			Top: []selector.Result{
				{Link: "vless://aaa@example.com:443"},
			},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel BEFORE starting

	p := &Probe{
		XrayKnifeBin: "definitely-not-a-real-binary-12345",
		Timeout:      100 * time.Millisecond,
	}
	done := make(chan struct{})
	go func() {
		_ = p.Filter(ctx, selections)
		close(done)
	}()
	select {
	case <-done:
		// good
	case <-time.After(5 * time.Second):
		t.Fatal("Filter did not return after ctx cancel within 5s")
	}
}
