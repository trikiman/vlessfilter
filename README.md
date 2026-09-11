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

Rotating exits: `subs/vless/rotating.txt` (145 configs)

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

Rotating exits: `subs/trojan/rotating.txt` (24 configs)

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

VlessFilter is a single Go binary. Two install paths, pick whichever:

### Option 1: `go install` (requires Go 1.26+)

```bash
go install github.com/trikiman/vlessfilter/cmd/vlessfilter@latest
```
Binary lands in `$GOPATH/bin` (or `$HOME/go/bin`). Make sure that's on your `$PATH`.

### Option 2: From source

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
| 🇦🇪 AE | 1114 | 14.0 | 2 |
| 🇦🇲 AM | 1652 | 14.9 | 1 |
| 🇦🇹 AT | 583 | 32.9 | 1 |
| 🇦🇺 AU | 705 | 21.5 | 2 |
| 🇨🇦 CA | 161 | 93.5 | 3 |
| 🇨🇭 CH | 544 | 16.8 | 1 |
| 🇩🇪 DE | 376 | 37.9 | 3 |
| 🇪🇪 EE | 565 | 29.9 | 2 |
| 🇪🇸 ES | 486 | 36.0 | 3 |
| 🇫🇮 FI | 526 | 34.9 | 3 |
| 🇫🇷 FR | 469 | 38.7 | 3 |
| 🇬🇧 GB | 419 | 38.9 | 3 |
| 🇭🇰 HK | 725 | 25.3 | 3 |
| 🇮🇳 IN | 866 | 20.4 | 2 |
| 🇮🇹 IT | 501 | 34.2 | 3 |
| 🇯🇵 JP | 558 | 33.2 | 3 |
| 🇰🇷 KR | 643 | 28.0 | 1 |
| 🇰🇿 KZ | 821 | 17.8 | 2 |
| 🇱🇻 LV | 626 | 31.5 | 2 |
| 🇲🇽 MX | 375 | 61.7 | 1 |
| 🇲🇾 MY | 807 | 18.5 | 2 |
| 🇳🇱 NL | 427 | 43.3 | 3 |
| 🇳🇴 NO | 495 | 36.5 | 3 |
| 🇵🇱 PL | 527 | 33.7 | 3 |
| 🇷🇴 RO | 629 | 26.2 | 2 |
| 🇸🇪 SE | 538 | 33.3 | 2 |
| 🇸🇬 SG | 805 | 22.1 | 3 |
| 🇹🇷 TR | 709 | 26.6 | 3 |
| 🇹🇼 TW | 713 | 25.0 | 3 |
| 🇺🇸 US | 98 | 136.0 | 3 |
| 🇺🇿 UZ | 1048 | 14.8 | 1 |

**Rotating-exit pool:** 145 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇨🇦 CA | 111 | 64.1 | 2 |
| 🇪🇸 ES | 417 | 17.1 | 1 |
| 🇬🇧 GB | 394 | 15.5 | 1 |
| 🇭🇰 HK | 1108 | 15.3 | 1 |
| 🇯🇵 JP | 405 | 17.2 | 3 |
| 🇰🇷 KR | 489 | 13.3 | 1 |
| 🇳🇱 NL | 573 | 18.9 | 1 |
| 🇸🇬 SG | 597 | 12.9 | 2 |
| 🇺🇸 US | 119 | 62.8 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇫🇷 FR | 625 | 20.3 | 3 |
| 🇳🇱 NL | 675 | 19.3 | 3 |
| 🇺🇸 US | 434 | 35.2 | 3 |

**Rotating-exit pool:** 24 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇨🇦 CA | 83 | 56.7 | 3 |
| 🇩🇪 DE | 354 | 19.8 | 3 |
| 🇩🇰 DK | 425 | 18.2 | 1 |
| 🇪🇸 ES | 431 | 16.4 | 2 |
| 🇫🇮 FI | 602 | 13.4 | 1 |
| 🇫🇷 FR | 378 | 18.0 | 1 |
| 🇬🇧 GB | 319 | 22.1 | 2 |
| 🇮🇹 IT | 372 | 17.1 | 2 |
| 🇯🇵 JP | 403 | 17.6 | 3 |
| 🇳🇱 NL | 325 | 21.6 | 3 |
| 🇵🇱 PL | 365 | 18.1 | 3 |
| 🇷🇺 RU | 374 | 13.8 | 3 |
| 🇺🇸 US | 50 | 125.8 | 3 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-09-11T02:30:15Z -->
