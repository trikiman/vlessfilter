// Package output writes the user-facing results: per-country subscription
// files (subs/<CC>.txt), a README.md summary table, and diagnostic outputs
// (all-results.csv, raw/dead.txt) at the chosen output directory root.
//
// Format matches D-08 (README columns), D-09 (subs files: plain vless URIs,
// one per line, UTF-8, LF) and D-16/D-17 (diagnostic + deterministic).
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/trikiman/vlessfilter/internal/selector"
)

// WriteAll writes everything: subs/<CC>.txt + subs/rotating.txt +
// README.md + (when diagnostics is non-nil) all-results.csv + raw/dead.txt.
//
// DEPRECATED in v2.0 multi-protocol mode. Use WriteProtocol per-protocol +
// WriteMultiProtocolReadme + WriteDiagnostics. Retained for tests + any
// single-protocol callers that haven't migrated.
//
// `selections`  — curated top-3-per-country (only stable-country configs).
// `allTested`   — full set for all-results.csv.
// `dead`        — failed set for raw/dead.txt.
// `rotating`    — multi-exit configs (CF Workers + load balancers etc.).
//
//	These can't be honestly tagged with one country; they go
//	to subs/rotating.txt with a "🌐 ROTATING" remark.
func WriteAll(outDir string, selections []selector.CountrySelection, allTested, dead, rotating []selector.Result, generatedAt time.Time) error {
	if err := Write(outDir, selections, rotating, generatedAt); err != nil {
		return err
	}
	return WriteDiagnostics(outDir, allTested, dead)
}

// ProtoReadme is the per-protocol summary fed into WriteMultiProtocolReadme.
type ProtoReadme struct {
	Protocol   string
	Selections []selector.CountrySelection
	Rotating   int // count, not the slice itself
}

