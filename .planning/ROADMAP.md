# Roadmap: VlessFilter

## Overview

VlessFilter goes from "empty repo + a list of subscription URLs" to "self-updating top-3-per-country VLESS subscription published from an ephemeral VPS or a GitHub Action". Phase 1 proves the pipeline works locally on Linux, Phase 2 makes it survive a 60-minute auto-deleting VPS with no manual intervention, and Phase 3 ships the GitHub Action / docs / release machinery.

## Phases

- [x] **Phase 1: MVP End-to-End** — sources.yaml + xray-knife integration + top-3-per-country output, runs locally on Linux ✅ shipped 2026-05-14
- [x] **Phase 2: Ephemeral-VPS Hardening** — kernel tuning, 60-min budget, checkpoint+commit loop, full output set ✅ shipped 2026-05-15
- [x] **Phase 3: CI + Polish** — GitHub Action cron workflow, docs, release automation ✅ shipped 2026-05-15

## Phase Details

### Phase 1: MVP End-to-End
**Goal**: Run the whole VLESS-filter pipeline on a local Linux box from a small `sources.yaml` to per-country top-3 output files. No git push, no resume, no 60-min budget — just prove the pipeline works end-to-end.
**Depends on**: Nothing (first phase)
**Requirements**: AGGR-01, AGGR-03, AGGR-04, TEST-01, TEST-03, TEST-04, SEL-01, SEL-02, SEL-03, OUT-01, OUT-02, DEP-01, DEP-02
**Success Criteria** (what must be TRUE):
  1. `vlessfilter run` reads `sources.yaml`, fetches subs, decodes them, populates xray-knife's SQLite library
  2. Stage 1 handshake filter visibly drops dead keys (≥80% reduction observed on a real public sub)
  3. Stage 2 speedtest emits per-key latency, throughput, and exit-IP country
  4. After the run, `subs/<CC>.txt` files exist with top 3 `vless://` URIs per country, sorted by composite score
  5. `README.md` summary table is written and shows at least 5 distinct countries
**Plans**: 2 plans

Plans:
- [x] 01-01: Project skeleton + xray-knife wrapper + sources.yaml ingest + smoke-run end-to-end on a small public sub
- [x] 01-02: Read xray-knife.db, group-by-country, top-3 selector, output formatters (`subs/<CC>.txt` + `README.md`)

### Phase 2: Ephemeral-VPS Hardening
**Goal**: Make the pipeline survive a 60-minute auto-deleting VPS unattended. Kernel tuning, checkpoint commits, GitHub auth, deterministic output, and the default `sources.yaml` shipped with the repo.
**Depends on**: Phase 1
**Requirements**: AGGR-02, AGGR-05, TEST-02, TEST-05, SEL-04, OUT-03, OUT-04, OUT-05, DEP-03, DEP-04, DEP-05
**Success Criteria** (what must be TRUE):
  1. Running on a 4 vCPU / 32 GB ephemeral VPS, the pipeline finishes successfully under 60 minutes for a realistic input (≥10k raw keys)
  2. If the VPS is killed at minute 30, the repo contains committed partial results from no later than minute 28 (≤2-minute checkpoint guarantee)
  3. `git push` from the VPS works using only `$GH_TOKEN` and leaves no residual creds in `~/.gitconfig`
  4. After Stage 1, `raw/dead.txt` lists every key that failed handshake; after Stage 2, `all-results.csv` contains every survivor with measurements
  5. Two consecutive runs with the same sources produce byte-identical output files (deterministic)
  6. Default `sources.yaml` ships in the repo and pulls from v2go's per-country files plus 2–3 well-known aggregator repos
**Plans**: 2 plans

Plans:
- [x] 02-01: Kernel tuning + 60-min budget enforcement + checkpoint/resume + GitHub-auth git push loop
- [x] 02-02: Composite score formula + deterministic ordering + diagnostic outputs (CSV, dead.txt) + default `sources.yaml`

### Phase 3: CI + Polish
**Goal**: Ship a GitHub Actions cron workflow as a no-VPS fallback, document the deployment, and automate releases so a `git tag v1.0.0` produces installable binaries.
**Depends on**: Phase 2
**Requirements**: DEP-06
**Success Criteria** (what must be TRUE):
  1. A `.github/workflows/refresh.yml` runs the full pipeline on a cron schedule and commits results back to the repo
  2. Repo `README.md` documents both deployment paths (ephemeral VPS + GitHub Action) with copy-pasteable commands
  3. Tagging a release produces downloadable Linux/macOS binaries via goreleaser (or equivalent)
  4. Repo includes a deployment runbook for the 2z2 Cloud Labs ephemeral VPS path
