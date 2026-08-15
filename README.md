# VlessFilter Results

Auto-curated top 3 fastest proxy keys per country, refreshed automatically. Multi-protocol: VLESS / VMess / Trojan / Shadowsocks.

## How to use

Pick the protocol your client supports best. Each has its own subscription URLs:

### VLESS

All VLESS countries (single subscription):

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/vless/all.txt
```

Specific country:

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/vless/<CC>.txt
```

Rotating exits: `subs/vless/rotating.txt` (7 configs)

### VMESS

All VMESS countries (single subscription):

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/vmess/all.txt
```

Specific country:

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/vmess/<CC>.txt
```

Rotating exits: `subs/vmess/rotating.txt` (0 configs)

### TROJAN

All TROJAN countries (single subscription):

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/trojan/all.txt
```

Specific country:

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/trojan/<CC>.txt
```

Rotating exits: `subs/trojan/rotating.txt` (2 configs)

### SS

All SS countries (single subscription):

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/ss/all.txt
```

Specific country:

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/ss/<CC>.txt
```

Rotating exits: `subs/ss/rotating.txt` (0 configs)

### All protocols combined (one URL → everything)

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/all.txt
```

Specific country across all protocols:

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/<CC>.txt
```

Rotating exits (all protocols): `subs/rotating.txt`

## Stability filter

Many public configs route through proxy chains, load balancers, or Cloudflare Workers — these have **rotating exit countries** (e.g., one connection lands in Sweden, the next in India). Tagging them with a single country would be misleading.

Each config's full test history is checked:
- **Stable** (always exits same country) → published in `subs/<protocol>/<CC>.txt` with that country code
- **Rotating** (varies across tests, OR is a `*.workers.dev` / `*.pages.dev` host) → published in `subs/<protocol>/rotating.txt` with `🌐 ROTATING` label
- **Dead** → not published

## Install

VlessFilter is a single Go binary. Three install paths, pick whichever:

### Option 1: Pre-built binary (fastest)

Each tagged release ships Linux + macOS binaries on GitHub Releases. Pick from <https://github.com/trikiman/vlessfilter/releases/latest>.

Linux (amd64):

```bash
curl -sSL https://github.com/trikiman/vlessfilter/releases/latest/download/vlessfilter_Linux_amd64.tar.gz \
  | tar -xz -C /tmp && sudo mv /tmp/vlessfilter /usr/local/bin/
```

### Option 2: `go install` (requires Go 1.22+)

```bash
go install github.com/trikiman/vlessfilter/cmd/vlessfilter@latest
```
Binary lands in `$GOPATH/bin` (or `$HOME/go/bin`). Make sure that's on your `$PATH`.

### Option 3: From source

```bash
git clone https://github.com/trikiman/vlessfilter.git
cd vlessfilter
go build -o bin/vlessfilter ./cmd/vlessfilter
```

### Verify it works

```bash
vlessfilter --help
# Quick smoke run against the default sources (writes ./subs/ + ./README.md):
vlessfilter run --threads1 50 --threads2 5 --limit 30 --budget-min 5
ls subs/
```

### Configuration

Edit `sources.yaml` to add or remove subscription sources. See comments in the file for the schema.

## Off-PC Deployment

Run the pipeline without using your own computer. The **primary, recommended path is GitHub Actions** (Option A) — free, fully automated, always-on, and already shipped in this repo (`.github/workflows/refresh.yml`). The other options are optional manual fallbacks.

### What each run does

1. **Stage 1 — alive/handshake check.** High-concurrency TLS handshake against the pool; dead keys are dropped.
2. **Stage 2 — speed connection test.** Survivors get a real-proxy speedtest, run 3 separate times, to measure throughput (Mbps) and latency (ms).
3. **Pre-publish probe.** Top-3-per-country selections are re-tested right before publishing so stale/dead keys never reach the results.

Results appear in `README.md` (per-country latency + median speed table) and `all-results.csv` (full raw results).

### Option A: GitHub Actions (PRIMARY — recommended)

**Cost:** $0 (public repos get unlimited Actions minutes). **Setup:** ~5 min. **Always-on:** yes.

1. Create a GitHub account (any throwaway email; no card needed for free-tier Actions on public repos).
2. **Fork** `https://github.com/trikiman/vlessfilter`.
3. Generate a PAT (Settings → Developer settings → Fine-grained tokens): repository access = your results repo, Permissions → Contents = **Read and write**.
4. In the repo: Settings → Secrets and variables → Actions → new secret `PUSH_TOKEN` = the PAT.
5. Enable Actions: Settings → Actions → General → Allow all.
6. Trigger the first run: Actions tab → **refresh** workflow → Run workflow.

`refresh.yml` runs every 4 hours, does the full alive-check + speed test, and commits fresh results — no involvement from your machine.

### Option B: h2.nexus 15-minute ephemeral VPS (manual fallback, no signup)

Free 15-min VPS (4 CPU / 8 GB / 1 Gbps), no account. Generate a PAT (as in Option A), open <https://h2.nexus/cli>, pick Debian 11, then in the web console run:

```bash
curl -sSL https://raw.githubusercontent.com/trikiman/vlessfilter/main/scripts/h2-quick.sh | bash -s -- ghp_xxx
```

