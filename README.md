# VlessFilter Results

Auto-curated top 3 fastest proxy keys per country, refreshed automatically. Multi-protocol: VLESS / VMess / Trojan / Shadowsocks.

## How to use

Pick the protocol your client supports best. Each has its own subscription URLs:

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

## VMESS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇨🇦 CA | 88 | 115.6 | 3 |
| 🇩🇪 DE | 1215 | 2.6 | 2 |
| 🇪🇸 ES | 334 | 48.3 | 3 |
| 🇬🇧 GB | 345 | 0.0 | 3 |
| 🇭🇰 HK | 626 | 0.0 | 3 |
| 🇮🇳 IN | 837 | 0.0 | 3 |
| 🇮🇹 IT | 406 | 43.7 | 3 |
| 🇯🇵 JP | 501 | 27.8 | 3 |
| 🇰🇷 KR | 750 | 21.3 | 3 |
| 🇰🇿 KZ | 819 | 0.0 | 2 |
| 🇲🇾 MY | 706 | 19.0 | 3 |
| 🇳🇱 NL | 271 | 0.0 | 3 |
| 🇸🇬 SG | 694 | 22.3 | 3 |
| 🇹🇷 TR | 406 | 36.9 | 3 |
| 🇹🇼 TW | 631 | 7.1 | 3 |
| 🇺🇸 US | 35 | 167.4 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇪 AE | 733 | 0.0 | 3 |
| 🇦🇺 AU | 775 | 23.1 | 3 |
| 🇧🇬 BG | 687 | 39.7 | 3 |
| 🇧🇷 BR | 555 | 0.0 | 3 |
| 🇨🇭 CH | 616 | 38.8 | 3 |
| 🇩🇪 DE | 500 | 64.0 | 2 |
| 🇪🇬 EG | 618 | 28.4 | 3 |
| 🇪🇸 ES | 364 | 50.8 | 3 |
| 🇫🇮 FI | 530 | 51.2 | 3 |
| 🇫🇷 FR | 359 | 71.9 | 3 |
| 🇬🇧 GB | 374 | 49.4 | 3 |
| 🇭🇰 HK | 944 | 19.1 | 3 |
| 🇭🇺 HU | 454 | 36.1 | 3 |
| 🇮🇩 ID | 950 | 2.1 | 3 |
| 🇮🇱 IL | 667 | 39.3 | 3 |
| 🇮🇳 IN | 2213 | 0.0 | 3 |
| 🇮🇹 IT | 467 | 44.9 | 3 |
| 🇯🇵 JP | 662 | 0.0 | 3 |
| 🇰🇷 KR | 789 | 23.2 | 3 |
| 🇰🇿 KZ | 999 | 1.1 | 3 |
| 🇱🇹 LT | 640 | 40.1 | 3 |
| 🇱🇻 LV | 605 | 29.2 | 3 |
| 🇲🇽 MX | 247 | 17.0 | 3 |
| 🇲🇾 MY | 927 | 14.4 | 3 |
| 🇳🇱 NL | 395 | 65.0 | 3 |
| 🇳🇴 NO | 556 | 41.0 | 3 |
| 🇵🇰 PK | 734 | 8.0 | 3 |
| 🇵🇱 PL | 459 | 55.2 | 3 |
| 🇷🇺 RU | 674 | 39.2 | 3 |
| 🇸🇦 SA | 771 | 13.5 | 3 |
| 🇸🇪 SE | 458 | 53.3 | 3 |
| 🇸🇬 SG | 918 | 19.4 | 3 |
| 🇹🇷 TR | 751 | 38.3 | 3 |
| 🇹🇼 TW | 1147 | 0.2 | 1 |
| 🇺🇸 US | 146 | 153.8 | 2 |
| 🇻🇳 VN | 1268 | 6.4 | 3 |
| 🇿🇦 ZA | 875 | 5.1 | 3 |

**Rotating-exit pool:** 29 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇪 AE | 827 | 7.1 | 3 |
| 🇦🇺 AU | 464 | 32.7 | 3 |
| 🇧🇬 BG | 609 | 0.0 | 3 |
| 🇧🇷 BR | 511 | 0.0 | 3 |
| 🇨🇦 CA | 80 | 171.3 | 3 |
| 🇨🇴 CO | 364 | 24.1 | 3 |
| 🇨🇿 CZ | 615 | 17.3 | 3 |
| 🇩🇪 DE | 468 | 30.1 | 3 |
| 🇪🇪 EE | 945 | 0.0 | 3 |
| 🇪🇸 ES | 509 | 0.0 | 3 |
| 🇫🇮 FI | 515 | 21.1 | 3 |
| 🇫🇷 FR | 455 | 0.0 | 1 |
| 🇬🇧 GB | 427 | 33.7 | 3 |
| 🇮🇩 ID | 549 | 11.1 | 3 |
| 🇮🇪 IE | 397 | 0.0 | 3 |
| 🇮🇳 IN | 711 | 0.0 | 3 |
| 🇮🇹 IT | 561 | 27.5 | 3 |
| 🇯🇵 JP | 322 | 0.0 | 3 |
| 🇰🇷 KR | 435 | 0.0 | 3 |
| 🇲🇾 MY | 566 | 26.4 | 1 |
| 🇳🇱 NL | 447 | 31.0 | 3 |
| 🇳🇴 NO | 477 | 0.0 | 3 |
| 🇵🇭 PH | 605 | 10.3 | 3 |
| 🇵🇱 PL | 566 | 27.2 | 3 |
| 🇷🇴 RO | 641 | 2.5 | 2 |
| 🇸🇦 SA | 842 | 15.2 | 3 |
| 🇸🇬 SG | 512 | 4.0 | 3 |
| 🇹🇷 TR | 590 | 25.2 | 3 |
| 🇹🇼 TW | 423 | 20.0 | 3 |
| 🇺🇸 US | 238 | 93.1 | 3 |
| 🇿🇦 ZA | 816 | 17.1 | 3 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-07-05T02:33:57Z -->
