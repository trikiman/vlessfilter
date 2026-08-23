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

Rotating exits: `subs/vless/rotating.txt` (204 configs)

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

Rotating exits: `subs/trojan/rotating.txt` (35 configs)

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
| 🇦🇪 AE | 770 | 13.7 | 1 |
| 🇦🇹 AT | 539 | 35.1 | 2 |
| 🇦🇺 AU | 711 | 22.1 | 2 |
| 🇧🇬 BG | 1566 | 18.3 | 1 |
| 🇧🇷 BR | 835 | 28.7 | 1 |
| 🇨🇦 CA | 159 | 106.5 | 3 |
| 🇨🇭 CH | 549 | 24.3 | 2 |
| 🇨🇿 CZ | 559 | 31.4 | 2 |
| 🇩🇪 DE | 364 | 40.6 | 3 |
| 🇪🇪 EE | 543 | 31.8 | 2 |
| 🇪🇸 ES | 503 | 36.3 | 1 |
| 🇫🇮 FI | 485 | 34.5 | 3 |
| 🇫🇷 FR | 361 | 38.1 | 3 |
| 🇬🇧 GB | 422 | 41.4 | 3 |
| 🇭🇰 HK | 720 | 24.9 | 3 |
| 🇮🇳 IN | 851 | 20.1 | 3 |
| 🇮🇹 IT | 574 | 30.4 | 3 |
| 🇯🇵 JP | 539 | 33.3 | 3 |
| 🇰🇷 KR | 652 | 23.2 | 2 |
| 🇰🇿 KZ | 1905 | 17.2 | 1 |
| 🇱🇹 LT | 625 | 32.7 | 1 |
| 🇳🇱 NL | 353 | 57.2 | 3 |
| 🇳🇴 NO | 467 | 34.9 | 2 |
| 🇵🇪 PE | 918 | 29.2 | 1 |
| 🇵🇭 PH | 954 | 20.3 | 1 |
| 🇵🇱 PL | 574 | 34.0 | 3 |
| 🇷🇴 RO | 775 | 29.3 | 1 |
| 🇷🇺 RU | 618 | 28.9 | 1 |
| 🇸🇪 SE | 463 | 33.9 | 3 |
| 🇸🇬 SG | 811 | 23.1 | 3 |
| 🇹🇷 TR | 721 | 23.7 | 2 |
| 🇹🇼 TW | 519 | 19.9 | 2 |
| 🇺🇸 US | 109 | 162.2 | 3 |
| 🇻🇳 VN | 1129 | 18.4 | 1 |

**Rotating-exit pool:** 204 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇨🇦 CA | 79 | 83.7 | 2 |
| 🇬🇧 GB | 344 | 24.2 | 2 |
| 🇭🇰 HK | 1115 | 19.1 | 2 |
| 🇯🇵 JP | 661 | 18.2 | 3 |
| 🇰🇷 KR | 561 | 22.7 | 1 |
| 🇲🇾 MY | 1357 | 14.3 | 2 |
| 🇳🇱 NL | 385 | 49.4 | 1 |
| 🇸🇬 SG | 675 | 20.1 | 2 |
| 🇺🇸 US | 110 | 72.8 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇺 AU | 581 | 14.9 | 3 |
| 🇨🇭 CH | 609 | 14.5 | 1 |
| 🇨🇿 CZ | 862 | 12.8 | 1 |
| 🇩🇪 DE | 734 | 17.0 | 3 |
| 🇪🇪 EE | 624 | 14.8 | 1 |
| 🇫🇷 FR | 611 | 19.8 | 3 |
| 🇭🇰 HK | 788 | 12.0 | 2 |
| 🇳🇱 NL | 533 | 16.8 | 3 |
| 🇵🇱 PL | 621 | 14.5 | 3 |
| 🇷🇺 RU | 779 | 12.8 | 1 |
| 🇸🇪 SE | 735 | 15.6 | 2 |
| 🇸🇬 SG | 821 | 15.1 | 1 |
| 🇺🇸 US | 169 | 58.1 | 3 |

**Rotating-exit pool:** 35 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇺 AU | 517 | 27.5 | 1 |
| 🇧🇷 BR | 431 | 31.7 | 3 |
| 🇨🇦 CA | 87 | 96.6 | 3 |
| 🇩🇪 DE | 345 | 39.8 | 3 |
| 🇪🇸 ES | 430 | 30.5 | 2 |
| 🇫🇮 FI | 460 | 30.8 | 1 |
| 🇫🇷 FR | 329 | 41.1 | 3 |
| 🇬🇧 GB | 332 | 42.6 | 3 |
| 🇮🇹 IT | 473 | 24.9 | 1 |
| 🇳🇱 NL | 317 | 43.7 | 3 |
| 🇵🇱 PL | 414 | 17.8 | 1 |
| 🇷🇺 RU | 454 | 29.3 | 3 |
| 🇸🇬 SG | 594 | 20.6 | 2 |
| 🇺🇸 US | 52 | 112.2 | 3 |
| 🇿🇦 ZA | 723 | 19.4 | 1 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-08-23T05:17:19Z -->
