# Phase 3: CI + Polish - Context

**Gathered:** 2026-05-15
**Mode:** `--auto`

<domain>
## Phase Boundary

Ship the project: GitHub Actions workflow that runs the pipeline on a schedule (no-VPS fallback), goreleaser config that builds Linux/macOS binaries on tag, and a deployment runbook for the 2z2 ephemeral VPS path.

</domain>

<decisions>
- **D-20** Refresh workflow: `.github/workflows/refresh.yml` runs every 6 hours via cron + manual dispatch. Installs Go 1.22, builds the binary, runs `vlessfilter run --git-push` with `${{ secrets.GH_TOKEN }}` (or `secrets.GITHUB_TOKEN`). Uses `VLESSFILTER_QUIET=1` so xray-knife progress doesn't pollute the action log.
- **D-21** Release workflow: `.github/workflows/release.yml` triggers on `v*.*.*` tags. Uses goreleaser to produce Linux amd64/arm64 + macOS amd64/arm64 zips. Publishes a GitHub Release with the assets.
- **D-22** Goreleaser config: `.goreleaser.yml` with single `cmd/vlessfilter` binary, no Docker/snap/etc. — keep it minimal.
- **D-23** Deployment runbook: `docs/DEPLOYMENT-VPS.md` walks the user through spinning up a 2z2 VPS, cloning the repo with the PAT pattern, and running the binary. Concrete commands, no theory.
- **D-24** No standalone `README.md` for the project itself — the pipeline-generated `README.md` is the project's primary face. We add a `<!-- vlessfilter-bootstrap -->` comment at the top so first-time visitors know to wait for the first run.
- **D-25** No automated test runs in CI for now (Linux-only test path, the same code we already validate locally). Lint via `gofmt -l` is enough.

</decisions>

<canonical_refs>
- `.planning/PROJECT.md`, `.planning/ROADMAP.md` §"Phase 3"
- `.planning/phases/02-vps-hardening/02-SUMMARY.md` — open follow-ups Phase 3 closes
- goreleaser docs: https://goreleaser.com/quick-start

</canonical_refs>

<deferred>
- Multi-VPS distributed CI matrix → v2
- Dockerfile / container distribution → out of scope (single-binary is the value prop)
- Homebrew tap → v1.1+ if user demand

</deferred>

---
*Phase: 03-ci-polish*
