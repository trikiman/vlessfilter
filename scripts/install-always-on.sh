#!/bin/bash
# VlessFilter always-on VPS install script.
#
# Designed for free-tier Linux VMs that stay on 24/7:
#   - Oracle Cloud Always-Free (4 ARM cores, 24GB RAM)
#   - Google Cloud Free Tier (e2-micro)
#   - AWS Free Tier (t2.micro, 750 hr/month)
#   - Any home server / Raspberry Pi / self-hosted Linux box
#
# What it does:
#   1. Installs Go + git if missing
#   2. Clones the repo
#   3. Builds binary
#   4. Sets up cron schedule (4× per day: 00:00, 06:00, 12:00, 18:00 UTC)
#   5. Stores PAT for git push
#
# Tested on: Ubuntu 22.04, Ubuntu 24.04, Debian 12.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/trikiman/vlessfilter/main/scripts/install-always-on.sh | bash -s -- <github_pat>
#
# Or with explicit args:
#   bash install-always-on.sh github_pat_xxx [user/repo]
#
# Defaults:
#   user/repo = trikiman/vlessfilter
#
set -e

PAT="${1:-${GH_TOKEN:-}}"
UPSTREAM="${2:-trikiman/vlessfilter}"

if [ -z "$PAT" ]; then
  echo "ERROR: GitHub Personal Access Token required as first argument"
  echo "Usage: bash install-always-on.sh <github_pat> [user/repo]"
  exit 1
fi

INSTALL_DIR="${INSTALL_DIR:-$HOME/vlessfilter}"

echo "=== VlessFilter always-on install ==="
echo "Install dir: $INSTALL_DIR"
echo "Upstream:    $UPSTREAM"
echo

# 1. System deps
if ! command -v git >/dev/null 2>&1; then
  echo "--- Installing git ---"
  sudo apt-get update -y >/dev/null 2>&1 || true
  sudo apt-get install -y git curl >/dev/null 2>&1 \
    || sudo dnf install -y git curl \
    || sudo yum install -y git curl
fi

# 2. Go (1.22+)
if ! command -v go >/dev/null 2>&1 || ! go version | grep -qE 'go1\.(2[2-9]|[3-9])'; then
  echo "--- Installing Go 1.22.10 ---"
  ARCH=$(uname -m)
  case "$ARCH" in
    x86_64)  GOARCH=amd64 ;;
    aarch64|arm64) GOARCH=arm64 ;;
    *) echo "Unsupported arch: $ARCH"; exit 1 ;;
  esac
  curl -sSL "https://go.dev/dl/go1.22.10.linux-${GOARCH}.tar.gz" -o /tmp/go.tgz
  rm -rf "$HOME/.local/go"
  mkdir -p "$HOME/.local"
  tar -C "$HOME/.local" -xzf /tmp/go.tgz
  rm /tmp/go.tgz
fi
export PATH="$HOME/.local/go/bin:$HOME/go/bin:$PATH"
export GOPATH="$HOME/go"
go version

# 3. Clone or update repo
if [ ! -d "$INSTALL_DIR" ]; then
  echo "--- Cloning $UPSTREAM ---"
  git clone "https://github.com/$UPSTREAM.git" "$INSTALL_DIR"
else
  echo "--- Updating $INSTALL_DIR ---"
  cd "$INSTALL_DIR"
  git pull --rebase || true
fi
cd "$INSTALL_DIR"

# 4. Build binary
echo "--- Building ---"
go build -o bin/vlessfilter ./cmd/vlessfilter
./bin/vlessfilter help 2>&1 | head -3

# 5. Install xray-knife
echo "--- Installing xray-knife ---"
go install github.com/lilendian0x00/xray-knife/v9@latest

# 6. Persist PAT to a file the cron job will source
mkdir -p "$HOME/.vlessfilter"
chmod 700 "$HOME/.vlessfilter"
cat > "$HOME/.vlessfilter/env" <<EOF
export GH_TOKEN="$PAT"
export UPSTREAM_REPO="$UPSTREAM"
export PATH="\$HOME/.local/go/bin:\$HOME/go/bin:\$PATH"
EOF
chmod 600 "$HOME/.vlessfilter/env"

# 7. Refresh script
cat > "$HOME/.vlessfilter/refresh.sh" <<'EOF'
#!/bin/bash
set -e
source "$HOME/.vlessfilter/env"
cd "$HOME/vlessfilter"
git pull --rebase || true
go build -o bin/vlessfilter ./cmd/vlessfilter
LOG="$HOME/.vlessfilter/refresh-$(date -u +%Y%m%d-%H%M%S).log"
./bin/vlessfilter run --threads1 1000 --budget-min 60 --accuracy-probe \
  --git-push --git-repo . --git-branch main \
  --checkpoint-min 5 \
  >"$LOG" 2>&1 || true
git add -f subs/ README.md all-results.csv raw/dead.txt 2>/dev/null || true
git -c user.name="VlessFilter Bot" -c user.email="vf-bot@$(hostname)" \
  commit -m "results: scheduled refresh $(date -u +%Y%m%d-%H%M%S)" 2>/dev/null || true
git push "https://x-access-token:${GH_TOKEN}@github.com/${UPSTREAM_REPO}.git" \
  HEAD:refs/heads/main 2>>"$LOG" || true
# Keep last 30 logs
ls -1t "$HOME/.vlessfilter"/refresh-*.log 2>/dev/null | tail -n +31 | xargs -r rm
EOF
chmod +x "$HOME/.vlessfilter/refresh.sh"

# 8. Cron schedule (4× daily at 00, 06, 12, 18 UTC)
CRON_LINE="0 0,6,12,18 * * * $HOME/.vlessfilter/refresh.sh"
( crontab -l 2>/dev/null | grep -v 'vlessfilter/refresh.sh'; echo "$CRON_LINE" ) | crontab -
echo "--- Cron installed: 4x daily ---"
crontab -l | grep refresh.sh

echo
echo "=== DONE ==="
echo "First run will trigger at next 6h boundary (UTC). To run NOW:"
echo "  $HOME/.vlessfilter/refresh.sh &"
echo
echo "To monitor:"
echo "  tail -f \$(ls -1t $HOME/.vlessfilter/refresh-*.log | head -1)"
echo
echo "Subscription URLs after first push:"
echo "  https://raw.githubusercontent.com/${UPSTREAM}/main/subs/all.txt"
