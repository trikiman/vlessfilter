// Package output writes the user-facing results: per-country subscription
// files (subs/<CC>.txt), a README.md summary table, and diagnostic outputs
// (all-results.csv, raw/dead.txt) at the chosen output directory root.
//
// Format matches D-08 (README columns), D-09 (subs files: plain vless URIs,
// one per line, UTF-8, LF) and D-16/D-17 (diagnostic + deterministic).
package output

import (
	"encoding/csv"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/trikiman/vlessfilter/internal/selector"
)

// WriteAll writes everything: subs/<CC>.txt + README.md + (when diagnostics
// is non-nil) all-results.csv + raw/dead.txt. This is the function the
// pipeline calls.
//
// `selections` is the curated top-3-per-country output.
// `allTested` and `dead` are the full and failed result sets for diagnostics.
// Pass nil/empty for either to skip that file.
func WriteAll(outDir string, selections []selector.CountrySelection, allTested, dead []selector.Result, generatedAt time.Time) error {
	if err := Write(outDir, selections, generatedAt); err != nil {
		return err
	}
	return WriteDiagnostics(outDir, allTested, dead)
}

// Write produces subs/<CC>.txt files, subs/all.txt, and README.md inside outDir.
//
// Each VLESS URI's `#fragment` (the human-readable remark shown by client UIs)
// is rewritten to `<flag-emoji> <CC>` so users see country flags next to each
// entry in their VLESS client. The original remark from the upstream
// subscription is replaced — clients display whatever's in the fragment.
//
// subs/all.txt aggregates the curated top-N from every country in alphabetical
// order. One subscription URL serves every country at once.
func Write(outDir string, selections []selector.CountrySelection, generatedAt time.Time) error {
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
			rewritten := rewriteRemark(r.Link, cs.Country)
			b.WriteString(rewritten)
			b.WriteByte('\n')
			allBuf.WriteString(rewritten)
			allBuf.WriteByte('\n')
		}
		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	// Combined "all countries" subscription file. One URL → every country.
	allPath := filepath.Join(subsDir, "all.txt")
	if err := os.WriteFile(allPath, []byte(allBuf.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", allPath, err)
	}

	readme := buildReadme(sortedSel, generatedAt)
	path := filepath.Join(outDir, "README.md")
	return os.WriteFile(path, []byte(readme), 0o644)
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
func rewriteRemark(link, cc string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}
	cc = strings.ToUpper(cc)
	flag := flagEmoji(cc)
	name := countryName(cc)
	// url.URL.String() URL-encodes the Fragment automatically.
	if name == cc {
		// Unknown country — don't duplicate the code (e.g., "🇿🇿 ZZ" not "🇿🇿 ZZ ZZ").
		u.Fragment = fmt.Sprintf("%s %s", flag, cc)
	} else {
		u.Fragment = fmt.Sprintf("%s %s %s", flag, cc, name)
	}
	return u.String()
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

	var b strings.Builder
	for _, l := range links {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(rawDir, "dead.txt"), []byte(b.String()), 0o644)
}

// buildReadme renders the summary README.md.
//
// D-17 determinism: the only non-input-derived content is a single HTML
// comment at the bottom carrying the timestamp. Two runs with identical
// `selections` produce README files that diff only on that one line.
func buildReadme(selections []selector.CountrySelection, generatedAt time.Time) string {
	var b strings.Builder
	stamp := generatedAt.UTC().Format(time.RFC3339)

	b.WriteString("# VlessFilter Results\n\n")
	b.WriteString("Auto-curated top 3 fastest VLESS keys per country, refreshed automatically.\n\n")
	b.WriteString("## How to use\n\n")
	b.WriteString("**Single subscription URL covering every country** — paste this into your VLESS client (Hiddify Next, v2rayN, Streisand, NekoBox, etc.):\n\n")
	b.WriteString("```\nhttps://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/all.txt\n```\n\n")
	b.WriteString("Or pick a specific country:\n\n")
	b.WriteString("```\nhttps://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/<CC>.txt\n```\n\n")
	b.WriteString("Replace `<CC>` with the 2-letter country code from the table. Each entry's name shows a flag emoji and country code so you can see at a glance which key you're connecting through.\n\n")
	b.WriteString("## Top 3 per country\n\n")
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

	b.WriteString("\n_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._\n\n")
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
