# Plan 01-01: Skeleton + Sources + xray-knife Wrapper — SUMMARY

**Completed:** 2026-05-14
**Plans:** 01-01 + 01-02 implemented together (small phase, single coherent codebase)
**Engine validated:** xray-knife v9.12.1 installed via auto-install path

## What was built

| Artifact | Lines | Notes |
|---|---|---|
| `cmd/vlessfilter/main.go` | 177 | CLI entry, flag parsing, exit codes |
| `internal/sources/sources.go` | 170 | sources.yaml loader, country-template expansion, base64/plain auto-decode, HTTP fetcher |
| `internal/sources/sources_test.go` | 176 | 6 tests covering all branches |
| `internal/xrayknife/xrayknife.go` | 226 | Runner interface, RealRunner subprocess wrapper, EnsureInstalled with `go install` fallback |
| `internal/xrayknife/xrayknife_test.go` | 138 | 5 tests + FakeRunner test double |
| `sources.yaml` | 21 | Default 2-source config (v2go US+DE templates + mahdibland-vless plain) |
| `Makefile` | 36 | build, test, lint, fmt, smoke, clean, tidy |
| `.gitignore` | 27 | bin/, *.db, subs/, etc. |
| `go.mod` | — | Go 1.26.3, deps: gopkg.in/yaml.v3, modernc.org/sqlite |

## Acceptance criteria (Plan 01-01) — verified

- ✅ `go.mod` has `module github.com/trikiman/vlessfilter` + 2+ deps
- ✅ `cmd/vlessfilter/main.go` has `package main`, `func main()`, `run` subcommand, `--stage` flag, `log/slog` import
- ✅ `Makefile` has `build`, `test`, `lint`, `smoke` targets
- ✅ `.gitignore` has `bin/` and `*.db` lines
- ✅ Internal package directories exist
- ✅ `go build ./cmd/vlessfilter` exits 0
- ✅ `./vlessfilter --help` prints usage with `run` and `--stage`
- ✅ `internal/sources/sources.go` defines `Config`, `Source`, `Subscription`, `ExpandedSource` types and `Load`, `Expand`, `Fetch` functions
- ✅ Country-template URL with `{CC}` placeholder expansion works
- ✅ Base64 auto-detect works (validated by `TestDecodeBody_Base64`)
- ✅ All sources package unit tests pass
- ✅ `internal/xrayknife/xrayknife.go` has Runner interface + RealRunner + auto-install logic
- ✅ Idempotent SubsAdd recognizes `UNIQUE constraint failed`, `already exists`, and `duplicate` patterns
- ✅ All xrayknife package unit tests pass

## Real-world validation

The `--stage fetch` execution against the actual public network revealed and fixed three behaviors not covered by the original RESEARCH.md:

1. **`xray-knife subs fetch` requires `--all` flag** (or `--id`/`--url`/`--file`). Was running without flag → exit 1. Fixed.
2. **`UNIQUE constraint failed` is the actual idempotency error** (not `already exists`). Fixed pattern-matching.
3. **`xray-knife subs fetch` exits 1 if ANY subscription fails** — barry-far/V2ray-Configs path returned 404. Pipeline now tolerates partial failure when at least one sub succeeded (logs warning instead of aborting).

After fixes: `--stage fetch` populated xray-knife.db with 2016 vless configs (1527 from v2go US, 489 from v2go DE) in well under 1 second.

## Deviations from plan

- **mahdibland-vless URL** also 404s (verified). Default `sources.yaml` now ships with it `enabled: true` for documentation but Phase 2 should swap it for a verified-working aggregator. Not phase-1-blocking — v2go US+DE alone gives 2k+ configs.
- **selector.go** uses `modernc.org/sqlite` (pure-Go) instead of `mattn/go-sqlite3` (CGO). Better for ephemeral VPS deploys (no gcc dependency). Confirmed working with both fixture and real xray-knife.db.

---
*Phase: 01-mvp-end-to-end*
*Plan: 01-01*
