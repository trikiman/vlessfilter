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

Rotating exits: `subs/vless/rotating.txt` (99 configs)

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

Rotating exits: `subs/trojan/rotating.txt` (29 configs)

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
| 🇦🇪 AE | 758 | 18.9 | 1 |
| 🇦🇱 AL | 480 | 38.0 | 1 |
| 🇧🇬 BG | 499 | 36.9 | 1 |
| 🇨🇦 CA | 93 | 136.3 | 3 |
| 🇨🇭 CH | 586 | 36.3 | 2 |
| 🇨🇿 CZ | 497 | 34.3 | 2 |
| 🇩🇪 DE | 377 | 49.7 | 3 |
| 🇪🇪 EE | 481 | 35.5 | 2 |
| 🇪🇸 ES | 394 | 48.5 | 2 |
| 🇫🇮 FI | 354 | 39.2 | 2 |
| 🇫🇷 FR | 316 | 54.8 | 3 |
| 🇬🇧 GB | 317 | 56.9 | 3 |
| 🇭🇰 HK | 680 | 22.0 | 3 |
| 🇭🇺 HU | 449 | 39.3 | 1 |
| 🇮🇳 IN | 850 | 20.5 | 2 |
| 🇮🇹 IT | 403 | 39.8 | 3 |
| 🇯🇵 JP | 657 | 27.3 | 3 |
| 🇰🇷 KR | 754 | 23.6 | 2 |
| 🇰🇿 KZ | 742 | 24.0 | 2 |
| 🇱🇹 LT | 546 | 36.9 | 1 |
| 🇱🇻 LV | 363 | 36.5 | 2 |
| 🇳🇱 NL | 204 | 53.1 | 3 |
| 🇵🇱 PL | 430 | 39.9 | 3 |
| 🇵🇹 PT | 338 | 43.9 | 1 |
| 🇷🇴 RO | 499 | 33.9 | 2 |
| 🇷🇺 RU | 491 | 36.0 | 2 |
| 🇸🇪 SE | 408 | 40.7 | 2 |
| 🇸🇬 SG | 895 | 19.5 | 3 |
| 🇹🇭 TH | 1198 | 17.3 | 1 |
| 🇹🇷 TR | 553 | 32.9 | 3 |
| 🇹🇼 TW | 792 | 23.3 | 1 |
| 🇺🇦 UA | 586 | 34.3 | 1 |
| 🇺🇸 US | 50 | 320.0 | 3 |

**Rotating-exit pool:** 99 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇨🇦 CA | 87 | 28.3 | 1 |
| 🇪🇸 ES | 318 | 48.7 | 1 |
| 🇫🇮 FI | 355 | 39.8 | 1 |
| 🇯🇵 JP | 652 | 18.5 | 2 |
| 🇰🇷 KR | 552 | 23.1 | 1 |
| 🇲🇾 MY | 691 | 19.9 | 1 |
| 🇺🇸 US | 54 | 69.9 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇺 AU | 581 | 15.4 | 3 |
| 🇩🇪 DE | 543 | 15.7 | 3 |
| 🇫🇷 FR | 538 | 16.5 | 3 |
| 🇭🇰 HK | 799 | 12.0 | 1 |
| 🇮🇪 IE | 483 | 18.7 | 3 |
| 🇯🇵 JP | 450 | 20.6 | 3 |
| 🇰🇷 KR | 559 | 16.3 | 3 |
| 🇳🇱 NL | 537 | 16.7 | 3 |
| 🇵🇱 PL | 601 | 14.9 | 3 |
| 🇸🇬 SG | 708 | 12.3 | 3 |
| 🇹🇷 TR | 807 | 14.1 | 2 |
| 🇺🇸 US | 167 | 57.7 | 3 |

**Rotating-exit pool:** 29 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇱 AL | 550 | 25.4 | 2 |
| 🇦🇺 AU | 482 | 14.9 | 1 |
| 🇧🇬 BG | 754 | 12.1 | 1 |
| 🇨🇦 CA | 30 | 123.0 | 2 |
| 🇩🇪 DE | 481 | 29.3 | 1 |
| 🇪🇸 ES | 543 | 13.2 | 1 |
| 🇫🇮 FI | 527 | 26.9 | 1 |
| 🇫🇷 FR | 477 | 31.4 | 1 |
| 🇬🇧 GB | 410 | 31.0 | 1 |
| 🇮🇹 IT | 528 | 26.3 | 2 |
| 🇯🇵 JP | 298 | 23.0 | 3 |
| 🇳🇱 NL | 471 | 14.8 | 1 |
| 🇵🇱 PL | 475 | 19.1 | 3 |
| 🇸🇬 SG | 528 | 27.1 | 1 |
| 🇺🇸 US | 159 | 81.3 | 2 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-08-14T01:43:17Z -->
