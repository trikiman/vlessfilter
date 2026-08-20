<!-- GSD:project-start source:PROJECT.md -->
## Project

**VlessFilter**

VlessFilter discovers and publishes the top 3 fastest VLESS proxy keys per country, refreshed automatically. It pulls VLESS configs from public subscription aggregators, runs real-proxy latency and speed tests, groups results by exit-IP country, and commits the top 3 per country to a git repo as ready-to-import subscription files. Built to run on ephemeral 60-minute cloud VPS instances and scheduled GitHub Actions.

**Core Value:** Always-fresh, auto-curated, geo-tagged top 3 VLESS keys per country, accessible as a static URL (`https://raw.githubusercontent.com/<user>/<repo>/main/subs/<CC>.txt`) — no client install, no manual testing, just paste-and-import.

### Constraints

- **Tech stack:** Go — matches xray-knife, single static binary, deploys in <10 s on a fresh Ubuntu VPS
- **Wall clock:** `--budget-min 50` per protocol job, `--checkpoint-min 5` (this
  read "≤60 minutes; checkpoint at ≤2-minute intervals" — the 2-minute figure
  was never shipped; see `refresh.yml`)
- **Stage 1 concurrency:** `--threads1` defaults to **3000** (`benchmark.yml`
  sweeps up to 6000); this said "~1000". Kernel tuning is mandatory either way —
  without `ulimit -n` raise + port-range expansion + `tcp_tw_reuse`, the OS
  kills the process via socket exhaustion
- **Stage 2 concurrency:** ≤20 threads — higher saturates the uplink and
  produces meaningless throughput numbers
- **Compute:** GitHub Actions runners, NOT a VPS. The "fresh Ubuntu VPS" framing
  below predates the hard constraint in `.planning/STATE.md` ("ALL compute must
  be free cloud: GitHub Actions runners"). Nothing runs on the operator's PC.
- **Licensing:** project itself MIT or Apache-2.0; respects upstream (GPL-3 for v2go means we link as a data source, not a code dependency)
- **No state on the VPS:** SQLite + temp files vanish at 60 min; everything that matters is in git
- **Privacy:** results are public proxy keys harvested from already-public subs; no user accounts, no PII
<!-- GSD:project-end -->

<!-- GSD:stack-start source:STACK.md -->
## Technology Stack

Technology stack not yet documented. Will populate after codebase mapping or first phase.
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

### Published key naming (remark / URI fragment) — MANDATORY

Every published key's name (the `#fragment` at the end of each proxy URI in
`subs/**`) **MUST begin with the country flag emoji, 100% of the time.** The
flag is always the first character(s) of the name — no icon, label, or speed
indicator may come before it.

Canonical order for a published (stable-country) key name:

```
<flag> <speed-icon?> <CC> <Country Name>
```

- `<flag>` — country flag emoji. **Always first. Never omitted for a
  country-tagged key.**
- `<speed-icon?>` — optional speed-tier icon, placed AFTER the flag (e.g.
  `📺` 1080p-ready ≥12 Mbps, `🎬` ≥25 Mbps, `⚡` ≥60 Mbps).
- `<speed?>` — bracketed measured speed, e.g. `[19.4 mb]`, emitted right after
  the icon when a positive speed was measured, omitted otherwise.
  This line previously read "Icon-only; no raw Mbps number in the name (the
  number lives in `README.md` and `all-results.csv`)" — but `rewriteRemark()`
  in `internal/output/output.go` has always emitted the number, and
  `TestRewriteRemark` asserts the `[15.0 mb]` form. Code and test agreed; only
  this doc disagreed, so it was telling contributors the shipped behaviour was
  a violation to "fix". Corrected to match reality.
  Caveat on the unit: `mb` is not a real unit, and the value is measured from a
  GitHub Actions runner, not from the subscriber's network — treat it as a
  relative hint, not a throughput promise.
- `<CC>` — 2-letter ISO country code (stable fallback when a client font has
  no emoji support).
- `<Country Name>` — human-readable name.

Rotating-exit keys are the one exception to the country flag: they use the
`🌐 ROTATING` label instead (no single country can be honestly claimed).

Implementation note: this rule is enforced in `rewriteRemark()` in
`internal/output/output.go`. Any change to the name format must keep the flag
first and must be covered by `TestRewriteRemark` in
`internal/output/output_test.go`.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, or `.github/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
