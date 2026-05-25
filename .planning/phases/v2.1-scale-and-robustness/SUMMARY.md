# Phase v2.1 — Scale and Robustness

**Status:** shipped (code + first benchmark run triggered).
**Tag:** v2.1.0 (pending validation of first matrix run).
**Commits:** 82bbff5 (code), c1574db (data refresh).

## Goals

1. Empirically determine the optimal `--threads1` value on the deployment
   target (GitHub Actions ubuntu-latest) instead of armchair-engineering.
2. Increase per-cycle DB-pool coverage from 80,000 keys to 120,000+.
3. Run all four protocols (vless/vmess/trojan/ss) in wall-clock parallel
   instead of letting vless eat the whole budget.
4. Tighten the cycle interval from every 6h to every 2h.
5. Goal: stably published 100 countries × 3 keys = 300 stable keys.
6. Add local-backup workflow against the case GitHub deletes the repo.
7. Add the canonical sub.pai.yt source.

## What shipped

### Code (commit 82bbff5)

| File | Change |
|---|---|
| `cmd/vlessfilter/main.go` | new `sources-list` subcommand; `--untested-batch` flag |
| `internal/pipeline/pipeline.go` | new `Opts.UntestedBatch` field, threaded through `runTestProtocol` |
| `.github/workflows/refresh.yml` | matrix parallelism (4 protocols × ubuntu-latest), every-2h schedule, --threads1 2000, --untested-batch 120000, separate merge-and-push job |
| `.github/workflows/benchmark.yml` | new workflow_dispatch sweep for thread sweet-spot data collection |
| `scripts/dump-sources-txt.sh` | regenerator for sources.txt |
| `scripts/backup-local.ps1` | timestamped snapshot of xray-knife.db, all-results.csv, subs/, sources.{yaml,txt}, git bundle |
| `sources.txt` | committed manifest of all 283 expanded URLs |
| `.gitignore` | added dev/, .bench/, .kiro/, scripts/logs/ |
| `CHANGELOG.md` | v2.1.0 entry prepended |

### Investigation findings (no code change)

- **28-country root cause:** `selector.LoadStableAndRotating`'s
  `minPassesForStable = 2` (line 379). Each config needs two passing
  tests to graduate from rotating to stable. With 6h cycles this
  meant ~12-18h before a freshly-tested key could be published. Left
  the threshold intact (robustness > velocity per user direction);
  speed comes from running more cycles, not lowering the bar.
- **sub.pai.yt/singbox:** 200 OK, ~46KB sing-box JSON with ~140+
  outbounds (HK/JP/SG/US dominant). Cannot be added as `kind: plain`
  because xray-knife reads URI subscriptions, not sing-box JSON.
  Deferred — needs a JSON→URI converter (~150 LOC of Go).
- **Account billing:** confirmed unblocked. Gross metered usage of
  $0.50/month is fully offset by free-tier discount; net = $0.
  Workflow uses ubuntu-latest standard runner — perpetually free
  on public repos.

## Math: why 2h+matrix is the right move

```
Old (v2.0):
  every 6h × 1 runner × 60min × 80k keys → 80k tests/cycle, 4 cycles/day
  → 320k tests/day (1 protocol; others starved)

New (v2.1):
  every 2h × 4 runners × 55min × 120k keys → 480k tests/cycle, 12 cycles/day
  → 5,760k tests/day
  → ~18× higher per-protocol test rate, all 4 protocols served fairly
```

DB pool turnover at this rate: ~1 day for vless's 1.14M pool. Stable-status
accrual should compress from days to hours.

## Open

- [ ] **Benchmark results** (run #26404559597 in progress as of commit
  time). Sweep at threads1=1000/2000/3000 will set our actual default.
- [ ] **First matrix-refresh run** (#26404569393 in progress). Verifies
  that the workflow is wired correctly across the 4 protocols + merge.
- [ ] **Country growth tracking.** Need to watch the next ~7 days of
  results to see if 28→100 trajectory materializes. If country count
  plateaus < 80, source-set diversification is the next lever.
- [ ] **sub.pai.yt source converter.** Deferred. Implementation note:
  parse `outbounds[]` array, emit one URI per outbound based on `type`
  field (vless/vmess/trojan/shadowsocks/hysteria2 → standard URI form).
  Estimate ~150 LOC + tests.
- [ ] **DB pool backup.** xray-knife.db lives on the user's
  pipeline-host machine, not on the local PC where backup-local.ps1
  runs. To get a true offsite backup of the DB, we'd need to add a
  step to refresh.yml that uploads xray-knife.db as an artifact each
  run. Non-trivial: DB is 200-500MB, artifacts have a 7-day retention
  + 5GB/month outbound cap on free plan. Consider `git-annex` or
  S3-equivalent (BackBlaze B2, Wasabi) for proper cold storage.

## How to verify the milestone is healthy

```bash
# Benchmark results (after job completes, ~25min):
gh run download 26404559597 -D .bench-out/
cat .bench-out/bench-t2000-b40000/summary.txt

# Matrix refresh result (after ~60min):
gh run watch 26404569393

# Country count after first successful run:
curl -s https://raw.githubusercontent.com/trikiman/vlessfilter/main/README.md \
  | grep -oP '\| 🇦?\K[A-Z]{2}' | sort -u | wc -l
```

## Dependencies / contract for next phase

- v2.2 should consume the benchmark results to *replace* the
  workflow defaults (`--threads1 2000`) with whatever empirically
  proves to use 70-90% of available RAM without OOMing — currently
  it's an educated guess.
- v2.2 should consume the country-count growth curve to decide
  whether the 100-country goal needs source expansion.
- The sing-box JSON converter is a v2.2 candidate; if it lands, expect
  HK/JP/SG/US country counts to jump.
