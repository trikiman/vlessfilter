# Phase 1: MVP End-to-End - Research

**Researched:** 2026-05-14
**Domain:** VLESS proxy testing & subscription management on Linux, Go orchestration
**Confidence:** HIGH (engine + upstream feed both verified by reading actual READMEs; CSV column names need a one-line confirmation in execution)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- D-01: xray-knife integration via CLI subprocess (not Go library)
- D-02: Auto-install xray-knife with `go install ...@<pinned>` if missing on `$PATH`
- D-03: sources.yaml uses typed entries `{name, url, kind, enabled}`; `kind: country-template` allows `{CC}` placeholder
- D-04: Default sources.yaml ships with 2 sources only for Phase 1 smoke test
- D-05: Single command `vlessfilter run` + `--stage {fetch,test,select}` flag for stage isolation
- D-06: Score = `0.6 * norm_speed - 0.4 * norm_latency`, both capped at sane upper bounds
- D-07: Tie-break by latency asc; partial output for countries with <3 keys; omit countries with 0 keys
- D-08: README columns: Country (flag + ISO) | Top latency (ms) | Median speed (Mbps) | Keys | Last tested UTC
- D-09: subs/<CC>.txt is plain `vless://` URIs, one per line, UTF-8, LF
- D-10: Two test tiers — unit tests (CI, mocked) + manual smoke (real network, `make smoke`)
- D-11: Standard Go layout: `cmd/vlessfilter/`, `internal/{sources,xrayknife,selector,output,pipeline}/`, `testdata/`

### Claude's Discretion
- Log format / log level (use `log/slog`)
- xray-knife version detection (parse `--version` output)
- Concurrency primitives in stage 1 driver (delegate to xray-knife `-t`)
- Exact CLI flag names beyond `--stage` (`--source`, `--limit`, `--out-dir`)
- Error wrapping conventions

### Deferred Ideas (OUT OF SCOPE for Phase 1)
- Kernel tuning (Phase 2)
- 60-min budget enforcement (Phase 2)
- Checkpoint / resume / git push (Phase 2)
- `all-results.csv`, `raw/dead.txt` diagnostics (Phase 2)
- Deterministic byte-identical output guarantees (Phase 2)
- GitHub Actions workflow (Phase 3)
- Other protocols, multi-region, Telegram scraping, web UI (v2 / out of scope)

</user_constraints>

<research_summary>
## Summary

Two pre-existing tools cover ~95% of Phase 1's heavy lifting:
1. **`xray-knife`** (lilendian0x00, Go, MIT) — does VLESS subscription parsing, SQLite-backed library, multi-threaded handshake ping, real-proxy speedtest with **built-in exit-IP geolocation**, dual-engine (xray-core + sing-box).
2. **`v2go`** (Danialsamadi, Go, GPL-3) — already aggregates 451k → 21k unique configs every 6 hours, splits by country into `Splitted-By-Country/<CC>.txt` files. We consume its outputs as upstream feed (data-only dependency, not code).

The orchestrator we're building is a thin Go layer that:
1. Reads `sources.yaml`, fetches subscription URLs (HTTPS GET, base64-decode if needed)
2. Feeds them to xray-knife via `subs add` + `subs fetch`
3. Runs `xray-knife http --from-db --protocol vless` for stage 1 (handshake) then again with `--speedtest` for stage 2
4. Reads `~/.xray-knife/xray-knife.db` (SQLite) directly to get per-key results with country
5. Groups by country, applies composite score (D-06), picks top 3 per country (D-07)
6. Writes `subs/<CC>.txt` + `README.md` table

**Primary recommendation:** Build the orchestrator as 2 plans — one for ingestion infrastructure (sources.yaml loader + xray-knife subprocess wrapper + auto-install), one for the result-reading + selector + output formatter. Both plans use mocked xray-knife in unit tests and rely on `make smoke` for real-network validation.

</research_summary>

<standard_stack>
## Standard Stack

### Core (Go stdlib + 2 deps)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.22+ | Language | Matches xray-knife's Go 1.25 install instructions; modern stdlib has `log/slog`, `errors.Join`, `slices` |
| `gopkg.in/yaml.v3` | latest | sources.yaml parsing | De facto standard YAML lib for Go; well-maintained |
| `github.com/mattn/go-sqlite3` | latest | Read xray-knife.db | Most-used Go SQLite driver, stable, CGO-based |
| `log/slog` (stdlib) | — | Structured logging | Stdlib since 1.21; no external dep needed |
| `os/exec` (stdlib) | — | xray-knife subprocess | Stdlib subprocess management |
| `net/http` (stdlib) | — | Subscription URL fetching | Stdlib HTTP client; no external dep needed |

### External binary (auto-installed)

| Binary | Install Command | Purpose |
|--------|----------------|---------|
| `xray-knife` | `go install github.com/lilendian0x00/xray-knife/v9@<PIN>` | Subscription manager + tester |

