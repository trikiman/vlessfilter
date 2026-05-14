# Roadmap: VlessFilter

## Overview

VlessFilter goes from "empty repo + a list of subscription URLs" to "self-updating top-3-per-country VLESS subscription published from an ephemeral VPS or a GitHub Action". Phase 1 proves the pipeline works locally on Linux, Phase 2 makes it survive a 60-minute auto-deleting VPS with no manual intervention, and Phase 3 ships the GitHub Action / docs / release machinery.

## Phases

- [ ] **Phase 1: MVP End-to-End** — sources.yaml + xray-knife integration + top-3-per-country output, runs locally on Linux
- [ ] **Phase 2: Ephemeral-VPS Hardening** — kernel tuning, 60-min budget, checkpoint+commit loop, full output set
- [ ] **Phase 3: CI + Polish** — GitHub Action cron workflow, docs, release automation

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
- [ ] 01-01: Project skeleton + xray-knife wrapper + sources.yaml ingest + smoke-run end-to-end on a small public sub
- [ ] 01-02: Read xray-knife.db, group-by-country, top-3 selector, output formatters (`subs/<CC>.txt` + `README.md`)

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
- [ ] 02-01: Kernel tuning + 60-min budget enforcement + checkpoint/resume + GitHub-auth git push loop
- [ ] 02-02: Composite score formula + deterministic ordering + diagnostic outputs (CSV, dead.txt) + default `sources.yaml`

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
- [ ] 03-01: GitHub Actions cron workflow (with PAT secret, kernel-tuning step, full pipeline)
- [ ] 03-02: README + deployment runbook + release automation (goreleaser config + tag-triggered workflow)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. MVP End-to-End | 0/2 | Not started | - |
| 2. Ephemeral-VPS Hardening | 0/2 | Not started | - |
| 3. CI + Polish | 0/2 | Not started | - |
