---
gsd_state_version: 1.0
milestone: v2.1.0
milestone_name: scale-and-robustness
status: shipped
last_updated: "2026-05-25T14:10:00Z"
last_activity: 2026-05-25 -- Matrix parallelism + benchmark workflow + sources.txt + local backup shipped
progress:
  total_phases: 9
  completed_phases: 9
  total_plans: 9
  completed_plans: 9
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md
**Core value:** Always-fresh, auto-curated, geo-tagged top 3 proxy keys per country (VLESS / VMess / Trojan / SS), accessible as static URLs, with **honest validation** including pre-publish probe.

## Shipped Milestones

- **v1.0** — MVP End-to-End (Phases 1-3) — 2026-05-15
- **v1.4** — Liveness Validation (3x retest, passes>=2) — 2026-05-24
- **v1.5** — Country Identification (consensus + post-publish probe) — 2026-05-24
- **v2.0** — Multi-Protocol Pivot (VLESS+VMess+Trojan+SS) — 2026-05-24
- **v2.0.1** — Pre-publish probe + always-on-checkpoint — 2026-05-25
- **v2.1.0** — Scale-out: matrix parallelism, benchmark, sources.txt, local backup — 2026-05-25

## Currently Running

GitHub Actions:
- benchmark.yml run #26404559597 (threads1=2000, batch=40k, vless, 20min budget)
- refresh.yml run #26404569393 (matrix: vless+vmess+trojan+ss in parallel)

Both triggered 2026-05-25T14:06Z via gh workflow_dispatch. Benchmark ETA ~20min, refresh ETA ~60min.

## Accumulated Decisions

### v1.0
- xray-knife as test engine, top-3-per-country composite score
- Composite: 0.6*norm_speed - 0.4*norm_latency

### v1.1-1.3
- Pre-classified country sources (SoliSpirit + V2Hive) — 88+ files
- Stability filter — separate stable from rotating exits
- Sources expanded to 105 entries → 1M+ unique configs

### v1.4
- 3x retest before marking alive (LIVE-01)
- Sticky-alive: full pass history evaluated, not latest-per-link

### v1.5
- ipinfo.io as ACCURACY PROBE URL (not stage-2 URL — kept cdn-cgi/trace there for speed)
- 2+ matching country tests required for stable label
- Post-publish accuracy probe with 80% threshold

### v2.0
- Multi-protocol: VLESS / VMess / Trojan / SS
- Per-protocol output: subs/{vless,vmess,trojan,ss}/<CC>.txt
- Top-level subs/<CC>.txt = VLESS mirror (v1 URL back-compat)
- ss<->shadowsocks naming asymmetry handled in selector

### v2.0 fixes (today)
- Pre-publish probe: re-test top-3 selections RIGHT BEFORE publish, drop dead
- Probe runs on every publish (checkpoint AND end-of-run) — prevents the
  "checkpoint published unprobed output" race
- Detection uses xray-knife's ✅ / ❌ markers + "Real Delay: NNNms" parsing
  (exit code alone is unreliable — xray-knife exits 0 even on connection
  failure since "config parsed" counts as success from its view)

## Open Issues

- v2.0 production validation pending — first end-to-end run with all v2.0
  fixes ongoing (08:35-09:35)
- VMess: 0 stable countries in last few runs — needs more test cycles
  (each run only tests ~80k vmess of 209k available)
- Probe detection on non-VLESS protocols unverified — xray-knife behavior
  for vmess/trojan/ss may differ

## Pending Todos

- Verify pre-publish probe drop rate matches expected ~50% on fresh data
- Monitor v2rayN result vs published — user expects timeout rate to drop
  from 80-90% (pre-fix) to ~30% (post-fix)
- After confirming probe works in prod: document v2.0 retroactively in
  phases/v2.0-multi-protocol/SUMMARY.md
- Consider GEO-04 auto-rollback (restore prev subs/ if probe fails) —
  currently checkpoint flow handles this implicitly (writes filtered output
  to disk, leaves prev state if probe drops everything)

## Session Continuity

Last session: 2026-05-25T08:35:00Z
Active mode: Autonomous execution per user directive ("go autonomous until project is fully ready")
Resume file: .planning/ROADMAP.md
