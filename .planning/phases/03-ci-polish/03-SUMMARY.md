# Phase 3: CI + Polish — SUMMARY

**Completed:** 2026-05-15
**Plan:** 03-01 (single plan, 4 tasks)

## What was added

| File | Purpose |
|---|---|
| `.github/workflows/refresh.yml` | Cron-driven (every 6h) + manual-dispatch workflow that builds the binary and runs `vlessfilter run --git-push` with a 50-min budget on a free GitHub-hosted Linux runner. Uses `secrets.VLESSFILTER_TOKEN || secrets.GITHUB_TOKEN` for auth and `permissions: contents: write` so it can push back to the repo. Sets `VLESSFILTER_QUIET=1` so xray-knife progress bar doesn't pollute the action log. Uploads `all-results.csv`/`raw/dead.txt` as 7-day-retention artifacts. |
| `.github/workflows/release.yml` | Triggers on `v*.*.*` tags. Runs goreleaser-action v6 (goreleaser ~> v2). |
| `.goreleaser.yml` | Builds `cmd/vlessfilter` for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`. CGO disabled (we use pure-Go `modernc.org/sqlite` per Phase 1 decision). Archives include `LICENSE*`, `README.md`, `sources.yaml`, `docs/`. SHA256 checksums file. Auto-changelog excludes docs/test/chore commits. |
| `docs/DEPLOYMENT-VPS.md` | Concrete 7-step runbook for the 2z2 Cloud Labs ephemeral-VPS path. Includes the no-sudo Go install one-liner, the `git clone -c http.extraheader=...` pattern, exact `vlessfilter run` invocation with budget+checkpoint+git-push flags, and a Troubleshooting section. |
| `docs/INSTALL.md` | 3 install paths (pre-built binary curl, `go install @latest`, from source). 5-line "verify it works" smoke. Brief GitHub Actions setup walkthrough. |

## Acceptance criteria — all verified

- ✅ refresh.yml is valid YAML, has `on.schedule.cron`, `on.workflow_dispatch`, `permissions: contents: write`, references `secrets.VLESSFILTER_TOKEN || secrets.GITHUB_TOKEN`
- ✅ release.yml triggers on `tags: [v*.*.*]` push and runs goreleaser
- ✅ .goreleaser.yml builds linux+darwin × amd64+arm64 archives with checksums
- ✅ DEPLOYMENT-VPS.md contains concrete `git clone -c http.extraheader=` command + `vlessfilter run --git-push` + Troubleshooting section + literal "2z2" reference
- ✅ INSTALL.md documents all 3 install methods + minimum-viable smoke test

## Phase 3 requirements coverage (1/1)

- ✅ DEP-06 (GitHub Actions cron workflow) — `.github/workflows/refresh.yml`

Plus extras not in REQUIREMENTS.md but in PROJECT.md "Constraints":
- Single static binary deployment → goreleaser config achieves this
- 60-minute VPS friendly → DEPLOYMENT-VPS.md walks the user through it

---
*Phase: 03-ci-polish*
*Files added: 5 (no Go code — purely CI/docs work)*
