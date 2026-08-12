#!/bin/bash
# VlessFilter 15-minute quick-run for h2.nexus / similar ephemeral VPS.
#
# Designed to complete in <15 minutes wall clock:
#   - Reduced stage 1 batch (5k untested vs 80k)
#   - Stage 2: 2 speedtest attempts (vs 3)
#   - Pre-publish probe runs (drops dead keys before publish)
#   - Skip post-publish accuracy probe (saves ~3 min)
#   - Fast checkpoint cadence (1 min) so partial progress lands
#   - Push to upstream trikiman/vlessfilter via PAT
#
# h2.nexus workflow:
#   1. Open https://h2.nexus/cli — get a free 15-min Debian 11 VM
#   2. Open the VM console (web terminal)
#   3. Paste this ONE LINE (replace ghp_xxx with your PAT):
#
#      curl -sSL https://raw.githubusercontent.com/trikiman/vlessfilter/main/scripts/h2-quick.sh | bash -s -- ghp_xxx
#
#   4. Wait ~12 minutes — results push to github automatically
#   5. VM auto-deletes at 15 min — no cleanup needed
#
# Repeat whenever you want fresh keys (e.g., 2-4× per day).

set -e

PAT="${1:-${GH_TOKEN:-}}"
UPSTREAM="${2:-trikiman/vlessfilter}"

if [ -z "$PAT" ]; then
  echo "ERROR: GitHub Personal Access Token required as first argument"
  echo "Usage: bash h2-quick.sh <github_pat>"
  exit 1
fi

START_TIME=$(date +%s)

echo "=== VlessFilter h2.nexus quick-refresh ==="
echo "Upstream: $UPSTREAM"
echo "Time budget: 13 minutes (fits inside h2.nexus 15-min VM lifetime)"
echo

# 1. System deps (Debian 11 / Ubuntu compatible)
echo "--- Installing dependencies ---"
sudo apt-get update -qq
sudo apt-get install -y -qq git curl ca-certificates
sudo sysctl -w net.ipv4.tcp_tw_reuse=1 >/dev/null
sudo sysctl -w "net.ipv4.ip_local_port_range=1024 65535" >/dev/null
ulimit -n 100000 || true

# 2. Go (1.22+) — cache to /tmp so re-runs are faster
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) GOARCH=amd64 ;;
  aarch64|arm64) GOARCH=arm64 ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

if [ ! -d /opt/go ] || ! /opt/go/bin/go version 2>/dev/null | grep -qE 'go1\.(2[2-9]|[3-9])'; then
  echo "--- Installing Go ---"
  curl -sSL "https://go.dev/dl/go1.22.10.linux-${GOARCH}.tar.gz" -o /tmp/go.tgz
  sudo rm -rf /opt/go
  sudo tar -C /opt -xzf /tmp/go.tgz
  rm /tmp/go.tgz
fi
export PATH="/opt/go/bin:$HOME/go/bin:$PATH"
export GOPATH="$HOME/go"
go version

# 3. Clone repo
cd /tmp
rm -rf vlessfilter
echo "--- Cloning repo ---"
git clone --depth=1 "https://github.com/$UPSTREAM.git" vlessfilter
cd vlessfilter

# 4. Build
echo "--- Building binary ---"
go build -o bin/vlessfilter ./cmd/vlessfilter

# 5. Install xray-knife (this can take ~2 min on first install)
echo "--- Installing xray-knife ---"
go install github.com/lilendian0x00/xray-knife/v10@v10.1.1

# 6. Run pipeline (tuned for 13-min budget)
ELAPSED=$(( $(date +%s) - START_TIME ))
REMAINING=$(( 13 * 60 - ELAPSED ))
BUDGET_MIN=$(( REMAINING / 60 ))
if [ $BUDGET_MIN -lt 5 ]; then BUDGET_MIN=5; fi
echo "--- Running pipeline (${BUDGET_MIN}-min budget, $ELAPSED seconds setup overhead) ---"

VLESSFILTER_QUIET=1 ./bin/vlessfilter run \
  --threads1 1000 \
  --threads2 20 \
  --budget-min $BUDGET_MIN \
  --checkpoint-min 1 \
  --limit 5000 \
  --git-push \
  --git-repo . \
  --git-branch main \
  || echo "(pipeline returned non-zero — partial outputs may still have committed)"

# 7. Manual push (in case pipeline's --git-push didn't fire on last batch)
echo "--- Final push ---"
git add -f subs/ README.md all-results.csv 2>/dev/null || true
git -c user.name="VlessFilter Bot (h2)" -c user.email="vf-bot@h2.nexus" \
  commit -m "results: h2.nexus refresh $(date -u +%Y%m%d-%H%M%S)" 2>/dev/null || echo "no changes to commit"
git push "https://x-access-token:${PAT}@github.com/${UPSTREAM}.git" \
  HEAD:refs/heads/main || true

# 8. Summary
TOTAL_TIME=$(( $(date +%s) - START_TIME ))
echo
echo "=== DONE in ${TOTAL_TIME}s ==="
echo
echo "Published countries (vless top-level):"
ls subs/*.txt 2>/dev/null | grep -v all.txt | grep -v rotating.txt | wc -l
echo
echo "Subscription URL:"
echo "  https://raw.githubusercontent.com/${UPSTREAM}/main/subs/all.txt"
echo
echo "VM auto-deletes at 15-min mark — no cleanup needed."
