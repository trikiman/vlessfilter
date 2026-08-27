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
	// Errored counts keys whose probe never produced a verdict — the binary was
	// missing, the engine crashed, the context was cancelled. These are NOT
	// evidence the key is dead, and they are excluded from DropRate.
	//
	// Without this split, an unusable prober yielded DropRate == 1.0 and tripped
	// the caller's 75% abort, which republishes stale output. That is the same
	// silent-staleness failure that kept 26-day-old trojan keys in production.
	Errored  int
	DropRate float64 // (verdicts - alive) / verdicts, where verdicts = input - errored
}

// probeOutcome is the tri-state result of probing one key.
type probeOutcome int

const (
	probeAlive   probeOutcome = iota // ran, real traffic flowed
	probeDead                        // ran, engine reported failure
	probeErrored                     // never ran, or died before reporting
)

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

	// Per-job outcome. Zero value is probeAlive, which would be the WRONG
	// default for a panicked goroutine, so seed everything to probeErrored:
	// "we never learned" is the honest default, and it neither drops the key
	// nor inflates DropRate.
	outcomes := make([]probeOutcome, len(jobs))
	for i := range outcomes {
		outcomes[i] = probeErrored
	}

	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup
	for ji, j := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func(ji int, j job) {
			defer wg.Done()
			defer func() { <-sem }()
			link := selections[j.csIdx].Top[j.linkIdx].Link
			outcomes[ji] = probeOne(ctx, bin, link, url, timeout)
		}(ji, j)
	}
	wg.Wait()

	// Build filtered selections.
	out := make([]selector.CountrySelection, 0, len(selections))
	alive := 0
	errored := 0
	for ci, cs := range selections {
		var kept []selector.Result
		for li, r := range cs.Top {
			// Find pass index for (ci, li) — same flatten order.
			for ji, j := range jobs {
				if j.csIdx == ci && j.linkIdx == li {
					switch outcomes[ji] {
					case probeAlive:
						kept = append(kept, r)
						alive++
					case probeErrored:
						// KEEP. A probe that never ran is not evidence the key
						// is dead. Dropping on error is how three transient
						// failures on a country's three keys deleted
						// subs/<CC>.txt outright and 404'd a subscriber's URL.
						kept = append(kept, r)
						errored++
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
		Errored:   errored,
	}
	// Denominator is VERDICTS, not inputs. Dividing by len(jobs) meant a
	// prober that could not run at all scored DropRate 1.0 and tripped the
	// caller's 75% abort, republishing stale output — the same silent-staleness
	// mode that served 26-day-old trojan keys.
	if verdicts := len(jobs) - errored; verdicts > 0 {
		res.DropRate = float64(verdicts-alive) / float64(verdicts)
	}
	if errored > 0 {
		slog.Warn("pre-publish probe: some keys never produced a verdict",
			"errored", errored, "of_input", len(jobs),
			"note", "counted as inconclusive, not dead; excluded from drop_rate")
	}
	return res
}

// probeOne attempts a single HTTP request through the proxy. Returns true
// on success (xray-knife exits 0 AND output contains "Real Delay: NNNms"
// — the success marker xray-knife emits after a successful HTTP fetch).
//
// Empirical xray-knife v9.12 output format:
//
//	Success:
//	  ...config dump...
//	  ✅ HH:MM:SS Real Delay: 1234ms
//
//	Failure (timeout, connection error, non-2xx upstream):
//	  ...config dump...
//	  ❌ HH:MM:SS Real Delay: -1ms / Failed: <reason>
//
// xray-knife exits 0 in BOTH cases (config parsed = success from xray-knife's
// view), so exit code alone is unreliable. The "Real Delay: NNNms" with a
// positive number, plus the green checkmark emoji, is the only reliable
// indicator that real traffic flowed.
func probeOne(ctx context.Context, bin, link, url string, timeout time.Duration) probeOutcome {
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
		"-b", // print response body so we can verify upstream returned data
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		// NOT "definitively dead". The doc comment above already establishes
		// that xray-knife exits 0 for BOTH pass and fail (handleSingleConfig
		// prints the failure and returns), so a non-zero exit almost never
		// means the key is dead — it means the binary is missing, the engine
		// crashed, or the context was cancelled. Counting these as deaths is
		// what let an unusable prober reach DropRate 1.0 and abort the publish
		// into republishing stale output.
		return probeErrored
	}
	body := out.String()
	// Hard fail markers — xray-knife explicitly reported failure.
	if strings.Contains(body, "❌") ||
		strings.Contains(body, "Real Delay: -1ms") ||
		strings.Contains(body, "Failed:") {
		return probeDead
	}
	// Hard success markers — xray-knife confirmed real traffic flowed.
	if strings.Contains(body, "✅") {
		return probeAlive
	}
	// Soft success: "Real Delay: NNNms" with a positive ms reading.
	if i := strings.Index(body, "Real Delay: "); i >= 0 {
		rest := body[i+len("Real Delay: "):]
		// First few chars should be digits if positive ms.
		for k := 0; k < len(rest) && k < 6; k++ {
			if rest[k] < '0' || rest[k] > '9' {
				if rest[k] == 'm' || rest[k] == 's' {
					// "Real Delay: 1234ms" — k pointing at 'm', positive
					return probeAlive
				}
				if rest[k] == '-' {
					// "Real Delay: -1ms" — negative
					return probeDead
				}
				break
			}
		}
	}
	// No clear marker either way. The engine ran and printed something we do
	// not recognise, so this is inconclusive rather than a confirmed death —
	// an output-format change upstream must not read as a mass key die-off and
	// trip the publish abort.
	return probeErrored
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
