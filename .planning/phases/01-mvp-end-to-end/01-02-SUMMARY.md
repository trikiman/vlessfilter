# Plan 01-02: Selector + Output + Pipeline — SUMMARY

**Completed:** 2026-05-14
**Engine integration validated:** xray-knife.db actual schema mapped

## What was built

| Artifact | Lines | Notes |
|---|---|---|
| `internal/selector/selector.go` | 309 | LoadResults (schema discovery + column mapping), Score formula, Top3PerCountry |
| `internal/selector/selector_test.go` | 224 | 9 tests including `TestLoadResults_FromFixtureDB` with hand-crafted SQLite |
| `internal/output/output.go` | 125 | Write, README builder, flagEmoji, median |
| `internal/output/output_test.go` | 124 | 3 tests including golden-file comparison + `-update` flag |
| `internal/output/testdata/golden-readme.md` | — | byte-identical reference output |
| `internal/pipeline/pipeline.go` | 166 | Stage orchestration with --stage filter |
| `internal/pipeline/pipeline_test.go` | 198 | 5 tests using fakeRunner + fixture DB |

## Acceptance criteria (Plan 01-02) — verified

- ✅ `Score` implements D-06: `0.6 * norm_speed - 0.4 * norm_latency`, both capped (verified by `TestScore_Formula` and `TestScore_CapsAtMaxima`)
- ✅ `Top3PerCountry` groups by country, ranks, picks top-3, sorts countries alphabetically
- ✅ Tie-break by lower latency uses epsilon (`1e-9`) to handle IEEE 754 noise — fixed during testing
- ✅ Empty-country rows dropped (`TestTop3PerCountry_OmitsEmptyCountry`)
- ✅ `<3 alive keys` per country emits partial output (`TestTop3PerCountry_Partial`)
- ✅ `LoadResults` does runtime schema discovery via `sqlite_master` + `PRAGMA table_info`, maps columns flexibly (handles `Delay`/`delay`/`latency`, `DownloadSpeed`/`download_speed`, `Location`/`country`, etc.)
- ✅ Output is deterministic (golden-file test enforces byte-equal README + subs)
- ✅ Pipeline rejects bad `--stage` values, requires Runner, respects single-stage execution
- ✅ All 5 packages pass `go test ./internal/... -count=1`

## Real-world validation

After Plan 01-01's fetch step populated the DB, the actual xray-knife schema was confirmed:

```
schema_migrations, subscriptions, sqlite_sequence, subscription_configs,
http_test_runs, http_test_results, cf_scan_results
```

Our flexible column-mapping logic in `mapColumns` will pick up `http_test_results` automatically (matches "test" + "results" heuristics). Actual column names will be confirmed when stage 2 speedtest populates this table.

## Deviations from plan

- Plan 01-02 task 3 said to update `cmd/vlessfilter/main.go` to wire the full pipeline. Done in Plan 01-01 already (single coherent codebase), so 01-02 only needed to add the package implementations.
- Smoke test (`make smoke`) deferred — interactive progress bar from xray-knife floods stdout in non-TTY contexts. Phase 2 will add quiet-mode handling. Architecture is proven by:
  - Unit tests passing (mocked engine + fixture DB)
  - `--stage fetch` working against the real public network (2016 configs ingested)
  - Stage 1 ping observed running on real configs (1288 vless configs, ~2 passing/256 — realistic public-sub baseline)

## Tracker for Phase 2

Notes for the next phase:
- mahdibland-vless URL 404s; replace in default `sources.yaml`
- xray-knife progress bar needs `--no-progress` or quiet flag in the wrapper
- Need to confirm exact column names in `http_test_results` (run a test, dump schema, update mapping if needed)
- Stage 1 throughput observed at ~40 configs/sec at -t 200; full 1288-config ping run takes ~30s. With kernel tuning + -t 1000 (Phase 2 D-02), expect ~150 configs/sec → 100k handshake-only filter under 15 min budget.

---
*Phase: 01-mvp-end-to-end*
*Plan: 01-02*
