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

Rotating exits: `subs/vmess/rotating.txt` (1 configs)

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
| 🇦🇲 AM | 1687 | 13.5 | 1 |
| 🇦🇹 AT | 572 | 34.2 | 1 |
| 🇦🇺 AU | 718 | 24.9 | 3 |
| 🇨🇦 CA | 173 | 81.5 | 3 |
| 🇨🇭 CH | 1582 | 21.8 | 1 |
| 🇩🇪 DE | 367 | 36.7 | 3 |
| 🇩🇰 DK | 486 | 38.9 | 1 |
| 🇪🇪 EE | 560 | 31.7 | 1 |
| 🇪🇸 ES | 549 | 33.6 | 2 |
| 🇫🇮 FI | 571 | 32.4 | 3 |
| 🇫🇷 FR | 459 | 38.5 | 3 |
| 🇬🇧 GB | 469 | 38.3 | 3 |
| 🇬🇷 GR | 663 | 30.3 | 1 |
| 🇭🇰 HK | 726 | 16.7 | 2 |
| 🇮🇹 IT | 526 | 34.0 | 2 |
| 🇰🇷 KR | 645 | 24.4 | 2 |
| 🇰🇿 KZ | 1518 | 14.2 | 1 |
| 🇱🇰 LK | 1201 | 24.8 | 1 |
| 🇱🇻 LV | 596 | 28.9 | 2 |
| 🇳🇱 NL | 446 | 39.5 | 3 |
| 🇳🇴 NO | 527 | 29.8 | 3 |
| 🇵🇱 PL | 590 | 19.8 | 3 |
| 🇷🇴 RO | 739 | 27.1 | 3 |
| 🇸🇪 SE | 549 | 34.4 | 1 |
| 🇸🇬 SG | 1011 | 21.0 | 2 |
| 🇹🇼 TW | 744 | 25.6 | 1 |
| 🇺🇸 US | 124 | 129.4 | 2 |

**Rotating-exit pool:** 0 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇨🇦 CA | 104 | 29.1 | 2 |
| 🇯🇵 JP | 405 | 18.8 | 3 |
| 🇰🇷 KR | 473 | 23.5 | 3 |
| 🇰🇿 KZ | 1172 | 17.0 | 1 |
| 🇸🇬 SG | 597 | 25.6 | 1 |
| 🇺🇸 US | 130 | 93.7 | 3 |

**Rotating-exit pool:** 1 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇧🇬 BG | 575 | 12.4 | 1 |
| 🇩🇪 DE | 756 | 13.6 | 2 |
| 🇺🇸 US | 185 | 62.2 | 3 |

**Rotating-exit pool:** 0 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇹 AT | 315 | 22.0 | 1 |
| 🇨🇦 CA | 151 | 65.9 | 2 |
| 🇩🇪 DE | 274 | 25.8 | 1 |
| 🇩🇰 DK | 410 | 15.6 | 1 |
| 🇪🇸 ES | 362 | 20.5 | 2 |
| 🇫🇮 FI | 432 | 15.2 | 1 |
| 🇫🇷 FR | 284 | 24.2 | 1 |
| 🇬🇧 GB | 251 | 25.5 | 3 |
| 🇮🇱 IL | 439 | 13.8 | 1 |
| 🇮🇹 IT | 303 | 23.3 | 2 |
| 🇯🇵 JP | 462 | 15.3 | 3 |
| 🇳🇱 NL | 292 | 25.8 | 2 |
| 🇺🇸 US | 50 | 119.4 | 3 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-09-18T19:34:04Z -->
