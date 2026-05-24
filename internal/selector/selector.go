// Package selector reads xray-knife's SQLite results database, applies the
// composite ranking score (per D-06), and groups results by exit-IP country
// to produce top-3-per-country selections.
//
// Schema discovery is dynamic: xray-knife may version its schema across
// releases. selector queries sqlite_master and PRAGMA table_info to find the
// results table and map its columns to our Result type.
package selector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"sort"
	"strings"

	_ "modernc.org/sqlite" // pure-Go sqlite driver (no CGO)
)

// Composite-score constants per D-06 (CONTEXT.md).
const (
	MaxLatencyMs = 1000.0
	MaxSpeedMbps = 100.0
	WSpeed       = 0.6
	WLatency     = 0.4
)

// Result is one tested key after we've populated all measured fields.
type Result struct {
	Link      string  // full vless:// URI
	LatencyMs int     // round-trip latency
	SpeedMbps float64 // download throughput
	Country   string  // 2-letter ISO; empty if xray-knife failed to detect
	Score     float64 // composite, computed by Score()
}

// CountrySelection is one row of the final output: country code + up to 3
// best Results sorted by score desc (latency asc as tie-breaker).
type CountrySelection struct {
	Country string
	Top     []Result
}

// Score implements D-06: 0.6*norm_speed - 0.4*norm_latency, both capped.
//
// Higher score = better. The caps prevent a single fast-but-far outlier from
// dominating the global ranking; everything above MaxSpeedMbps maps to 1.0
// for the speed term and everything above MaxLatencyMs maps to 1.0 for the
// latency penalty.
func Score(r Result) float64 {
	normSpeed := math.Min(r.SpeedMbps, MaxSpeedMbps) / MaxSpeedMbps
	normLatency := math.Min(float64(r.LatencyMs), MaxLatencyMs) / MaxLatencyMs
	return WSpeed*normSpeed - WLatency*normLatency
}

// Top3PerCountry computes scores for each result, groups by Country, sorts
// each group, and returns top 3 per country sorted alphabetically by code.
//
// Behavior per D-07:
//   - Empty Country → row dropped
//   - <3 alive keys per country → partial output (1 or 2 entries)
//   - Tie on Score → lower LatencyMs wins
//   - Output countries listed alphabetically by ISO code
func Top3PerCountry(results []Result) []CountrySelection {
	groups := make(map[string][]Result)
	for _, r := range results {
		if r.Country == "" {
			continue
		}
		r.Score = Score(r)
		groups[r.Country] = append(groups[r.Country], r)
	}

	out := make([]CountrySelection, 0, len(groups))
	const scoreEps = 1e-9 // tolerate IEEE 754 noise so identical math doesn't desync ordering
	for cc, rs := range groups {
		sort.SliceStable(rs, func(i, j int) bool {
			if math.Abs(rs[i].Score-rs[j].Score) > scoreEps {
				return rs[i].Score > rs[j].Score
			}
			return rs[i].LatencyMs < rs[j].LatencyMs
		})
		if len(rs) > 3 {
			rs = rs[:3]
		}
		out = append(out, CountrySelection{Country: cc, Top: rs})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Country < out[j].Country })
	return out
}

// LoadResults opens the xray-knife SQLite DB, discovers the results table and
// columns, and returns Result rows from the most recent test run only.
//
// LoadUntestedLinks returns config_links from subscription_configs that have
// no row in http_test_results yet, capped to `limit` rows.
//
// Used by the pipeline to incrementally cover the pool: each scheduled run
// tests the next batch of NEW configs instead of trying to retest the entire
// (now 775k+) pool and getting killed by the budget.
//
// Returns nil when limit <= 0 (caller should fall back to whole-pool test).
func LoadUntestedLinks(ctx context.Context, dbPath string, limit int) ([]string, error) {
	if limit <= 0 {
		return nil, nil
	}
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dbPath, err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping %s: %w", dbPath, err)
	}

	q := `
		SELECT sc.config_link FROM subscription_configs sc
		LEFT JOIN http_test_results r ON r.config_link = sc.config_link
		WHERE sc.protocol = 'vless' AND r.config_link IS NULL
		LIMIT ?
	`
	rows, err := db.QueryContext(ctx, q, limit)
	if err != nil {
		// Tolerate missing subscription_configs table (e.g., test DBs that
		// only seed http_test_results). Caller falls back to retesting
		// the existing alive set without an "untested" batch.
		if strings.Contains(err.Error(), "no such table") {
			return nil, nil
		}
		return nil, fmt.Errorf("query untested: %w", err)
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var link string
		if err := rows.Scan(&link); err != nil {
			return nil, err
		}
		if link != "" {
			out = append(out, link)
		}
	}
	return out, rows.Err()
}

