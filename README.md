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

Rotating exits: `subs/vless/rotating.txt` (483 configs)

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

Rotating exits: `subs/trojan/rotating.txt` (71 configs)

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
| 🇦🇱 AL | 697 | 0.0 | 3 |
| 🇧🇷 BR | 486 | 29.0 | 3 |
| 🇨🇦 CA | 194 | 35.8 | 3 |
| 🇩🇪 DE | 433 | 0.0 | 3 |
| 🇪🇸 ES | 369 | 0.0 | 3 |
| 🇬🇧 GB | 395 | 0.0 | 3 |
| 🇭🇺 HU | 643 | 0.0 | 3 |
| 🇮🇳 IN | 631 | 0.0 | 3 |
| 🇮🇹 IT | 420 | 0.0 | 3 |
| 🇯🇵 JP | 347 | 40.4 | 2 |
| 🇰🇷 KR | 413 | 24.6 | 3 |
| 🇰🇿 KZ | 993 | 0.0 | 3 |
| 🇲🇽 MX | 139 | 103.5 | 3 |
| 🇲🇾 MY | 735 | 24.0 | 3 |
| 🇵🇭 PH | 742 | 2.2 | 1 |
| 🇸🇬 SG | 540 | 12.4 | 2 |
| 🇹🇷 TR | 510 | 0.0 | 3 |
| 🇹🇼 TW | 468 | 0.0 | 3 |
| 🇺🇸 US | 50 | 163.6 | 2 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇩🇪 DE | 466 | 39.5 | 2 |
| 🇫🇮 FI | 500 | 35.1 | 3 |
| 🇫🇷 FR | 408 | 0.0 | 2 |
| 🇭🇰 HK | 1456 | 0.4 | 2 |
| 🇯🇵 JP | 578 | 33.1 | 3 |
| 🇱🇻 LV | 524 | 7.4 | 3 |
| 🇳🇱 NL | 441 | 41.9 | 3 |
| 🇵🇱 PL | 677 | 0.0 | 3 |
| 🇷🇺 RU | 580 | 0.9 | 3 |
| 🇸🇪 SE | 463 | 0.0 | 3 |
| 🇸🇬 SG | 978 | 0.0 | 3 |
| 🇺🇸 US | 116 | 31.5 | 3 |
| 🇿🇦 ZA | 1199 | 18.1 | 2 |

**Rotating-exit pool:** 71 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇪 AE | 540 | 25.3 | 3 |
| 🇦🇷 AR | 442 | 32.9 | 3 |
| 🇦🇹 AT | 346 | 0.5 | 3 |
| 🇦🇺 AU | 671 | 20.7 | 3 |
| 🇧🇬 BG | 509 | 36.8 | 3 |
| 🇧🇭 BH | 532 | 1.0 | 3 |
| 🇧🇴 BO | 602 | 0.1 | 3 |
| 🇧🇷 BR | 356 | 0.8 | 3 |
| 🇨🇦 CA | 71 | 180.6 | 3 |
| 🇨🇭 CH | 330 | 0.3 | 3 |
| 🇨🇱 CL | 378 | 0.7 | 3 |
| 🇨🇳 CN | 1224 | 16.8 | 3 |
| 🇨🇾 CY | 459 | 1.5 | 3 |
| 🇩🇪 DE | 345 | 50.2 | 3 |
| 🇪🇨 EC | 359 | 44.0 | 3 |
| 🇪🇪 EE | 382 | 0.0 | 3 |
| 🇪🇬 EG | 498 | 0.5 | 3 |
| 🇪🇸 ES | 411 | 1.6 | 3 |
| 🇫🇮 FI | 329 | 47.1 | 3 |
| 🇬🇧 GB | 244 | 60.6 | 3 |
| 🇭🇷 HR | 311 | 45.5 | 3 |
| 🇮🇩 ID | 731 | 0.8 | 3 |
| 🇮🇪 IE | 530 | 0.2 | 3 |
| 🇮🇳 IN | 611 | 9.0 | 3 |
| 🇮🇹 IT | 405 | 0.4 | 3 |
| 🇯🇵 JP | 495 | 29.5 | 3 |
| 🇰🇭 KH | 923 | 0.8 | 3 |
| 🇰🇷 KR | 735 | 0.0 | 3 |
| 🇱🇹 LT | 351 | 27.0 | 3 |
| 🇱🇻 LV | 342 | 42.7 | 3 |
| 🇲🇰 MK | 390 | 1.2 | 3 |
| 🇲🇹 MT | 429 | 0.3 | 3 |
| 🇳🇬 NG | 1446 | 0.0 | 3 |
| 🇳🇱 NL | 270 | 53.1 | 3 |
| 🇵🇦 PA | 258 | 56.1 | 3 |
| 🇵🇪 PE | 511 | 0.0 | 3 |
| 🇵🇰 PK | 734 | 0.5 | 3 |
| 🇵🇷 PR | 281 | 0.4 | 3 |
| 🇵🇹 PT | 329 | 14.3 | 3 |
| 🇷🇴 RO | 2524 | 2.4 | 3 |
| 🇷🇺 RU | 754 | 21.7 | 1 |
| 🇸🇬 SG | 692 | 20.1 | 3 |
| T1 T1 | 930 | 2.7 | 3 |
| 🇹🇭 TH | 764 | 18.3 | 3 |
| 🇹🇷 TR | 425 | 25.3 | 3 |
| 🇹🇼 TW | 605 | 23.5 | 3 |
| 🇺🇸 US | 17 | 459.8 | 3 |
| 🇿🇦 ZA | 673 | 20.9 | 3 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-05-25T23:10:58Z -->
