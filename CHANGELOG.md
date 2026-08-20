# Changelog

All notable changes to VlessFilter are documented here.

## Unreleased — v2.2 (in progress)

This entry was missing entirely: the changelog stopped at v2.1.0 while ~3 months
of work shipped. Reconstructed from git history 2026-05-26..2026-08-20.

**RU verification (the v2.2 milestone).**
- `verify-russia.yml` — bridge-verify candidates through an RU-datacenter SOCKS
  proxy every 30 min; publishes `subs/verified-russia.txt`.
- Per-key reliability history in `.readme-data/verified-russia-history.jsonl`.
- Probing parallelised (xargs fan-out) and the caps made workflow inputs. The
  old serial loop plus a stop-at-15 early exit meant the published count was a
  ceiling, not a survival rate; the pool also widened from `subs/all.txt` alone
  to `all-results.csv` as well (115 → 643 unique endpoints on the same data).
- **Protocol tiers measured for the first time** (run 31562489245, full pool,
  565/565 attempted): ss 60.3% pass / 582ms median, vless 50.6% / 769ms,
  trojan 0/36 (stale keys, see below). Shadowsocks — previously excluded as
  "TSPU blocks AEAD" — leads on both. That exclusion had never been measured:
  tier E sat at index ~502 of 565 while the sweep exited at 150 passes, so its
  record was 0/0, not 0/63. That run took `verified-russia.txt` from 157 keys
  (100% vless) to 274 (236 vless + 38 ss) — the first SS keys ever published
  there. The count churns per run as keys die and new ones verify (153 at the
  time of writing); the protocol MIX is the durable change. Failures are now
  recorded too, so a pass rate is derivable at all — every one of the 19,132
  prior records was `alive:true`, making it a survivors log.

**Correctness fixes.**
- Endpoint-aware selection: "top 3 per country" was one host under three UUIDs
  (`subs/vless/DE.txt` shipped `104.18.32.47:2096` three times), so a single
  outage killed every key for that country. Now dedups on host:port and prefers
  distinct /24s and domains, with backfill so diversity never costs supply.
- `xray-knife` pinned to v10.1.1. v9's bundled xray-core SIGSEGVs on trojan
  splithttp configs, which silently froze `subs/trojan/` for 26 days while the
  job still reported success. Pinned at all five real install sites, with a test
  that fails on drift — the Go constant alone was never read by CI.
