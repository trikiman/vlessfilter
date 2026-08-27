---
gsd_state_version: 1.0
milestone: v2.3.0
milestone_name: subscription-95-in-client
status: in-progress
last_updated: "2026-08-20T12:10:00Z"
last_activity: 2026-08-20 -- v2.2 archived; v2.3 opened. Goal moves the measurement instrument to the operator's own client.
progress:
  total_phases: 5
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md · Requirements: .planning/REQUIREMENTS.md · Phases 23-27: .planning/ROADMAP.md

**v2.3 goal:** a subscription URL where **≥95% of keys connect in the operator's
own v2rayN test, from their residential Russian line.**

## Shipped Milestones

See `.planning/MILESTONES.md`. Most recent: v2.2.0 ru-verified-10 (2026-08-20),
archived to `.planning/milestones/`.

## Why v2.3 exists

v2.2 asked for "99% alive" and measured it from GitHub Actions runners and an
RU-datacenter bridge — both on the wrong side of TSPU. Phase 22.1 *proved* that
gap (95% pass from Frankfurt vs 0% from Iskratelecom residential) and the
acceptance criteria then measured the wrong side of it anyway.

Observed when the operator finally tested `subs/all.txt`: **18 of 257 keys
responded (7%)**, and all 18 were Shadowsocks — the protocol v2.2 had excluded
as TSPU-blocked without ever measuring it.

## Carried-forward decisions (still binding)

### HARD CONSTRAINT: NO user PC, ever
- No self-hosted runner, no local PowerShell verifier, no "run this on your PC".
- ALL compute is free cloud: GitHub Actions + Oracle VPS (Frankfurt).
- Consequence: we CANNOT test from a true RU residential IP. The RU-datacenter
  bridge is the closest available approximation and is NOT identical.
- ALIVE-02 works within this: the operator *reports* a result, they do not host
  compute. The 7 dead `dev/*.ps1` scripts and the two Scheduled-Task scripts
  were deleted 2026-08-20 for violating this.

### Protocol pass rates — MEASURED, not asserted (run 31562489245)
Full pool, 565/565 attempted, tier E hoisted to the front:

| proto  | pass | fail | attempts | pass rate | median |
|--------|-----:|-----:|---------:|----------:|-------:|
| ss     |   38 |   25 |       63 | **60.3%** |  582ms |
| vless  |  236 |  230 |      466 |     50.6% |  769ms |
| trojan |    0 |   36 |       36 |      0.0% |      — |

Shadowsocks leads on both rate and latency. The old "SS dropped entirely (TSPU
blocks AEAD)" tier order is **superseded** — it was never measured: tier E sat
at index ~502 of 565 while the sweep exited at 150 passes, so SS's record was
0/0, not 0/63.

**trojan 0/36 is a harness artifact, not a protocol verdict.** Those were the
stale 2026-07-17 keys; xray-knife v9.12.1 SIGSEGVs on trojan splithttp configs.
Trojan now has 229,968 fresh rows. ALIVE-06 re-measures before re-tiering — the
old order baked in an anecdote and the new one must not bake in a crash.

The TCP-gate hypothesis for SS does NOT hold: SS's delay floor is the *highest*
of the three (143ms vs trojan 66ms, vless 73ms), all 58 distinct SS hosts sustain
real throughput, and 46,881 vless rows report 0 Mbps.

### Key lifetime — the physics
131,232 history records, 1,259 distinct keys: p25 3.7h, **median 21h**, p75 93.8h.
**53% die within 24h.** So 95%-alive is only reachable for a small, recently
verified list. This is why ALIVE-01 caps at ≤30 keys / ≤30 min old.

### Why bridge from `subs/RU.txt`
Free and self-bootstrapping. `RU_BRIDGE_FALLBACK` covers the chicken-and-egg.
Tradeoff: when all bridges decay the job skips and preserves previous output —
which is also how 33 runs published nothing while reporting green (ALIVE-04).

## Currently Running

- refresh.yml: every 4h cron. **Verified green 2026-08-20** (run 32361058687,
  all 6 jobs, all-results.csv 347,470 rows / 66 MB under the 100 MB limit).
- verify-russia.yml: cron moved `*/30` → `7,37` on 2026-08-20. 708 of 985 slots
  (72%) previously never fired — GitHub never created them, `*/30` being the most
  contended cron expression. Effect measurable within a day by counting fired
  slots.
- benchmark.yml: dispatch-only, dormant since 2026-05-25, output consumed by
  nothing.
- Dashboard: https://trikiman.github.io/vlessfilter/dashboard.html

## Known-open, carried into v2.3 phases

1. **Speedtest truncation** (ALIVE-03) — 10 MB body / 5 s timeout ⇒ exactly
   16 Mbps ceiling; 54.2% of published values are timeout artifacts, the 14.0 Mbps
   bin alone is 20.24%. `--min-speed 12` filters on that, so 6 countries with live
   keys publish nothing. Fix needs `HTTPOpts` to expose a timeout field.
2. **Bridge pool depth** (ALIVE-04) — `verify-russia.yml:134` truncates with `>`
   inside a loop; pool lands at 6-11 and is effectively 1-2 endpoints.
3. **prepublish false aborts** (ALIVE-05) — infra failures counted as dead keys;
   an unusable prober yields `DropRate == 1.0` and trips the 75% abort into
   republishing stale data. Its own test asserts this as correct.
4. **accuracy probe** — fully serial, up to 400 spawns × 15 s, runs on *every*
   checkpoint; also probes only vless from top-level `subs/`, so 3 of 4 matrix
   jobs measure the previous run's other-protocol data. Not yet in a phase.
5. **`sources.yaml` working-tree diff** may contain a credentialed URL. NOT
   reviewed, NOT staged. Rotate the token if real.

## Pending Todos (v2.3)

- [ ] Confirm ONE ss key end-to-end from the residential line by browsing, not by
      reading its ms: `149.22.95.183:443` (survived both the operator's v2rayN run
      and the bridge, alive 146ms). Only CI cannot answer this.
- [ ] Verify the cron change recovered fired slots (count over 24h).
