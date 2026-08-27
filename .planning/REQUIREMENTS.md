# Requirements: v2.3 — subscription-95-in-client

**Goal:** a subscription URL where **≥95% of keys connect in the operator's own
v2rayN test, from their residential Russian line.**

## Why this is a different goal from v2.2

v2.2 asked for "99% alive" and measured it from GitHub Actions runners and an
RU-datacenter bridge. Both are on the wrong side of TSPU. When the operator
finally tested `subs/all.txt`, **18 of 257 keys responded (7%)** — and all 18
were Shadowsocks, the protocol v2.2 had excluded as blocked.

v2.3 moves the measurement instrument to the operator's client. Nothing counts
as "alive" until it connects there.

## The physics we cannot change

Measured from 131,232 history records (1,259 distinct keys):

| percentile | key lifetime |
|---|---|
| p25 | 3.7 h |
| **median** | **21 h** |
| p75 | 93.8 h |

**53% of keys die within 24 h.** So 95%-alive is achievable only for a list that
is **small and recently verified**. 95% of 250 keys is not reachable with free
public sources; 95% of ~25 keys verified in the last 30 minutes is.

This is the central design decision of the milestone: **trade breadth for
freshness.** A large list is what produced the 7% experience.

---

## Requirements

### ALIVE-01 — A freshness-bounded primary subscription
- [ ] A single URL publishes **≤30 keys**, every one verified within the last
      **30 minutes**, ordered best-first.
- [ ] Each key's remark carries its verification age and measured bridge latency
      so a stale list is visible in the client, not silent.
- **Acceptance:** operator imports the URL, runs v2rayN's real-delay test, and
      ≥95% of keys return a delay (not `-1`). Repeated on 3 separate days.

### ALIVE-02 — Close the measurement loop
- [ ] The operator's own test result is recorded so the pipeline learns which
      keys/protocols survive their line, not just the bridge.
- [ ] Divergence between bridge-pass and client-pass is reported per protocol.
- **Why:** every prior tier decision came from anecdote or from CI. This is the
      only requirement that makes 95% verifiable rather than asserted.
- **Note:** must not require the operator's PC to run pipeline compute — the
      NO-PC constraint holds. A pasted result or a small committed file is fine.

### ALIVE-03 — Stop discarding good keys on a broken measurement
- [ ] The speedtest must not be truncated by its own timeout.
- **Evidence:** the test pulls 10 MB with a 5 s client timeout. 10 MB in 5 s is
      **exactly 16 Mbps**, so slower keys are cut off mid-download and the read
      error is discarded (`b, _ := io.ReadAll`) — a partial read reports success.
      **54.2% of published values sit at or below 16 Mbps**, and the 14.0 Mbps
      bin alone is **20.24%** of all rows. `--min-speed 12` then filters on that
      number, and **6 countries with live keys publish nothing** as a result.
- [ ] Keys with no usable speed measurement are ranked in a latency-ordered
      second tier rather than dropped.
- **Acceptance:** the 14.0 Mbps spike disappears from the distribution; the
      6 empty countries publish again.

### ALIVE-04 — Remove the single-point-of-failure bridge
- [ ] The RU bridge pool holds **≥10 distinct endpoints** per run.
- **Evidence:** `verify-russia.yml:134` writes igareck candidates with `>` inside
      a loop, truncating the file each iteration, so the pool lands at 6-11 and
      working bridges are nearly always the same host. **33 runs (8.5% of
      successes) published nothing** because all candidates failed — and a
      `skip=true` run still reports green.
- [ ] A skipped publish is visibly non-green, not a silent success.

### ALIVE-05 — Never abort a publish because the prober itself broke
- [ ] Infrastructure failures (binary missing, engine crash, ctx cancel) are
      counted separately from keys that ran and failed.
- **Evidence:** `prepublish.go:180-184` returns `false` on any `cmd.Run()` error,
      and `DropRate` has no error term, so an unusable prober yields
      `DropRate == 1.0` and trips the 75% abort — republishing stale data. The
      existing test *asserts* this behaviour as correct.

### ALIVE-06 — Protocol order from measurement, on fresh keys
- [ ] Re-measure all four protocols now that trojan has live data again
      (229,968 rows, after 26 days at zero), then set the tier order from
      measured pass rates.
- **Current measurement:** ss 60.3% · vless 50.6% · trojan 0/36 (stale keys).
      Trojan's zero is a harness artifact, not a protocol verdict — re-tiering
      before re-measuring would bake in the crash the way the old order baked in
      an anecdote.

### ALIVE-07 — Honest speed provenance
- [ ] Anywhere a speed is shown, state that it is measured from a datacenter
      runner and is an upper bound, not a prediction for the subscriber.
- **Evidence:** the value is genuine decimal Mbps (traced to
      `examiner.go:576`), but it is Azure→Cloudflare, capped by the timeout in
      ALIVE-03 and 20-way contention. A key labelled 250 Mbps and one labelled
      40 Mbps can deliver identical throughput through TSPU.

---

## Out of scope

- Making free public keys durable. A 21 h median is the ecosystem, not a bug.
- Any requirement that runs compute on the operator's PC (hard constraint).
- Paid providers.
