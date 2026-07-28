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

Rotating exits: `subs/vless/rotating.txt` (55 configs)

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
| 🇦🇪 AE | 726 | 13.5 | 3 |
| 🇦🇹 AT | 1058 | 30.7 | 3 |
| 🇧🇪 BE | 583 | 11.4 | 3 |
| 🇧🇬 BG | 569 | 30.4 | 2 |
| 🇧🇷 BR | 481 | 16.2 | 3 |
| 🇨🇦 CA | 117 | 173.9 | 3 |
| 🇨🇭 CH | 493 | 41.0 | 3 |
| 🇨🇳 CN | 1358 | 8.1 | 1 |
| 🇨🇴 CO | 571 | 31.0 | 2 |
| 🇨🇾 CY | 367 | 8.4 | 3 |
| 🇨🇿 CZ | 455 | 41.4 | 3 |
| 🇩🇪 DE | 284 | 49.9 | 2 |
| 🇪🇸 ES | 498 | 0.0 | 1 |
| 🇫🇮 FI | 457 | 42.5 | 3 |
| 🇫🇷 FR | 369 | 53.0 | 3 |
| 🇬🇧 GB | 247 | 53.8 | 3 |
| 🇭🇰 HK | 628 | 37.8 | 3 |
| 🇭🇺 HU | 445 | 42.3 | 2 |
| 🇮🇪 IE | 465 | 4.0 | 3 |
| 🇮🇳 IN | 769 | 23.1 | 3 |
| 🇮🇸 IS | 572 | 34.7 | 2 |
| 🇮🇹 IT | 354 | 60.8 | 3 |
| 🇯🇵 JP | 655 | 28.3 | 2 |
| 🇰🇷 KR | 751 | 24.1 | 3 |
| 🇰🇿 KZ | 1570 | 0.3 | 3 |
| 🇱🇹 LT | 503 | 40.1 | 3 |
| 🇱🇻 LV | 383 | 38.7 | 3 |
| 🇳🇱 NL | 203 | 67.6 | 3 |
| 🇳🇴 NO | 490 | 0.0 | 3 |
| 🇵🇭 PH | 1155 | 9.1 | 1 |
| 🇵🇱 PL | 450 | 44.3 | 3 |
| 🇷🇴 RO | 589 | 1.7 | 1 |
| 🇷🇸 RS | 883 | 0.0 | 3 |
| 🇷🇺 RU | 635 | 35.2 | 3 |
| 🇸🇪 SE | 424 | 45.1 | 3 |
| 🇸🇬 SG | 795 | 44.5 | 3 |
| 🇹🇭 TH | 1164 | 12.7 | 3 |
| 🇹🇷 TR | 438 | 33.6 | 3 |
| 🇹🇼 TW | 597 | 0.0 | 3 |
| 🇺🇦 UA | 671 | 31.8 | 1 |
| 🇺🇸 US | 42 | 192.8 | 3 |
| 🇻🇳 VN | 1044 | 0.0 | 3 |
| 🇿🇦 ZA | 921 | 0.2 | 3 |

**Rotating-exit pool:** 55 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇨🇦 CA | 194 | 0.0 | 3 |
| 🇩🇪 DE | 477 | 0.0 | 3 |
| 🇪🇸 ES | 515 | 26.7 | 3 |
| 🇬🇧 GB | 416 | 29.6 | 3 |
| 🇭🇰 HK | 758 | 28.1 | 3 |
| 🇮🇳 IN | 887 | 0.0 | 3 |
| 🇯🇵 JP | 432 | 40.8 | 3 |
| 🇰🇷 KR | 396 | 19.9 | 2 |
| 🇰🇿 KZ | 960 | 18.2 | 3 |
| 🇲🇽 MX | 1692 | 0.0 | 1 |
| 🇲🇾 MY | 623 | 22.9 | 3 |
| 🇵🇭 PH | 652 | 0.0 | 3 |
| 🇸🇬 SG | 517 | 29.9 | 3 |
| 🇺🇸 US | 28 | 588.2 | 1 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇱 AL | 467 | 31.7 | 3 |
| 🇦🇹 AT | 440 | 40.1 | 1 |
| 🇦🇺 AU | 604 | 23.1 | 3 |
| 🇧🇪 BE | 382 | 42.6 | 3 |
| 🇧🇬 BG | 740 | 22.5 | 1 |
| 🇨🇦 CA | 100 | 107.1 | 1 |
| 🇨🇭 CH | 364 | 45.4 | 3 |
| 🇩🇪 DE | 274 | 50.2 | 3 |
| 🇩🇰 DK | 338 | 30.1 | 3 |
| 🇪🇪 EE | 419 | 12.9 | 2 |
| 🇪🇸 ES | 590 | 33.7 | 3 |
| 🇫🇮 FI | 384 | 37.8 | 3 |
| 🇫🇷 FR | 283 | 53.1 | 3 |
| 🇬🇧 GB | 290 | 51.1 | 3 |
| 🇭🇺 HU | 349 | 4.4 | 1 |
| 🇮🇪 IE | 237 | 0.2 | 3 |
| 🇮🇱 IL | 571 | 31.2 | 2 |
| 🇮🇳 IN | 907 | 20.1 | 2 |
| 🇮🇹 IT | 365 | 43.9 | 3 |
| 🇯🇵 JP | 465 | 14.1 | 2 |
| 🇰🇷 KR | 642 | 0.0 | 3 |
| 🇳🇱 NL | 280 | 49.8 | 2 |
| 🇳🇴 NO | 307 | 0.0 | 3 |
| 🇵🇱 PL | 429 | 38.9 | 3 |
| 🇷🇸 RS | 437 | 39.5 | 3 |
| 🇷🇺 RU | 406 | 16.2 | 3 |
| 🇸🇪 SE | 346 | 30.8 | 2 |
| 🇸🇬 SG | 687 | 10.0 | 3 |
| 🇸🇰 SK | 375 | 41.5 | 3 |
| 🇹🇭 TH | 814 | 0.4 | 3 |
| 🇹🇷 TR | 425 | 29.6 | 3 |
| 🇹🇼 TW | 561 | 0.0 | 3 |
| 🇺🇦 UA | 500 | 0.0 | 3 |
| 🇺🇸 US | 66 | 184.8 | 3 |
| 🇿🇦 ZA | 640 | 0.0 | 3 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-07-28T21:47:21Z -->
