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

Rotating exits: `subs/vless/rotating.txt` (471 configs)

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

## VLESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇪 AE | 776 | 14.7 | 3 |
| 🇦🇲 AM | 998 | 21.0 | 3 |
| 🇦🇹 AT | 522 | 44.0 | 3 |
| 🇧🇬 BG | 616 | 29.8 | 3 |
| 🇧🇷 BR | 538 | 13.5 | 3 |
| 🇨🇦 CA | 150 | 141.8 | 3 |
| 🇨🇭 CH | 534 | 37.7 | 3 |
| 🇨🇾 CY | 428 | 0.0 | 3 |
| 🇨🇿 CZ | 505 | 33.2 | 3 |
| 🇩🇪 DE | 337 | 42.8 | 3 |
| 🇩🇰 DK | 455 | 17.9 | 2 |
| 🇪🇪 EE | 553 | 1.4 | 3 |
| 🇪🇸 ES | 845 | 15.6 | 2 |
| 🇫🇮 FI | 517 | 32.9 | 3 |
| 🇫🇷 FR | 429 | 43.8 | 3 |
| 🇬🇧 GB | 448 | 37.4 | 3 |
| 🇭🇰 HK | 763 | 8.8 | 3 |
| 🇭🇺 HU | 1690 | 21.3 | 1 |
| 🇮🇳 IN | 1078 | 2.2 | 3 |
| 🇮🇸 IS | 701 | 31.8 | 1 |
| 🇮🇹 IT | 274 | 49.0 | 3 |
| 🇯🇵 JP | 587 | 30.7 | 3 |
| 🇰🇷 KR | 703 | 26.0 | 3 |
| 🇰🇿 KZ | 1181 | 3.4 | 3 |
| 🇱🇹 LT | 539 | 35.4 | 3 |
| 🇱🇻 LV | 445 | 23.2 | 3 |
| 🇳🇱 NL | 322 | 48.4 | 3 |
| 🇳🇴 NO | 473 | 19.9 | 2 |
| 🇵🇱 PL | 525 | 38.6 | 2 |
| 🇷🇴 RO | 611 | 30.5 | 2 |
| 🇷🇸 RS | 933 | 0.0 | 3 |
| 🇷🇺 RU | 620 | 31.4 | 2 |
| 🇸🇦 SA | 929 | 2.1 | 3 |
| 🇸🇪 SE | 458 | 39.1 | 3 |
| 🇸🇬 SG | 719 | 60.2 | 3 |
| 🇹🇷 TR | 599 | 31.3 | 3 |
| 🇹🇼 TW | 561 | 0.0 | 3 |
| 🇺🇦 UA | 741 | 15.8 | 1 |
| 🇺🇸 US | 102 | 206.2 | 3 |
| 🇿🇦 ZA | 1045 | 0.2 | 3 |

**Rotating-exit pool:** 471 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇨🇦 CA | 150 | 87.3 | 3 |
| 🇩🇪 DE | 436 | 0.0 | 3 |
| 🇪🇸 ES | 451 | 30.4 | 3 |
| 🇭🇰 HK | 838 | 0.3 | 2 |
| 🇮🇳 IN | 923 | 0.0 | 3 |
| 🇯🇵 JP | 490 | 2.5 | 3 |
| 🇰🇷 KR | 572 | 22.9 | 3 |
| 🇲🇾 MY | 583 | 23.5 | 1 |
| 🇳🇱 NL | 3736 | 0.0 | 1 |
| 🇵🇭 PH | 729 | 0.0 | 3 |
| 🇸🇬 SG | 565 | 26.6 | 3 |
| 🇹🇷 TR | 510 | 28.3 | 3 |
| 🇺🇸 US | 143 | 107.4 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇱 AL | 441 | 30.0 | 3 |
| 🇦🇹 AT | 579 | 31.6 | 1 |
| 🇦🇺 AU | 556 | 26.4 | 3 |
| 🇧🇪 BE | 415 | 36.0 | 3 |
| 🇧🇬 BG | 803 | 0.0 | 3 |
| 🇨🇦 CA | 141 | 91.8 | 2 |
| 🇨🇭 CH | 433 | 35.4 | 3 |
| 🇨🇿 CZ | 497 | 0.0 | 3 |
| 🇩🇪 DE | 368 | 40.0 | 3 |
| 🇩🇰 DK | 414 | 36.2 | 2 |
| 🇪🇪 EE | 1163 | 0.0 | 1 |
| 🇪🇸 ES | 603 | 29.8 | 3 |
| 🇫🇮 FI | 458 | 33.8 | 3 |
| 🇫🇷 FR | 369 | 40.5 | 3 |
| 🇭🇺 HU | 864 | 0.0 | 3 |
| 🇮🇱 IL | 577 | 27.3 | 2 |
| 🇮🇳 IN | 671 | 0.0 | 3 |
| 🇮🇹 IT | 458 | 35.5 | 3 |
| 🇯🇵 JP | 406 | 0.0 | 3 |
| 🇰🇷 KR | 560 | 0.0 | 3 |
| 🇱🇻 LV | 494 | 33.4 | 1 |
| 🇳🇱 NL | 353 | 39.8 | 3 |
| 🇳🇴 NO | 362 | 8.9 | 3 |
| 🇵🇱 PL | 448 | 34.3 | 3 |
| 🇷🇴 RO | 520 | 30.4 | 3 |
| 🇷🇸 RS | 511 | 31.9 | 2 |
| 🇷🇺 RU | 502 | 29.3 | 2 |
| 🇸🇪 SE | 413 | 33.0 | 2 |
| 🇸🇬 SG | 602 | 17.2 | 1 |
| 🇸🇰 SK | 453 | 33.3 | 3 |
| 🇹🇭 TH | 741 | 0.8 | 3 |
| 🇹🇷 TR | 513 | 0.0 | 2 |
| 🇹🇼 TW | 505 | 0.0 | 3 |
| 🇺🇦 UA | 709 | 28.4 | 3 |
| 🇺🇸 US | 135 | 113.3 | 3 |
| 🇿🇦 ZA | 729 | 0.0 | 3 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-07-28T14:16:41Z -->
