# Phase 1: MVP End-to-End - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in `01-CONTEXT.md` — this log preserves the alternatives considered.

**Date:** 2026-05-14
**Phase:** 01-mvp-end-to-end
**Mode:** `--auto` (no interactive Q&A; decisions auto-picked from prior conversation context, PROJECT.md, REQUIREMENTS.md, ROADMAP.md, and teammate research)
**Areas auto-discussed:** xray-knife integration, xray-knife installation, sources.yaml schema, default sources count, run command structure, score formula, tie-breaking, README schema, subs-file schema, test strategy, project layout

---

## xray-knife integration approach

| Option | Description | Selected |
|--------|-------------|----------|
| A) CLI subprocess | `os/exec` xray-knife binary, parse stdout/CSV | ✓ |
| B) Library import | Vendor xray-knife as Go module dep | |

**Auto choice:** A. **Why:** binary-first project, clean process boundary, easy to mock in tests, every xray-knife feature available without re-vendoring. Subprocess overhead is negligible at our scale (Phase 2 may revisit if benchmarks show >5% loss).

---

## xray-knife installation strategy

| Option | Description | Selected |
|--------|-------------|----------|
| A) Auto-install on first run | `go install ...@<pinned>` if missing on $PATH | ✓ |
| B) Require user to pre-install | Document in README, fail fast otherwise | |
| C) Bundle binary in repo | Vendor release artifact | |

**Auto choice:** A. **Why:** zero-touch deployment is a Phase 2 hard requirement; establishing the pattern in Phase 1 means Phase 2 only adds kernel tuning, not install logic.

---

## sources.yaml schema

| Option | Description | Selected |
|--------|-------------|----------|
| A) Flat URL list | `sources: [url1, url2]` | |
| B) Typed entries | `{name, url, kind, enabled}` per source | ✓ |
| C) Mixed | URL list + optional metadata block | |

**Auto choice:** B. **Why:** `kind: country-template` lets one source URL expand to 100+ countries via `{CC}` placeholder. `name` makes logs readable. `enabled` lets users toggle without deleting. Cost vs flat is trivial (one struct).

---

## Default sources.yaml content

| Option | Description | Selected |
|--------|-------------|----------|
| 2 sources (smoke-test friendly) | v2go US + barry-far vless | ✓ |
| 5+ sources from day one | Full upstream catalog | |

**Auto choice:** 2 sources. **Why:** keeps Phase 1 smoke test fast (~30s end-to-end). Phase 2 expands to ~5 sources alongside multi-country expansion.

---

## Run command structure

| Option | Description | Selected |
|--------|-------------|----------|
| A) Single command | `vlessfilter run` only | |
| B) Subcommands | `fetch`, `test`, `select`, `publish` | |
| C) Single command + `--stage` flag | Default end-to-end, optional stage isolation | ✓ |

**Auto choice:** C. **Why:** one keystroke for the happy path; `--stage` is invaluable for debugging output formatting without re-running 5-min speedtests.

---

## Composite score formula

| Option | Description | Selected |
|--------|-------------|----------|
| A) Latency-only | Sort purely by ping ms | |
| B) Weighted normalized | `0.6 * norm_speed - 0.4 * norm_latency`, both capped | ✓ |
| C) Speed-only | Sort purely by Mbps | |

**Auto choice:** B. **Why:** explicit weights are easy to tune empirically in Phase 2. 60/40 toward speed because most users care about throughput once latency is "good enough". Caps prevent a single fast outlier from dominating.

---

## Tie-breaking + low-population countries

| Option | Description | Selected |
|--------|-------------|----------|
| Sort by score desc, then latency asc; emit partials for <3-key countries | Deterministic | ✓ |
| Random tiebreak | Non-deterministic, rejected | |

**Auto choice:** deterministic. **Why:** reproducibility is a Phase 2 requirement (OUT-05) and we want zero rework when we get there.

---

## README.md output schema

**Auto choice:** Country (flag emoji + ISO) | Top latency (ms) | Median speed (Mbps) | Keys | Last tested (UTC). Sorted alphabetically by country code. **Why:** standard, scannable, no scope creep. Deterministic ordering deferred to Phase 2 OUT-05.

---

## subs/<CC>.txt output schema

**Auto choice:** plain `vless://` URIs, one per line, UTF-8, LF endings, no headers. **Why:** canonical subscription format; paste-into-client without transformation per user's spec.

---

## Test strategy for Phase 1

| Option | Description | Selected |
|--------|-------------|----------|
| A) Real-network smoke only | One integration test, no unit tests | |
| B) Mock xray-knife everywhere | Pure unit tests, no real network | |
| C) Both: unit tests in CI + manual smoke | Two tiers | ✓ |

**Auto choice:** C. **Why:** unit tests give CI signal without flakiness; manual smoke validates real integration without making CI depend on flaky public subs.

---

## Project layout

**Auto choice:** standard Go layout with `cmd/vlessfilter/` + `internal/{sources,xrayknife,selector,output,pipeline}/` + `testdata/`. **Why:** zero novelty, every Go developer recognizes it instantly, `internal/` package boundary prevents accidental external API exposure.

---

## Claude's Discretion

The following decisions were left to implementer discretion (low-risk, easy to change):
- Log format / log level (`log/slog` defaults)
- xray-knife version detection (parse `--version` output)
- Concurrency primitive in stage 1 (likely delegated to xray-knife `-t`)
- Exact CLI flag names beyond `--stage`
- Error wrapping conventions

---

## Deferred Ideas

Captured in `01-CONTEXT.md` `<deferred>` section:
- Multi-region tester orchestration (v2)
- Telegram scraping (out of scope permanently)
- Hysteria2 / Trojan / SS / VMess (v2)
- Repeated-runs stability scoring (v2 QUAL-01)
- Web UI (out of scope)
- Library-mode xray-knife (possibly Phase 4+)

---

*Generated by `gsd-discuss-phase 1 --auto`*
