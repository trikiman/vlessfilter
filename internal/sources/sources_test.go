package sources

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoad_Valid covers basic happy-path parsing of a mixed sources.yaml.
func TestLoad_Valid(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "sources.yaml")
	if err := os.WriteFile(yamlPath, []byte(`
countries: [US, DE, JP]
sources:
  - name: v2go-country
    url: https://example.com/cc/{CC}.txt
    kind: country-template
    enabled: true
  - name: barry-far
    url: https://example.com/vless.txt
    kind: plain
    enabled: true
  - name: disabled-one
    url: https://example.com/x.txt
    kind: plain
    enabled: false
`), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	cfg, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := len(cfg.Countries); got != 3 {
		t.Errorf("countries len = %d, want 3", got)
	}
	if got := len(cfg.Sources); got != 2 {
		t.Errorf("enabled sources len = %d, want 2 (disabled should be filtered)", got)
	}
	if cfg.Sources[0].Name != "v2go-country" || cfg.Sources[1].Name != "barry-far" {
		t.Errorf("source order: %v", cfg.Sources)
	}
}

// TestLoad_DefaultsCountries: when countries is omitted AND a country-template
// source is enabled, fall back to DefaultCountries (~100 entries).
func TestLoad_DefaultsCountries(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sources.yaml")
	_ = os.WriteFile(p, []byte(`
sources:
  - name: x
    url: https://example.com/{CC}.txt
    kind: country-template
    enabled: true
`), 0o644)

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Countries) < 50 {
		t.Errorf("expected DefaultCountries() (>=50 entries), got %d", len(cfg.Countries))
	}
	hasUS, hasDE := false, false
	for _, c := range cfg.Countries {
		if c == "US" {
			hasUS = true
		}
		if c == "DE" {
			hasDE = true
		}
	}
	if !hasUS || !hasDE {
		t.Errorf("DefaultCountries should contain at least US and DE")
	}
}

// TestLoad_NoCountriesNoTemplate: empty countries + no template source = stays empty.
func TestLoad_NoCountriesNoTemplate(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sources.yaml")
	_ = os.WriteFile(p, []byte(`
sources:
  - name: x
    url: https://example.com/x.txt
    kind: plain
    enabled: true
`), 0o644)

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Countries) != 0 {
		t.Errorf("expected empty countries when no template source; got %d", len(cfg.Countries))
	}
}

// TestLoad_RejectsBadKind ensures unknown kind produces a clear error.
func TestLoad_RejectsBadKind(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sources.yaml")
	_ = os.WriteFile(p, []byte(`
sources:
  - name: x
    url: https://example.com/x.txt
    kind: bogus
    enabled: true
`), 0o644)

	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for bad kind, got nil")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("err should mention bad kind value: %v", err)
	}
}

// TestLoad_RejectsCountryTemplateWithoutCC ensures country-template URLs require {CC}.
func TestLoad_RejectsCountryTemplateWithoutCC(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sources.yaml")
	_ = os.WriteFile(p, []byte(`
sources:
  - name: x
    url: https://example.com/no-placeholder.txt
    kind: country-template
    enabled: true
`), 0o644)

	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for country-template without {CC}")
	}
}

// TestExpand_CountryTemplate verifies template expansion against country list.
func TestExpand_CountryTemplate(t *testing.T) {
	cfg := &Config{
		Countries: []string{"US", "DE", "JP"},
		Sources: []Source{
			{Name: "v2go", URL: "https://example.com/cc/{CC}.txt", Kind: KindCountryTemplate, Enabled: true},
			{Name: "static", URL: "https://example.com/vless.txt", Kind: KindPlain, Enabled: true},
		},
	}
	exp := Expand(cfg)
	if got := len(exp); got != 4 {
		t.Fatalf("expanded count = %d, want 4 (3 from template + 1 plain)", got)
	}
	wantNames := map[string]string{
		"v2go/US": "https://example.com/cc/US.txt",
		"v2go/DE": "https://example.com/cc/DE.txt",
		"v2go/JP": "https://example.com/cc/JP.txt",
		"static":  "https://example.com/vless.txt",
	}
	for _, e := range exp {
		want, ok := wantNames[e.Name]
		if !ok {
			t.Errorf("unexpected expanded name: %q", e.Name)
			continue
		}
		if e.URL != want {
			t.Errorf("name %q: url = %q, want %q", e.Name, e.URL, want)
		}
	}
}

// TestDecodeBody_Base64 verifies auto-detection of base64-encoded subs.
func TestDecodeBody_Base64(t *testing.T) {
	uri := "vless://00000000-0000-0000-0000-000000000000@example.com:443?type=tcp&security=reality#fake-test-key"
	body := []byte(base64.StdEncoding.EncodeToString([]byte(uri)))
	got, err := DecodeBody(body)
	if err != nil {
		t.Fatalf("DecodeBody: %v", err)
	}
	if !strings.Contains(string(got), "vless://") {
		t.Errorf("decoded body should contain vless://, got: %q", got)
	}
}

// TestDecodeBody_Plain ensures plain-text input is returned unchanged.
func TestDecodeBody_Plain(t *testing.T) {
	uri := "vless://00000000-0000-0000-0000-000000000000@example.com:443?type=tcp&security=reality#fake-test-key\n"
	got, err := DecodeBody([]byte(uri))
	if err != nil {
		t.Fatalf("DecodeBody: %v", err)
	}
	if string(got) != uri {
		t.Errorf("plain body should be returned unchanged; got len=%d want len=%d", len(got), len(uri))
	}
}

// TestDecodeBody_Empty surfaces an error on empty bodies.
func TestDecodeBody_Empty(t *testing.T) {
	if _, err := DecodeBody([]byte("   \n")); err == nil {
		t.Error("expected error for empty body")
	}
}
