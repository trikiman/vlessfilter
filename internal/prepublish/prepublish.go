// Package prepublish runs a final connectivity probe on each top-3 selection
// IMMEDIATELY BEFORE publication, dropping keys that fail.
//
// The motivating bug: stage 2 speedtest happens early in the pipeline run
// (~5-10 min in), but publication happens 50+ min later. By the time the
// user opens v2rayN, configs have churned for hours. Publishing keys that
// were alive an hour ago produces the "80-90% -1 timeout" experience the
// user reported.
//
// This package re-validates each curated selection in the seconds before
// WriteAll, ensuring published keys were alive within the last ~30s.
//
// Design:
//   - Concurrent probes (configurable, default 20 — same as stage 2)
//   - Single 5s HTTP attempt per key (no retries — dead is dead)
//   - Returns filtered selections (failed keys removed)
//   - Returns drop-rate so caller can decide whether to abort the run
package prepublish

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/trikiman/vlessfilter/internal/selector"
)

// Probe re-tests selected keys for connectivity right before publication.
type Probe struct {
	// XrayKnifeBin is the binary path. Empty defaults to "xray-knife".
	XrayKnifeBin string
	// Timeout per key. 0 = 5s default.
	Timeout time.Duration
	// Concurrency caps parallel probes. 0 = 20 default.
	Concurrency int
	// TestURL is the upstream the probe routes through each key.
	// Empty = default https://www.cloudflare.com/cdn-cgi/trace.
	TestURL string
}

// Result describes one set of selections after probing.
type Result struct {
	Filtered  []selector.CountrySelection // selections with failed keys removed
	InputKeys int                         // count before probe
	AlivKeys  int                         // count that passed probe
	DropRate  float64                     // (input - alive) / input
}

// Filter probes each key in `selections`. Returns filtered results plus
// summary stats. `selections` is not mutated.
//
// On context cancellation, returns whatever it has so far (caller can
// decide to abort or proceed with partial filter).
func (p *Probe) Filter(ctx context.Context, selections []selector.CountrySelection) Result {
	bin := p.XrayKnifeBin
	if bin == "" {
		bin = "xray-knife"
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	conc := p.Concurrency
	if conc <= 0 {
		conc = 20
	}
	url := p.TestURL
	if url == "" {
		url = "https://www.cloudflare.com/cdn-cgi/trace"
	}

	// Flatten all keys into one queue with back-pointers to their CountrySelection.
	type job struct {
		csIdx   int
		linkIdx int
	}
	var jobs []job
	for ci, cs := range selections {
		for li := range cs.Top {
			jobs = append(jobs, job{ci, li})
		}
	}
	if len(jobs) == 0 {
		return Result{}
	}

	// Per-job pass/fail. Default = false (failed) so a panicked goroutine
	// is treated as a probe failure (drop the key, safe default).
	pass := make([]bool, len(jobs))

	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup
	for ji, j := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func(ji int, j job) {
			defer wg.Done()
			defer func() { <-sem }()
			link := selections[j.csIdx].Top[j.linkIdx].Link
			pass[ji] = probeOne(ctx, bin, link, url, timeout)
		}(ji, j)
	}
	wg.Wait()

	// Build filtered selections.
	out := make([]selector.CountrySelection, 0, len(selections))
	alive := 0
	for ci, cs := range selections {
		var kept []selector.Result
		for li, r := range cs.Top {
			// Find pass index for (ci, li) — same flatten order.
			for ji, j := range jobs {
				if j.csIdx == ci && j.linkIdx == li {
					if pass[ji] {
						kept = append(kept, r)
						alive++
					}
					break
				}
			}
		}
		if len(kept) > 0 {
			out = append(out, selector.CountrySelection{
				Country: cs.Country,
				Top:     kept,
			})
		}
	}

	res := Result{
		Filtered:  out,
		InputKeys: len(jobs),
		AlivKeys:  alive,
	}
	if len(jobs) > 0 {
		res.DropRate = float64(len(jobs)-alive) / float64(len(jobs))
	}
	return res
}

// probeOne attempts a single HTTP request through the proxy. Returns true
// on success (2xx response received), false on any failure (timeout,
// connection reset, non-2xx, exec error).
//
// Uses xray-knife http -c <link> -u <url> -d <ms> for a one-shot test.
func probeOne(ctx context.Context, bin, link, url string, timeout time.Duration) bool {
	timeoutMs := int(timeout / time.Millisecond)
	if timeoutMs < 1000 {
		timeoutMs = 1000
	}
	cctx, cancel := context.WithTimeout(ctx, timeout+5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cctx, bin, "http",
		"-c", link,
		"-u", url,
		"-d", fmt.Sprintf("%d", timeoutMs),
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return false
	}
	body := strings.ToLower(out.String())
	// xray-knife exits 0 even when HTTP fails. Look for success markers
	// in stderr/stdout: "delay:" with a positive ms reading.
	if strings.Contains(body, "failed") || strings.Contains(body, "error") {
		return false
	}
	// Positive signal: "delay: 123ms" or "ok" or response body present.
	if strings.Contains(body, "delay:") || strings.Contains(body, "200 ok") {
		return true
	}
	return false
}

// LogResult emits a structured summary at INFO level.
func LogResult(r Result) {
	slog.Info("pre-publish probe complete",
		"input_keys", r.InputKeys,
		"alive_keys", r.AlivKeys,
		"dropped", r.InputKeys-r.AlivKeys,
		"drop_rate", fmt.Sprintf("%.1f%%", r.DropRate*100),
	)
}
