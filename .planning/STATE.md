---
gsd_state_version: 1.0
milestone: v2.2.0
milestone_name: ru-verified-10
status: in-progress
last_updated: "2026-08-12T04:30:00Z"
last_activity: 2026-08-12 -- measured per-protocol RU-bridge pass rates for the first time (ss 60.3% > vless 50.6%); tier order above is superseded, see "ANSWERED 2026-08-12"
progress:
  total_phases: 5
  completed_phases: 4
  total_plans: 5
  completed_plans: 4
  percent: 90
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

- ✅ **22.4** — Validate 10 × 99% over 7 days [BREAKTHROUGH 2026-05-31]
  - Per-key reliability tracking (`.readme-data/verified-russia-history.jsonl`)
  - Dashboard surfaces "Seen N/M" + "Alive %" per key over last 24h
  - **Protocol-tier ranking** based on 2026 community evidence:
    - Tier A: VLESS Reality + Vision (~95% RU residential pass)
    - Tier B: VLESS Reality, VLESS+WS+TLS, Trojan+WS+TLS
    - Tier C: everything else (plain VLESS, vmess)
    - SS dropped entirely (TSPU blocks AEAD)
  - **igareck whitelist used as bridge source** (their in-RU probe gives fresh bridges)
  - **Result at the time:** verified-russia.txt = 15 VLESS keys, **100% Reality+Vision**
  - "SS dropped entirely (TSPU blocks AEAD)" on line 63 was never measured — see
    "ANSWERED 2026-08-12". SS was removed from the candidate grep on 2026-05-31,
    so the claim could not be tested; when finally probed it led on pass rate.
  - Current: 274 keys (236 vless + 38 ss) as of run 31562489245

- ⏳ **22.5** — Document & ship as stable v2.2
  - Update README.md, CHANGELOG, dashboard footer, mention current methodology
  - Then explicit go/no-go from user before scaling to "all keys"

## Currently Running

GitHub Actions:
- refresh.yml: every 4h cron
- verify-russia.yml: every 30min cron + after refresh + workflow_dispatch
- Dashboard live at https://trikiman.github.io/vlessfilter/dashboard.html

Both are stable; cron is enabled. No long ad-hoc runs in flight.

## Accumulated Decisions (v2.2)

