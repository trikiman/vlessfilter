#!/bin/bash
# VlessFilter one-paste VPS deploy script.
#
# Usage on a fresh Ubuntu 24.04 VPS (e.g., 2z2 ephemeral):
#
#   export GH_TOKEN="github_pat_..."           # optional, only if --git-push
#   export GH_USER="your-github-username"      # optional, only if --git-push
#   export GH_REPO="vlessfilter-results"       # optional, only if --git-push
#   curl -sSL https://raw.githubusercontent.com/<you>/vlessfilter/master/scripts/deploy-vps.sh | bash
#
# Or run locally on the VPS after `git clone` of this repo:
#   bash scripts/deploy-vps.sh

set -e

echo "=== VlessFilter VPS deploy ==="
echo

# 1. Install Go if missing
if ! command -v go >/dev/null 2>&1; then
  echo "--- Installing Go 1.22 ---"
  GO_VER=1.22.10
  curl -sSL "https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz" \
    | tar -C "$HOME/.local" -xzf - 2>/dev/null \
    || tar -C /tmp -xzf - 2>/dev/null
  if [ -d "$HOME/.local/go" ]; then
    GOROOT="$HOME/.local/go"
  elif [ -d /tmp/go ]; then
    mkdir -p "$HOME/.local"
    mv /tmp/go "$HOME/.local/go"
    GOROOT="$HOME/.local/go"
  fi
  export PATH="$GOROOT/bin:$HOME/go/bin:$PATH"
  export GOPATH="$HOME/go"
fi
go version

# 2. Clone or update repo
if [ ! -d vlessfilter ]; then
  echo "--- Cloning vlessfilter ---"
  if [ -n "${GH_USER:-}" ] && [ -n "${GH_REPO:-}" ] && [ -n "${GH_TOKEN:-}" ]; then
    AUTH="Authorization: Basic $(echo -n "oauth2:${GH_TOKEN}" | base64 -w0)"
    git clone -c http.extraheader="${AUTH}" \
      "https://github.com/${GH_USER}/${GH_REPO}.git" vlessfilter
  else
    # Public clone fallback (results push will need PAT later)
    git clone https://github.com/trikiman/vlessfilter.git
  fi
fi
cd vlessfilter

# 3. Build
echo "--- Building binary ---"
go build -o bin/vlessfilter ./cmd/vlessfilter

# 4. Run
echo "--- Running pipeline (60-min budget, 1000 stage1 threads) ---"
ARGS=(
  run
  --threads1 1000
  --threads2 20
  --budget-min 55
  --checkpoint-min 2
)
if [ -n "${GH_TOKEN:-}" ]; then
  ARGS+=(--git-push --git-branch main)
  echo "GH_TOKEN set → results will be committed and pushed"
else
  echo "GH_TOKEN not set → results stay local in ./subs/"
fi

VLESSFILTER_QUIET=1 ./bin/vlessfilter "${ARGS[@]}"

# 5. Show what landed
echo
echo "=== Output files ==="
ls -la subs/ 2>/dev/null
echo
echo "=== Country count ==="
ls subs/*.txt 2>/dev/null | wc -l
echo
echo "=== Done. Subscription URLs (after push): ==="
if [ -n "${GH_USER:-}" ] && [ -n "${GH_REPO:-}" ]; then
  echo "https://raw.githubusercontent.com/${GH_USER}/${GH_REPO}/main/subs/<CC>.txt"
else
  echo "(set GH_USER/GH_REPO/GH_TOKEN env vars and re-run with --git-push to publish)"
fi
