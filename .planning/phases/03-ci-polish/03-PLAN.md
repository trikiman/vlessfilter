---
phase: 03-ci-polish
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - .github/workflows/refresh.yml
  - .github/workflows/release.yml
  - .goreleaser.yml
  - docs/DEPLOYMENT-VPS.md
  - docs/INSTALL.md
autonomous: true
requirements:
  - DEP-06
must_haves:
  truths:
    - "refresh.yml is valid GitHub Actions YAML and uses ${{ secrets.GH_TOKEN }} (or secrets.GITHUB_TOKEN) to authenticate the git push"
    - "release.yml triggers on tags matching v*.*.* and runs goreleaser"
    - ".goreleaser.yml produces linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 archives"
    - "docs/DEPLOYMENT-VPS.md contains concrete copy-pasteable commands for the 2z2 ephemeral VPS path"
  artifacts:
    - ".github/workflows/refresh.yml"
    - ".github/workflows/release.yml"
    - ".goreleaser.yml"
    - "docs/DEPLOYMENT-VPS.md"
    - "docs/INSTALL.md"
---

<objective>
Ship Phase 1+2 work as a usable project with scheduled CI, tagged releases, and a one-page deployment runbook.
</objective>

<tasks>

<task type="auto">
  <name>Task 1: Refresh workflow</name>
  <action>
    Write .github/workflows/refresh.yml: cron `0 */6 * * *` + workflow_dispatch, ubuntu-latest runner, actions/checkout, actions/setup-go (1.22), `go install ./cmd/vlessfilter`, then `VLESSFILTER_QUIET=1 vlessfilter run --git-push --git-branch main`. Pass `GH_TOKEN: ${{ secrets.VLESSFILTER_TOKEN || secrets.GITHUB_TOKEN }}`. Set `permissions: contents: write` so the GITHUB_TOKEN can push.
  </action>
  <acceptance_criteria>
    - File parses as YAML
    - Has `on.schedule.cron` and `on.workflow_dispatch`
    - References `secrets.GH_TOKEN` or `secrets.GITHUB_TOKEN`
    - Has `permissions: contents: write`
  </acceptance_criteria>
</task>

<task type="auto">
  <name>Task 2: Release workflow + goreleaser</name>
  <action>
    .github/workflows/release.yml triggers on `tags: v*` push, runs goreleaser/goreleaser-action.
    .goreleaser.yml: builds=cmd/vlessfilter; archives=tar.gz on linux/macOS; checksums; release notes from CHANGELOG (or auto). Linux: amd64+arm64. macOS: amd64+arm64.
  </action>
  <acceptance_criteria>
    - release.yml triggers on v*
    - .goreleaser.yml has goos: [linux, darwin] and goarch: [amd64, arm64]
  </acceptance_criteria>
</task>

<task type="auto">
  <name>Task 3: docs/DEPLOYMENT-VPS.md</name>
  <action>
    One-page runbook for 2z2 ephemeral VPS path. Steps:
    1. Spin VPS at gcp.2z2.top/dashboard (Ubuntu 24.04, 4vCPU/32GB)
    2. SSH in
    3. Install Go 1.22+ (link to one-liner)
    4. clone repo with PAT pattern (paste the `git clone -c http.extraheader=...` command)
    5. cd repo; go install ./cmd/vlessfilter
    6. export GH_TOKEN; ./bin/vlessfilter run --git-push --budget-min 55
    7. VPS auto-deletes; results in repo
    Include a troubleshooting section for: kernel-tune permission errors (run as root or accept warnings), git push failures (check PAT scope), xray-knife install timeouts.
  </action>
  <acceptance_criteria>
    - File exists with `# Ephemeral VPS Deployment` h1
    - Contains `git clone -c http.extraheader=` and `vlessfilter run --git-push`
    - Has Troubleshooting section
  </acceptance_criteria>
</task>

<task type="auto">
  <name>Task 4: docs/INSTALL.md (quick install)</name>
  <action>
    Brief: 3 install paths
    1. Pre-built binary: `curl -sSL ...releases/latest/...`
    2. go install: `go install github.com/trikiman/vlessfilter/cmd/vlessfilter@latest`
    3. From source: `git clone && go build`
    Plus a 5-line "minimum to test it works" section.
  </action>
  <acceptance_criteria>File exists with 3 install methods documented</acceptance_criteria>
</task>

</tasks>

<verification>
- All 5 files exist
- YAML files parse cleanly
- DEPLOYMENT-VPS.md includes literal "2z2" reference + the http.extraheader pattern
</verification>

<output>SUMMARY at .planning/phases/03-ci-polish/03-SUMMARY.md</output>