Results push automatically; the VM auto-deletes at 15 min. Manual trigger only (no schedule), reduced run scope to fit the window.

### Option C: Termux on Android (manual fallback)

Install Termux from F-Droid, then:

```bash
pkg update && pkg upgrade -y
pkg install -y golang git curl
curl -sSL https://raw.githubusercontent.com/trikiman/vlessfilter/main/scripts/install-always-on.sh | bash -s -- github_pat_xxx
```

cron may fail under Termux — use `termux-job-scheduler --period-ms 21600000 --script $HOME/.vlessfilter/refresh.sh` (6h). Keep the phone charging and out of deep sleep for a full run.

### Which to pick

| Your situation | Pick |
|----------------|------|
| Fully autonomous, always-on, zero maintenance | **Option A** (GitHub Actions) — default |
| One-off manual refresh, no account | **Option B** (h2.nexus) |
| Only a phone available | **Option C** (Termux) |

> Note: the earlier 2Z2 Cloud Labs VPS runbook is retired (the service shut down). GitHub Actions is the recommended replacement.

## VLESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇱 AL | 729 | 23.6 | 1 |
| 🇦🇹 AT | 695 | 26.2 | 1 |
| 🇦🇺 AU | 612 | 24.3 | 2 |
| 🇧🇦 BA | 885 | 23.2 | 1 |
| 🇧🇬 BG | 719 | 23.3 | 1 |
| 🇨🇦 CA | 265 | 54.4 | 3 |
| 🇨🇭 CH | 718 | 27.1 | 1 |
| 🇨🇿 CZ | 703 | 24.7 | 2 |
| 🇩🇪 DE | 491 | 33.6 | 2 |
| 🇪🇪 EE | 695 | 18.1 | 1 |
| 🇪🇸 ES | 606 | 28.6 | 3 |
| 🇫🇮 FI | 697 | 26.2 | 3 |
| 🇫🇷 FR | 383 | 30.8 | 3 |
| 🇬🇧 GB | 563 | 30.8 | 3 |
| 🇭🇰 HK | 610 | 30.0 | 3 |
| 🇮🇳 IN | 936 | 19.3 | 2 |
| 🇮🇹 IT | 657 | 27.2 | 1 |
| 🇯🇵 JP | 451 | 37.8 | 3 |
| 🇰🇷 KR | 533 | 30.5 | 2 |
| 🇱🇻 LV | 588 | 22.3 | 3 |
| 🇳🇱 NL | 578 | 30.9 | 3 |
| 🇵🇱 PL | 669 | 24.7 | 3 |
| 🇵🇹 PT | 520 | 27.6 | 1 |
| 🇷🇴 RO | 1865 | 21.4 | 1 |
| 🇷🇺 RU | 807 | 23.3 | 1 |
| 🇸🇪 SE | 721 | 26.4 | 1 |
| 🇸🇬 SG | 705 | 26.5 | 3 |
| 🇹🇷 TR | 832 | 22.9 | 2 |
| 🇹🇼 TW | 583 | 23.6 | 2 |
| 🇺🇸 US | 47 | 193.7 | 3 |

**Rotating-exit pool:** 7 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇨🇦 CA | 80 | 23.6 | 1 |
| 🇩🇪 DE | 388 | 16.5 | 1 |
| 🇬🇧 GB | 529 | 39.1 | 1 |
| 🇭🇰 HK | 980 | 22.5 | 1 |
| 🇰🇷 KR | 473 | 26.8 | 2 |
| 🇲🇾 MY | 623 | 21.9 | 1 |
| 🇵🇱 PL | 617 | 43.6 | 1 |
| 🇸🇬 SG | 599 | 12.7 | 2 |
| 🇺🇸 US | 126 | 63.8 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇺 AU | 774 | 21.0 | 3 |
| 🇩🇪 DE | 370 | 46.2 | 3 |
| 🇫🇷 FR | 355 | 48.3 | 3 |
| 🇭🇰 HK | 961 | 17.6 | 3 |
| 🇮🇩 ID | 1637 | 13.6 | 1 |
| 🇮🇪 IE | 286 | 59.6 | 3 |
| 🇯🇵 JP | 640 | 27.4 | 3 |
| 🇰🇷 KR | 746 | 23.4 | 3 |
| 🇳🇱 NL | 353 | 47.6 | 3 |
| 🇵🇱 PL | 433 | 39.0 | 3 |
| 🇸🇬 SG | 896 | 19.6 | 3 |
| 🇺🇸 US | 326 | 54.8 | 3 |

**Rotating-exit pool:** 2 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇺 AU | 555 | 19.1 | 2 |
| 🇧🇷 BR | 404 | 27.4 | 3 |
| 🇩🇪 DE | 330 | 30.8 | 2 |
| 🇪🇸 ES | 405 | 17.2 | 2 |
| 🇫🇮 FI | 583 | 20.4 | 1 |
| 🇬🇧 GB | 301 | 23.7 | 3 |
| 🇳🇱 NL | 343 | 12.3 | 1 |
| 🇵🇭 PH | 763 | 14.0 | 1 |
| 🇹🇷 TR | 765 | 24.5 | 1 |
| 🇺🇸 US | 85 | 76.8 | 3 |
| 🇿🇦 ZA | 702 | 20.0 | 2 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-08-15T13:00:45Z -->