We pin a specific tag in code (start with `v9.0.0` or whatever's latest stable; check at implementation time). Pinning gives us reproducible test results.

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `gopkg.in/yaml.v3` | `github.com/goccy/go-yaml` | Faster, but yaml.v3 is enough for our small files |
| Subprocess | xray-knife as Go library | Requires vendoring xray-core + sing-box transitive deps; rejected per D-01 |
| `mattn/go-sqlite3` | `modernc.org/sqlite` | Pure-Go (no CGO), but slower; CGO is fine for a CLI |

### Installation

```bash
go mod init github.com/<user>/vlessfilter
go get gopkg.in/yaml.v3@latest
go get github.com/mattn/go-sqlite3@latest
```

</standard_stack>

<xray_knife_interface>
## xray-knife CLI surface (what we wrap)

From the xray-knife README (lilendian0x00/xray-knife/blob/master/README.md):

### Subscription management
```bash
xray-knife subs add --url "<URL>" --remark "<NAME>"   # add a subscription
xray-knife subs show                                   # list all subscriptions
xray-knife subs fetch --id <N>                         # fetch one
xray-knife subs fetch                                  # fetch all
```

Subscriptions populate the SQLite library at `~/.xray-knife/xray-knife.db` with deduped, parsed configs.

### Testing
```bash
# From the database (preferred — uses subs we added):
xray-knife http --from-db --protocol vless                              # stage 1: handshake/latency
xray-knife http --from-db --protocol vless --speedtest -t 20            # stage 2: latency + speed + IP location
xray-knife http --from-db --protocol vless --speedtest --limit 5000     # cap input size

# View results from latest run:
xray-knife http list-results --limit 20
```

### Database location
`~/.xray-knife/xray-knife.db` (SQLite, queryable directly via Go's `database/sql`).

### CSV export (alternative to direct SQLite read)
```bash
xray-knife http -f configs.txt --speedtest --sort --type csv -o results.csv
```

### Schema unknowns to verify in execution
The exact column names in the SQLite results table need to be confirmed by running `xray-knife http list-results --limit 1` once during plan 01-02 implementation. Expected columns based on the README: `Link`, `Delay` (latency ms), `DownloadSpeed` (Mbps), `UploadSpeed`, `IpAddress`, `Location`, plus a `subscription_id` foreign key. **Action item embedded in 01-02 task 1**: read the actual schema and adapt the Go query.

</xray_knife_interface>

<v2go_interface>
## v2go upstream feed (what we consume)

From the v2go README (Danialsamadi/v2go/blob/main/README.md):

- Repo refreshes every 6 hours via its own GitHub Action
- `Splitted-By-Country/<CC>.txt` files contain plain-text vless/vmess/trojan URIs, one per line
- 100+ countries with files
- Examples:
  - `https://raw.githubusercontent.com/Danialsamadi/v2go/main/Splitted-By-Country/US.txt`
  - `https://raw.githubusercontent.com/Danialsamadi/v2go/main/Splitted-By-Country/DE.txt`
- Also has `Splitted-By-Protocol/vless.txt` for protocol-only filtering

In `sources.yaml`, we declare a single `kind: country-template` entry:
```yaml
- name: v2go-country
  url: https://raw.githubusercontent.com/Danialsamadi/v2go/main/Splitted-By-Country/{CC}.txt
  kind: country-template
  enabled: true
```

For Phase 1 we expand this template to **2 specific countries (US, DE)** to keep the smoke test fast. Phase 2 expands to ~100 countries.

</v2go_interface>

<gotchas>
## Known Gotchas

1. **xray-knife auto-install requires Go on PATH.** If `go` is missing, we must error clearly. Check with `which go` before attempting `go install`.
2. **xray-knife creates `~/.xray-knife/` on first run.** We don't need to pre-create it; just verify the dir exists after first invocation.
3. **xray-knife's `subs add` is idempotent by URL.** Re-running with the same URL gives a "duplicate" warning but doesn't fail. Plan handles this by ignoring the warning in stdout.
4. **Subscription URL responses can be plain text OR base64.** Auto-detect: try base64-decode first; if result has `vless://` substring, use it; else treat input as plain text.
5. **xray-knife may take 10-30s to fetch a sub for the first time** (DNS, TLS, parsing). Phase 1 smoke test should expect that latency.
6. **Country codes in xray-knife results are likely 2-letter ISO** (US, DE, etc.) but unverified — confirm in plan 01-02 task 1.

</gotchas>

<test_strategy>
## Test Strategy

### Unit tests (Phase 1, CI-runnable)
- `internal/sources`: parsing yaml, country-template expansion, base64 vs plain detection (pure logic, no I/O)
- `internal/selector`: composite score, group-by, top-3 with ties (pure logic, fed by fixture rows)
- `internal/output`: README + subs/<CC>.txt formatters (golden file comparison via `testdata/`)
- `internal/xrayknife`: subprocess wrapper using a mock command (interface-based; `Runner` interface with real and fake impls)

Target: 70%+ coverage on `selector` + `output`. `xrayknife` and `sources` get spot-coverage via mocks.

### Smoke test (Phase 1, manual gate, real network)
`make smoke` invokes `vlessfilter run` against the default 2-source `sources.yaml`. Asserts:
1. `subs/` directory exists with at least 1 file
2. `README.md` exists at repo root and contains `| Country |` header
3. Exit code 0
4. Total wall time < 5 minutes on a developer machine

Not run in CI (requires network + ~5 min, public subs are flaky).

</test_strategy>

---

*Researched: 2026-05-14*
*Sources: xray-knife README (master branch, MIT), v2go README (main branch, GPL-3), GSD teammate channel (4 rounds)*
