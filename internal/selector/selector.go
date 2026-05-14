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
// Filters out rows with LatencyMs == 0 || LatencyMs > 10000 (treated as
// failed/unmeasurable).
func LoadResults(ctx context.Context, dbPath string) ([]Result, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dbPath, err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping %s: %w", dbPath, err)
	}

	tables, err := listTables(ctx, db)
	if err != nil {
		return nil, err
	}
	slog.Debug("xray-knife db tables", "tables", tables)

	resultsTable, err := pickResultsTable(tables)
	if err != nil {
		return nil, err
	}
	slog.Debug("results table", "name", resultsTable)

	cols, err := tableColumns(ctx, db, resultsTable)
	if err != nil {
		return nil, err
	}
	slog.Debug("results columns", "table", resultsTable, "columns", cols)

	mapping, err := mapColumns(cols)
	if err != nil {
		return nil, err
	}

	return queryResults(ctx, db, resultsTable, mapping)
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
// contains "result" or "test".
func pickResultsTable(tables []string) (string, error) {
	preferred := []string{"test_results", "results", "http_results", "http_tests"}
	for _, p := range preferred {
		for _, t := range tables {
			if strings.EqualFold(t, p) {
				return t, nil
			}
		}
	}
	for _, t := range tables {
		lt := strings.ToLower(t)
		if strings.Contains(lt, "result") || strings.Contains(lt, "test") {
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
		Link:    pick("link", "config_link", "config", "uri", "url"),
		Latency: pick("delay", "latency", "latency_ms", "ping", "ping_ms"),
		Speed:   pick("downloadspeed", "download_speed", "speed_mbps", "speed", "download"),
		Country: pick("location", "country", "country_code", "ipaddress", "ip_country"),
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
