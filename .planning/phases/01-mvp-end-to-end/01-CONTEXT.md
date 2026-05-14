# Phase 1: MVP End-to-End - Context

**Gathered:** 2026-05-14
**Status:** Ready for planning
**Mode:** `--auto` (decisions auto-picked from PROJECT.md + research; review and override below if needed)

<domain>
## Phase Boundary

Deliver an end-to-end VlessFilter pipeline that runs locally on Linux: read `sources.yaml` → fetch and decode subscriptions → filter dead keys with a handshake stage → speedtest survivors via xray-knife → group by exit-IP country → write top-3-per-country output files (`subs/<CC>.txt` + `README.md` summary).

**In scope for Phase 1**
- Single binary that runs the full happy path on a developer's local Linux machine
- xray-knife integration (CLI subprocess, not library)
- sources.yaml ingest with typed source entries
- Read xray-knife's SQLite results, group by country, pick top 3, write outputs
- Smoke-test against a small real public sub (manual gate, not CI)

**Explicitly NOT in Phase 1** (deferred to Phase 2/3)
- Kernel tuning (`ulimit`, `tcp_tw_reuse`, port range expansion)
- 60-minute wall-clock budget enforcement
- Checkpoint / resume / partial-result commits
- GitHub auth + git push automation
- `all-results.csv` and `raw/dead.txt` diagnostic outputs
- Deterministic output ordering guarantees
- GitHub Actions cron workflow
- Goreleaser / release automation

</domain>

<decisions>
## Implementation Decisions

### xray-knife integration approach
- **D-01:** Use xray-knife as a **CLI subprocess** (not as a Go library import). Spawn `xray-knife subs add/fetch/show` and `xray-knife http -f --speedtest` via `os/exec`, parse JSON or CSV output.
  - *Why:* xray-knife is a binary-first project; importing as a library risks Go module incompatibility, version drift, and pulls in xray-core + sing-box transitively. Subprocess gives a clean process boundary, easier mocking for tests, and we get every xray-knife feature for free including future additions. Cost is process-spawn overhead, which is negligible at our scale.
  - *Revisit if:* Phase 2 benchmarks show subprocess overhead is >5% of pipeline time, or we need stage-2-during-stage-1 streaming results.
  - `[auto] Selected: A) CLI subprocess (recommended)`

### xray-knife installation
- **D-02:** Auto-install xray-knife on first run if `xray-knife` binary is missing on `$PATH`. Use `go install github.com/lilendian0x00/xray-knife/v9@<pinned-version>`. Pinned version stored as a constant in code.
  - *Why:* zero-touch deployment is a hard requirement for ephemeral VPSes (Phase 2), so let's establish the pattern in Phase 1. Version pin keeps results reproducible.
  - `[auto] Selected: A) auto-install on startup if missing (recommended)`

