# Find-keys-that-work-from-Russia.
# Runs on user's PC. Tests against 3 URLs. Keeps only keys passing ALL 3.
# Goal: 10-20 keys with 99%+ confidence they work for the user.
#
# Strategy:
#   1. Pull candidate pool from multiple aggregators (more raw input)
#   2. Test each against ipinfo + google + cf-trace
#   3. Keep only 3/3 passers (this is the 99%-confidence bar)
#   4. Sort by avg latency, take top 20
#   5. Save as out\verified-from-rus.txt

$xkExe = "$env:TEMP\xk\xray-knife.exe"

# Verify our IP is Russian
try {
    $ipinfo = Invoke-RestMethod -Uri "https://ipinfo.io/json" -TimeoutSec 5
    Write-Host ("Network: {0}  Country: {1}" -f $ipinfo.ip, $ipinfo.country)
    if ($ipinfo.country -ne "RU") {
        Write-Host "WARNING: not on Russian IP, results won't be representative" -ForegroundColor Yellow
    }
} catch {}
Write-Host ""

# --- 1. Build candidate pool ---
$sources = @(
    "https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/vless/all.txt",
    "https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/vmess/all.txt",
    "https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/trojan/all.txt",
    "https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/ss/all.txt"
)

$keys = New-Object System.Collections.Generic.HashSet[string]
foreach ($url in $sources) {
    try {
        $body = (Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 30).Content
        foreach ($line in ($body -split "`n")) {
            $line = $line.Trim()
            if ($line -match "^(vless|vmess|trojan|ss)://") { [void]$keys.Add($line) }
        }
    } catch {
        Write-Host "  failed to fetch $url" -ForegroundColor Yellow
    }
}
$candidates = @($keys)
Write-Host ("Candidate pool: {0} unique keys" -f $candidates.Count)
Write-Host ""

# --- 2. Test each against 3 URLs (early-exit on first failure to save time) ---
$urls = @(
    @{ Name = "ipinfo";    Url = "https://ipinfo.io/json" },
    @{ Name = "google";    Url = "https://www.google.com/" },
    @{ Name = "cf-trace";  Url = "https://www.cloudflare.com/cdn-cgi/trace" }
)

function Test-Key($key, $url) {
    $output = & $xkExe http -c $key -u $url -d 5000 -b 2>&1 | Out-String
    if ($output -match '\u2705' -and $output -match "Real Delay:\s*(\d+)\s*ms") {
        return @{ pass = $true; delay = [int]$Matches[1] }
    }
    return @{ pass = $false; delay = -1 }
}

Write-Host "Testing $($candidates.Count) candidates against 3 URLs (early-exit on fail)..."
Write-Host ""

$verified = New-Object System.Collections.ArrayList
$counter = 0
$startTime = Get-Date

foreach ($key in $candidates) {
    $counter++
    if ($counter % 25 -eq 0) {
        $elapsed = ((Get-Date) - $startTime).TotalSeconds
        Write-Host ("  [{0}/{1}] verified={2}, elapsed={3:N0}s" -f $counter, $candidates.Count, $verified.Count, $elapsed)
    }

    $totalDelay = 0
    $allPass = $true
    foreach ($u in $urls) {
        $r = Test-Key $key $u.Url
        if (-not $r.pass) { $allPass = $false; break }
        $totalDelay += $r.delay
    }
    if ($allPass) {
        $avgDelay = [int]($totalDelay / $urls.Count)
        $proto = if ($key -match "^([a-z]+)://") { $Matches[1] } else { "?" }
        [void]$verified.Add(@{ Key = $key; AvgDelay = $avgDelay; Proto = $proto })
        Write-Host ("    [VERIFIED] {0,-7} avg={1}ms" -f $proto, $avgDelay) -ForegroundColor Green
    }
}

# --- 3. Sort by latency, take top 20 ---
$top = $verified | Sort-Object { $_.AvgDelay } | Select-Object -First 20

Write-Host ""
Write-Host "============================================================"
Write-Host ("RESULT: {0} verified keys (3/3 URLs pass), top 20 by latency:" -f $verified.Count)
Write-Host "============================================================"
foreach ($v in $top) {
    $shortKey = $v.Key.Substring(0, [Math]::Min(80, $v.Key.Length))
    Write-Host ("  {0,-7} {1,5}ms  {2}..." -f $v.Proto, $v.AvgDelay, $shortKey)
}

# --- 4. Save ---
$outDir = "E:\Projects\VlessFilter\dev\verified-from-rus"
if (-not (Test-Path $outDir)) { New-Item -ItemType Directory -Path $outDir | Out-Null }
$outPath = Join-Path $outDir ("top20-{0}.txt" -f (Get-Date -Format "yyyyMMdd-HHmmss"))
$top | ForEach-Object { $_.Key } | Set-Content $outPath -Encoding UTF8
Write-Host ""
Write-Host "Saved to: $outPath"
