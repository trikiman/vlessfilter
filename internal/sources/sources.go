// Package sources loads sources.yaml, expands country-templates, fetches HTTP
// subscription bodies, and auto-detects base64 vs plain-text payloads.
//
// It does not call into xray-knife; that's the xrayknife package's job. This
// package is pure I/O and parsing so it can be unit-tested without a network
// or external binary.
package sources

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Kind values allowed in sources.yaml.
const (
	KindPlain           = "plain"
	KindCountryTemplate = "country-template"
)

// Config is the deserialized sources.yaml.
type Config struct {
	Sources   []Source `yaml:"sources"`
	Countries []string `yaml:"countries,omitempty"`
}

// Source is a single subscription declaration.
type Source struct {
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
	Kind    string `yaml:"kind"`
	Enabled bool   `yaml:"enabled"`
}

// ExpandedSource is a Source after country-template expansion.
type ExpandedSource struct {
	Name string // includes country suffix, e.g., "v2go-country/US"
	URL  string // {CC} placeholder substituted
}

// Subscription is the result of fetching an ExpandedSource.
type Subscription struct {
	SourceName string
	URL        string
	Body       []byte // always plain text after auto-decode
}

// Load reads sources.yaml from disk, validates structure, applies defaults,
// and returns enabled sources only.
//
// Default countries (when omitted) are ["US","DE"] per Phase 1 D-04.
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(cfg.Countries) == 0 {
		cfg.Countries = []string{"US", "DE"}
	}
	enabled := make([]Source, 0, len(cfg.Sources))
	for i, s := range cfg.Sources {
		if s.Name == "" {
			return nil, fmt.Errorf("source[%d]: name is required", i)
		}
		if s.URL == "" {
			return nil, fmt.Errorf("source %q: url is required", s.Name)
		}
		switch s.Kind {
		case KindPlain, KindCountryTemplate:
		default:
			return nil, fmt.Errorf("source %q: kind must be %q or %q (got %q)", s.Name, KindPlain, KindCountryTemplate, s.Kind)
		}
		if s.Kind == KindCountryTemplate && !strings.Contains(s.URL, "{CC}") {
			return nil, fmt.Errorf("source %q: country-template url must contain {CC} placeholder", s.Name)
		}
		if !s.Enabled {
			continue
		}
		enabled = append(enabled, s)
	}
	cfg.Sources = enabled
	return &cfg, nil
}

// Expand resolves country-template URLs using the country list and returns
// a flat list of fetchable sources.
func Expand(cfg *Config) []ExpandedSource {
	out := make([]ExpandedSource, 0, len(cfg.Sources))
	for _, s := range cfg.Sources {
		switch s.Kind {
		case KindPlain:
			out = append(out, ExpandedSource{Name: s.Name, URL: s.URL})
		case KindCountryTemplate:
			for _, cc := range cfg.Countries {
				out = append(out, ExpandedSource{
					Name: s.Name + "/" + cc,
					URL:  strings.ReplaceAll(s.URL, "{CC}", cc),
				})
			}
		}
	}
	return out
}

// userAgent identifies the tool to upstream subscription hosts.
var userAgent = "vlessfilter/0.1 (+https://github.com/trikiman/vlessfilter)"

// httpClient is a package-level client with a 30s timeout. Tests override it.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// Fetch performs an HTTP GET against the source URL and returns a Subscription
// whose Body is plain text (base64 is auto-decoded).
func Fetch(ctx context.Context, src ExpandedSource) (Subscription, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.URL, nil)
	if err != nil {
		return Subscription{}, fmt.Errorf("build request for %s: %w", src.URL, err)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return Subscription{}, fmt.Errorf("fetch %s: %w", src.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Subscription{}, fmt.Errorf("fetch %s: status %d", src.URL, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20)) // 64 MiB safety cap
	if err != nil {
		return Subscription{}, fmt.Errorf("read body from %s: %w", src.URL, err)
	}
	plain, err := DecodeBody(body)
	if err != nil {
		return Subscription{}, fmt.Errorf("decode body from %s: %w", src.URL, err)
	}
	return Subscription{SourceName: src.Name, URL: src.URL, Body: plain}, nil
}

// DecodeBody auto-detects base64 vs plain text. If the trimmed input
// base64-decodes cleanly AND the decoded bytes contain "vless://", the
// decoded form is returned. Otherwise the input is returned unchanged.
func DecodeBody(body []byte) ([]byte, error) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil, errors.New("empty body")
	}
	// Try standard base64 first, then URL-safe.
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.RawURLEncoding} {
		decoded, err := enc.DecodeString(trimmed)
		if err == nil && bytesContainsVless(decoded) {
			return decoded, nil
		}
	}
	return body, nil
}

func bytesContainsVless(b []byte) bool {
	return strings.Contains(string(b), "vless://")
}
