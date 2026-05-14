# Requirements: VlessFilter

**Defined:** 2026-05-14
**Core Value:** Always-fresh, auto-curated, geo-tagged top 3 VLESS keys per country, accessible as a static URL.

## v1 Requirements

### Aggregation

- [ ] **AGGR-01**: Tool can pull VLESS configs from a list of subscription URLs declared in `sources.yaml`
- [ ] **AGGR-02**: Default `sources.yaml` ships with sensible upstream feed (v2go's per-country files + 2–3 well-known aggregator repos)
- [ ] **AGGR-03**: Tool decodes both plain-text and base64-encoded subscription responses
- [ ] **AGGR-04**: Tool deduplicates configs by host+port identity (delegated to xray-knife's SQLite library)
- [ ] **AGGR-05**: User can add or remove sources by editing `sources.yaml`, no code changes required

### Testing

- [ ] **TEST-01**: Stage 1 handshake filter (TCP/TLS) drops dead keys at high concurrency; observed reduction ≥80%
- [ ] **TEST-02**: Tool applies kernel tuning before Stage 1 (`ulimit -n`, `tcp_tw_reuse=1`, expanded `ip_local_port_range`) so the OS doesn't kill the process from socket exhaustion
- [ ] **TEST-03**: Stage 2 speedtest runs through xray-knife with concurrency capped at ≤20
- [ ] **TEST-04**: Stage 2 records latency, throughput, and exit-IP country per key (all native xray-knife output)
- [ ] **TEST-05**: Each tested key gets a single composite ranking score combining latency and throughput

### Selection & Geo

- [ ] **SEL-01**: Results are grouped by exit-IP country (2-letter ISO code)
- [ ] **SEL-02**: Within each country, keys are ranked by composite score (lower latency + higher speed = better)
- [ ] **SEL-03**: Tool selects top 3 per country, ties broken deterministically by latency
- [ ] **SEL-04**: Countries with fewer than 3 alive keys still produce partial output (1 or 2 keys)

### Output

- [ ] **OUT-01**: Tool writes `subs/<CC>.txt` per country containing top 3 `vless://` URIs, one per line
- [ ] **OUT-02**: Tool writes a `README.md` summary table at repo root: country, top latencies, top speed, last-tested UTC timestamp
- [ ] **OUT-03**: Tool writes `all-results.csv` with every tested key plus measurements (for debugging/analysis)
- [ ] **OUT-04**: Tool writes `raw/dead.txt` listing keys that failed Stage 1 handshake (diagnostics)
- [ ] **OUT-05**: Output is deterministic — identical inputs produce byte-identical outputs (stable sort, rounded numbers)

### Deployment

- [ ] **DEP-01**: Project builds as a single static Go binary via `go install ./cmd/vlessfilter`
- [ ] **DEP-02**: `vlessfilter run` executes the full pipeline end-to-end with no required flags (sensible defaults)
- [ ] **DEP-03**: Auth via `$GH_TOKEN` env var; uses `git -c http.extraheader=...` so secrets never land in `~/.gitconfig` or `ps aux`
- [ ] **DEP-04**: Full pipeline finishes within 60-minute wall-clock on a 4 vCPU / 32 GB Linux VPS, given a reasonable input size
- [ ] **DEP-05**: Resumable — partial results are committed and pushed every ≤2 minutes, so a VPS death loses ≤2 minutes of work
- [ ] **DEP-06**: A GitHub Actions cron workflow runs the same binary on a schedule as a no-VPS fallback

## v2 Requirements

Deferred to a future milestone. Tracked but not in current roadmap.

### Multi-protocol

- **PROTO-01**: Hysteria2 keys tested and published alongside VLESS
- **PROTO-02**: Trojan keys tested and published
- **PROTO-03**: Shadowsocks keys tested and published

### Multi-region orchestration

- **MULTI-01**: Multi-VPS runs publish per-tester-region branches and a merged "best-from-all-regions" output
- **MULTI-02**: Tester-region geo metadata appears in `all-results.csv`

### Discovery

- **DISC-01**: Pull from Telegram channels directly (auth + scraping)
- **DISC-02**: Auto-discover new aggregator repos from GitHub topic search

### Quality

- **QUAL-01**: Repeated runs flag unstable keys (high variance) and demote them
- **QUAL-02**: TLS fingerprint and Reality params validated separately from speed

## Out of Scope

| Feature | Reason |
|---------|--------|
| Telegram channel scraping in v1 | Upstream aggregators (v2go, V2Hub2) already cover that surface |
| Custom proxy engine | xray-knife handles VLESS Reality + speedtest + geo |
| Custom GeoIP database | xray-knife speedtest output already includes exit-IP country |
| Web UI / live dashboard | The product is static files in git; no service to host |
| Real-time monitoring | Cron + git history sufficient for v1 |
| Multi-VPS distributed orchestration | Single VPS per run in v1; multi-region deferred to v2 |
| User accounts / authentication | Public proxy-key publication, no users |
| Persistent state outside the repo | VPS dies at 60 min; the repo is the only persistence |

## Traceability

Filled in after roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| AGGR-01 | Phase 1 | Pending |
| AGGR-02 | Phase 2 | Pending |
| AGGR-03 | Phase 1 | Pending |
| AGGR-04 | Phase 1 | Pending |
| AGGR-05 | Phase 2 | Pending |
| TEST-01 | Phase 1 | Pending |
| TEST-02 | Phase 2 | Pending |
| TEST-03 | Phase 1 | Pending |
| TEST-04 | Phase 1 | Pending |
| TEST-05 | Phase 2 | Pending |
| SEL-01 | Phase 1 | Pending |
| SEL-02 | Phase 1 | Pending |
| SEL-03 | Phase 1 | Pending |
| SEL-04 | Phase 2 | Pending |
| OUT-01 | Phase 1 | Pending |
| OUT-02 | Phase 1 | Pending |
| OUT-03 | Phase 2 | Pending |
| OUT-04 | Phase 2 | Pending |
| OUT-05 | Phase 2 | Pending |
| DEP-01 | Phase 1 | Pending |
| DEP-02 | Phase 1 | Pending |
| DEP-03 | Phase 2 | Pending |
| DEP-04 | Phase 2 | Pending |
| DEP-05 | Phase 2 | Pending |
| DEP-06 | Phase 3 | Pending |

**Coverage:**
- v1 requirements: 25 total
- Mapped to phases: 25
- Unmapped: 0 ✓

---
*Requirements defined: 2026-05-14*
*Last updated: 2026-05-14 after initial definition*
