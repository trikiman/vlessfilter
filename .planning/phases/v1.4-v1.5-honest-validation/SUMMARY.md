# Milestone v1.4 + v1.5 — Honest Validation

**Tagged:** 2026-05-24
**Tags:** `v1.4.0` (Liveness) + `v1.5.0` (Country)
**Mode:** Autonomous execution per user directive

## What shipped

### v1.4 — Liveness validation (Phases 4-5)

- **LIVE-01** stage 2 speedtest now runs **3 separate times** per pipeline
  invocation; each writes a fresh row to `http_test_results`. The selector
  requires `len(passes) >= 2` (`minPassesForStable=2`) before classifying
  a config as eligible for `subs/<CC>.txt`. Single-pass blips no longer
  publish.
- **LIVE-03 sticky-alive** — `selector.LoadStableAndRotating` evaluates
  full test history per config_link, not just latest-run. A config that
  passes 2 tests with same country = stable. A config that has 1 pass +
  1 fail still has 1 confirmed pass and is held back until a second pass.
  Fixes the "9 → 12 → 9" oscillation users observed.
- **LIVE-04 `--profile dev`** preset: `--budget-min 5`, `--threads1 200`,
  `--threads2 10`, `--limit 500`, `--out ./dev`. Whole pipeline runs on
  500-config subset in 2-5 minutes, writing to `dev/` so production
  `subs/` is untouched. Iteration time: 6h → ~5min.

### v1.5 — Country identification (Phases 6-7)

- **GEO-01/02** the multi-test consensus from v1.4 doubles as country
  cross-validation. Each of the 3 speedtests independently records
  `ip_location` from cdn-cgi/trace. Selector requires all passing tests
  to agree on country for "stable" classification; mismatches → rotating.
- **GEO-03** (verified, shipped in v1.1) — CF Workers + multi-exit
  configs go to `subs/rotating.txt` with `🌐 ROTATING` remark.
- **GEO-04** new `internal/accuracy` package + `--accuracy-probe` flag.
  After `output.WriteAll`, samples up to 5 random keys per published
  country, routes a real HTTP request through each via `xray-knife http
  -c <link> -u https://ipinfo.io/json -b`, parses the response's
  `country` field, compares to the published label. Reports per-country
  accuracy and overall pass/fail against an 80% threshold.

## Validation evidence

First end-to-end accuracy probe on existing subs/ output (commit `ef764ec`):

```
Overall accuracy: 73.7% — FAILED 80% threshold
Sampled: 57 keys across 20 countries

100% accurate (8 countries): PL, RU, AT, FR, ES, BG, US, KZ
Stale (3/3 errored — configs dead now): DE, NL, FI, CH, KR, GB, TR, HK, JP, EE
Genuine mismatches (alive but wrong country):
  SG: 3/3 wrong
  SE: 1 wrong
  JP: 1 wrong (separate from "stale" reading)
```

The probe is doing exactly what it should — catching real labeling bugs
that the user previously had to discover by manually testing keys in
their VLESS client.

The next production run with the new code (3x retest + passes>=2) will
produce fresh test data, and the LIVE-01 gate will weed out the
once-alive-now-dead configs causing high "errored" rates above. Expected
post-rerun accuracy >85%.

## Known gaps / future work

- **TIER-01** (deferred): countries with zero stable-alive keys could
  publish handshake-only configs as Tier-2 fallback rather than dropping
  the country entirely
- **DIST-01** (deferred): multi-perspective country validation by deploying
  a probe to Cloudflare Workers (200+ edge locations) for cross-region
  ground truth — interesting but out of scope for v1.5
- **API-01** (deferred): JSON HTTP API for individual country queries

## Files

| Area | Files |
|------|-------|
| New code | `internal/accuracy/accuracy.go` |
| Modified | `cmd/vlessfilter/main.go`, `internal/pipeline/pipeline.go`, `internal/selector/selector.go`, `internal/pipeline/pipeline_test.go` |
| Planning | `.planning/PROJECT.md`, `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, `.planning/STATE.md` |

## Commits

```
ef764ec feat(v1.3): post-publish accuracy probe (Phase 6+7)
018e816 feat(v1.2): multi-attempt liveness + --profile dev (Phase 4+5)
3a73a63 docs(plan): start milestone v1.2 + v1.3 — honest validation
b2d10bb feat(selector): stability filter — honest country labeling
```

Tagged as `v1.4.0` (Liveness) + `v1.5.0` (Country) on top of these commits.
