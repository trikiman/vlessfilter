# Re-test the keys currently in verified-russia.txt with the new stability filter.
# Drops any that fail 2-pass stability test or exceed 1500ms.
# This is the key SURVIVOR step — keep only those still alive AND stable.

$xkExe = "$env:TEMP\xk\xray-knife.exe"
$verifiedFile = "E:\Projects\VlessFilter\subs\verified-russia.txt"

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

function Test-KeyStable($key, $url) {
    $r1 = Test-Key $key $url
    if (-not $r1.pass) { return @{ pass = $false; delay = -1 } }
    return @{ pass = $true; delay = $r1.delay }
}

$keys = Get-Content $verifiedFile
Write-Host "Re-testing $($keys.Count) keys with 2-pass stability filter..."
Write-Host ""

$survivors = New-Object System.Collections.ArrayList
foreach ($key in $keys) {
    if ($key -notmatch "^[a-z]+://") { continue }
    $hostPort = if ($key -match "@([^:?/#]+):(\d+)") { "$($Matches[1]):$($Matches[2])" } else { "?" }

    $totalDelay = 0
    $allPass = $true
    foreach ($u in $urls) {
        $r = Test-KeyStable $key $u.Url
        if (-not $r.pass) { $allPass = $false; break }
        $totalDelay += $r.delay
    }

    if ($allPass) {
        $avg = [int]($totalDelay / $urls.Count)
        if ($avg -le 1500) {
            [void]$survivors.Add(@{ Key = $key; Avg = $avg; HP = $hostPort })
            Write-Host ("  [SURVIVE] {0}  avg={1}ms" -f $hostPort, $avg) -ForegroundColor Green
        } else {
            Write-Host ("  [DROP-SLOW] {0}  avg={1}ms" -f $hostPort, $avg) -ForegroundColor Yellow
        }
    } else {
        Write-Host ("  [DEAD]    {0}" -f $hostPort) -ForegroundColor Red
    }
}

Write-Host ""
Write-Host ("==== {0} / {1} survived ====" -f $survivors.Count, $keys.Count)
$ranked = $survivors | Sort-Object { $_.Avg }
$ranked | ForEach-Object { $_.Key } | Set-Content $verifiedFile -Encoding UTF8
$ranked | ForEach-Object { Write-Host ("  {0,5}ms  {1}" -f $_.Avg, $_.HP) }