**Plans**: 2 plans

Plans:
- [x] 03-01: GitHub Actions cron workflow (with PAT secret, kernel-tuning step, full pipeline)
- [x] 03-02: README + deployment runbook + release automation (goreleaser config + tag-triggered workflow)
  *Note: implemented as a single combined plan (03-01) covering refresh.yml + release.yml + .goreleaser.yml + docs/DEPLOYMENT-VPS.md + docs/INSTALL.md*

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. MVP End-to-End | 2/2 | ✅ Complete | 2026-05-14 |
| 2. Ephemeral-VPS Hardening | 2/2 | ✅ Complete | 2026-05-15 |
| 3. CI + Polish | 2/2 | ✅ Complete | 2026-05-15 |
| 4. Multi-attempt Liveness | 0/1 | 🚧 Active | — |
| 5. `--profile dev` Fast Iteration | 0/1 | ⏳ Planned | — |
| 6. Country Cross-Validation | 0/1 | ⏳ Planned | — |
| 7. Auto-Accuracy Probe | 0/1 | ⏳ Planned | — |

## Milestone v1.2 — Liveness Validation

### Phase 4: Multi-Attempt Liveness Validation
**Goal:** Eliminate handshake-passes-but-no-traffic false positives by
requiring 3 successful real-HTTP-traffic tests before marking alive.

**Depends on:** Phase 3
**Requirements:** LIVE-01, LIVE-02, LIVE-03
**Success criteria:**
1. After stage 2 speedtest, each survivor is retested 2 more times via
   the same speedtest path
2. `selector.LoadStableAndRotating` only includes configs with passes >= 2
3. The "9 → 12 → 9" published-count oscillation stops; counts move
   monotonically

### Phase 5: `--profile dev` Fast Iteration
**Goal:** Code change to verified-correct in under 2 min wall-clock,
not 6 hours.

**Depends on:** Phase 4
**Requirements:** LIVE-04
**Success criteria:**
1. `vlessfilter run --profile dev` completes in <120 seconds
2. Output written to a separate dev/ subdirectory (no pollution of
   production subs/ files)
3. Same code paths exercised as production run, just with --limit caps
   (500 configs total instead of 1M+)

## Milestone v1.3 — Country Identification

### Phase 6: Country Cross-Validation via ipinfo.io
**Goal:** Replace xray-knife's first-hop IP geolocation with actual
exit-IP lookup through the proxy.

**Depends on:** Phase 4 (need verified-alive configs)
**Requirements:** GEO-01, GEO-02, GEO-03 (verify)
**Success criteria:**
1. xray-knife configured with `-u https://ipinfo.io/json`
2. xray-knife's `ip_location` field reflects ipinfo.io's `country` response
3. Configs with 2+ matching country tests = stable; mismatched = rotating
4. Sample test: at least 90% of keys in `subs/<CC>.txt` actually exit
   through CC when manually tested

### Phase 7: Auto-Accuracy Probe
**Goal:** Continuously verify the published output matches reality;
refuse to push bad data.

**Depends on:** Phase 6
**Requirements:** GEO-04
**Success criteria:**
1. After select stage produces output, before commit/push, the probe runs
2. 5 random keys per country are routed to ipinfo.io and country compared
3. Run logs accuracy% per country and overall
4. If overall accuracy < 80%: ERROR, do not commit/push, alert in log

## Plans

| Phase | Plan | Status |
|-------|------|--------|
| 4 | 04-01-multi-attempt-liveness | ⏳ Planned |
| 5 | 05-01-dev-profile | ⏳ Planned |
| 6 | 06-01-ipinfo-country | ⏳ Planned |
| 7 | 07-01-accuracy-probe | ⏳ Planned |

## Milestone v1.0 — SHIPPED 2026-05-15

All 25 v1 requirements covered. Artifacts:
- Single Go binary (~13 MB, pure-Go, CGO-free)
- 7 internal packages, 2867 LOC, 5 test files (all green)
- 2 GitHub Actions workflows (6h cron refresh + tag release)
- Goreleaser config (linux/darwin × amd64/arm64)
- Deployment runbook for 2z2 ephemeral VPS

Bugs caught during execution (not predicted by research):
- xray-knife `subs fetch` requires `--all` flag explicitly
- xray-knife `subs add` UNIQUE-constraint message format (broadened idempotency check)
- xray-knife exits 1 on partial sub-fetch failure (added partial-tolerance)
- `git.sanitize` infinite loop when replacement string contained search pattern (switched to regexp)