// WriteProtocol writes per-protocol output: subs/<protocol>/<CC>.txt,
// subs/<protocol>/all.txt, subs/<protocol>/rotating.txt.
//
// Mirrors Write() but rooted at subs/<protocol>/ so each protocol gets its
// own subscription URL space (e.g., subs/vmess/all.txt). Top-level subs/
// remains for back-compat (vless mirror — handled by callers via Write).
func WriteProtocol(outDir, protocol string, selections []selector.CountrySelection, rotating []selector.Result, generatedAt time.Time) error {
	subsDir := filepath.Join(outDir, "subs", protocol)
	if err := os.MkdirAll(subsDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", subsDir, err)
	}

	// Clean stale CC.txt files for this protocol so countries that fell
	// off (e.g., zero stable-alive this run) don't keep showing.
	stale, _ := filepath.Glob(filepath.Join(subsDir, "*.txt"))
	for _, p := range stale {
		base := filepath.Base(p)
		if len(base) == 6 && strings.HasSuffix(base, ".txt") &&
			base[0] >= 'A' && base[0] <= 'Z' && base[1] >= 'A' && base[1] <= 'Z' {
			_ = os.Remove(p)
		}
	}

	sortedSel := make([]selector.CountrySelection, len(selections))
	copy(sortedSel, selections)
	sort.SliceStable(sortedSel, func(i, j int) bool {
		return sortedSel[i].Country < sortedSel[j].Country
	})

	var allBuf strings.Builder
	for _, cs := range sortedSel {
		path := filepath.Join(subsDir, cs.Country+".txt")
		var b strings.Builder
		for _, r := range cs.Top {
			rewritten := rewriteRemark(r.Link, cs.Country, r.SpeedMbps)
			b.WriteString(rewritten)
			b.WriteByte('\n')
			allBuf.WriteString(rewritten)
			allBuf.WriteByte('\n')
		}
		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	allPath := filepath.Join(subsDir, "all.txt")
	if err := os.WriteFile(allPath, []byte(allBuf.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", allPath, err)
	}

	if err := writeRotating(subsDir, rotating); err != nil {
		return err
	}
	_ = generatedAt // reserved for future per-protocol README per-dir

	// Write a JSON sidecar with the data the merge step needs to rebuild
	// a multi-protocol README. We can't pass selector.Result directly —
	// the merge runner doesn't have xray-knife.db — so we serialize the
	// minimum that WriteMultiProtocolReadme needs: per-country top entries
	// and rotating count.
	//
	// Path: <outDir>/.readme-data/<proto>.json (NOT under subs/ so the
	// workflow's git add -f subs/ doesn't pick them up — they're build
	// artifacts, not subscription content).
	sidecarDir := filepath.Join(outDir, ".readme-data")
	if err := os.MkdirAll(sidecarDir, 0o755); err == nil {
		sidecar := struct {
			Protocol    string                       `json:"protocol"`
			Selections  []selector.CountrySelection  `json:"selections"`
			Rotating    int                          `json:"rotating"`
			GeneratedAt time.Time                    `json:"generated_at"`
		}{
			Protocol:    protocol,
			Selections:  sortedSel,
			Rotating:    len(rotating),
			GeneratedAt: generatedAt,
		}
		// A dropped sidecar is not cosmetic: regen-readme builds the README
		// purely from these, so losing one silently deletes that protocol's
		// entire section and its subscription URLs from the docs while the
		// keys stay published. Surface the failure instead of swallowing it.
		buf, err := json.MarshalIndent(sidecar, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal readme sidecar [%s]: %w", protocol, err)
		}
		if err := os.WriteFile(filepath.Join(sidecarDir, protocol+".json"), buf, 0o644); err != nil {
			return fmt.Errorf("write readme sidecar [%s]: %w", protocol, err)
		}
	}
	return nil
}

// Write produces subs/<CC>.txt files, subs/all.txt, subs/rotating.txt
// and README.md inside outDir.
func Write(outDir string, selections []selector.CountrySelection, rotating []selector.Result, generatedAt time.Time) error {
	subsDir := filepath.Join(outDir, "subs")
	if err := os.MkdirAll(subsDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", subsDir, err)
	}

	// Delete stale subs/<CC>.txt files. Reason: when a country has 0 alive
	// keys in this run but had keys in a previous run, its old file would
	// otherwise stay around forever with stale/broken keys. Glob and remove
	// every <CC>.txt before writing fresh content. all.txt is also rewritten
	// below so it gets overwritten regardless.
	stale, _ := filepath.Glob(filepath.Join(subsDir, "*.txt"))
	for _, p := range stale {
		// Keep all.txt — it gets rewritten below; deleting first is fine
		// but unnecessary. Same for any other non-CC file we add later.
		base := filepath.Base(p)
		// Heuristic: 2-letter uppercase CC.txt = our country file
		if len(base) == 6 && strings.HasSuffix(base, ".txt") &&
			base[0] >= 'A' && base[0] <= 'Z' && base[1] >= 'A' && base[1] <= 'Z' {
			_ = os.Remove(p)
		}
	}

	// Sort once so per-country files AND all.txt are deterministic.
	sortedSel := make([]selector.CountrySelection, len(selections))
	copy(sortedSel, selections)
	sort.SliceStable(sortedSel, func(i, j int) bool {
		return sortedSel[i].Country < sortedSel[j].Country
	})

	var allBuf strings.Builder
	for _, cs := range sortedSel {
		path := filepath.Join(subsDir, cs.Country+".txt")
		var b strings.Builder
		for _, r := range cs.Top {
			rewritten := rewriteRemark(r.Link, cs.Country, r.SpeedMbps)
			b.WriteString(rewritten)
			b.WriteByte('\n')
			allBuf.WriteString(rewritten)
			allBuf.WriteByte('\n')
		}
		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	// Combined "all countries" subscription file. One URL → every country
	// (stable only — rotating goes to its own file).
	allPath := filepath.Join(subsDir, "all.txt")
	if err := os.WriteFile(allPath, []byte(allBuf.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", allPath, err)
	}

	// Rotating bucket: configs whose exit country varies across tests
	// (CF Workers, load balancers, multi-exit chains). Honest disclosure
	// that these aren't country-stable.
	if err := writeRotating(subsDir, rotating); err != nil {
		return err
	}

	readme := buildReadme(sortedSel, len(rotating), generatedAt)
	path := filepath.Join(outDir, "README.md")
	return os.WriteFile(path, []byte(readme), 0o644)
}

// writeRotating creates subs/rotating.txt — multi-exit configs labeled with
// "🌐 ROTATING" so the user knows the country varies per connection.
//
// Sorts deterministically by link for diff stability.
func writeRotating(subsDir string, rotating []selector.Result) error {
	if len(rotating) == 0 {
		// Still write empty file so the URL doesn't 404.
		return os.WriteFile(filepath.Join(subsDir, "rotating.txt"), nil, 0o644)
	}
	rows := make([]selector.Result, len(rotating))
	copy(rows, rotating)
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Link < rows[j].Link })

	var b strings.Builder
	for _, r := range rows {
		// Rewrite remark to "🌐 ROTATING" — make it unmistakable in clients.
		u, err := url.Parse(r.Link)
		if err != nil {
			b.WriteString(r.Link)
			b.WriteByte('\n')
			continue
		}
		u.Fragment = "🌐 ROTATING"
		b.WriteString(u.String())
		b.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(subsDir, "rotating.txt"), []byte(b.String()), 0o644)
}

// rewriteRemark replaces (or sets) the URI fragment of a vless:// link with
// "<flag-emoji> <CC> <Country Name in Russian>" (e.g., "🇵🇱 PL Польша").
// The rest of the URI is preserved verbatim.
//
// Format chosen so that:
//   - Flag emoji renders as a country flag in modern clients.
//   - 2-letter CC is a stable fallback when client font has no emoji support.
//   - Russian name makes entries scannable for the project's primary audience.
//
// On unparseable input, returns the original link unchanged — better to ship
// a working key with a bad-looking name than drop it from the output.
func rewriteRemark(link, cc string, speedMbps float64) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}
	cc = strings.ToUpper(cc)
	flag := flagEmoji(cc)
	name := countryName(cc)

	// Middle segments: speed-tier icon + bracketed Mbps. Both omitted when
	// there is no positive speed measurement, so the flag stays adjacent to
	// the CC. Flag is ALWAYS first (mandatory convention -- see AGENTS.md).
	var mid string
	if speedMbps > 0 {
		if icon := speedIcon(speedMbps); icon != "" {
			mid += icon + " "
		}
		mid += fmt.Sprintf("[%.1f mb] ", speedMbps)
	}

	// url.URL.String() URL-encodes the Fragment automatically.
	if name == cc {
		// Unknown country — don't duplicate the code (e.g., "🇿🇿 ZZ" not "🇿🇿 ZZ ZZ").
		u.Fragment = strings.TrimSpace(fmt.Sprintf("%s %s%s", flag, mid, cc))
	} else {
		u.Fragment = strings.TrimSpace(fmt.Sprintf("%s %s%s %s", flag, mid, cc, name))
	}
	return u.String()
}

