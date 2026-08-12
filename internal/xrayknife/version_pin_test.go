package xrayknife

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The CI workflows install xray-knife in their own shell step and never read
// XrayKnifeVersion, so the constant and the workflows can silently disagree.
// That is exactly what hid a 26-day trojan outage: the constant was bumped to
// v10.1.1 while every workflow still ran `v9@latest` (v9.12.1), whose bundled
// xray-core panics on trojan splithttp configs. Fail loudly on any drift.
func TestWorkflowsPinSameXrayKnifeVersion(t *testing.T) {
	root := filepath.Join("..", "..")
	want := xrayKnifeModule + "@" + XrayKnifeVersion

	if strings.Contains(XrayKnifeVersion, "latest") {
		t.Fatalf("XrayKnifeVersion must be a fixed tag, got %q: an unpinned "+
			"engine is how a protocol goes dark without failing a check", XrayKnifeVersion)
	}

	installRE := regexp.MustCompile(`go install (github\.com/lilendian0x00/xray-knife/\S+)`)
	targets := []string{
		filepath.Join(root, ".github", "workflows", "refresh.yml"),
		filepath.Join(root, ".github", "workflows", "verify-russia.yml"),
		filepath.Join(root, ".github", "workflows", "benchmark.yml"),
		filepath.Join(root, "scripts", "h2-quick.sh"),
		filepath.Join(root, "scripts", "install-always-on.sh"),
	}
	for _, path := range targets {
		raw, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue // script may legitimately be removed later
			}
			t.Fatalf("read %s: %v", path, err)
		}
		matches := installRE.FindAllStringSubmatch(string(raw), -1)
		if len(matches) == 0 {
			continue // file doesn't install xray-knife
		}
		for _, m := range matches {
			if got := m[1]; got != want {
				t.Errorf("%s installs %q, want %q (version drift between CI and XrayKnifeVersion)",
					filepath.Base(path), got, want)
			}
		}
	}
}
