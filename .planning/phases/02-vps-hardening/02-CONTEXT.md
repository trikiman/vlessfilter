# Phase 2: Ephemeral-VPS Hardening - Context

**Gathered:** 2026-05-14
**Mode:** `--auto` (decisions auto-picked from PROJECT.md + Phase 1 SUMMARY findings)

<domain>
## Phase Boundary

Make the Phase 1 pipeline survive an unattended 60-minute run on an ephemeral GCP VPS:
- Apply Linux kernel tuning before stage 1 so 1000-thread ping doesn't kill the OS
- Enforce 60-min wall-clock budget; abort cleanly before VPS auto-delete
- Checkpoint partial results to git every ≤2 min so a sudden VPS death never loses >2 min of work
- Auth via `$GH_TOKEN` env var; no residual creds on VPS
- Add diagnostic outputs (`all-results.csv`, `raw/dead.txt`)
- Make output deterministic (byte-identical for identical inputs)
- Replace dead `mahdibland-vless` URL with verified-working aggregator
- Fix xray-knife progress-bar pollution in non-TTY runs

</domain>

<decisions>
- **D-12 Kernel tuning:** new `internal/kerntune` package; tries `ulimit -n 100000`, `sysctl net.ipv4.tcp_tw_reuse=1`, `sysctl net.ipv4.ip_local_port_range="1024 65535"`. Logs WARN on failure, never aborts. No-op on non-Linux. Invoked at top of stage 1 only.
- **D-13 Budget enforcement:** new `--budget-min` flag (default `55`, leaves a 5-min safety margin under the 60-min VPS auto-delete). Pipeline wraps `Run` in `context.WithDeadline`. On expiry, gracefully cancels in-flight xray-knife child processes and proceeds straight to stage 3 (select+output) so partial results still ship.
- **D-14 Checkpoint pattern:** in full-pipeline mode, the pipeline writes outputs (subs/, README.md, all-results.csv) every `--checkpoint-min` (default 2). If `--git-push` flag is on AND `$GH_TOKEN` is set, it commits + pushes after each checkpoint. Implementation: simple goroutine ticker that calls `runSelect` periodically.
- **D-15 git auth:** new `internal/git` package wraps `git -c http.extraheader="Authorization: Basic $(b64 oauth2:$GH_TOKEN)" ...`. Provides `Configure(repoDir, token)`, `CommitAll(msg)`, `Push(branch)`. No persistent config writes — uses `-c` per-call so secrets never land in `.git/config` or `~/.gitconfig`.
- **D-16 Diagnostic outputs:**
  - `all-results.csv` — every tested key with columns: link, latency_ms, speed_mbps, country, score (sorted by country asc, score desc — deterministic)
  - `raw/dead.txt` — links that failed handshake (latency=0 or >10000), one per line
- **D-17 Deterministic output:** all generators sort their inputs deterministically before writing. Floats rounded to 1 decimal in human-facing files. Timestamps only in `<!-- last-tested: ... -->` HTML comments at the bottom of README.md (not in the table) so a re-run with identical results produces byte-identical files except for that comment.
- **D-18 Dead-URL fix:** Phase 2 default `sources.yaml` uses ONLY `v2go-country` (verified working) plus a `peasoup-aggregator` (Epodonios/v2ray-configs) entry. The 100-country expansion is enabled in Phase 2 default config (was 2 countries in Phase 1).
- **D-19 xray-knife progress bar:** The xray-knife `http` command has no `--quiet` flag in v9.12. Workaround: detect `os.Getenv("CI") == "true"` or non-TTY stdout and pipe xray-knife's stdout through a line-filter that drops the carriage-return progress lines.

### Claude's Discretion
- Exact retry policy on transient git push failures
- Whether to crash or continue on kernel-tune permission errors (decided: log+continue)
- Specific sysctl values for additional tuning beyond the 3 from D-12

</decisions>

<canonical_refs>
- `.planning/PROJECT.md` — engine + upstream + constraints
- `.planning/REQUIREMENTS.md` — Phase 2 covers AGGR-02, AGGR-05, TEST-02, TEST-05, SEL-04, OUT-03, OUT-04, OUT-05, DEP-03, DEP-04, DEP-05
- `.planning/phases/01-mvp-end-to-end/01-01-SUMMARY.md` — what Phase 1 shipped + open follow-ups
- `.planning/phases/01-mvp-end-to-end/01-02-SUMMARY.md` — selector schema notes for Phase 2
- xray-knife README §"Testing Configs" — flags surface for our wrapper
- Linux man-pages: sysctl(8), ulimit(1) — kernel-tuning targets

</canonical_refs>

<code_context>
- Phase 1 packages (sources, xrayknife, selector, output, pipeline) are all in place and tested
- `xrayknife.RealRunner.HTTPTest` already accepts `HTTPOpts.Threads`; we just need to raise the ceiling
- `pipeline.Opts` already has `Threads1`/`Threads2`/`Limit` so adding a few more fields is mechanical
- `output.Write` produces deterministic file content already (sort.SliceStable everywhere, Go map iter only via slices); the timestamp in the README is the one moving piece
</code_context>

<deferred>
- Multi-VPS distributed orchestration → v2 (multi-region branches, merging)
- Other protocols → v2
- Web UI → out of scope permanently

</deferred>

---
*Phase: 02-ephemeral-vps-hardening (slug TBD; using `02-vps-hardening`)*
*Context gathered: 2026-05-14*
