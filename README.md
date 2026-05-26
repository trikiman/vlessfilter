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

Rotating exits: `subs/vless/rotating.txt` (1271 configs)

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

Rotating exits: `subs/trojan/rotating.txt` (81 configs)

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

**Back-compat (v1 URLs):** `subs/all.txt`, `subs/<CC>.txt`, `subs/rotating.txt` continue to work — they mirror the VLESS protocol files.

## Stability filter

Many public configs route through proxy chains, load balancers, or Cloudflare Workers — these have **rotating exit countries** (e.g., one connection lands in Sweden, the next in India). Tagging them with a single country would be misleading.

Each config's full test history is checked:
- **Stable** (always exits same country) → published in `subs/<protocol>/<CC>.txt` with that country code
- **Rotating** (varies across tests, OR is a `*.workers.dev` / `*.pages.dev` host) → published in `subs/<protocol>/rotating.txt` with `🌐 ROTATING` label
- **Dead** → not published

## VLESS — top 3 per country (stable only)

_No stable countries this run._

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇱 AL | 626 | 0.0 | 3 |
| 🇧🇷 BR | 464 | 0.0 | 3 |
| 🇨🇦 CA | 180 | 37.0 | 3 |
| 🇩🇪 DE | 590 | 46.0 | 3 |
| 🇪🇸 ES | 406 | 0.0 | 3 |
| 🇬🇧 GB | 356 | 0.0 | 3 |
| 🇭🇰 HK | 507 | 28.7 | 1 |
| 🇭🇺 HU | 580 | 0.0 | 2 |
| 🇮🇳 IN | 647 | 0.0 | 3 |
| 🇮🇹 IT | 385 | 0.0 | 3 |
| 🇰🇷 KR | 435 | 19.5 | 3 |
| 🇰🇿 KZ | 942 | 0.0 | 3 |
| 🇲🇽 MX | 134 | 0.0 | 3 |
| 🇲🇾 MY | 751 | 0.0 | 3 |
| 🇵🇭 PH | 760 | 0.8 | 3 |
| 🇸🇬 SG | 561 | 0.0 | 1 |
| 🇹🇷 TR | 485 | 0.0 | 3 |
| 🇹🇼 TW | 496 | 0.0 | 3 |
| 🇺🇸 US | 152 | 93.1 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇩🇪 DE | 390 | 48.0 | 2 |
| 🇫🇮 FI | 480 | 35.2 | 3 |
| 🇫🇷 FR | 356 | 0.0 | 2 |
| 🇭🇰 HK | 1108 | 25.3 | 2 |
| 🇯🇵 JP | 647 | 28.7 | 3 |
| 🇰🇿 KZ | 926 | 0.0 | 1 |
| 🇱🇻 LV | 456 | 0.0 | 3 |
| 🇳🇱 NL | 372 | 50.5 | 3 |
| 🇵🇱 PL | 453 | 0.0 | 3 |
| 🇷🇺 RU | 544 | 35.1 | 3 |
| 🇸🇪 SE | 406 | 0.0 | 3 |
| 🇸🇬 SG | 1381 | 0.0 | 3 |
| 🇺🇸 US | 313 | 31.4 | 2 |
| 🇿🇦 ZA | 1121 | 19.1 | 2 |

**Rotating-exit pool:** 81 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇪 AE | 592 | 22.5 | 3 |
| 🇦🇷 AR | 492 | 0.0 | 3 |
| 🇦🇹 AT | 422 | 0.8 | 3 |
| 🇦🇺 AU | 623 | 19.8 | 3 |
| 🇧🇬 BG | 598 | 29.9 | 3 |
| 🇧🇭 BH | 592 | 1.2 | 3 |
| 🇧🇴 BO | 613 | 0.8 | 3 |
| 🇧🇷 BR | 465 | 0.0 | 3 |
| 🇨🇦 CA | 76 | 195.6 | 3 |
| 🇨🇭 CH | 2371 | 0.2 | 3 |
| 🇨🇱 CL | 405 | 0.0 | 3 |
| 🇨🇾 CY | 535 | 2.0 | 3 |
| 🇩🇪 DE | 361 | 39.7 | 3 |
| 🇪🇨 EC | 373 | 41.1 | 3 |
| 🇪🇪 EE | 2139 | 0.2 | 3 |
| 🇪🇬 EG | 576 | 1.9 | 3 |
| 🇪🇸 ES | 495 | 1.5 | 3 |
| 🇫🇮 FI | 406 | 36.0 | 3 |
| 🇬🇧 GB | 315 | 43.1 | 3 |
| 🇭🇰 HK | 633 | 21.6 | 3 |
| 🇭🇷 HR | 385 | 36.1 | 3 |
| 🇮🇩 ID | 656 | 0.7 | 3 |
| 🇮🇪 IE | 2316 | 0.0 | 2 |
| 🇮🇳 IN | 666 | 5.6 | 3 |
| 🇮🇹 IT | 398 | 1.8 | 3 |
| 🇯🇵 JP | 407 | 34.3 | 3 |
| 🇰🇭 KH | 846 | 0.6 | 3 |
| 🇰🇷 KR | 813 | 0.0 | 3 |
| 🇱🇹 LT | 418 | 23.0 | 3 |
| 🇱🇻 LV | 411 | 33.1 | 3 |
| 🇲🇰 MK | 464 | 0.4 | 3 |
| 🇲🇹 MT | 526 | 0.0 | 3 |
| 🇳🇬 NG | 1872 | 0.0 | 3 |
| 🇳🇱 NL | 360 | 40.1 | 3 |
| 🇵🇦 PA | 311 | 45.7 | 3 |
| 🇵🇪 PE | 561 | 0.9 | 3 |
| 🇵🇰 PK | 836 | 0.7 | 3 |
| 🇵🇷 PR | 299 | 1.2 | 3 |
| 🇵🇹 PT | 417 | 13.0 | 3 |
| 🇷🇺 RU | 3975 | 17.0 | 1 |
| 🇸🇬 SG | 600 | 9.3 | 3 |
| 🇹🇭 TH | 669 | 21.0 | 3 |
| 🇹🇷 TR | 499 | 25.4 | 3 |
| 🇹🇼 TW | 525 | 17.3 | 3 |
| 🇺🇸 US | 52 | 278.7 | 3 |
| 🇿🇦 ZA | 741 | 18.6 | 3 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-05-26T04:44:00Z -->
