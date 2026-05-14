# Changelog

All notable changes to VlessFilter are documented here.

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
