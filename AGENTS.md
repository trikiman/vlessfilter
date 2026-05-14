<!-- GSD:project-start source:PROJECT.md -->
## Project

**VlessFilter**

VlessFilter discovers and publishes the top 3 fastest VLESS proxy keys per country, refreshed automatically. It pulls VLESS configs from public subscription aggregators, runs real-proxy latency and speed tests, groups results by exit-IP country, and commits the top 3 per country to a git repo as ready-to-import subscription files. Built to run on ephemeral 60-minute cloud VPS instances and scheduled GitHub Actions.

**Core Value:** Always-fresh, auto-curated, geo-tagged top 3 VLESS keys per country, accessible as a static URL (`https://raw.githubusercontent.com/<user>/<repo>/main/subs/<CC>.txt`) — no client install, no manual testing, just paste-and-import.

### Constraints

- **Tech stack:** Go — matches xray-knife, single static binary, deploys in <10 s on a fresh Ubuntu VPS
- **Wall clock:** ≤60 minutes per VPS run; checkpoint at ≤2-minute intervals
- **Stage 1 concurrency:** up to ~1000 threads — kernel tuning is mandatory (without `ulimit -n` raise + port-range expansion + `tcp_tw_reuse`, the OS kills the process via socket exhaustion)
- **Stage 2 concurrency:** ≤20 threads — higher saturates the VPS uplink and produces meaningless throughput numbers
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

Conventions not yet established. Will populate as patterns emerge during development.
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
