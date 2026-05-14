# VlessFilter

## What This Is

VlessFilter discovers and publishes the top 3 fastest VLESS proxy keys per country, refreshed automatically. It pulls VLESS configs from public subscription aggregators, runs real-proxy latency and speed tests, groups results by exit-IP country, and commits the top 3 per country to a git repo as ready-to-import subscription files. Built to run on ephemeral 60-minute cloud VPS instances and scheduled GitHub Actions.

## Core Value

Always-fresh, auto-curated, geo-tagged top 3 VLESS keys per country, accessible as a static URL (`https://raw.githubusercontent.com/<user>/<repo>/main/subs/<CC>.txt`) — no client install, no manual testing, just paste-and-import.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Pull VLESS keys from a configurable list of subscription URLs and aggregator repos (default seed: v2go's per-country outputs at `Danialsamadi/v2go/main/Splitted-By-Country/`)
- [ ] Decode + dedupe configs into xray-knife's local SQLite library via `subs add` / `subs fetch`
- [ ] Stage 1: high-concurrency TCP/TLS handshake filter (drops ~80–90% dead keys cheaply)
- [ ] Stage 2: real-proxy speedtest through `xray-knife http --speedtest` (≤20 concurrency to avoid VPS bandwidth saturation)
- [ ] Stage 3: read `xray-knife.db`, group by exit-IP country (built into xray-knife's results), pick top 3 per country by combined latency+speed score
- [ ] Output: `README.md` summary table, per-country `subs/<CC>.txt`, full `all-results.csv`
- [ ] Auto-commit + push results to repo using a PAT injected at runtime (no residual creds on VPS)
- [ ] Run end-to-end within 60-minute budget on an ephemeral VPS, including kernel tuning (`ulimit -n 100000`, `tcp_tw_reuse=1`, expanded `ip_local_port_range`)
- [ ] Resumable: checkpoint partial results to git every ~2 minutes so a VPS death never loses >2 minutes of work
- [ ] Single-binary deployment: `git clone && go install ./cmd/vlessfilter && vlessfilter run`
- [ ] Optional GitHub Actions cron workflow as a second deployment target (same binary, no external VPS)

### Out of Scope

- Telegram channel scraping — upstream aggregators (v2go, V2Hub2, etc.) already cover this
- Custom proxy engine — xray-knife handles VLESS Reality, latency, speed, geo
- Custom GeoIP database — exit-IP country comes free from xray-knife speedtest output
- Web UI / live dashboard — results are static files in git
- Non-VLESS protocols (Hysteria2, Trojan, Shadowsocks, VMess) — VLESS-only in v1
- Real-time monitoring — cron + git history is enough for v1
- Multi-VPS distributed orchestration — single-VPS-per-run in v1; per-region results merged via separate runs if useful later

## Context

- **Engine:** `lilendian0x00/xray-knife` (MIT, Go) — handles fetch, parse, dedupe via SQLite, ping, speedtest, exit-IP geo; dual-core (xray-core + sing-box) so it covers VLESS Reality + Hysteria2/Trojan/etc. when we expand
- **Upstream feed:** `Danialsamadi/v2go` (GPL-3, Go) — pre-aggregated, deduped, geo-tagged config lists refreshed every 6 hours via its own GitHub Action; we consume its `Splitted-By-Country/` outputs as a baseline
- **Deployment target A:** 2z2 Cloud Labs ephemeral GCP VPSes (4 vCPU / 32 GB / 250 GB SSD / all ports open / 60-min auto-delete)
- **Deployment target B:** GitHub Actions runners (free cron, ~2k minutes/month, but variable network)
- **Auth from VPS:** PAT in `$GH_TOKEN`, used via `git -c http.extraheader=...` to keep secrets out of `~/.gitconfig` and `ps aux`
- **Result publishing:** user-owned GitHub repo

## Constraints

- **Tech stack:** Go — matches xray-knife, single static binary, deploys in <10 s on a fresh Ubuntu VPS
- **Wall clock:** ≤60 minutes per VPS run; checkpoint at ≤2-minute intervals
- **Stage 1 concurrency:** up to ~1000 threads — kernel tuning is mandatory (without `ulimit -n` raise + port-range expansion + `tcp_tw_reuse`, the OS kills the process via socket exhaustion)
- **Stage 2 concurrency:** ≤20 threads — higher saturates the VPS uplink and produces meaningless throughput numbers
- **Licensing:** project itself MIT or Apache-2.0; respects upstream (GPL-3 for v2go means we link as a data source, not a code dependency)
- **No state on the VPS:** SQLite + temp files vanish at 60 min; everything that matters is in git
- **Privacy:** results are public proxy keys harvested from already-public subs; no user accounts, no PII

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| xray-knife as test engine | MIT, Go, single binary, covers fetch + ping + speedtest + geo + persistent SQLite; writing our own re-implements months of mature code | — Pending |
| v2go's outputs as primary upstream | Already aggregates 451k raw → 21k unique configs every 6h with country split; saves us building an aggregator | — Pending |
| Go for orchestrator | Matches engine, single static binary, easy `go install` on fresh VPS | — Pending |
| Top 3 per country | Per user's spec; small enough to be a useful curation, large enough for failover within a region | — Pending |
| Ephemeral VPS + GitHub Action dual deploy | VPS gives raw network access from random regions for ground truth; Action gives free always-on cron when no VPS is available | — Pending |
| Skip Telegram scraping in v1 | Upstream aggregators already cover that surface; avoids Telegram bot/auth complexity | — Pending |
| Score = f(latency, speed) | Single ranking column simpler than multi-key sort; exact formula determined empirically in Phase 2 | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-14 after initialization*
