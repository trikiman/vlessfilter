// Package accuracy implements the post-publish ground-truth probe (GEO-04).
//
// After the pipeline writes subs/<CC>.txt files, the probe samples up to
// MaxPerCountry random keys per country, routes a real HTTP request through
// each via xray-knife to ipinfo.io/json, and compares the actual exit
// country to the published label.
//
// Returns a Report with per-country accuracy and an overall pass/fail
// against the configured threshold. The pipeline can use this to refuse
// publishing bad data (for example: revert subs/ from git instead of
// pushing if accuracy is below threshold).
package accuracy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Result is one sampled key's verification outcome.
type Result struct {
	Link            string
	PublishedCC     string // what we labeled it as
	ActualCC        string // what ipinfo.io reported
	Match           bool   // PublishedCC == ActualCC (both upper-cased)
	Err             string // empty if ok, error message if probe failed
	DurationSeconds float64
}

// Report aggregates per-country accuracy across all sampled keys.
type Report struct {
	GeneratedAt    time.Time
	OverallPercent float64
	TotalSampled   int
	TotalMatched   int
	TotalErrored   int
	PerCountry     map[string]CountryReport
	Threshold      float64 // pass/fail threshold (e.g., 0.80 for 80%)
	Passed         bool    // OverallPercent >= Threshold
	Samples        []Result
}

// CountryReport is per-country breakdown.
type CountryReport struct {
	Sampled  int
	Matched  int
	Errored  int
	Mismatch []Result // first few mismatches for diagnostics
}

// Probe samples keys from each subs/<CC>.txt and verifies their actual
// exit country.
//
// MaxPerCountry caps the sample size per country (default 5).
// Threshold is the minimum overall accuracy to mark Passed=true (default 0.80).
// Timeout caps each individual proxy probe (default 15s).
type Probe struct {
	SubsDir       string  // path to subs/ directory
	MaxPerCountry int     // 0 → 5
	Threshold     float64 // 0 → 0.80
	Timeout       time.Duration
	XrayKnifeBin  string // path to xray-knife binary; empty = "xray-knife"
}

// Run performs the probe and returns a Report.
func (p *Probe) Run(ctx context.Context) (*Report, error) {
	if p.MaxPerCountry == 0 {
		p.MaxPerCountry = 5
	}
	if p.Threshold == 0 {
		p.Threshold = 0.80
	}
	if p.Timeout == 0 {
		p.Timeout = 15 * time.Second
	}
	if p.XrayKnifeBin == "" {
		p.XrayKnifeBin = "xray-knife"
	}

	// Collect candidates: one slice of (cc, link) per country.
	candidates, err := loadCandidates(p.SubsDir)
	if err != nil {
		return nil, fmt.Errorf("load candidates: %w", err)
	}
	if len(candidates) == 0 {
		return nil, errors.New("no subs/<CC>.txt files found to probe")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	report := &Report{
		GeneratedAt: time.Now().UTC(),
		Threshold:   p.Threshold,
		PerCountry:  make(map[string]CountryReport),
	}

	for cc, links := range candidates {
		// Random sample without replacement up to MaxPerCountry.
		rng.Shuffle(len(links), func(i, j int) { links[i], links[j] = links[j], links[i] })
		sample := links
		if len(sample) > p.MaxPerCountry {
			sample = sample[:p.MaxPerCountry]
		}
		cr := CountryReport{}
		for _, link := range sample {
			r := p.probeOne(ctx, cc, link)
			report.Samples = append(report.Samples, r)
			cr.Sampled++
			report.TotalSampled++
			if r.Err != "" {
				cr.Errored++
				report.TotalErrored++
				continue
			}
			if r.Match {
				cr.Matched++
				report.TotalMatched++
			} else {
				if len(cr.Mismatch) < 3 {
					cr.Mismatch = append(cr.Mismatch, r)
				}
			}
		}
		report.PerCountry[cc] = cr
	}

	// Overall = matched / (sampled - errored). Errored samples don't count
	// against accuracy because we couldn't verify them.
	denom := report.TotalSampled - report.TotalErrored
	if denom > 0 {
		report.OverallPercent = float64(report.TotalMatched) / float64(denom)
	}
	report.Passed = report.OverallPercent >= report.Threshold
	return report, nil
}

// probeOne routes a single HTTP request through `link` to ipinfo.io/json
// and parses the country field.
func (p *Probe) probeOne(ctx context.Context, expectedCC, link string) Result {
	r := Result{
		Link:        link,
		PublishedCC: strings.ToUpper(strings.TrimSpace(expectedCC)),
	}
	start := time.Now()
	defer func() { r.DurationSeconds = time.Since(start).Seconds() }()

	// Use xray-knife http -c <link> -u https://ipinfo.io/json -d Timeout
	// to route the request through the proxy. Capture stdout — xray-knife
	// prints the response body when --speedtest is omitted.
	cctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx,
		p.XrayKnifeBin, "http",
		"-c", link,
		"-u", "https://ipinfo.io/json",
		"-d", fmt.Sprintf("%d", int(p.Timeout/time.Millisecond)),
		"-b", // print response body
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.Err = fmt.Sprintf("xray-knife: %v", err)
		return r
	}

	cc, parseErr := parseIPInfoCountry(out)
	if parseErr != nil {
		r.Err = fmt.Sprintf("parse: %v (out len=%d)", parseErr, len(out))
		return r
	}
	r.ActualCC = strings.ToUpper(cc)
	r.Match = r.ActualCC == r.PublishedCC
	return r
}