// speedIcon maps a measured speed (Mbps) to a scannable tier icon, placed
// AFTER the flag in a key's name. Empty string when below the video-ready
// floor (or unknown), so slow/unmeasured keys carry no icon.
//
//	⚡ >= 60 Mbps   (very fast)
//	🎬 >= 25 Mbps   (4K / heavy streaming headroom)
//	📺 >= 12 Mbps   (1080p-ready, incl. ~2x playback)
func speedIcon(mbps float64) string {
	switch {
	case mbps >= 60:
		return "⚡"
	case mbps >= 25:
		return "🎬"
	case mbps >= 12:
		return "📺"
	default:
		return ""
	}
}

// countryName returns the Russian name for an ISO 3166-1 alpha-2 code.
// Falls back to the code itself for unknown values.
//
// Coverage matches DefaultCountries() in internal/sources plus a few extras
// observed in xray-knife's ip_location output (e.g., LT, BR, RO).
func countryName(cc string) string {
	if name, ok := countryNames[cc]; ok {
		return name
	}
	return cc
}

var countryNames = map[string]string{
	"AE": "UAE", "AF": "Afghanistan", "AL": "Albania", "AM": "Armenia",
	"AR": "Argentina", "AT": "Austria", "AU": "Australia", "AZ": "Azerbaijan",
	"BA": "Bosnia", "BD": "Bangladesh", "BE": "Belgium", "BG": "Bulgaria",
	"BH": "Bahrain", "BO": "Bolivia", "BR": "Brazil", "BY": "Belarus",
	"BZ": "Belize", "CA": "Canada", "CH": "Switzerland", "CL": "Chile",
	"CN": "China", "CO": "Colombia", "CR": "Costa Rica", "CY": "Cyprus",
	"CZ": "Czechia", "DE": "Germany", "DK": "Denmark", "DO": "Dominican Republic",
	"DZ": "Algeria", "EC": "Ecuador", "EE": "Estonia", "EG": "Egypt",
	"ES": "Spain", "ET": "Ethiopia", "FI": "Finland", "FR": "France",
	"GB": "United Kingdom", "GE": "Georgia", "GH": "Ghana", "GR": "Greece",
	"GT": "Guatemala", "HK": "Hong Kong", "HN": "Honduras", "HR": "Croatia",
	"HU": "Hungary", "ID": "Indonesia", "IE": "Ireland", "IL": "Israel",
	"IM": "Isle of Man", "IN": "India", "IQ": "Iraq", "IR": "Iran",
	"IS": "Iceland", "IT": "Italy", "JM": "Jamaica", "JO": "Jordan",
	"JP": "Japan", "KE": "Kenya", "KG": "Kyrgyzstan", "KH": "Cambodia",
	"KR": "South Korea", "KW": "Kuwait", "KZ": "Kazakhstan", "LB": "Lebanon",
	"LK": "Sri Lanka", "LT": "Lithuania", "LU": "Luxembourg", "LV": "Latvia",
	"MA": "Morocco", "MD": "Moldova", "MK": "North Macedonia", "MN": "Mongolia",
	"MU": "Mauritius", "MX": "Mexico", "MY": "Malaysia", "NG": "Nigeria",
	"NL": "Netherlands", "NO": "Norway", "NP": "Nepal", "NZ": "New Zealand",
	"OM": "Oman", "PA": "Panama", "PE": "Peru", "PH": "Philippines",
	"PK": "Pakistan", "PL": "Poland", "PR": "Puerto Rico", "PT": "Portugal",
	"QA": "Qatar", "RO": "Romania", "RS": "Serbia", "RU": "Russia",
	"SA": "Saudi Arabia", "SE": "Sweden", "SG": "Singapore", "SI": "Slovenia",
	"SK": "Slovakia", "SY": "Syria", "TH": "Thailand", "TJ": "Tajikistan",
	"TM": "Turkmenistan", "TN": "Tunisia", "TR": "Turkey", "TW": "Taiwan",
	"UA": "Ukraine", "US": "United States", "UY": "Uruguay", "UZ": "Uzbekistan",
	"VE": "Venezuela", "VN": "Vietnam", "ZA": "South Africa",
}

