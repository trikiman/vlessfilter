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

Rotating exits: `subs/vless/rotating.txt` (26 configs)

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

Rotating exits: `subs/trojan/rotating.txt` (17 configs)

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
| 🇦🇱 AL | 586 | 32.2 | 1 |
| 🇦🇺 AU | 1417 | 16.0 | 1 |
| 🇧🇬 BG | 549 | 33.2 | 1 |
| 🇨🇦 CA | 72 | 165.0 | 3 |
| 🇨🇭 CH | 462 | 37.9 | 2 |
| 🇨🇿 CZ | 1020 | 28.8 | 1 |
| 🇩🇪 DE | 332 | 41.9 | 3 |
| 🇪🇪 EE | 534 | 31.9 | 3 |
| 🇪🇸 ES | 497 | 36.2 | 3 |
| 🇫🇮 FI | 515 | 41.5 | 3 |
| 🇫🇷 FR | 413 | 43.2 | 3 |
| 🇬🇧 GB | 397 | 58.7 | 3 |
| 🇭🇰 HK | 740 | 24.2 | 3 |
| 🇭🇺 HU | 499 | 35.7 | 1 |
| 🇮🇳 IN | 860 | 21.2 | 1 |
| 🇮🇹 IT | 395 | 37.7 | 3 |
| 🇯🇵 JP | 593 | 23.8 | 2 |
| 🇰🇷 KR | 693 | 20.2 | 2 |
| 🇰🇿 KZ | 839 | 17.8 | 2 |
| 🇱🇹 LT | 749 | 27.1 | 1 |
| 🇱🇻 LV | 419 | 35.6 | 1 |
| 🇳🇱 NL | 398 | 55.2 | 3 |
| 🇵🇱 PL | 378 | 39.0 | 3 |
| 🇵🇹 PT | 390 | 37.5 | 1 |
| 🇷🇴 RO | 742 | 28.7 | 3 |
| 🇸🇪 SE | 531 | 35.3 | 2 |
| 🇸🇬 SG | 833 | 21.5 | 3 |
| 🇹🇭 TH | 1117 | 18.4 | 1 |
| 🇹🇷 TR | 640 | 28.9 | 3 |
| 🇹🇼 TW | 741 | 25.3 | 1 |
| 🇺🇦 UA | 612 | 32.6 | 1 |
| 🇺🇸 US | 43 | 256.0 | 3 |

**Rotating-exit pool:** 26 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇰🇷 KR | 471 | 23.2 | 3 |
| 🇵🇱 PL | 584 | 24.3 | 1 |
| 🇸🇬 SG | 602 | 12.4 | 1 |
| 🇺🇸 US | 163 | 47.0 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇺 AU | 722 | 23.4 | 2 |
| 🇩🇪 DE | 490 | 40.9 | 2 |
| 🇫🇮 FI | 430 | 33.3 | 2 |
| 🇫🇷 FR | 610 | 24.8 | 3 |
| 🇬🇧 GB | 795 | 52.7 | 1 |
| 🇮🇪 IE | 388 | 45.0 | 3 |
| 🇯🇵 JP | 557 | 32.3 | 3 |
| 🇰🇷 KR | 660 | 27.2 | 3 |
| 🇳🇱 NL | 474 | 35.5 | 1 |
| 🇵🇱 PL | 564 | 30.0 | 2 |
| 🇸🇬 SG | 818 | 20.4 | 1 |
| 🇹🇷 TR | 663 | 17.9 | 1 |
| 🇺🇸 US | 248 | 77.2 | 2 |

**Rotating-exit pool:** 17 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇱 AL | 362 | 39.0 | 2 |
| 🇦🇺 AU | 563 | 12.1 | 2 |
| 🇧🇬 BG | 540 | 13.7 | 1 |
| 🇨🇦 CA | 73 | 41.9 | 3 |
| 🇩🇪 DE | 270 | 50.5 | 3 |
| 🇪🇸 ES | 370 | 19.8 | 1 |
| 🇫🇮 FI | 357 | 20.7 | 2 |
| 🇫🇷 FR | 264 | 31.7 | 2 |
| 🇮🇹 IT | 385 | 37.3 | 2 |
| 🇯🇵 JP | 461 | 15.3 | 3 |
| 🇳🇱 NL | 304 | 22.5 | 1 |
| 🇵🇱 PL | 412 | 16.2 | 1 |
| 🇹🇷 TR | 536 | 14.0 | 1 |
| 🇺🇸 US | 45 | 131.6 | 3 |
| 🇿🇦 ZA | 640 | 21.9 | 1 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-08-13T01:43:52Z -->