### HARD CONSTRAINT (user 2026-05-31): NO user PC, ever
- No self-hosted GitHub Actions runner on user's PC
- No local PowerShell verifier (dev/*.ps1 are dead — do not run them)
- No "run this on your PC" requests
- ALL compute must be free cloud: GitHub Actions runners + Oracle VPS (Frankfurt)
- Consequence: we CANNOT test from a true RU residential IP. Best available
  approximation is the RU-datacenter bridge. We accept the residential gap and
  compensate by biasing toward protocols that empirically survive the user's
  residential TSPU (Trojan-TLS > VLESS Reality for their Iskratelecom network).

### Why "RU residential ≠ RU datacenter"
- Frankfurt + GH Actions runners pass 95% of probe tests on the same keys that Russian Iskratelecom-residential sees as 0%
- TSPU/RKN treats residential and datacenter traffic differently
- Bridge architecture (route GH Actions through RU datacenter) approximates residential but is NOT identical
- 2026-05-31 user test: Reality+Vision 1/15 alive residential, Trojan-TLS ~70% alive

### Protocol tier order — SUPERSEDED 2026-08-12 by measurement, see below
Kept for provenance. This order came from a single 2026-05-31 anecdote and the
first real measurement inverts it.
- Tier A: Trojan TLS/TCP — best on user's residential
- Tier B: VLESS Reality — passes datacenter bridge, mixed residential
- Tier C: WS+TLS / xhttp (CDN-fronted)
- Tier D: plain VLESS TCP, vmess
- Tier E: Shadowsocks AEAD — **under measurement, no longer excluded** (see below)
- verify-russia.yml interleaves A+B so output carries both protocols as a hedge

### OPEN QUESTION (2026-07-26): the Shadowsocks exclusion is contradicted
A v2rayN run from the operator's actual residential connection returned 18 alive
keys that were **100% Shadowsocks**, with zero trojan/vless surviving — the exact
inverse of the tier order above, which was set from a 2026-05-31 test.

Both results cannot be right. Rather than pick one, verify-russia.yml now probes
SS at lowest priority (Tier E), so it cannot displace a better-tiered key but its
true pass rate lands in `.readme-data/verified-russia-history.jsonl`. **Re-settle
the tier order from that history, not from either anecdote.**

Caveat worth resolving first: SS has no TLS handshake, so a v2rayN delay probe can
succeed on a bare TCP connect that passes real traffic nowhere. The 18 "alive" SS
keys may be false positives. Confirm by browsing through one, not by reading its ms.

### ANSWERED 2026-08-12 — the operator's anecdote was right, the tier order was wrong

Run 31562489245 probed the **full** pool (565/565 attempted, no early exit) with
tier E hoisted to the front. First per-protocol pass rates this project has ever
had, because failures are now recorded too:

| proto  | pass | fail | attempts | pass rate | median |
|--------|-----:|-----:|---------:|----------:|-------:|
| ss     |   38 |   25 |       63 | **60.3%** |  582ms |
| vless  |  236 |  230 |      466 |     50.6% |  769ms |
| trojan |    0 |   36 |       36 |      0.0% |      — |

**Shadowsocks has the highest pass rate and the lowest latency of the three.**
`subs/verified-russia.txt` went 157 keys (100% vless) -> 274 (236 vless + 38 ss);
those 38 are the first SS keys ever published there.

Why the old order survived so long — three compounding measurement faults, not a
protocol fact:
1. Tier E was appended LAST, at index ~502 of 565, while the sweep exits at
   `target_keys` (150). It was never reached: SS's record was 0/0, not 0/63.
2. The history file only ever recorded passers (19,132/19,132 `alive:true`), so no
   pass rate was derivable from it — an unmeasured protocol read as settled.
3. The run summary printed pool size as "probed", so a sweep that stopped at ~360
   still claimed 565, making tier E's silence look like measured failure.

The TCP-gate caveat above does NOT hold. SS's delay FLOOR is the highest of the
three (143ms vs trojan 66ms, vless 73ms) — the opposite of a cheap gate — all 58
distinct SS hosts sustain real throughput, and 46,881 vless rows report 0 Mbps.

**trojan 0/36 is not a protocol verdict.** Those are the stale 2026-07-17 keys;
`subs/trojan/all.txt` has not moved in 26 days because xray-knife v9.12.1 (bundled
xray-core v1.260327.0) SIGSEGVs on splithttp trojan configs. 1,129 trojan configs
passed handshake before the crash — the servers work, the harness died. Tier A was
ranked "EMPIRICALLY BEST" on dead credentials, probed first every run, while the
protocol that actually works sat where the sweep never reached.

Corroborated single datum: `149.22.95.183:443` (ss) survived the operator's
residential test AND passes the bridge — alive at 146ms in this run, 24 records.

STILL OPEN (needs the operator, not CI): whether TSPU treats SS AEAD differently
on residential vs datacenter paths. Every CI vantage point — GH runner, Frankfurt,
RU-datacenter bridge — is on the wrong side of that boundary. Browse through
`149.22.95.183:443` and load a page; do not read its ms.

Re-tiering is deliberately NOT done yet: trojan must be re-measured on FRESH keys
once the v10.1.1 pin is confirmed, or the new order would bake in a harness bug
the same way the old one baked in an anecdote.

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

- [x] ~~Wait for next refresh.yml cycle to populate subs/RU.txt~~ — resolved
      ~440 cycles ago; this todo predates 2026-05-31
- [ ] Confirm ONE ss key end-to-end from the residential line by browsing, not by
      reading its ms: `149.22.95.183:443` (survived both the operator's v2rayN run
      and the bridge, alive 146ms). This is the only remaining question CI cannot
      answer — every runner/Frankfurt/RU-datacenter vantage point is on the wrong
      side of the TSPU boundary. Replaces the old "PowerShell verifier daily for a
      week" item, which contradicted the HARD CONSTRAINT on line 86 and needed the
      dead dev/*.ps1 scripts.
- [ ] Re-measure trojan on FRESH keys once the v10.1.1 pin is confirmed, then
      re-tier from pass rates. Do not re-tier before that: trojan's 0/36 is a
      harness crash, not a protocol result.
- [ ] Add per-key reliability tracking to dashboard (last 7 days alive %) — now
      actually computable, since failures are recorded from 2026-08-12 on
- [ ] Decide: self-hosted runner (residential exact) vs current DC-bridge (approximation) vs Russian VPS

## Session Continuity

Last session: 2026-05-28T13:25:00Z
Active mode: User-driven iteration on v2.2 small-scope goal
Resume: `.planning/STATE.md` then `git log --oneline -20`
