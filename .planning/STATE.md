---
gsd_state_version: 1.0
milestone: v2.2.0
milestone_name: ru-verified-10
status: in-progress
last_updated: "2026-05-29T06:05:00Z"
last_activity: 2026-05-29 -- 22.4 reliability tracking shipped (per-key history + dashboard table)
progress:
  total_phases: 5
  completed_phases: 3
  total_plans: 5
  completed_plans: 3
  percent: 75
---

# Project State

## Project Reference

See: .planning/PROJECT.md
**Core value:** Always-fresh, auto-curated, geo-tagged top 3 proxy keys per country (VLESS / VMess / Trojan / SS), accessible as static URLs, with **honest validation** including pre-publish probe.

**v2.2 milestone goal (per user 2026-05-27):** *"10 keys is just small test, when it will work perfect we will do all keys but its not worked perfect"* — ship a small `subs/verified-russia.txt` with **10 keys at 99% alive from a Russian residential IP**. Prove the architecture works at small scale before scaling to all countries.

## Shipped Milestones

- **v1.0** — MVP End-to-End (Phases 1-3) — 2026-05-15
- **v1.4** — Liveness Validation (3x retest, passes>=2) — 2026-05-24
- **v1.5** — Country Identification (consensus + post-publish probe) — 2026-05-24
- **v2.0** — Multi-Protocol Pivot (VLESS+VMess+Trojan+SS) — 2026-05-24
- **v2.0.1** — Pre-publish probe + always-on-checkpoint — 2026-05-25
- **v2.1.0** — Scale-out: matrix parallelism, benchmark, sources.txt, local backup — 2026-05-25

## Currently Active: v2.2 — RU-Verified-10

**Scope (deliberately small):** ~10 keys in `subs/verified-russia.txt` that work for the operator's residential RU IP (Iskratelecom). When the architecture proves stable at 10 × 99% over a week, scale up.

### Phases

- ✅ **22.1** — RU verification investigation
  - Proved 95% pass from Frankfurt vs 0% from residential RU IP
  - Root cause: same key tested differently from datacenter vs residential

- ✅ **22.2** — saleapp source integration
  - 9 RU-curated aggregator sources added to sources.yaml
  - igareck × 7 + kort0881 sni_filtered + v2nodes paid
  - +1129 URI input for richer RU bridge pool

- ✅ **22.3** — verify-russia.yml bridge architecture
  - Picks alive RU bridge from subs/RU.txt or RU_BRIDGE_FALLBACK secret
  - xray-knife proxy inbound → SOCKS5 → proxychains4
  - 2-stage filter: cf-trace gate, then ipinfo+google confirm
  - Outputs subs/verified-russia.txt (auto-merged, never empty-overwritten)
  - Schedule: every 30 min + after refresh.yml + workflow_dispatch

- ⏳ **22.4** — Validate 10 × 99% over 7 days [IN PROGRESS]
  - ✅ Per-key reliability tracking (`.readme-data/verified-russia-history.jsonl`)
  - ✅ Dashboard surfaces "Seen N/M" + "Alive %" per key over last 24h
  - ✅ Round-robin protocol selection so verified-russia.txt has mixed vless/trojan/ss (was 100% ss earlier)
  - ✅ Bridge pick+start unified with fallback: tries each candidate end-to-end (liveness probe + start as SOCKS + curl through SOCKS to verify it actually proxies)
  - ⏳ Need 7-day baseline data accumulated (~336 verify-russia runs)
  - ⏳ Decision pending: residential always-on (self-hosted runner) or accept DC-bridge approximation
  - Current observed: ~25 keys verified per cycle, 30min cron, 2-4min runtime each

- ⏳ **22.5** — Document & ship as stable v2.2
  - Update README.md to point at verified-russia.txt as recommended for RU users
  - SUMMARY.md for the milestone
  - Then explicit go/no-go from user before scaling to "all keys"

## Currently Running

GitHub Actions:
- refresh.yml: every 4h cron
- verify-russia.yml: every 30min cron + after refresh + workflow_dispatch
- Dashboard live at https://trikiman.github.io/vlessfilter/dashboard.html

Both are stable; cron is enabled. No long ad-hoc runs in flight.

## Accumulated Decisions (v2.2)

### Why "RU residential ≠ RU datacenter"
- Frankfurt + GH Actions runners pass 95% of probe tests on the same keys that Russian Iskratelecom-residential sees as 0%
- TSPU/RKN treats residential and datacenter traffic differently
- Bridge architecture (route GH Actions through RU datacenter) approximates residential but is NOT identical

### Why bridge from `subs/RU.txt` instead of dedicated RU VPS
- Free, self-bootstrapping (no external infra)
- Fallback secret `RU_BRIDGE_FALLBACK` covers chicken-and-egg when RU.txt is stale
- Tradeoff: when ALL bridges decay, verify-russia.yml skips and preserves previous output (graceful degrade)

### Why 30 min cycle vs continuous
- Public free keys have ~70% mortality per 12h
- 30 min cycle puts decay window << v2rayN auto-update interval (15-60 min)
- Net effect: user always sees keys verified in last 30 min

### What "99% alive" means in v2.2 scope
- Snapshot reliability: keys-alive-now / keys-published-now ≥ 90% (acceptable)
- Subscription reliability: when v2rayN auto-pulls, ≥ 99% of pulled keys connect ≥ once
- Achieved via: small list (10), high refresh rate (30 min), client auto-rotation

## Open Issues

- **22.4 OPEN**: 7-day stability test not started — need to run with current architecture and measure
- **saleapp manual seeds**: 6 hand-curated trojan seeds tested via v2rayN: 2/6 worked (saleapp-3, saleapp-5 at ~55ms). Other 4 were stale at fetch time.
- **xray-knife password decoding bug**: trojan URIs with `#` in password don't round-trip cleanly. Workaround: use xray-core directly. Filed as known limitation.
- **verify-russia.yml has not yet seen a "fresh refresh" cycle** with the new saleapp sources. First cycle picking up the new 1129-URI pool will land at next refresh.yml (4h cron).

## Pending Todos (v2.2)

- [ ] Wait for next refresh.yml cycle to populate subs/RU.txt with new sources
- [ ] Spot-check verified-russia.txt reliability via PowerShell verifier on user's PC daily for 1 week
- [ ] Add per-key reliability tracking to dashboard (last 7 days alive %)
- [ ] Decide: self-hosted runner (residential exact) vs current DC-bridge (approximation) vs Russian VPS

## Session Continuity

Last session: 2026-05-28T13:25:00Z
Active mode: User-driven iteration on v2.2 small-scope goal
Resume: `.planning/STATE.md` then `git log --oneline -20`
