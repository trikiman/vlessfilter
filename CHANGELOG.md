# Changelog

All notable changes to VlessFilter are documented here.

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
