---
gsd_state_version: 1.0
milestone: v1.2.0
milestone_name: liveness-validation
status: executing
last_updated: "2026-05-24T05:48:00Z"
last_activity: 2026-05-24 -- Milestone v1.2/v1.3 started, autonomous execution
progress:
  total_phases: 4
  completed_phases: 0
  total_plans: 4
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md
**Core value:** Always-fresh, auto-curated, geo-tagged top 3 VLESS keys per country, accessible as a static URL, with **honest validation**.
**Current focus:** Phase 4 — Multi-Attempt Liveness Validation

## Current Position

Phase: 4 of 7 (Multi-Attempt Liveness Validation)
Plan: 0 of 1 in current phase
Status: Executing autonomously
Last activity: 2026-05-24 -- Milestone v1.2/v1.3 kicked off

Progress: [▓▓▓░░░░░░░] 43% overall (3 of 7 phases done from prior milestones)

## Accumulated Context

### Decisions (carried forward + new)
- v1.0: xray-knife as test engine, top-3-per-country composite score
- v1.1: Pre-classified country sources (SoliSpirit + V2Hive) added — 88+ files
- v1.1: Stability filter — separate stable from rotating exits
- v1.2 NEW: 3x retest before marking alive (LIVE-01)
- v1.2 NEW: Sticky-alive — once-passed configs survive single retest fail (LIVE-03)
- v1.3 NEW: ipinfo.io as test URL for actual exit IP, not first-hop (GEO-01)
- v1.3 NEW: 2+ matching country tests required for stable label (GEO-02)
- v1.3 NEW: Post-publish accuracy probe with 80% threshold (GEO-04)

### Open Bugs Being Fixed
- "IN config exits Sweden" — Will be caught by GEO-01 (ipinfo.io probe sees actual exit, not first-hop)
- "9 → 12 → 9 oscillation" — Will be fixed by LIVE-03 (sticky alive)
- "47k configs but only 12 countries" — partially v1.1 (more sources), now LIVE-01 catches handshake-only-not-real-traffic

### Blockers/Concerns
- GitHub Actions billing-locked (irrelevant — local-only deployment confirmed)
- Pool grew to 1M unique vless; ingest takes ~36s, full stage 1 ~50min (acceptable)
- h2.nexus / Cloudflare Workers explored as multi-perspective validators; deferred to v1.4

### Pending Todos
- Tag v1.2.0 after phases 4+5 done
- Tag v1.3.0 after phases 6+7 done
- Run accuracy probe end-to-end before declaring milestone complete

## Session Continuity

Last session: 2026-05-24T05:48:00Z
Active mode: Autonomous execution per user directive
Resume file: .planning/ROADMAP.md
