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

### TROJAN

All TROJAN countries (single subscription):

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/trojan/all.txt
```

Specific country:

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/trojan/<CC>.txt
```

Rotating exits: `subs/trojan/rotating.txt` (3 configs)

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
| 🇦🇱 AL | 499 | 0.0 | 2 |
| 🇦🇲 AM | 1421 | 0.9 | 3 |
| 🇦🇹 AT | 459 | 42.5 | 3 |
| 🇦🇺 AU | 805 | 21.7 | 3 |
| 🇧🇦 BA | 491 | 3.8 | 3 |
| 🇨🇦 CA | 64 | 193.2 | 3 |
| 🇨🇭 CH | 591 | 55.4 | 3 |
| 🇨🇿 CZ | 908 | 52.3 | 3 |
| 🇩🇪 DE | 275 | 97.7 | 3 |
| 🇪🇪 EE | 427 | 42.5 | 3 |
| 🇫🇮 FI | 392 | 43.0 | 3 |
| 🇫🇷 FR | 240 | 64.3 | 3 |
| 🇬🇧 GB | 326 | 58.8 | 3 |
| 🇬🇷 GR | 816 | 9.9 | 3 |
| 🇭🇰 HK | 818 | 22.6 | 3 |
| 🇮🇪 IE | 374 | 51.3 | 3 |
| 🇮🇱 IL | 626 | 32.9 | 3 |
| 🇮🇹 IT | 312 | 46.4 | 3 |
| 🇯🇵 JP | 661 | 27.9 | 3 |
| 🇰🇷 KR | 762 | 24.2 | 3 |
| 🇰🇿 KZ | 937 | 8.5 | 3 |
| 🇱🇹 LT | 470 | 39.9 | 3 |
| 🇱🇺 LU | 842 | 31.6 | 3 |
| 🇱🇻 LV | 492 | 25.3 | 2 |
| 🇳🇱 NL | 227 | 112.2 | 3 |
| 🇳🇴 NO | 589 | 39.5 | 3 |
| 🇵🇱 PL | 449 | 39.2 | 1 |
| 🇷🇸 RS | 553 | 29.4 | 3 |
| 🇷🇺 RU | 902 | 30.4 | 1 |
| 🇸🇦 SA | 924 | 1.9 | 3 |
| 🇸🇪 SE | 688 | 36.4 | 1 |
| 🇸🇬 SG | 899 | 20.3 | 3 |
| 🇹🇭 TH | 1066 | 0.3 | 2 |
| 🇹🇷 TR | 632 | 34.7 | 3 |
| 🇺🇸 US | 44 | 396.0 | 3 |

**Rotating-exit pool:** 55 configs in `subs/vless/rotating.txt`

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇱 AL | 525 | 0.0 | 3 |
| 🇨🇦 CA | 93 | 55.7 | 3 |
| 🇪🇸 ES | 295 | 39.8 | 3 |
| 🇬🇧 GB | 248 | 28.1 | 3 |
| 🇬🇷 GR | 1456 | 0.0 | 1 |
| 🇭🇰 HK | 599 | 0.0 | 3 |
| 🇮🇳 IN | 645 | 0.0 | 3 |
| 🇮🇹 IT | 290 | 0.0 | 2 |
| 🇯🇵 JP | 463 | 0.0 | 3 |
| 🇰🇷 KR | 567 | 21.1 | 1 |
| 🇰🇿 KZ | 2620 | 1.2 | 1 |
| 🇲🇽 MX | 166 | 77.1 | 3 |
| 🇵🇭 PH | 910 | 0.0 | 3 |
| 🇵🇱 PL | 1000 | 1.4 | 1 |
| 🇸🇪 SE | 431 | 37.8 | 1 |
| 🇸🇬 SG | 676 | 22.6 | 3 |
| 🇹🇼 TW | 600 | 0.0 | 3 |
| 🇺🇸 US | 65 | 108.4 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇩🇪 DE | 309 | 46.2 | 3 |
| 🇫🇮 FI | 509 | 24.3 | 3 |
| 🇫🇷 FR | 426 | 59.9 | 1 |
| 🇭🇰 HK | 828 | 15.1 | 1 |
| 🇰🇿 KZ | 2233 | 3.5 | 1 |
| 🇱🇻 LV | 550 | 21.0 | 2 |
| 🇳🇱 NL | 435 | 40.7 | 3 |
| 🇸🇬 SG | 852 | 15.3 | 3 |
| T1 T1 | 1182 | 3.0 | 3 |
| 🇺🇸 US | 108 | 138.4 | 3 |

**Rotating-exit pool:** 3 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇹 AT | 1307 | 0.3 | 3 |
| 🇨🇦 CA | 155 | 91.5 | 3 |
| 🇨🇿 CZ | 377 | 35.3 | 3 |
| 🇩🇪 DE | 370 | 40.4 | 3 |
| 🇪🇪 EE | 2829 | 0.0 | 2 |
| 🇪🇸 ES | 344 | 0.0 | 3 |
| 🇫🇮 FI | 374 | 0.7 | 3 |
| 🇬🇧 GB | 313 | 46.1 | 3 |
| 🇮🇪 IE | 390 | 0.2 | 3 |
| 🇮🇳 IN | 656 | 0.0 | 3 |
| 🇯🇵 JP | 432 | 0.0 | 3 |
| 🇳🇱 NL | 318 | 4.3 | 1 |
| 🇸🇪 SE | 410 | 0.3 | 3 |
| 🇸🇬 SG | 637 | 0.0 | 3 |
| 🇹🇷 TR | 496 | 29.3 | 3 |
| 🇹🇼 TW | 552 | 20.6 | 3 |
| 🇺🇦 UA | 477 | 33.8 | 1 |
| 🇺🇸 US | 72 | 189.1 | 3 |
| 🇿🇦 ZA | 714 | 0.0 | 3 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-05-28T12:17:48Z -->
