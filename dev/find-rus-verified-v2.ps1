# Smarter RU-IP verifier
# - Pulls broader candidate pool from raw aggregator URLs
# - Dedupes by fingerprint (host:port + uuid/password)
# - Tests with 1st-URL early-exit (drops ~95% of dead keys quickly)
# - Caps total runtime at 5 minutes
# - Stops when N verified found
#
# Goal: 10-20 verified keys with 99% pass-rate from the user's RU IP.

$ErrorActionPreference = 'Continue'
$xkExe = "$env:TEMP\xk\xray-knife.exe"

# Tunables
$TARGET_VERIFIED = 25
$TIME_BUDGET_SEC = 300   # 5 minutes
$MAX_CANDIDATES  = 800   # cap candidate pool size

# --- 1. Build BROAD candidate pool from raw aggregator sources ---
# Sources curated by saleapp/vless/sources.py - empirically validated for RU exits
$sources = @(
    # Already-published lists (known-good 2.5% pass rate)
    "https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/all.txt",
    # IGARECK repo - "vpn-configs-for-russia" - PRIMARY saleapp source
    "https://raw.githubusercontent.com/igareck/vpn-configs-for-russia/main/BLACK_VLESS_RUS.txt",
    "https://raw.githubusercontent.com/igareck/vpn-configs-for-russia/main/BLACK_VLESS_RUS_mobile.txt",
    "https://raw.githubusercontent.com/igareck/vpn-configs-for-russia/main/Vless-Reality-White-Lists-Rus-Mobile.txt",
    "https://raw.githubusercontent.com/igareck/vpn-configs-for-russia/main/Vless-Reality-White-Lists-Rus-Mobile-2.txt",
    "https://raw.githubusercontent.com/igareck/vpn-configs-for-russia/main/WHITE-CIDR-RU-all.txt",
    "https://raw.githubusercontent.com/igareck/vpn-configs-for-russia/main/WHITE-CIDR-RU-checked.txt",
    "https://raw.githubusercontent.com/igareck/vpn-configs-for-russia/main/WHITE-SNI-RU-all.txt",
    # kort0881/vpn-vless-configs-russia - "595 RU-flagged"
    "https://raw.githubusercontent.com/kort0881/vpn-vless-configs-russia/main/subscriptions/sni_filtered.txt",
    # SoliSpirit Russia - 15-min refresh
    "https://raw.githubusercontent.com/SoliSpirit/v2ray-configs/main/Countries/Russia.txt"
)

function Get-Fingerprint($key) {
    if ($key -match "^vless://([^@]+)@([^:?/#]+):(\d+)") {
        return "vless:$($Matches[2].ToLower()):$($Matches[3]):$($Matches[1])"
    }
    if ($key -match "^trojan://([^@]+)@([^:?/#]+):(\d+)") {
        return "trojan:$($Matches[2].ToLower()):$($Matches[3])"
    }
    if ($key -match "^ss://([^@]+)@([^:?/#]+):(\d+)") {
        return "ss:$($Matches[2].ToLower()):$($Matches[3])"
    }
    if ($key -match "^vmess://(.+?)(\#|$)") {
        try {
            $b64 = $Matches[1].Replace('-', '+').Replace('_', '/')
            while (($b64.Length % 4) -ne 0) { $b64 += '=' }
            $json = [System.Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($b64))
            if ($json -match '"add"\s*:\s*"([^"]+)".*?"port"\s*:\s*"?(\d+)"?.*?"id"\s*:\s*"([^"]+)"') {
                return "vmess:$($Matches[1].ToLower()):$($Matches[2]):$($Matches[3])"
            }
        } catch {}
        return "vmess:$($Matches[1].Substring(0,[Math]::Min(40,$Matches[1].Length)))"
    }
    return $key
}

# Verify our IP
try {
    $ipinfo = Invoke-RestMethod -Uri "https://ipinfo.io/json" -TimeoutSec 5
    Write-Host ("Network: {0}  Country: {1}  ISP: {2}" -f $ipinfo.ip, $ipinfo.country, $ipinfo.org)
} catch {}
Write-Host ""

Write-Host "Fetching candidate pool from $($sources.Count) sources..."
$seen = @{}
$candidates = New-Object System.Collections.ArrayList
foreach ($url in $sources) {
    try {
        $body = (Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 20).Content
        # Auto-decode base64 if needed
        $first = ($body.Substring(0, [Math]::Min(200, $body.Length))).Trim()
        if ($first -match '^[A-Za-z0-9+/=\s]+$' -and $first.Length -gt 50) {
            try {
                $decoded = [System.Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($first.Replace("`n","").Replace("`r","")))
                if ($decoded -match "vless://|vmess://|trojan://|ss://") { $body = $decoded }
            } catch {}
        }
        $added = 0
        foreach ($line in ($body -split "`n")) {
            $line = $line.Trim()
            if ($line -notmatch "^(vless|vmess|trojan|ss)://") { continue }
            $fp = Get-Fingerprint $line
            if ($seen.ContainsKey($fp)) { continue }
            $seen[$fp] = $true
            [void]$candidates.Add($line)
            $added++
            if ($candidates.Count -ge $MAX_CANDIDATES) { break }
        }
        Write-Host ("  +{0,5} from {1}" -f $added, ([uri]$url).Host)
        if ($candidates.Count -ge $MAX_CANDIDATES) { Write-Host "  cap reached"; break }
    } catch {
        Write-Host "  failed: $url"
    }
}
Write-Host ("Total unique candidates (post-dedup): {0}" -f $candidates.Count)
Write-Host ""