- `all-results.csv` is now the union of all four protocols (it copied the first
  artifact then `break`ed, discarding three protocols' measurements per run),
  and capped so it cannot breach GitHub's 100 MB blob limit.
- `subs/T1.txt` leak: the merge wiped country files with `[A-Z][A-Z].txt` but
  appended with `??.txt`, so `T1` accumulated forever — 2042 lines, 2 servers.
- README generation refuses to publish a README that omits a protocol with live
  keys (VLESS and Trojan had vanished from the docs while being served).

**Docs/infra.** Go 1.26 (setup-go@v6 defaults `GOTOOLCHAIN=local`); Node-24
action versions; the 404ing binary-install path dropped for `go install`;
`sub.pai.yt` age-gated after 82 days of HTTP 521; 326 tracked-but-gitignored
files un-ignored; 7 PC-only PowerShell scripts removed.

## v2.1.0 — 2026-05-25

**Theme: scale-out via matrix parallelism + benchmarks + local backup.**

The v2.0 lineage proved out the multi-protocol pipeline. v2.1 takes the
scaling lid off: GitHub Actions runs each protocol on its own runner, so
the four protocols progress in wall-clock parallel rather than stealing
each other's budget. A separate benchmark workflow finds the empirical
thread sweet-spot per runner type instead of guessing.

### Benchmark findings (commit 777a173)

Ran 4 sweeps at threads1 = 1000, 2000, 3000, 5000 (untested-batch 40k,
20-min budget, vless protocol, on ubuntu-latest). All 4 runs hit the
20-min budget without exhausting any resource:

| threads1 | peak RAM | peak FDs | peak procs | wall  |
|----------|----------|----------|------------|-------|
| 1000     | 2.85 GB  | 11,800   | 170        | 20:00 |
| 2000     | 3.41 GB  | 13,458   | 168        | 20:00 |
| 3000     | 3.58 GB  | 16,870   | 168        | 20:00 |
| 5000     | 3.70 GB  | 16,813   | 172        | 20:00 |

Conclusions:
- **Diminishing returns past 3000 threads** — going 3k→5k added only
  100 MB RAM and FDs were flat (xray-knife internally caps).
- **RAM is not the constraint** at any tested level (max 23% of 16 GB).
- **Stage 1 isn't the bottleneck.** Wall clock is dominated by Stage 2
  speedtests (bandwidth-bound, --threads2 20 maxes the runner uplink)
  and the pre-publish probe (tests each top-3 selection again).
- **New default: --threads1 3000, --untested-batch 240000.** Up from
  the previous v2.0 default of 1000/80000.


### Added

- **`--untested-batch` flag** on `vlessfilter run` (default 80000 keeps
  parity; raise to 120000-240000 on powerful runners). Wired through to
  `pipeline.runTestProtocol` so per-protocol DB-pool coverage scales
  with available compute. Hardcoded literal previously buried in
  pipeline.go is now operator-tunable.
- **`vlessfilter sources-list` subcommand** dumps every fetchable URL
  after country/range-template expansion. Default `--format=plain`
  emits one URL per line; `--format=name-url` emits TSV with the
  declaring source name. Used to materialize `sources.txt`.
- **`sources.txt`** — committed manifest of all 283 expanded source
  URLs (397 as of 2026-08-20). Diffable in PRs, useful for manual probing,
  lets a fresh machine re-create the source set without parsing YAML.
  Note: nothing reads it — the pipeline loads `sources.yaml`.
- **`scripts/dump-sources-txt.sh`** — regenerator for sources.txt.
- **`scripts/backup-local.ps1`** — point-in-time snapshot of
  xray-knife.db, all-results.csv, subs/, sources.{yaml,txt}, and a
  full `git bundle --all`. Output rotates by timestamp under
  E:\Backups\vlessfilter\. Restores via `git clone vlessfilter.bundle`.
- **`.github/workflows/benchmark.yml`** — workflow_dispatch sweep that
  records peak RAM, FD count, /usr/bin/time -v RSS, and stage 1 wall
  clock at varying `--threads1` (500/1000/1500/2000/2500/3000) and
  `--untested-batch` (20k/40k/80k/160k). Background monitor samples
  every 2s. Results uploaded as artifacts (14d retention). Gives us
  data to set thread defaults instead of armchair-engineering them.

### Changed

- **`refresh.yml` → matrix parallelism**: 4 protocols × ubuntu-latest
  in parallel + final merge-and-push job. Schedule moved from every
  6h to every 2h (still $0/month on public repos). Defaults bumped:
  `--threads1 2000` (was 1000), `--untested-batch 120000` (was 80000),
  `--budget-min 55` (per protocol). With parallelism, total wall
  clock per cycle is now ~60min while exercising 4× the keys.

### Investigated (no change)

- **28-country ceiling root cause**: `selector.minPassesForStable = 2`
  requires two passing tests before publication. With every-6h cycles
  this took multiple days for a key to "graduate". The matrix + 2h
  schedule + 50% larger batch should compress this to under 12h.
  Robustness threshold left intact per user direction; speed comes
  from running more cycles, not from lowering the bar.

### Late additions (post-initial-ship)

- **sub.pai.yt sing-box source** integrated. New `cmd/singboxconv`
  parses sing-box client JSON and emits vless/vmess/trojan/ss URI
  lines. Refresh workflow gained a `prep-sources` job that fetches
  https://sub.pai.yt/singbox, runs the converter, and commits the
  result to `dev/sources-fetched/paiyt.txt`. The 4 protocol matrix
  jobs depend on prep-sources via `needs:`. First run extracted 122
  URIs from 131 outbounds (vless 36 / vmess 5 / trojan 73 / ss 4).
  Hysteria2 and v2ray-plugin SS skipped (xray-knife doesn't support
  them).
- **Pipeline ctx-cancellation fix**: when stage 2 finishes with budget
  on the edge of expiry, the parent ctx becomes dead between
  `runTest()` returning and `runSelect()` starting, causing the DB
  ping to fail with "context deadline exceeded". Added a retry with a
  fresh 5-min ctx, mirroring the existing fallback path. Manifested
  as test (vless) and test (ss) failing in matrix runs while smaller
  pools (vmess/trojan) succeeded — these jobs now publish reliably.
- **Push race fix in merge-and-push**: added `git fetch + rebase`
  before push, with force-overwrite-with-ours fallback if rebase
  conflicts. Handles the case where main moved during a 60min run.

## v2.0.1 — 2026-05-25 (in progress)

**Critical fix: pre-publish probe + always-probe-on-checkpoint.**

Root cause of user-visible "80-90% -1 timeouts in v2rayN" complaint:
the publish output was coming from a CHECKPOINT runSelect, not the
end-of-run runSelect. Checkpoints had `SkipPrePublishProbe=true` to
keep them fast, so the published output was never re-validated. Stage
2 results were 50min stale by publish time — configs had churned.

### Fixed

- **prepublish.probeOne success detection**: now uses xray-knife's
  visual markers (✅ / ❌ / "Real Delay: NNNms" parsing). Previous
  detection trusted exit code, but xray-knife exits 0 even on
  connection failure ("config parsed" = success from its view).
- **Probe runs on EVERY publish** including checkpoints. ~30s overhead
  per checkpoint at 100-key scale; acceptable for 2min interval.
- **Budget-exhausted fallback runSelect** uses a fresh 5min ctx instead
  of the cancelled budget ctx, so the probe can finish even after
  main pipeline budget expires.

### Empirical validation

Manual probe of 8 random keys from currently-published `subs/all.txt`:
4 PASS, 4 FAIL. Detection markers reliable. After deploying the fix,
expected v2rayN `-1` rate drops from 80-90% to ~30-40%.

### Architecture note

`SkipPrePublishProbe` flag retained on `pipeline.Opts` for unit tests
(test fixtures use synthetic vless URIs that can't pass real probes).

## v2.0.0 — 2026-05-24

**Multi-protocol pivot.** VlessFilter now curates **VLESS, VMess, Trojan,
and Shadowsocks** keys per country. Previously VLESS-only.

### Added
- `--protocols vless,vmess,trojan,ss` CLI flag — comma-separated list of
  proxy schemes to test + publish. Default tests all four.
- `subs/<protocol>/<CC>.txt` output structure. Each protocol gets its
  own subscription URL space:
  - `subs/vless/all.txt`, `subs/vless/<CC>.txt`, `subs/vless/rotating.txt`
  - `subs/vmess/all.txt`, `subs/vmess/<CC>.txt`, `subs/vmess/rotating.txt`
  - `subs/trojan/all.txt`, `subs/trojan/<CC>.txt`, `subs/trojan/rotating.txt`
  - `subs/ss/all.txt`, `subs/ss/<CC>.txt`, `subs/ss/rotating.txt`
- README.md now lists per-protocol subscription URLs and per-protocol
  per-country top-3 tables.
- `selector.SupportedProtocols` + `selector.ProtocolFromLink()` helpers.

### Changed
- `pipeline.runTest` now loops over `opts.Protocols`, running stage 1
  (handshake) + stage 2 (3x speedtest) for each protocol independently.
- `pipeline.runSelect` produces one set of subscription files per protocol.
- Selector functions now take a `protocol` parameter:
  - `LoadStableAndRotating(ctx, dbPath, protocol string)`
  - `LoadAliveLinks(ctx, dbPath, protocol string)`
  - `LoadUntestedLinks(ctx, dbPath, limit int, protocol string)`
  - Pass `""` (empty) for legacy any-protocol behavior.
- Existing v1 URLs (`subs/all.txt`, `subs/<CC>.txt`, `subs/rotating.txt`)
  still work — they mirror the VLESS protocol output for back-compat.

### Pool growth
- Single-protocol VLESS pool: ~1,019,913 unique configs
- Estimated four-protocol pool from same sources: 3M+ configs

### Why the pivot
Author noticed that aggregator files publish all four protocols mixed
(e.g., MatinGhanbari super-sub: 15 vless + 73 vmess + 71 ss + 37 trojan),
and our v1 filter discarded ~75% of available proxy keys. Multi-protocol
mode preserves that signal — users pick whichever protocol their client
prefers.

## v1.3.0 — 2026-05-15

**100k target crossed: 113,031 unique VLESS configs ingested.**

After v1.2 doubled the reach to 75k, drilling deeper into the
aggregator dirs found 4 more high-volume sources, pushing the total
past the user's 100k target.

### Added sources
- **kort0881/vpn-aggregator/out/by_type/vless.txt** — 26,576 plain
  vless. Largest single-file vless source in the project (bigger than
  sevcator's protocols/vl.txt).
- **cybersecplayground/V2Hive/by-protocol/all_vless.txt** — 11,994 plain
  vless. Multi-protocol aggregator with structured by-protocol layout.
- **ninjastrikers/nexus-nodes/configs/all.txt** — 1,385 base64 vless.
  Active aggregator updated within hours of probe.
- **YawStar/Proxy-Hunter/configs/proxy_configs.txt** — 419 plain vless.
  Telegram + base64 + SSCONF aggregator.

### Validated against live network
- **159 subscriptions** ingested (~36s wall time)
- **113,031 unique vless configs** in DB after dedup ✅ exceeds 100k
- Stage 1 at 1000 threads on the full 113k pool: pipeline runs to
  completion within budget
- **1,164 alive keys across 32 countries**:
  - Sub-200ms latency to RU (58), FI (91), DE (123), EE (139), CH (154),
    GB (159), PL (171), ES (174), NL (182), LV (183)
  - Mid-range: AT (218), HU (267), IE (311), KZ (341), CZ (369), BG (368)
  - Asia/Oceania: HK (747), JP (623), SG, KR (2820), TW, ID (894), AU (1359)
- Subscription files written for all 32 countries (`subs/<CC>.txt`),
  README.md generated with country flags

### Method that finally got us past 100k
Drilling structured directories instead of guessing single-file paths:
- API list `/contents/<dir>` → see actual filenames
- Each aggregator has its own naming scheme (e.g., `all_vless.txt`,
  `out/by_type/vless.txt`, `configs/all.txt`)
- The 30+ "DEAD" candidates from earlier rounds were all path-guess
  failures, not actually missing data

## v1.2.0 — 2026-05-15

The big one. Doubled vless reach to **75,391 unique configs** by adding
sevcator/5ubscrpt10n (37 files, the largest single-repo VLESS supply
publicly available) and a new `range-template` source kind.

### Added
- **`range-template` source kind** for enumerated `{N}` URLs. Replaces
  what would otherwise be 36 boilerplate sources.yaml entries with one
  entry having `from: 1` and `to: 36`.
- **sevcator/5ubscrpt10n** as 2 sources:
  - `sevcator-mini` (range-template 1..36): 36 mini-sub files, mix of
    plain and base64, ~30k+ unique vless after dedup
  - `sevcator-vl`: 25,706 plain vless lines in protocols/vl.txt — the
    largest single-file vless source on GitHub
- Default `--threads1` raised to 1000 to keep ingest-to-completion time
  under 5 minutes on kernel-tuned hardware.

### Validated against live network (May 2026)
- **151 subscriptions** ingested in 36 seconds wall time
- **228,888 raw configs** ingested (all protocols)
- **75,391 unique vless** in subscription_configs after xray-knife dedup
- All previous v1.1 sources still working (additive only)

### How we got here (the meta-lesson)
v1.1's CHANGELOG claimed "100k unique VLESS does not exist publicly today".
That was wrong — based on probing only ~40 candidate URLs and failing to
recognize that several aggregators (notably sevcator/5ubscrpt10n) split
their archives across many enumerated files. The user pushed back with
two specific URLs from sevcator's mini/ directory, which led to finding
36 such files plus a 25k-line protocols/vl.txt. Lesson: when a number
"feels low", probe deeper directory structures and don't dismiss
aggregators that returned 404 on guessed paths.

## v1.1.0 — 2026-05-15

Massive source expansion. Pipeline now ingests from 8 aggregators instead
of 2, validated against the full set on real network.

### Changed
- **`sources.yaml` now lists 8 confirmed-alive VLESS aggregators** (was 2):
  v2go-country, v2go-vless, ebrasha-vless (~22.5k configs, biggest single
  source), nscl5-allsub, nscl5-vless, epodonios-allsub, ndsphonemy-default,
  ndsphonemy-speed.
- **Default `--threads1` raised from 200 to 500** to match the larger
  ingest volume. Kernel-tuned Linux easily handles this; xray-knife
  observed at ~165 it/s during stage 1 testing.

### Validated against live network (May 2026)
- **45,557 raw configs / 36,927 unique vless** ingested across 114
  subscriptions in 25 seconds.
- Stage 1 (handshake, 500 threads) on 36,927 configs: 3 min 42 sec wall
  time, 259 alive across 15 countries (AT/CH/DE/ES/FI/HK/JP/LV/NL/RS/RU/
  SE/SG/TW/US).
- Top countries by alive count: DE 74, LV 55, RU 48, FI 26, NL 14.
- README + subs/<CC>.txt + all-results.csv + raw/dead.txt all produced
  correctly with the expected determinism guarantees.

### Notes on the "100k keys" target
The public VLESS supply across the major aggregators is ~37k unique. 100k
unique VLESS keys do not exist publicly today — what looks like 100k+ in
naive line counts is heavily duplicated across aggregators. Reaching
volumes >50k would require Telegram channel scraping (deferred to v2)
or paid feed sources.

### Known issues (deferred)
- Stage 2 speedtest with `--limit 0` (default) tests all configs at 20
  threads which can exceed the 60-min budget on full 36k ingest. With
  the current xray-knife API there's no way to limit stage 2 to "alive
  from stage 1 only" — `--limit N` selects from raw `subscription_configs`,
  not previous-run survivors. Workaround: pass `--threads2 0` to skip
  stage 2 entirely and rank by latency alone (still produces a valid
  top-3-per-country output as shown above).
- Setting `--threads2 0` currently still runs stage 2 at default 20
  threads (CLI flag default-zero override bug). Fix planned for v1.2.

## v1.0.0 — 2026-05-15 (initial release)

First end-to-end working release. Validated against live network: ingests
~2000 VLESS configs from public sources, runs handshake + speedtest filter,
groups by exit-IP country, publishes top 3 per country.

### Added
- VLESS subscription ingestion from `sources.yaml` via xray-knife
- Stage 1 handshake/ping filter (high concurrency, default 200 threads)
- Stage 2 speedtest filter (low concurrency, default 20 threads)
- Top-3-per-country selector with composite score (latency + speed + tie-breakers)
- Output writers: `subs/<CC>.txt`, `README.md`, `all-results.csv`, `raw/dead.txt`
- Deterministic byte-identical output across consecutive runs
- 60-minute wall-clock budget with checkpoint loop (commits every N minutes)
- Linux kernel tuning (RLIMIT_NOFILE + sysctl tcp_tw_reuse + ip_local_port_range)
- Git push with `http.extraheader` PAT pattern (no creds in `~/.gitconfig`)
- `--git-push --git-branch` CLI flags for autonomous CI runs
- Quiet-mode log filtering for non-TTY (CI, ephemeral VPS) contexts
- GitHub Actions workflow (`refresh.yml`) — 6-hour cron with manual dispatch
- GitHub Actions release workflow + goreleaser config
- Deployment runbook for 2z2 Cloud Labs ephemeral VPS

### Known limitations
- **Stage 2 speedtest doesn't filter to stage 1 survivors.**
  xray-knife's `--limit N` selects N raw configs from `subscription_configs`,
  not from previous-run survivors in `http_test_results`. As a result, when
  the user passes `--limit`, stage 2 mostly retests dead configs. Workaround:
  omit `--limit` so stage 2 runs on the full set; the selector's per-link
  latest-wins query then prefers the stage-2 result for any config tested in
  both stages and falls back to the stage-1 result for the rest.
- **Default git identity** is `VlessFilter Dev <vlessfilter@localhost>`.
  Override with `git config user.email/user.name` before first push.
- **Module path** is `github.com/trikiman/vlessfilter`. Fork users should
  edit `go.mod` to match their own GitHub username.

### Bugs caught during self-validation (none predicted by docs)
- xray-knife `subs fetch` requires explicit `--all` flag
- xray-knife `subs add` returns UNIQUE-constraint error message that needs
  a regex match (not equality) for idempotency
- xray-knife `subs fetch` exits 1 on partial failure — added partial tolerance
- xray-knife `http` writes only to `valid.txt` unless `--save-db` is passed —
  the entire DB-driven pipeline silently produced empty output until this
  was caught
- xray-knife 9.x table is `http_test_results` (not `test_results`); columns
  are `delay_ms`, `download_mbps`, `ip_location` (not the Phase 1 guesses)
- Selector `MAX(run_id)` filter dropped survivors not present in the latest
  run; switched to latest-per-config_link semantics
- `git.sanitize` had infinite loop when replacement contained search marker;
  switched to regexp

### Stack
- Go 1.22+, single static binary (~13 MB, CGO-disabled)
- xray-knife 9.x (auto-installed via `go install` if not on PATH)
- modernc.org/sqlite (pure-Go SQLite driver, no gcc required)
- 7 internal packages, ~2900 LOC, 5 test files, all green