// parseIPInfoCountry extracts the `country` field from ipinfo.io/json's
// response, possibly with extra log lines mixed in (xray-knife prefixes
// connection/timing logs in the output).
func parseIPInfoCountry(out []byte) (string, error) {
	// Find a JSON object containing "country".
	s := string(out)
	jsonStart := strings.Index(s, "{")
	if jsonStart < 0 {
		return "", errors.New("no JSON object in output")
	}
	candidate := s[jsonStart:]
	// Trim to last closing brace.
	if i := strings.LastIndex(candidate, "}"); i >= 0 {
		candidate = candidate[:i+1]
	}
	var resp struct {
		Country string `json:"country"`
	}
	if err := json.Unmarshal([]byte(candidate), &resp); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}
	if resp.Country == "" {
		return "", errors.New("ipinfo.io response missing country field")
	}
	return resp.Country, nil
}

// loadCandidates reads every subs/<CC>.txt under subsDir and returns
// {CC: [link, link, ...]}. Skips all.txt and rotating.txt.
func loadCandidates(subsDir string) (map[string][]string, error) {
	entries, err := os.ReadDir(subsDir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", subsDir, err)
	}
	out := make(map[string][]string)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Want exactly XX.txt where XX is uppercase letters.
		if !strings.HasSuffix(name, ".txt") {
			continue
		}
		base := strings.TrimSuffix(name, ".txt")
		if len(base) != 2 {
			continue
		}
		if base[0] < 'A' || base[0] > 'Z' || base[1] < 'A' || base[1] > 'Z' {
			continue
		}
		body, err := os.ReadFile(filepath.Join(subsDir, name))
		if err != nil {
			slog.Warn("accuracy: read country file failed", "name", name, "err", err)
			continue
		}
		var links []string
		for _, line := range strings.Split(string(body), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "vless://") {
				links = append(links, line)
			}
		}
		if len(links) > 0 {
			out[base] = links
		}
	}
	return out, nil
}

// LogReport prints the report as structured slog INFO/ERROR lines.
func LogReport(r *Report) {
	level := slog.LevelInfo
	msg := "accuracy probe: PASSED"
	if !r.Passed {
		level = slog.LevelError
		msg = "accuracy probe: FAILED — published labels diverge from actual exits"
	}
	slog.Log(context.Background(), level, msg,
		"overall_percent", fmt.Sprintf("%.1f%%", r.OverallPercent*100),
		"threshold_percent", fmt.Sprintf("%.1f%%", r.Threshold*100),
		"sampled", r.TotalSampled,
		"matched", r.TotalMatched,
		"errored", r.TotalErrored)

	for cc, cr := range r.PerCountry {
		acc := 0.0
		denom := cr.Sampled - cr.Errored
		if denom > 0 {
			acc = float64(cr.Matched) / float64(denom)
		}
		slog.Info("accuracy per-country",
			"cc", cc,
			"sampled", cr.Sampled,
			"matched", cr.Matched,
			"errored", cr.Errored,
			"accuracy", fmt.Sprintf("%.0f%%", acc*100),
			"mismatches", len(cr.Mismatch))
	}
}