# --- 2. Shuffle so we don't always test the same source first ---
$candidates = $candidates | Sort-Object { Get-Random }

# --- 3. Test with early-exit, stop when target reached or budget expires ---
$urls = @(
    @{ Name = "ipinfo";    Url = "https://ipinfo.io/json" },
    @{ Name = "google";    Url = "https://www.google.com/" },
    @{ Name = "cf-trace";  Url = "https://www.cloudflare.com/cdn-cgi/trace" }
)

function Test-Key($key, $url) {
    $output = & $xkExe http -c $key -u $url -d 3000 -b 2>&1 | Out-String
    if ($output -match '\u2705' -and $output -match "Real Delay:\s*(\d+)\s*ms") {
        return @{ pass = $true; delay = [int]$Matches[1] }
    }
    return @{ pass = $false; delay = -1 }
}

# Stability test: require 2 consecutive passes per URL (catches flaky keys)
function Test-KeyStable($key, $url) {
    $r1 = Test-Key $key $url
    if (-not $r1.pass) { return @{ pass = $false; delay = -1 } }
    Start-Sleep -Milliseconds 200
    $r2 = Test-Key $key $url
    if (-not $r2.pass) { return @{ pass = $false; delay = -1 } }
    # Reject high variance (>50% delta) - sign of flakiness
    $maxDelay = [Math]::Max($r1.delay, $r2.delay)
    $minDelay = [Math]::Min($r1.delay, $r2.delay)
    if ($maxDelay -gt ($minDelay * 1.5) -and $maxDelay -gt 500) {
        return @{ pass = $false; delay = -1 }
    }
    return @{ pass = $true; delay = [int](($r1.delay + $r2.delay) / 2) }
}

Write-Host "Testing... (target=$TARGET_VERIFIED verified, budget=$TIME_BUDGET_SEC sec)"
$verified = New-Object System.Collections.ArrayList
$counter = 0
$startTime = Get-Date

foreach ($key in $candidates) {
    $counter++
    $elapsed = ((Get-Date) - $startTime).TotalSeconds
    if ($elapsed -gt $TIME_BUDGET_SEC) { Write-Host "  [BUDGET]"; break }
    if ($verified.Count -ge $TARGET_VERIFIED) { Write-Host "  [TARGET REACHED]"; break }

    $totalDelay = 0
    $allPass = $true
    foreach ($u in $urls) {
        $r = Test-KeyStable $key $u.Url
        if (-not $r.pass) { $allPass = $false; break }
        $totalDelay += $r.delay
    }
    if ($allPass) {
        $avgDelay = [int]($totalDelay / $urls.Count)
        # Reject high-latency keys (>1500ms avg = unreliable for real use)
        if ($avgDelay -gt 1500) { continue }
        $proto = if ($key -match "^([a-z]+)://") { $Matches[1] } else { "?" }
        [void]$verified.Add(@{ Key = $key; AvgDelay = $avgDelay; Proto = $proto })
        Write-Host ("  [+] [{0}/{1}] {2,-7} avg={3,5}ms  ({4:N0}s)" -f $verified.Count, $TARGET_VERIFIED, $proto, $avgDelay, $elapsed) -ForegroundColor Green
    }
    if ($counter % 50 -eq 0 -and $verified.Count -lt $TARGET_VERIFIED) {
        Write-Host ("  ...tested {0}, verified={1}, elapsed={2:N0}s" -f $counter, $verified.Count, $elapsed)
    }
}

# --- 4. Sort + display + save ---
$top = $verified | Sort-Object { $_.AvgDelay }

Write-Host ""
Write-Host ("======= RESULT: {0} verified in {1:N0}s (tested {2}) =======" -f $verified.Count, ((Get-Date)-$startTime).TotalSeconds, $counter)
$idx = 0
foreach ($v in $top) {
    $idx++
    $shortKey = $v.Key.Substring(0, [Math]::Min(90, $v.Key.Length))
    Write-Host ("  {0,2}. {1,-7} {2,5}ms  {3}..." -f $idx, $v.Proto, $v.AvgDelay, $shortKey)
}

# Save (only if we found at least one — never wipe the existing list)
if ($top.Count -gt 0) {
    $outPath = "E:\Projects\VlessFilter\subs\verified-russia.txt"

    # Merge with existing verified-russia.txt - dedup by host:port
    $existing = @()
    if (Test-Path $outPath) { $existing = @(Get-Content $outPath) }
    $allKeys = $existing + ($top | ForEach-Object { $_.Key })
    $seen2 = @{}
    $merged = @()
    foreach ($k in $allKeys) {
        if ($k -match "^[a-z]+://[^@]+@([^:?/#]+):(\d+)") {
            $fp2 = "$($Matches[1].ToLower()):$($Matches[2])"
            if (-not $seen2.ContainsKey($fp2)) { $seen2[$fp2] = $true; $merged += $k }
        }
    }
    $merged | Set-Content $outPath -Encoding UTF8

    $ts = Get-Date -Format "yyyyMMdd-HHmmss"
    $top | ForEach-Object { $_.Key } | Set-Content "E:\Projects\VlessFilter\dev\verified-from-rus\top-$ts.txt" -Encoding UTF8

    Write-Host ""
    Write-Host ("Updated: subs\verified-russia.txt  (was {0} keys, now {1} unique merged)" -f $existing.Count, $merged.Count)
} else {
    Write-Host ""
    Write-Host "No verified keys this run - existing subs\verified-russia.txt left UNCHANGED" -ForegroundColor Yellow
}
