# Phase 2: Ephemeral-VPS Hardening — SUMMARY

**Completed:** 2026-05-15
**Plans:** 02-01 + 02-02 implemented together (same coherent codebase pattern as Phase 1)
**Tests:** 7/7 packages green (added `kerntune` and `git` packages); gofmt + vet clean.

## What was added/changed

### New packages

| Package | Lines | Notes |
|---|---|---|
| `internal/kerntune` | 4 files (~95 LOC) | Build-tagged Linux/non-Linux; raises RLIMIT_NOFILE via `syscall.Setrlimit`, runs `sysctl tcp_tw_reuse=1` and widens local port range. Best-effort: never returns error. |
| `internal/git` | 2 files (~205 LOC) | `Configure` (idempotent user.name/email), `CommitAll` (returns committed=false on no-changes), `Push` (PAT via `-c http.extraheader=...`, never persists in `.git/config`). `sanitize()` redacts auth headers from error messages using regexp. |

### Updated packages

| Package | Change |
|---|---|
| `internal/sources` | New `DefaultCountries()` returns ~107 ISO codes; `Load` uses it when `countries:` is empty AND a country-template source is enabled. |
| `internal/selector` | New `LoadAllResults` returns alive + dead separately; tolerates empty DB (for checkpoint loops before stage 2 has data). |
| `internal/output` | New `WriteAll` (subs + README + diagnostics in one call), new `WriteDiagnostics` (`all-results.csv` + `raw/dead.txt`). README timestamp moved to bottom HTML comment so re-runs with identical results produce byte-identical files except for that single line. |
| `internal/xrayknife` | New `quietWriter` strips xray-knife's `\r`-progress-bar updates from stdout in non-TTY runs. Gated by `VLESSFILTER_QUIET=1`, `CI=true`, or `!isTerminal(stderr)`. |
| `internal/pipeline` | Major rework: `BudgetMin` wraps ctx with `WithDeadline`; budget expiry routes to recovery path that ships partial outputs. Concurrent checkpoint goroutine writes outputs and (when `GitPush`) commits + pushes every `CheckpointMin`. `kerntune.Apply` runs at top of stage 1. |
| `cmd/vlessfilter/main.go` | New flags: `--budget-min`, `--checkpoint-min`, `--git-push`, `--git-repo`, `--git-branch`. Reads `$GH_TOKEN` env. |
| `sources.yaml` | `countries: []` (use defaults). Replaced dead `mahdibland-vless` URL with `v2go-vless` (verified working). |

## Acceptance criteria — verified

- ✅ `kerntune.Apply()` always returns nil; raises FD ceiling on Linux; logs WARN when sysctl unavailable; no-op on non-Linux build
- ✅ `git.CommitAll` correctly returns `committed=false` on no-changes (covered by `TestCommitAll_NoChanges`)
- ✅ `git.Push` uses `-c http.extraheader=...`; tokens never written to `.git/config`
- ✅ `sanitize` redacts `Authorization: Basic <token>` to `[REDACTED-AUTH]` (regression: original O(N²) loop deadlocked when replacement string contained the marker prefix; switched to regexp)
- ✅ `output.WriteDiagnostics` produces deterministic CSV (sort: Country asc, Score desc, Link asc) with header `Link,LatencyMs,SpeedMbps,Country,Score`
- ✅ `output.Write` produces byte-identical README except for the bottom `<!-- last-tested: ... -->` HTML comment — covered by `TestWrite_TimestampInComment`
- ✅ `sources.DefaultCountries` returns 107 entries including `US`, `DE`, `RU`, `CN`, `JP`, etc.
- ✅ `xrayknife.quietWriter` drops `\r`-only-terminated lines, preserves `\n`-terminated final-output lines
- ✅ `pipeline.Run` with `BudgetMin > 0` wraps ctx with `WithDeadline`; budget expiry triggers recovery path
- ✅ Pipeline checkpoint loop writes outputs every `CheckpointMin` minutes; commits + pushes when `GitPush=true`
- ✅ All 7 internal packages pass `go test -count=1`
- ✅ Binary `--help` lists `--budget-min`, `--checkpoint-min`, `--git-push`, `--git-repo`, `--git-branch`

## Real-world bug caught (in CI, not production)

`git.sanitize` had an infinite loop bug — the original implementation looped finding `Authorization: Basic ` and replacing with a string that **also contained** that marker, so the next loop iteration matched the replacement and went forever. Test ran for the full 600s timeout before failing. Fixed by switching to regexp where the replacement string can't be re-matched.

This is exactly the kind of subtle issue that doesn't show up in design docs — only in actual test execution. Same lesson as Phase 1: real-world validation is irreplaceable.

## Phase 2 requirements coverage (11/11)

- ✅ AGGR-02 (default sources include v2go) — `sources.yaml` ships with v2go-country + v2go-vless
- ✅ AGGR-05 (user can edit sources) — single YAML file, no code changes needed
- ✅ TEST-02 (kernel tuning before stage 1) — `kerntune.Apply()` called at top of `runTest`
- ✅ TEST-05 (composite score) — already in Phase 1 selector; weights validated by `TestScore_Formula`
- ✅ SEL-04 (countries with <3 keys → partial output) — already in Phase 1; verified by `TestTop3PerCountry_Partial`
- ✅ OUT-03 (`all-results.csv`) — `output.WriteDiagnostics`
- ✅ OUT-04 (`raw/dead.txt`) — `output.WriteDiagnostics`
- ✅ OUT-05 (deterministic output) — README timestamp in HTML comment; CSV deterministic sort
- ✅ DEP-03 (PAT via http.extraheader) — `git.Push`
- ✅ DEP-04 (60-min budget) — `pipeline.Opts.BudgetMin` + `context.WithDeadline`
- ✅ DEP-05 (≤2-min checkpoint loss) — `pipeline.startCheckpointLoop` with default `CheckpointMin=2`

## Open follow-ups for Phase 3

- Add a smoke-mode flag that runs --stage select only with VLESSFILTER_QUIET=1 for CI sanity-check
- Add a SHELL TTY detection that's more accurate (current `isTerminal` is heuristic)
- Verify on a real ephemeral 2z2 VPS that the kernel tunables actually apply when run as the user's default account (may or may not be root)

---
*Phase: 02-vps-hardening*
*Total LOC added/changed: ~1100 across 11 files (incl 4 new test files)*