// WriteDiagnostics writes all-results.csv and raw/dead.txt.
//
// all-results.csv columns: Link, LatencyMs, SpeedMbps, Country, Score
// Sort: Country asc, Score desc, Link asc (full deterministic).
//
// raw/dead.txt: one Link per line, sorted alphabetically, no header.
func WriteDiagnostics(outDir string, allTested, dead []selector.Result) error {
	if err := writeAllResultsCSV(outDir, allTested); err != nil {
		return err
	}
	return writeDeadList(outDir, dead)
}

func writeAllResultsCSV(outDir string, all []selector.Result) error {
	if len(all) == 0 {
		return nil
	}
	rows := make([]selector.Result, len(all))
	copy(rows, all)
	for i := range rows {
		if rows[i].Score == 0 {
			rows[i].Score = selector.Score(rows[i])
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Country != rows[j].Country {
			return rows[i].Country < rows[j].Country
		}
		if rows[i].Score != rows[j].Score {
			return rows[i].Score > rows[j].Score
		}
		return rows[i].Link < rows[j].Link
	})

	path := filepath.Join(outDir, "all-results.csv")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"Link", "LatencyMs", "SpeedMbps", "Country", "Score"}); err != nil {
		return err
	}
	for _, r := range rows {
		row := []string{
			r.Link,
			fmt.Sprintf("%d", r.LatencyMs),
			fmt.Sprintf("%.2f", r.SpeedMbps),
			r.Country,
			fmt.Sprintf("%.4f", r.Score),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func writeDeadList(outDir string, dead []selector.Result) error {
	if len(dead) == 0 {
		return nil
	}
	rawDir := filepath.Join(outDir, "raw")
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", rawDir, err)
	}

	links := make([]string, 0, len(dead))
	for _, r := range dead {
		links = append(links, r.Link)
	}
	sort.Strings(links)

	// Cap at 5000 to keep the file small and well under GitHub's 100 MB
	// hard limit. Full 300k+ dead set would blow past 80 MB and hit the
	// recommended-max warning. The first 5000 alphabetically is enough as
	// a diagnostic sample; full data lives in xray-knife.db locally.
	const maxDead = 5000
	if len(links) > maxDead {
		links = links[:maxDead]
	}

	var b strings.Builder
	for _, l := range links {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(rawDir, "dead.txt"), []byte(b.String()), 0o644)
}

// WriteMultiProtocolReadme writes the top-level README.md describing all
// curated protocols. For v2.0 multi-protocol mode. Lists subscription URLs
// per protocol and shows the per-country top-3 table for each.
func WriteMultiProtocolReadme(outDir string, protos []ProtoReadme, generatedAt time.Time) error {
	readme := buildMultiProtocolReadme(protos, generatedAt)
	return os.WriteFile(filepath.Join(outDir, "README.md"), []byte(readme), 0o644)
}

func buildMultiProtocolReadme(protos []ProtoReadme, generatedAt time.Time) string {
	var b strings.Builder
	stamp := generatedAt.UTC().Format(time.RFC3339)

	b.WriteString("# VlessFilter Results\n\n")
	b.WriteString("Auto-curated top 3 fastest proxy keys per country, refreshed automatically. ")
	b.WriteString("Multi-protocol: VLESS / VMess / Trojan / Shadowsocks.\n\n")

	b.WriteString("## How to use\n\n")
	b.WriteString("Pick the protocol your client supports best. Each has its own subscription URLs:\n\n")

	for _, p := range protos {
		title := strings.ToUpper(p.Protocol)
		b.WriteString(fmt.Sprintf("### %s\n\n", title))
		b.WriteString(fmt.Sprintf("All %s countries (single subscription):\n\n", title))
		b.WriteString(fmt.Sprintf("```\nhttps://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/%s/all.txt\n```\n\n",
			p.Protocol))
		b.WriteString("Specific country:\n\n")
		b.WriteString(fmt.Sprintf("```\nhttps://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/%s/<CC>.txt\n```\n\n",
			p.Protocol))
		b.WriteString(fmt.Sprintf("Rotating exits: `subs/%s/rotating.txt` (%d configs)\n\n",
			p.Protocol, p.Rotating))
	}

	// Top-level subs/ is the multi-protocol union (every working key,
	// every protocol, every country). Useful when the client supports
	// all 4 schemes and you want maximum coverage from one URL.
	b.WriteString("### All protocols combined (one URL → everything)\n\n")
	b.WriteString("```\nhttps://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/all.txt\n```\n\n")
	b.WriteString("Specific country across all protocols:\n\n")
	b.WriteString("```\nhttps://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/<CC>.txt\n```\n\n")
	b.WriteString("Rotating exits (all protocols): `subs/rotating.txt`\n\n")

	b.WriteString("## Stability filter\n\n")
	b.WriteString("Many public configs route through proxy chains, load balancers, or Cloudflare Workers — these have **rotating exit countries** (e.g., one connection lands in Sweden, the next in India). Tagging them with a single country would be misleading.\n\n")
	b.WriteString("Each config's full test history is checked:\n")
	b.WriteString("- **Stable** (always exits same country) → published in `subs/<protocol>/<CC>.txt` with that country code\n")
	b.WriteString("- **Rotating** (varies across tests, OR is a `*.workers.dev` / `*.pages.dev` host) → published in `subs/<protocol>/rotating.txt` with `🌐 ROTATING` label\n")
	b.WriteString("- **Dead** → not published\n\n")

	writeInstallSection(&b)
	writeOffPCDeploymentSection(&b)
	for _, p := range protos {
		title := strings.ToUpper(p.Protocol)
		fmt.Fprintf(&b, "## %s — top 3 per country (stable only)\n\n", title)
		if len(p.Selections) == 0 {
			b.WriteString("_No stable countries this run._\n\n")
			continue
		}
		b.WriteString("| Country | Top latency (ms) | Median speed (Mbps) | Keys |\n")
		b.WriteString("|---------|------------------|---------------------|------|\n")

		rows := make([]selector.CountrySelection, len(p.Selections))
		copy(rows, p.Selections)
		sort.SliceStable(rows, func(i, j int) bool { return rows[i].Country < rows[j].Country })

		for _, cs := range rows {
			topLat := 0
			speeds := make([]float64, 0, len(cs.Top))
			for _, r := range cs.Top {
				if topLat == 0 || r.LatencyMs < topLat {
					topLat = r.LatencyMs
				}
				speeds = append(speeds, r.SpeedMbps)
			}
			fmt.Fprintf(&b, "| %s %s | %d | %.1f | %d |\n",
				flagEmoji(cs.Country), cs.Country, topLat, median(speeds), len(cs.Top))
		}
		fmt.Fprintf(&b, "\n**Rotating-exit pool:** %d configs in `subs/%s/rotating.txt`\n\n",
			p.Rotating, p.Protocol)
	}

	b.WriteString("_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._\n\n")
	fmt.Fprintf(&b, "<!-- last-tested: %s -->\n", stamp)
	return b.String()
}

// writeInstallSection appends the install instructions to the generated
// top-level README. Kept in the generator (not a static docs/ file) so the
// info always survives README regeneration on every workflow run.
func writeInstallSection(b *strings.Builder) {
	b.WriteString("## Install\n\n")
	b.WriteString("VlessFilter is a single Go binary. Three install paths, pick whichever:\n\n")

	b.WriteString("### Option 1: Pre-built binary (fastest)\n\n")
	b.WriteString("Each tagged release ships Linux + macOS binaries on GitHub Releases. Pick from <https://github.com/trikiman/vlessfilter/releases/latest>.\n\n")
	b.WriteString("Linux (amd64):\n\n")
	b.WriteString("```bash\n")
	b.WriteString("curl -sSL https://github.com/trikiman/vlessfilter/releases/latest/download/vlessfilter_Linux_amd64.tar.gz \\\n")
	b.WriteString("  | tar -xz -C /tmp && sudo mv /tmp/vlessfilter /usr/local/bin/\n")
	b.WriteString("```\n\n")

	b.WriteString("### Option 2: `go install` (requires Go 1.22+)\n\n")
	b.WriteString("```bash\n")
	b.WriteString("go install github.com/trikiman/vlessfilter/cmd/vlessfilter@latest\n")
	b.WriteString("```\n")
	b.WriteString("Binary lands in `$GOPATH/bin` (or `$HOME/go/bin`). Make sure that's on your `$PATH`.\n\n")

	b.WriteString("### Option 3: From source\n\n")
	b.WriteString("```bash\n")
	b.WriteString("git clone https://github.com/trikiman/vlessfilter.git\n")
	b.WriteString("cd vlessfilter\n")
	b.WriteString("go build -o bin/vlessfilter ./cmd/vlessfilter\n")
	b.WriteString("```\n\n")

	b.WriteString("### Verify it works\n\n")
	b.WriteString("```bash\n")
	b.WriteString("vlessfilter --help\n")
	b.WriteString("# Quick smoke run against the default sources (writes ./subs/ + ./README.md):\n")
	b.WriteString("vlessfilter run --threads1 50 --threads2 5 --limit 30 --budget-min 5\n")
	b.WriteString("ls subs/\n")
	b.WriteString("```\n\n")

	b.WriteString("### Configuration\n\n")
	b.WriteString("Edit `sources.yaml` to add or remove subscription sources. See comments in the file for the schema.\n\n")
}

// writeOffPCDeploymentSection appends the off-PC / autonomous deployment guide
// to the generated top-level README. GitHub Actions (already shipped in this
// repo) is the primary recommended path; the rest are optional fallbacks.
func writeOffPCDeploymentSection(b *strings.Builder) {
	b.WriteString("## Off-PC Deployment\n\n")
	b.WriteString("Run the pipeline without using your own computer. The **primary, recommended path is GitHub Actions** (Option A) — free, fully automated, always-on, and already shipped in this repo (`.github/workflows/refresh.yml`). The other options are optional manual fallbacks.\n\n")

	b.WriteString("### What each run does\n\n")
	b.WriteString("1. **Stage 1 — alive/handshake check.** High-concurrency TLS handshake against the pool; dead keys are dropped.\n")
	b.WriteString("2. **Stage 2 — speed connection test.** Survivors get a real-proxy speedtest, run 3 separate times, to measure throughput (Mbps) and latency (ms).\n")
	b.WriteString("3. **Pre-publish probe.** Top-3-per-country selections are re-tested right before publishing so stale/dead keys never reach the results.\n\n")
	b.WriteString("Results appear in `README.md` (per-country latency + median speed table) and `all-results.csv` (full raw results).\n\n")

	b.WriteString("### Option A: GitHub Actions (PRIMARY — recommended)\n\n")
	b.WriteString("**Cost:** $0 (public repos get unlimited Actions minutes). **Setup:** ~5 min. **Always-on:** yes.\n\n")
	b.WriteString("1. Create a GitHub account (any throwaway email; no card needed for free-tier Actions on public repos).\n")
	b.WriteString("2. **Fork** `https://github.com/trikiman/vlessfilter`.\n")
	b.WriteString("3. Generate a PAT (Settings → Developer settings → Fine-grained tokens): repository access = your results repo, Permissions → Contents = **Read and write**.\n")
	b.WriteString("4. In the repo: Settings → Secrets and variables → Actions → new secret `PUSH_TOKEN` = the PAT.\n")
	b.WriteString("5. Enable Actions: Settings → Actions → General → Allow all.\n")
	b.WriteString("6. Trigger the first run: Actions tab → **refresh** workflow → Run workflow.\n\n")
	b.WriteString("`refresh.yml` runs every 4 hours, does the full alive-check + speed test, and commits fresh results — no involvement from your machine.\n\n")

	b.WriteString("### Option B: h2.nexus 15-minute ephemeral VPS (manual fallback, no signup)\n\n")
	b.WriteString("Free 15-min VPS (4 CPU / 8 GB / 1 Gbps), no account. Generate a PAT (as in Option A), open <https://h2.nexus/cli>, pick Debian 11, then in the web console run:\n\n")
	b.WriteString("```bash\n")
	b.WriteString("curl -sSL https://raw.githubusercontent.com/trikiman/vlessfilter/main/scripts/h2-quick.sh | bash -s -- ghp_xxx\n")
	b.WriteString("```\n\n")
	b.WriteString("Results push automatically; the VM auto-deletes at 15 min. Manual trigger only (no schedule), reduced run scope to fit the window.\n\n")

	b.WriteString("### Option C: Termux on Android (manual fallback)\n\n")
	b.WriteString("Install Termux from F-Droid, then:\n\n")
	b.WriteString("```bash\n")
	b.WriteString("pkg update && pkg upgrade -y\n")
	b.WriteString("pkg install -y golang git curl\n")
	b.WriteString("curl -sSL https://raw.githubusercontent.com/trikiman/vlessfilter/main/scripts/install-always-on.sh | bash -s -- github_pat_xxx\n")
	b.WriteString("```\n\n")
	b.WriteString("cron may fail under Termux — use `termux-job-scheduler --period-ms 21600000 --script $HOME/.vlessfilter/refresh.sh` (6h). Keep the phone charging and out of deep sleep for a full run.\n\n")

	b.WriteString("### Which to pick\n\n")
	b.WriteString("| Your situation | Pick |\n")
	b.WriteString("|----------------|------|\n")
	b.WriteString("| Fully autonomous, always-on, zero maintenance | **Option A** (GitHub Actions) — default |\n")
	b.WriteString("| One-off manual refresh, no account | **Option B** (h2.nexus) |\n")
	b.WriteString("| Only a phone available | **Option C** (Termux) |\n\n")

	b.WriteString("> Note: the earlier 2Z2 Cloud Labs VPS runbook is retired (the service shut down). GitHub Actions is the recommended replacement.\n\n")
}

// buildReadme renders the summary README.md.
//
// D-17 determinism: the only non-input-derived content is a single HTML
// comment at the bottom carrying the timestamp. Two runs with identical
// `selections` produce README files that diff only on that one line.
func buildReadme(selections []selector.CountrySelection, rotatingCount int, generatedAt time.Time) string {
	var b strings.Builder
	stamp := generatedAt.UTC().Format(time.RFC3339)

	b.WriteString("# VlessFilter Results\n\n")
	b.WriteString("Auto-curated top 3 fastest VLESS keys per country, refreshed automatically.\n\n")
	b.WriteString("## How to use\n\n")
	b.WriteString("**Single subscription URL covering every stable country** — paste this into your VLESS client (Hiddify Next, v2rayN, Streisand, NekoBox, etc.):\n\n")
	b.WriteString("```\nhttps://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/all.txt\n```\n\n")
	b.WriteString("Or pick a specific country:\n\n")
	b.WriteString("```\nhttps://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/<CC>.txt\n```\n\n")
	b.WriteString("**Rotating-exit configs** (Cloudflare Workers + multi-exit load balancers — country varies per connection):\n\n")
	b.WriteString("```\nhttps://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/rotating.txt\n```\n\n")
	b.WriteString("Replace `<CC>` with the 2-letter country code from the table. Entries in country files have been verified to consistently exit through that country across multiple tests.\n\n")
	b.WriteString("## Stability filter\n\n")
	b.WriteString("Many public VLESS configs route through proxy chains, load balancers, or Cloudflare Workers — these have **rotating exit countries** (e.g., one connection lands in Sweden, the next in India). Tagging them with a single country would be misleading.\n\n")
	b.WriteString("Each config's full test history is checked:\n")
	b.WriteString("- **Stable** (always exits same country) → published in `subs/<CC>.txt` with that country code\n")
	b.WriteString("- **Rotating** (varies across tests, OR is a `*.workers.dev` / `*.pages.dev` host) → published in `subs/rotating.txt` with `🌐 ROTATING` label\n")
	b.WriteString("- **Dead** → not published\n\n")
	b.WriteString("## Top 3 per country (stable only)\n\n")
	b.WriteString("| Country | Top latency (ms) | Median speed (Mbps) | Keys |\n")
	b.WriteString("|---------|------------------|---------------------|------|\n")

	rows := make([]selector.CountrySelection, len(selections))
	copy(rows, selections)
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Country < rows[j].Country })

	for _, cs := range rows {
		topLat := 0
		speeds := make([]float64, 0, len(cs.Top))
		for _, r := range cs.Top {
			if topLat == 0 || r.LatencyMs < topLat {
				topLat = r.LatencyMs
			}
			speeds = append(speeds, r.SpeedMbps)
		}
		fmt.Fprintf(&b, "| %s %s | %d | %.1f | %d |\n",
			flagEmoji(cs.Country), cs.Country, topLat, median(speeds), len(cs.Top))
	}

	fmt.Fprintf(&b, "\n**Rotating-exit pool:** %d additional configs in `subs/rotating.txt` (no country guarantee).\n\n", rotatingCount)
	b.WriteString("_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._\n\n")
	fmt.Fprintf(&b, "<!-- last-tested: %s -->\n", stamp)
	return b.String()
}

// flagEmoji maps an ISO 3166-1 alpha-2 code to its flag emoji.
func flagEmoji(cc string) string {
	if len(cc) != 2 {
		return cc
	}
	a, b := cc[0], cc[1]
	if !(a >= 'A' && a <= 'Z') || !(b >= 'A' && b <= 'Z') {
		return cc
	}
	return string([]rune{rune(a) + 0x1F1A5, rune(b) + 0x1F1A5})
}

// median returns the median of xs rounded to 1 decimal. Empty input → 0.
func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sorted := make([]float64, len(xs))
	copy(sorted, xs)
	sort.Float64s(sorted)
	n := len(sorted)
	var m float64
	if n%2 == 1 {
		m = sorted[n/2]
	} else {
		m = (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return float64(int64(m*10+0.5)) / 10
}
