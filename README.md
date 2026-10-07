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

Rotating exits: `subs/vless/rotating.txt` (0 configs)

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

Rotating exits: `subs/trojan/rotating.txt` (0 configs)

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
| 🇦🇺 AU | 708 | 26.0 | 2 |
| 🇧🇪 BE | 499 | 35.8 | 1 |
| 🇧🇬 BG | 613 | 27.1 | 1 |
| 🇨🇦 CA | 164 | 96.0 | 3 |
| 🇨🇭 CH | 509 | 34.3 | 2 |
| 🇨🇿 CZ | 636 | 34.2 | 1 |
| 🇩🇪 DE | 404 | 36.6 | 3 |
| 🇪🇪 EE | 581 | 25.7 | 3 |
| 🇪🇸 ES | 534 | 33.7 | 2 |
| 🇫🇮 FI | 572 | 32.1 | 3 |
| 🇫🇷 FR | 433 | 40.5 | 3 |
| 🇬🇧 GB | 450 | 40.7 | 2 |
| 🇬🇷 GR | 680 | 28.5 | 1 |
| 🇭🇰 HK | 799 | 29.7 | 1 |
| 🇮🇳 IN | 867 | 14.7 | 1 |
| 🇮🇹 IT | 429 | 35.0 | 3 |
| 🇯🇵 JP | 564 | 32.5 | 3 |
| 🇰🇷 KR | 659 | 22.7 | 2 |
| 🇰🇿 KZ | 1471 | 19.0 | 1 |
| 🇱🇻 LV | 617 | 29.0 | 3 |
| 🇳🇱 NL | 456 | 41.5 | 3 |
| 🇵🇱 PL | 495 | 34.7 | 3 |
| 🇷🇴 RO | 594 | 31.2 | 1 |
| 🇸🇬 SG | 984 | 19.4 | 1 |
| 🇹🇷 TR | 750 | 26.6 | 2 |
| 🇺🇦 UA | 674 | 35.9 | 1 |
| 🇺🇸 US | 105 | 132.2 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇩🇪 DE | 918 | 13.2 | 1 |
| 🇭🇰 HK | 796 | 12.8 | 3 |
| 🇯🇵 JP | 502 | 14.9 | 1 |
| 🇰🇷 KR | 407 | 28.0 | 2 |
| 🇳🇱 NL | 607 | 14.3 | 1 |
| 🇵🇱 PL | 980 | 18.6 | 1 |
| 🇹🇭 TH | 621 | 21.9 | 1 |
| 🇺🇸 US | 92 | 85.0 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇫🇷 FR | 877 | 15.0 | 1 |
| 🇮🇪 IE | 411 | 21.0 | 1 |
| 🇯🇵 JP | 548 | 16.2 | 2 |
| 🇰🇷 KR | 656 | 13.6 | 2 |
| 🇵🇱 PL | 548 | 30.6 | 2 |
| 🇸🇬 SG | 2346 | 19.5 | 1 |
| 🇺🇸 US | 244 | 38.2 | 2 |

**Rotating-exit pool:** 0 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇹 AT | 459 | 29.9 | 1 |
| 🇨🇦 CA | 214 | 65.6 | 1 |
| 🇩🇪 DE | 451 | 32.6 | 3 |
| 🇪🇸 ES | 486 | 26.5 | 1 |
| 🇫🇮 FI | 487 | 20.9 | 2 |
| 🇬🇧 GB | 434 | 31.1 | 3 |
| 🇭🇰 HK | 552 | 21.0 | 1 |
| 🇮🇹 IT | 474 | 28.6 | 1 |
| 🇯🇵 JP | 335 | 42.1 | 3 |
| 🇳🇱 NL | 412 | 32.1 | 3 |
| 🇺🇸 US | 87 | 102.4 | 3 |
| 🇿🇦 ZA | 774 | 17.3 | 1 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-10-07T18:54:21Z -->