// LoadAliveLinks returns just the config_link strings for currently-alive
// configs (latest result per link with status=passed). Used by the pipeline
// to feed stage 2 (speedtest) only the stage-1 survivors instead of
// re-testing the whole pool.
//
// Returns empty slice (not error) when no alive configs exist.
func LoadAliveLinks(ctx context.Context, dbPath string) ([]string, error) {
	alive, _, err := LoadAllResults(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(alive))
	for _, r := range alive {
		if r.Link != "" {
			out = append(out, r.Link)
		}
	}
	return out, nil
}

// LoadStableAndRotating returns alive configs split into two buckets:
//
//   - stable:   configs where every passing test across history reported
//     the same country. Safe to publish under that country.
//   - rotating: configs that exited via 2+ different countries across
//     history (proxy chains, load balancers, CF Workers).
//     Useful but cannot be honestly tagged with one country.
//
// Cloudflare Worker / Pages configs are forced into rotating regardless
// of test history because their exit is determined by upstream selection
// per-connection (i.e., even if past tests happened to all hit one country,
// a future connection might exit elsewhere).
//
// Each Result returned uses the latest passing test as the snapshot
// (latency, speed, country at that test), but the Country field reflects
// the consensus determined here.
func LoadStableAndRotating(ctx context.Context, dbPath string) (stable, rotating []Result, err error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_busy_timeout=5000")
	if err != nil {
		return nil, nil, fmt.Errorf("open %s: %w", dbPath, err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("ping %s: %w", dbPath, err)
	}

	tables, err := listTables(ctx, db)
	if err != nil {
		return nil, nil, err
	}
	resultsTable, terr := pickResultsTable(tables)
	if terr != nil {
		return nil, nil, nil
	}
	cols, err := tableColumns(ctx, db, resultsTable)
	if err != nil {
		return nil, nil, err
	}
	mapping, merr := mapColumns(cols)
	if merr != nil {
		return nil, nil, merr
	}

	// Pull every passing test result (not just latest-per-link) so we can
	// see the full country history per config.
	//
	// "Passing" is determined by status='passed' if the column exists,
	// else by latency_ms > 0 && <= 10000 (legacy/fixture-test compatibility).
	statusCol := statusColumn(cols)
	hasStatus := false
	for _, c := range cols {
		if strings.EqualFold(c, statusCol) {
			hasStatus = true
			break
		}
	}
	whereClause := fmt.Sprintf("%s > 0 AND %s <= 10000",
		quote(mapping.Latency), quote(mapping.Latency))
	if hasStatus {
		whereClause = fmt.Sprintf("%s = 'passed' AND %s",
			quote(statusCol), whereClause)
	}
	q := fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s
		FROM %s
		WHERE %s
	`,
		quote(mapping.Link),
		quote(mapping.Latency),
		quote(mapping.Speed),
		func() string {
			if mapping.Country != "" {
				return quote(mapping.Country)
			}
			return "''"
		}(),
		func() string {
			if mapping.RunID != "" {
				return quote(mapping.RunID)
			}
			return "0"
		}(),
		quote(resultsTable),
		whereClause,
	)

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, nil, fmt.Errorf("query passing rows: %w", err)
	}
	defer rows.Close()

	type pass struct {
		latency int
		speed   float64
		country string
		runID   int
	}
	history := make(map[string][]pass) // config_link -> all passing tests
	for rows.Next() {
		var (
			link    sql.NullString
			latency sql.NullInt64
			speed   sql.NullFloat64
			country sql.NullString
			runID   sql.NullInt64
		)
		if err := rows.Scan(&link, &latency, &speed, &country, &runID); err != nil {
			return nil, nil, err
		}
		if link.String == "" {
			continue
		}
		history[link.String] = append(history[link.String], pass{
			latency: int(latency.Int64),
			speed:   speed.Float64,
			country: strings.ToUpper(strings.TrimSpace(country.String)),
			runID:   int(runID.Int64),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	for link, passes := range history {
		// Latest pass = highest run_id (snapshot for latency/speed values).
		latest := passes[0]
		for _, p := range passes[1:] {
			if p.runID > latest.runID {
				latest = p
			}
		}

		// Determine consensus country: count distinct non-empty countries.
		countries := make(map[string]bool)
		for _, p := range passes {
			if p.country != "" {
				countries[p.country] = true
			}
		}

		r := Result{
			Link:      link,
			LatencyMs: latest.latency,
			SpeedMbps: latest.speed,
			Country:   latest.country,
		}

		// CF Worker / Pages = always rotating regardless of history
		// (next connection might exit somewhere else).
		if isCFWorker(link) {
			rotating = append(rotating, r)
			continue
		}

		// LIVE-01 / GEO-02: require 2+ passing tests for a config to be
		// considered "alive enough" to publish. Single-test passes can
		// be flukes (handshake-passes-but-no-traffic, or transient
		// server bursts). Two passes is much harder to fake.
		const minPassesForStable = 2
		if len(passes) < minPassesForStable {
			// Not enough evidence yet — held back from publication.
			// Will surface in a future run when more tests confirm.
			continue
		}

		switch len(countries) {
		case 0:
			// No country info ever → can't honestly classify. Skip.
			continue
		case 1:
			// Same country every pass → stable. Use that country.
			for c := range countries {
				r.Country = c
			}
			stable = append(stable, r)
		default:
			// Rotated across 2+ countries → not stable for any single one.
			// Keep latest country in the field for diagnostics but caller
			// should treat as rotating.
			rotating = append(rotating, r)
		}
	}
	return stable, rotating, nil
}

// isCFWorker reports whether a vless:// link points at a Cloudflare Workers
// or Pages domain. Detection is by host header (the WebSocket Host=) since
// that's how the worker routes — the IP address is just any CF edge.
func isCFWorker(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	q := u.Query()
	host := strings.ToLower(q.Get("host"))
	if host == "" {
		host = strings.ToLower(u.Host)
	}
	return strings.Contains(host, ".workers.dev") ||
		strings.Contains(host, ".pages.dev")
}

// statusColumn returns the actual column name for "status" in the results
// table. xray-knife uses 'status'; some forks use 'state'. Falls back to
// 'status' which is always-present in xray-knife 9.x.
func statusColumn(cols []string) string {
	for _, c := range cols {
		if strings.EqualFold(c, "status") {
			return c
		}
	}
	return "status"
}

// Filters out rows with LatencyMs == 0 || LatencyMs > 10000 (treated as
// failed/unmeasurable).
func LoadResults(ctx context.Context, dbPath string) ([]Result, error) {
	alive, _, err := LoadAllResults(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	if len(alive) == 0 {
		return nil, errors.New("no usable results in xray-knife db; did stage 2 speedtest run?")
	}
	return alive, nil
}

// LoadAllResults returns both alive (latency in (0, 10000]) and dead
// (latency=0 or >10000) rows from the most-recent test run.
//
// Unlike LoadResults, returns no error when the DB is empty — pipeline
// checkpoint loops call this before stage 2 has any data.
func LoadAllResults(ctx context.Context, dbPath string) (alive, dead []Result, err error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_busy_timeout=5000")
	if err != nil {
		return nil, nil, fmt.Errorf("open %s: %w", dbPath, err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("ping %s: %w", dbPath, err)
	}

	tables, err := listTables(ctx, db)
	if err != nil {
		return nil, nil, err
	}
	resultsTable, terr := pickResultsTable(tables)
	if terr != nil {
		// No results table yet → empty, not an error.
		slog.Debug("LoadAllResults: no results table yet", "tables", tables)
		return nil, nil, nil
	}

	cols, err := tableColumns(ctx, db, resultsTable)
	if err != nil {
		return nil, nil, err
	}
	mapping, merr := mapColumns(cols)
	if merr != nil {
		return nil, nil, merr
	}

	rows, err := queryAllResults(ctx, db, resultsTable, mapping)
	if err != nil {
		return nil, nil, err
	}
	for _, r := range rows {
		if r.LatencyMs > 0 && r.LatencyMs <= 10000 {
			alive = append(alive, r)
		} else {
			dead = append(dead, r)
		}
	}
	return alive, dead, nil
}

// queryAllResults returns the latest row PER config_link across all runs.
//
// Why per-link instead of "MAX(run_id) only"? Stages can run on different
// subsets of configs (xray-knife's --limit doesn't filter by previous-run
// status), so a config that passed in run 1 might never appear in run 2.
// We want to keep its alive result, not discard it because run 2 is newer.
// Conversely, if a link IS retested and now fails, the newer (failed)
// result correctly overrides.
//
// Implementation: GROUP BY config_link with MAX(run_id) per link. Works
// even when the run-id column is missing (we fall back to a plain SELECT).
func queryAllResults(ctx context.Context, db *sql.DB, table string, m columnMapping) ([]Result, error) {
	cols := []string{quote(m.Link), quote(m.Latency), quote(m.Speed)}
	if m.Country != "" {
		cols = append(cols, quote(m.Country))
	} else {
		cols = append(cols, "''")
	}

	var q string
	if m.RunID != "" {
		// SQLite respects "max trick": GROUP BY x with bare aggregates over
		// other cols returns the row corresponding to MAX(run_id) per link.
		// Reference: https://sqlite.org/lang_select.html#bareagg
		inner := fmt.Sprintf(
			"SELECT %s, MAX(%s) AS _rid, %s, %s, %s FROM %s GROUP BY %s",
			quote(m.Link), quote(m.RunID), quote(m.Latency), quote(m.Speed),
			func() string {
				if m.Country != "" {
					return quote(m.Country)
				}
				return "''"
			}(),
			quote(table), quote(m.Link),
		)
		q = fmt.Sprintf(
			"SELECT %s, %s, %s, %s FROM (%s)",
			quote(m.Link), quote(m.Latency), quote(m.Speed),
			func() string {
				if m.Country != "" {
					return quote(m.Country)
				}
				return "''"
			}(),
			inner,
		)
	} else {
		q = fmt.Sprintf("SELECT %s FROM %s", strings.Join(cols, ", "), quote(table))
	}

	r, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", table, err)
	}
	defer r.Close()

	var out []Result
	for r.Next() {
		var (
			link    sql.NullString
			latency sql.NullInt64
			speed   sql.NullFloat64
			country sql.NullString
		)
		if err := r.Scan(&link, &latency, &speed, &country); err != nil {
			return nil, err
		}
		out = append(out, Result{
			Link:      link.String,
			LatencyMs: int(latency.Int64),
			SpeedMbps: speed.Float64,
			Country:   strings.ToUpper(strings.TrimSpace(country.String)),
		})
	}
	return out, r.Err()
}

// listTables returns the names of all user tables in the DB.
func listTables(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// pickResultsTable picks the most likely results table from the candidates.
// Heuristic: prefer exact match on common names, else any table whose name
// contains "result" but not "run" (xray-knife has both http_test_results
// and http_test_runs; we only want the former).
func pickResultsTable(tables []string) (string, error) {
	// Order matters: most-specific first. Real xray-knife (≥9.x) uses
	// http_test_results; older variants used the simpler names.
	preferred := []string{"http_test_results", "test_results", "results", "http_results", "http_tests"}
	for _, p := range preferred {
		for _, t := range tables {
			if strings.EqualFold(t, p) {
				return t, nil
			}
		}
	}
	for _, t := range tables {
		lt := strings.ToLower(t)
		// Skip the runs table — it has metadata, not per-config rows.
		if strings.Contains(lt, "run") && !strings.Contains(lt, "result") {
			continue
		}
		if strings.Contains(lt, "result") {
			return t, nil
		}
	}
	return "", fmt.Errorf("could not find xray-knife results table; tables present: %v. Run `xray-knife http list-results` to verify schema", tables)
}

// tableColumns returns column names for the given table via PRAGMA.
func tableColumns(ctx context.Context, db *sql.DB, table string) ([]string, error) {
	q := fmt.Sprintf("PRAGMA table_info(%q)", table)
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("table_info(%s): %w", table, err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var (
			cid     int
			name    string
			cType   string
			notNull int
			defVal  sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &cType, &notNull, &defVal, &pk); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// columnMapping resolves logical fields to actual column names.
type columnMapping struct {
	Link    string
	Latency string
	Speed   string
	Country string
	RunID   string // optional — empty if no run/created_at column found
}

// mapColumns finds xray-knife's column names by trying common variants.
func mapColumns(cols []string) (columnMapping, error) {
	colSet := make(map[string]string, len(cols)) // lowercase -> actual case
	for _, c := range cols {
		colSet[strings.ToLower(c)] = c
	}
	pick := func(opts ...string) string {
		for _, o := range opts {
			if actual, ok := colSet[strings.ToLower(o)]; ok {
				return actual
			}
		}
		return ""
	}

	m := columnMapping{
		Link:    pick("config_link", "link", "config", "uri", "url"),
		Latency: pick("delay_ms", "delay", "latency_ms", "latency", "ping_ms", "ping"),
		Speed:   pick("download_mbps", "downloadspeed", "download_speed", "speed_mbps", "speed", "download"),
		Country: pick("ip_location", "location", "country", "country_code", "ipaddress", "ip_country"),
		RunID:   pick("run_id", "test_run_id", "created_at", "tested_at", "timestamp"),
	}

	var missing []string
	if m.Link == "" {
		missing = append(missing, "link")
	}
	if m.Latency == "" {
		missing = append(missing, "latency")
	}
	if m.Speed == "" {
		missing = append(missing, "speed")
	}
	if m.Country == "" {
		// Country may legitimately be missing if speedtest hasn't run; we
		// proceed but log a warning. Empty Country rows get dropped in
		// Top3PerCountry anyway.
		slog.Warn("results table has no country/location column; rows will be dropped")
	}
	if len(missing) > 0 {
		return m, fmt.Errorf("xray-knife results table missing required columns: %v (have: %v)", missing, cols)
	}
	return m, nil
}

// queryResults executes the actual SELECT, scoped to the most recent run when
// a RunID column is available.
func queryResults(ctx context.Context, db *sql.DB, table string, m columnMapping) ([]Result, error) {
	cols := []string{quote(m.Link), quote(m.Latency), quote(m.Speed)}
	if m.Country != "" {
		cols = append(cols, quote(m.Country))
	} else {
		cols = append(cols, "''")
	}
	q := fmt.Sprintf("SELECT %s FROM %s", strings.Join(cols, ", "), quote(table))
	if m.RunID != "" {
		q += fmt.Sprintf(" WHERE %s = (SELECT MAX(%s) FROM %s)", quote(m.RunID), quote(m.RunID), quote(table))
	}

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", table, err)
	}
	defer rows.Close()

	var out []Result
	for rows.Next() {
		var (
			link    sql.NullString
			latency sql.NullInt64
			speed   sql.NullFloat64
			country sql.NullString
		)
		if err := rows.Scan(&link, &latency, &speed, &country); err != nil {
			return nil, err
		}
		l := int(latency.Int64)
		if l == 0 || l > 10000 {
			continue
		}
		out = append(out, Result{
			Link:      link.String,
			LatencyMs: l,
			SpeedMbps: speed.Float64,
			Country:   strings.ToUpper(strings.TrimSpace(country.String)),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, errors.New("no usable results in xray-knife db; did stage 2 speedtest run?")
	}
	return out, nil
}

// quote wraps a SQLite identifier in double quotes, escaping any embedded quotes.
func quote(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}
