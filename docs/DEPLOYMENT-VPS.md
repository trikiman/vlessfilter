# Ephemeral VPS Deployment

Run VlessFilter on a free 60-minute Google Cloud VPS via [2z2 Cloud Labs](https://gcp.2z2.top/dashboard). Each run produces a fresh `subs/<CC>.txt` and `README.md` in your GitHub repo.

## Prerequisites

- A GitHub repo to publish results into (your fork of this project, or a separate `vlessfilter-results` repo).
- A GitHub Personal Access Token (PAT) — fine-grained, scoped to that one repo, with **Contents: Write** permission.
- An account at [2z2.top](https://gcp.2z2.top).

## One-time setup (1 minute)

1. Create the PAT: <https://github.com/settings/personal-access-tokens/new>
   - Repository access: **Only select repositories** → pick your results repo
   - Permissions: **Repository permissions → Contents → Read and write**
   - Copy the token (starts with `github_pat_...`)

## Per-run procedure (every time you need fresh keys)

1. **Spin up the VPS.** At <https://gcp.2z2.top/dashboard>, click **Request VPS Access** with Ubuntu 24.04. Wait ~30 seconds for the IP and SSH command.

2. **SSH in** using the command shown on the dashboard.

3. **Install Go** (one-liner, no sudo):
   ```bash
   GO_VER=$(curl -sSL "https://go.dev/VERSION?m=text" | head -1 | sed 's/^go//')
   curl -sSL "https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz" | tar -C "$HOME/.local" -xzf -
   export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"
   export GOPATH="$HOME/go"
   go version
   ```

4. **Set the PAT** (replace with your token):
   ```bash
   export GH_TOKEN="github_pat_11AAAA..."
   ```

5. **Clone the repo using the PAT** (no creds in `.gitconfig`):
   ```bash
   AUTH="Authorization: Basic $(echo -n "oauth2:${GH_TOKEN}" | base64 -w0)"
   git clone -c http.extraheader="${AUTH}" \
     https://github.com/<you>/<your-results-repo>.git vlessfilter
   cd vlessfilter
   ```

6. **Build and run.** The binary auto-installs xray-knife on first run:
   ```bash
   go install ./cmd/vlessfilter
   $HOME/go/bin/vlessfilter run \
     --sources sources.yaml \
     --git-push \
     --git-branch main \
     --budget-min 55 \
     --checkpoint-min 2
   ```

7. **Watch logs.** The pipeline:
   - Pulls fresh subs from `sources.yaml`
   - Runs stage 1 (handshake) on ~thousands of keys at concurrency 200
   - Runs stage 2 (real-proxy speedtest) on survivors at concurrency 20
   - Every 2 minutes commits + pushes partial results
   - At end of run, commits the final `subs/`, `README.md`, `all-results.csv`, `raw/dead.txt`

8. **Done.** The VPS auto-deletes at 60 minutes. Your GitHub repo's `subs/<CC>.txt` files are ready to paste into Hiddify Next, v2rayN, Streisand, etc.

## Subscription URL pattern

```
https://raw.githubusercontent.com/<you>/<your-results-repo>/main/subs/<CC>.txt
```

Examples:
- `https://raw.githubusercontent.com/<you>/vlessfilter-results/main/subs/US.txt`
- `https://raw.githubusercontent.com/<you>/vlessfilter-results/main/subs/DE.txt`

## Troubleshooting

**"sysctl: permission denied" warnings during stage 1.**
The VPS user doesn't have root. The pipeline still runs but stage 1 throughput will be lower (sockets exhaust at the default 1024 FD limit). Either: (a) accept the warning and use `--threads1 100` to stay under the limit; or (b) request a root-shell VPS (some 2z2 instance types provide it).

**"git push: 403" or "remote: Permission denied".**
PAT scope is wrong. Recreate it with **Contents: Write** for that one repo.

**"xray-knife installation succeeded but binary not found on PATH".**
`$GOPATH/bin` isn't on PATH. Run `export PATH="$HOME/go/bin:$PATH"` and retry.

**xray-knife progress bar floods the log.**
Set `VLESSFILTER_QUIET=1` before running. The pipeline auto-detects non-TTY contexts but the auto-detect is heuristic; the env var forces filtering.

**Run hits 55-minute budget without finishing stage 2.**
Add `--limit 5000` to cap stage 2 input size. With 5k keys at 20 threads, stage 2 finishes in ~10 minutes.

## Alternative: GitHub Actions instead of VPS

If you don't want to spin a VPS, the repo's `.github/workflows/refresh.yml` runs the same pipeline on a free GitHub-hosted runner every 6 hours. See `docs/INSTALL.md` for the one-time setup.
