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

Rotating exits: `subs/vless/rotating.txt` (35 configs)

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

Rotating exits: `subs/trojan/rotating.txt` (12 configs)

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
| 🇦🇪 AE | 769 | 21.0 | 1 |
| 🇦🇺 AU | 1490 | 16.9 | 1 |
| 🇨🇦 CA | 81 | 181.3 | 2 |
| 🇨🇭 CH | 448 | 42.8 | 2 |
| 🇨🇿 CZ | 983 | 30.1 | 1 |
| 🇩🇪 DE | 304 | 54.2 | 3 |
| 🇪🇪 EE | 769 | 25.8 | 2 |
| 🇫🇮 FI | 496 | 37.3 | 3 |
| 🇫🇷 FR | 256 | 58.4 | 2 |
| 🇬🇧 GB | 252 | 24.3 | 3 |
| 🇭🇰 HK | 804 | 22.5 | 3 |
| 🇮🇳 IN | 979 | 13.6 | 1 |
| 🇮🇹 IT | 708 | 39.2 | 1 |
| 🇯🇵 JP | 648 | 26.4 | 2 |
| 🇰🇷 KR | 756 | 23.2 | 1 |
| 🇱🇻 LV | 416 | 30.2 | 1 |
| 🇳🇱 NL | 356 | 49.9 | 3 |
| 🇵🇱 PL | 438 | 40.2 | 3 |
| 🇵🇹 PT | 353 | 41.5 | 1 |
| 🇷🇴 RO | 842 | 26.8 | 1 |
| 🇷🇺 RU | 610 | 31.8 | 3 |
| 🇸🇪 SE | 416 | 39.5 | 2 |
| 🇸🇬 SG | 917 | 18.9 | 3 |
| 🇹🇷 TR | 968 | 24.1 | 1 |
| 🇹🇼 TW | 806 | 23.0 | 1 |
| 🇺🇸 US | 37 | 157.0 | 2 |

**Rotating-exit pool:** 35 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇨🇦 CA | 62 | 32.7 | 1 |
| 🇩🇪 DE | 485 | 21.1 | 2 |
| 🇬🇧 GB | 500 | 18.3 | 2 |
| 🇯🇵 JP | 1879 | 15.5 | 1 |
| 🇰🇷 KR | 562 | 20.4 | 2 |
| 🇵🇱 PL | 512 | 51.2 | 1 |
| 🇺🇸 US | 18 | 80.0 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇩🇪 DE | 580 | 32.8 | 2 |
| 🇫🇷 FR | 558 | 27.4 | 1 |
| 🇷🇺 RU | 850 | 28.5 | 3 |
| 🇸🇪 SE | 847 | 40.1 | 1 |

**Rotating-exit pool:** 12 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇱 AL | 523 | 25.6 | 2 |
| 🇦🇺 AU | 435 | 15.8 | 2 |
| 🇨🇦 CA | 286 | 60.3 | 1 |
| 🇩🇪 DE | 435 | 15.4 | 2 |
| 🇪🇸 ES | 483 | 14.6 | 1 |
| 🇫🇮 FI | 575 | 19.8 | 1 |
| 🇫🇷 FR | 476 | 29.8 | 2 |
| 🇬🇧 GB | 409 | 33.8 | 2 |
| 🇮🇹 IT | 538 | 20.0 | 1 |
| 🇯🇵 JP | 334 | 31.2 | 2 |
| 🇳🇱 NL | 403 | 31.3 | 3 |
| 🇵🇱 PL | 581 | 12.4 | 1 |
| 🇸🇬 SG | 557 | 15.0 | 1 |
| 🇹🇷 TR | 1313 | 21.1 | 1 |
| 🇺🇸 US | 65 | 137.2 | 3 |
| 🇿🇦 ZA | 785 | 17.5 | 1 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-08-12T09:46:46Z -->