### sources.yaml schema
- **D-03:** Typed entries with metadata, not a flat URL list:
  ```yaml
  sources:
    - name: v2go-by-country
      url: https://raw.githubusercontent.com/Danialsamadi/v2go/main/Splitted-By-Country/{CC}.txt
      kind: country-template   # url has {CC} placeholder; expand for each country we care about
      enabled: true
    - name: barry-far-vless
      url: https://raw.githubusercontent.com/barry-far/V2ray-Configs/main/Splitted-By-Protocol/vless.txt
      kind: plain              # plain text or base64; auto-detected
      enabled: true
  ```
  - *Why:* `name` makes logs readable, `kind` lets us treat country-template URLs differently (don't have to manually list 100 countries), `enabled` lets users toggle without deleting. Cost is one extra struct vs flat list — trivial.
  - `[auto] Selected: B) typed entries with metadata (recommended)`

### Phase 1 ships with a tiny default sources.yaml
- **D-04:** Phase 1 ships with `sources.yaml` containing **2 sources only** for the smoke test:
  - `v2go` `Splitted-By-Country/US.txt` (a known-populated country)
  - `barry-far/V2ray-Configs/main/Splitted-By-Protocol/vless.txt` (a known-populated protocol file)
  - Default sources.yaml is in repo, **expanded to ~5 sources in Phase 2 alongside ~100-country expansion**.
  - `[auto] Smaller default keeps Phase 1 smoke test fast (~30s end-to-end)`

### Run command structure
- **D-05:** Single command default `vlessfilter run`, plus `--stage` flag for debugging individual stages:
  - `vlessfilter run` — full pipeline
  - `vlessfilter run --stage fetch` — only ingest sources into xray-knife db
  - `vlessfilter run --stage test` — only run handshake+speedtest on what's in db
  - `vlessfilter run --stage select` — only re-read db, regenerate outputs
  - *Why:* default is one keystroke; `--stage` is invaluable when debugging output formatting without re-running 5-min speedtests.
  - `[auto] Selected: C) single command + --stage flag (recommended)`

### Composite score formula
- **D-06:** Score = `0.6 * normalized_speed - 0.4 * normalized_latency` where:
  - `normalized_speed = min(speed_mbps, 100) / 100`  *(cap at 100 Mbps so a single fast outlier doesn't dominate)*
  - `normalized_latency = min(latency_ms, 1000) / 1000`  *(cap at 1s)*
  - Higher score = better. Ties broken by lower latency.
  - *Why:* explicit weights are easy to tune empirically in Phase 2 once we have real measurements. Speed-weighted (60/40) because user said "ping/speed" and most users care about throughput once latency is "good enough".
  - **Revisit Phase 2** based on real distribution: if 95%+ of alive keys cluster in <200ms latency, may want to weight differently.
  - `[auto] Selected: B) weighted normalized (recommended)`

### Top-3 selection ties + low-population countries
- **D-07:** Within a country, sort by score (desc), then by latency (asc) for tie-breaking. If a country has <3 alive keys, output what's available (1 or 2 entries). Countries with 0 alive keys are omitted from output entirely.
  - `[auto] Selected: deterministic tie-break by latency (recommended)`

### README.md output schema
- **D-08:** Repo-root `README.md` table columns (one row per country):
  | Country | Top latency (ms) | Median speed (Mbps) | Keys | Last tested (UTC) |
  Plus a "How to use" section above the table linking to `subs/<CC>.txt`. Countries shown with `🇺🇸 US` style flag-emoji + ISO code. Sorted by country code alphabetically.
  - `[auto] Standard table; deterministic ordering deferred to Phase 2`

### subs/<CC>.txt output schema
- **D-09:** One vless URI per line, plain text, no headers, no comments. UTF-8, LF line endings.
  - This format is paste-into-Hiddify/v2rayN/Streisand compatible.
  - `[auto] Plain URIs is the canonical subscription format`

### Test strategy for Phase 1
- **D-10:** Two test tiers:
  - **Unit tests** (CI-runnable, no network): mock xray-knife subprocess via interface, fixture files for sample subscription responses, assert parsing/grouping/selection logic. Target: ≥70% line coverage on selector + output formatter.
  - **Smoke test** (manual gate, real network): `make smoke` runs `vlessfilter run` against the default 2-source sources.yaml on the developer's machine. Asserts output files exist and contain at least 1 country. Not run in CI.
  - `[auto] Selected: C) both, unit tests in CI + manual smoke (recommended)`

### Project layout
- **D-11:** Standard Go layout:
  ```
  cmd/vlessfilter/main.go     # CLI entry, flag parsing, dispatch
  internal/
    sources/                   # sources.yaml loader + URL expansion
    xrayknife/                 # subprocess wrapper, stdout parser
    selector/                  # group by country, score, top-3
    output/                    # writers for subs/<CC>.txt + README.md
    pipeline/                  # stage orchestration
  testdata/                    # fixture sub responses, golden output files
  sources.yaml                 # default 2-source seed
  Makefile                     # smoke + unit + lint targets
  go.mod
  ```
  - `[auto] Standard Go project layout`

### Claude's Discretion
- Exact log format / log level handling (use `slog` from stdlib, default `info`)
- How to detect xray-knife binary version (parse `xray-knife --version` output)
- Concurrency primitive choice for stage 1 driver (likely just delegated to xray-knife `-t`)
- Exact CLI flag names beyond `--stage` (e.g. `--source`, `--limit`, `--out-dir`)
- Error wrapping / sentinel errors

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project context
- `.planning/PROJECT.md` — full project vision, constraints, key decisions
- `.planning/REQUIREMENTS.md` — v1 requirements (Phase 1 covers AGGR-01, AGGR-03, AGGR-04, TEST-01, TEST-03, TEST-04, SEL-01, SEL-02, SEL-03, OUT-01, OUT-02, DEP-01, DEP-02)
- `.planning/ROADMAP.md` §"Phase 1: MVP End-to-End" — phase goal and 5 success criteria

### Engine
- xray-knife README: https://github.com/lilendian0x00/xray-knife/blob/master/README.md — `subs add/fetch/show`, `http -f --speedtest`, SQLite location at `~/.xray-knife/xray-knife.db`, dual-engine (xray-core + sing-box), MIT license
- xray-knife schema: subprocess returns CSV with at minimum `Link, Delay, DownloadSpeed, UploadSpeed, IpAddress, Location` (verify exact column names in plan-phase by running `xray-knife http --help`)

### Upstream feed
- v2go README: https://github.com/Danialsamadi/v2go/blob/main/README.md — output structure, country split format, GPL-3 license (we link as data source, not code dep)
- v2go country files: https://raw.githubusercontent.com/Danialsamadi/v2go/main/Splitted-By-Country/`<CC>`.txt — base format for `kind: country-template` source entries

### Go ecosystem (likely deps)
- `gopkg.in/yaml.v3` — sources.yaml parsing
- `database/sql` + `github.com/mattn/go-sqlite3` — read xray-knife.db
- `log/slog` (stdlib) — structured logging

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- None — greenfield project, no existing code in repo (only `.git`, `.planning/`, `AGENTS.md`)

### Established Patterns
- None yet — Phase 1 establishes the project layout and conventions per D-11

### Integration Points
- xray-knife: external binary on `$PATH` (auto-installed by D-02)
- xray-knife.db: SQLite file at `~/.xray-knife/xray-knife.db` is our read interface for results
- v2go upstream: HTTPS GETs against githubusercontent.com — no auth needed
- Filesystem: writes to repo root (`README.md`, `subs/`, `testdata/` for fixtures)

</code_context>

<specifics>
## Specific Ideas

- User explicitly wants "top 3 by ping/speed per country" — not top-N, not "all alive", not "best one". Three. (PROJECT.md Key Decision row 4)
- Output files must be paste-into-client compatible without any transformation — that locked D-09 to plain URIs.
- The whole architecture must be ephemeral-VPS-friendly even though Phase 1 doesn't use that path yet — Phase 2 should require zero rewrites of Phase 1 code, only additions. That's why D-02 (auto-install) is in Phase 1 not Phase 2.

</specifics>

<deferred>
## Deferred Ideas

- **Multi-region tester orchestration** — defer to v2 milestone (per PROJECT.md Out of Scope)
- **Telegram channel scraping** — defer permanently (per PROJECT.md Out of Scope, upstream covers it)
- **Other protocols (Hysteria2, Trojan, SS, VMess)** — defer to v2 milestone (per REQUIREMENTS.md v2)
- **Repeated-runs stability scoring** — v2 (REQUIREMENTS.md QUAL-01)
- **Web UI / dashboard** — out of scope permanently
- **Custom GeoIP DB** — never; xray-knife covers it
- **Library-mode xray-knife integration** — possibly Phase 4+ if subprocess overhead becomes measurable

</deferred>

---

*Phase: 01-mvp-end-to-end*
*Context gathered: 2026-05-14*
