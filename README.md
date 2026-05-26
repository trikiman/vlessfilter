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

Rotating exits: `subs/vless/rotating.txt` (1174 configs)

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

Rotating exits: `subs/trojan/rotating.txt` (104 configs)

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
| 🇦🇱 AL | 699 | 0.0 | 3 |
| 🇧🇷 BR | 506 | 28.0 | 3 |
| 🇨🇦 CA | 226 | 0.0 | 3 |
| 🇩🇪 DE | 454 | 38.3 | 3 |
| 🇬🇧 GB | 407 | 0.0 | 3 |
| 🇭🇰 HK | 467 | 26.4 | 3 |
| 🇭🇺 HU | 666 | 0.0 | 3 |
| 🇮🇳 IN | 606 | 23.1 | 3 |
| 🇮🇹 IT | 456 | 0.0 | 3 |
| 🇯🇵 JP | 325 | 42.3 | 3 |
| 🇰🇷 KR | 397 | 22.7 | 3 |
| 🇰🇿 KZ | 2484 | 9.1 | 3 |
| 🇲🇽 MX | 158 | 0.0 | 3 |
| 🇲🇾 MY | 700 | 0.0 | 1 |
| 🇵🇭 PH | 706 | 2.6 | 3 |
| 🇸🇬 SG | 515 | 0.0 | 3 |
| 🇹🇷 TR | 542 | 0.0 | 3 |
| 🇹🇼 TW | 456 | 1.5 | 3 |
| 🇺🇸 US | 23 | 547.9 | 3 |

**Rotating-exit pool:** 0 configs in `subs/vmess/rotating.txt`

## TROJAN — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇩🇪 DE | 397 | 47.5 | 2 |
| 🇫🇮 FI | 431 | 42.9 | 3 |
| 🇫🇷 FR | 352 | 56.7 | 3 |
| 🇭🇰 HK | 977 | 24.6 | 3 |
| 🇯🇵 JP | 641 | 28.0 | 3 |
| 🇰🇿 KZ | 935 | 0.0 | 2 |
| 🇱🇻 LV | 457 | 0.0 | 3 |
| 🇳🇱 NL | 369 | 0.0 | 3 |
| 🇵🇱 PL | 452 | 0.0 | 3 |
| 🇷🇺 RU | 537 | 0.0 | 3 |
| 🇸🇪 SE | 419 | 0.0 | 3 |
| 🇸🇬 SG | 1052 | 0.0 | 3 |
| 🇺🇸 US | 170 | 31.4 | 3 |
| 🇿🇦 ZA | 1121 | 19.2 | 2 |

**Rotating-exit pool:** 104 configs in `subs/trojan/rotating.txt`

## SS — top 3 per country (stable only)

| Country | Top latency (ms) | Median speed (Mbps) | Keys |
|---------|------------------|---------------------|------|
| 🇦🇪 AE | 604 | 0.0 | 3 |
| 🇦🇷 AR | 438 | 0.0 | 3 |
| 🇦🇹 AT | 321 | 0.0 | 3 |
| 🇦🇺 AU | 617 | 0.0 | 3 |
| 🇧🇬 BG | 784 | 31.2 | 3 |
| 🇧🇭 BH | 600 | 0.0 | 3 |
| 🇧🇴 BO | 829 | 0.0 | 3 |
| 🇧🇷 BR | 654 | 21.9 | 3 |
| 🇨🇦 CA | 85 | 154.4 | 3 |
| 🇨🇭 CH | 741 | 0.0 | 3 |
| 🇨🇱 CL | 372 | 0.0 | 3 |
| 🇨🇳 CN | 1212 | 12.4 | 3 |
| 🇨🇾 CY | 472 | 0.0 | 3 |
| 🇩🇪 DE | 277 | 0.0 | 3 |
| 🇪🇨 EC | 304 | 0.0 | 3 |
| 🇪🇪 EE | 377 | 0.6 | 3 |
| 🇪🇬 EG | 531 | 0.0 | 3 |
| 🇪🇸 ES | 398 | 0.0 | 3 |
| 🇫🇮 FI | 346 | 42.2 | 3 |
| 🇬🇧 GB | 290 | 31.7 | 3 |
| 🇭🇰 HK | 654 | 0.0 | 3 |
| 🇭🇷 HR | 322 | 0.0 | 3 |
| 🇮🇩 ID | 743 | 0.0 | 3 |
| 🇮🇪 IE | 419 | 0.5 | 3 |
| 🇮🇳 IN | 606 | 0.0 | 3 |
| 🇮🇹 IT | 344 | 0.0 | 3 |
| 🇯🇵 JP | 520 | 29.8 | 3 |
| 🇰🇭 KH | 893 | 0.0 | 3 |
| 🇰🇷 KR | 602 | 0.0 | 3 |
| 🇱🇹 LT | 359 | 0.0 | 3 |
| 🇱🇻 LV | 356 | 0.0 | 3 |
| 🇲🇰 MK | 394 | 0.0 | 3 |
| 🇲🇹 MT | 564 | 0.0 | 3 |
| 🇳🇬 NG | 945 | 0.0 | 3 |
| 🇳🇱 NL | 307 | 48.8 | 3 |
| 🇵🇦 PA | 258 | 0.0 | 3 |
| 🇵🇪 PE | 492 | 0.0 | 3 |
| 🇵🇰 PK | 793 | 0.0 | 3 |
| 🇵🇷 PR | 259 | 0.0 | 3 |
| 🇵🇹 PT | 329 | 0.0 | 3 |
| 🇷🇴 RO | 385 | 0.0 | 3 |
| 🇷🇺 RU | 577 | 0.0 | 3 |
| 🇸🇬 SG | 687 | 2.1 | 3 |
| T1 T1 | 635 | 4.7 | 3 |
| 🇹🇭 TH | 729 | 0.0 | 3 |
| 🇹🇷 TR | 435 | 35.4 | 3 |
| 🇹🇼 TW | 572 | 0.0 | 3 |
| 🇺🇸 US | 48 | 276.8 | 1 |
| 🇿🇦 ZA | 655 | 0.0 | 3 |

**Rotating-exit pool:** 0 configs in `subs/ss/rotating.txt`

_Generated by [vlessfilter](https://github.com/trikiman/vlessfilter). Source list: `sources.yaml`._

<!-- last-tested: 2026-05-26T01:00:33Z -->
