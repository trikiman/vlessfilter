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

Rotating exits: `subs/vless/rotating.txt` (68 configs)

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

Rotating exits: `subs/trojan/rotating.txt` (38 configs)

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
| 🇦🇱 AL | 485 | 32.5 | 2 |
| 🇦🇹 AT | 420 | 42.8 | 3 |
| 🇦🇺 AU | 768 | 23.6 | 1 |
| 🇧🇷 BR | 500 | 35.7 | 2 |
| 🇧🇾 BY | 656 | 25.3 | 1 |
| 🇨🇦 CA | 128 | 122.4 | 3 |
| 🇨🇭 CH | 413 | 32.4 | 2 |
| 🇨🇿 CZ | 1093 | 31.2 | 1 |
| 🇩🇪 DE | 276 | 51.1 | 3 |
| 🇪🇪 EE | 472 | 37.1 | 2 |
| 🇪🇸 ES | 366 | 54.5 | 1 |
| 🇫🇮 FI | 479 | 38.8 | 3 |
| 🇫🇷 FR | 262 | 50.3 | 3 |
| 🇬🇧 GB | 342 | 51.9 | 3 |
| 🇭🇰 HK | 793 | 22.9 | 3 |
| 🇮🇳 IN | 773 | 22.6 | 2 |
| 🇮🇸 IS | 499 | 47.5 | 1 |
| 🇮🇹 IT | 410 | 45.5 | 3 |
| 🇯🇵 JP | 934 | 20.3 | 2 |
| 🇰🇷 KR | 723 | 22.4 | 2 |
| 🇰🇿 KZ | 737 | 16.0 | 2 |
| 🇱🇹 LT | 585 | 35.9 | 1 |
| 🇳🇱 NL | 237 | 59.2 | 3 |
| 🇳🇴 NO | 407 | 41.1 | 3 |
| 🇵🇪 PE | 582 | 35.3 | 1 |
| 🇵🇱 PL | 441 | 37.2 | 3 |
| 🇷🇴 RO | 462 | 32.4 | 2 |
| 🇷🇺 RU | 584 | 17.2 | 1 |
| 🇸🇪 SE | 469 | 34.7 | 2 |
| 🇸🇬 SG | 914 | 19.9 | 3 |
| 🇹🇷 TR | 508 | 32.1 | 3 |
| 🇹🇼 TW | 581 | 23.7 | 2 |
| 🇺🇸 US | 51 | 300.3 | 3 |

**Rotating-exit pool:** 68 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇰🇷 KR | 393 | 30.1 | 1 |
| 🇺🇸 US | 49 | 49.1 | 2 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇺 AU | 703 | 36.2 | 1 |
| 🇩🇪 DE | 795 | 24.2 | 3 |
| 🇫🇮 FI | 702 | 33.5 | 1 |
| 🇫🇷 FR | 763 | 24.8 | 3 |
| 🇭🇰 HK | 621 | 28.7 | 2 |
| 🇱🇻 LV | 888 | 25.6 | 1 |
| 🇳🇱 NL | 759 | 33.8 | 1 |
| 🇷🇺 RU | 812 | 13.0 | 1 |
| 🇸🇪 SE | 721 | 30.5 | 2 |
| 🇸🇬 SG | 753 | 15.9 | 1 |
| 🇺🇸 US | 516 | 65.7 | 2 |

**Rotating-exit pool:** 38 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇱 AL | 481 | 28.7 | 1 |
| 🇦🇺 AU | 522 | 27.1 | 1 |
| 🇨🇦 CA | 108 | 67.8 | 2 |
| 🇨🇱 CL | 430 | 30.1 | 1 |
| 🇩🇪 DE | 374 | 16.9 | 3 |
| 🇪🇸 ES | 465 | 33.3 | 1 |
| 🇫🇷 FR | 324 | 37.3 | 1 |
| 🇬🇧 GB | 342 | 39.6 | 3 |
| 🇮🇹 IT | 371 | 37.3 | 1 |
| 🇳🇱 NL | 324 | 40.1 | 3 |
| 🇳🇴 NO | 376 | 13.2 | 1 |
| 🇵🇱 PL | 392 | 34.1 | 1 |
| 🇷🇺 RU | 588 | 29.6 | 1 |
| 🇸🇬 SG | 595 | 23.5 | 3 |
| 🇺🇸 US | 73 | 108.5 | 3 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-08-25T01:20:43Z -->
