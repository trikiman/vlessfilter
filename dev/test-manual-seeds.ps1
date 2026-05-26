# Test the 6 manual trojan seeds from saleapp's manual_seeds.py
# These were operator-supplied 2026-05-23 as "always-works fallback floor"
# for the VkusVill scraper. If they still work, they go straight into our
# verified-russia.txt as proven good keys.

Add-Type -AssemblyName System.Web

$xkExe = "$env:TEMP\xk\xray-knife.exe"

$password = "8r<[9'l6hAO#8ZQi"
$sni = "Koma-YT.PAGeS.Dev"
$wsPath = "/trTelegram🇨🇳+@WangCai2"
$wsPathSpace = "/trTelegram🇨🇳 @WangCai2"  # 2nd seed has space instead of +

$seeds = @(
    @{ Host = "93.77.177.164";   Port = 443;  Path = $wsPath },
    @{ Host = "150.241.74.98";   Port = 8443; Path = $wsPathSpace },
    @{ Host = "212.113.112.236"; Port = 8443; Path = $wsPath },
    @{ Host = "85.193.90.131";   Port = 2053; Path = $wsPath },
    @{ Host = "85.193.91.193";   Port = 8443; Path = $wsPath },
    @{ Host = "91.196.32.171";   Port = 8443; Path = $wsPath }
)

# Build trojan URIs
$keys = @()
foreach ($s in $seeds) {
    $encPath = [System.Web.HttpUtility]::UrlEncode($s.Path)
    $encPwd  = [System.Uri]::EscapeDataString($password)
    $uri = "trojan://$encPwd@$($s.Host):$($s.Port)?security=tls&type=ws&path=$encPath&host=$sni&sni=$sni&fp=chrome&allowInsecure=0#manual-$($s.Host)"
    $keys += $uri
}

Add-Type -AssemblyName System.Web

$urls = @(
    @{ Name = "google";   Url = "https://www.google.com/" },
    @{ Name = "cf-trace"; Url = "https://www.cloudflare.com/cdn-cgi/trace" }
)

Write-Host "Testing 6 manual trojan seeds against 2 URLs..."
$verified = New-Object System.Collections.ArrayList
foreach ($k in $keys) {
    $hostPort = if ($k -match "@([^:]+):(\d+)") { "$($Matches[1]):$($Matches[2])" } else { "?" }
    Write-Host ("--- {0} ---" -f $hostPort)
    $allPass = $true
    $delays = 0
    foreach ($u in $urls) {
        $output = & $xkExe http -c $k -u $u.Url -d 8000 -b 2>&1 | Out-String
        if ($output -match '\u2705' -and $output -match "Real Delay:\s*(\d+)\s*ms") {
            $d = [int]$Matches[1]
            $delays += $d
            Write-Host ("  {0,-9} PASS {1}ms" -f $u.Name, $d) -ForegroundColor Green
        } else {
            $allPass = $false
            $errLine = ($output -split "`n" | Where-Object {$_ -match "Failed|Error|❌|signal"} | Select-Object -First 1)
            Write-Host ("  {0,-9} FAIL ({1})" -f $u.Name, $errLine) -ForegroundColor Red
            break
        }
    }
    if ($allPass) {
        $avg = [int]($delays / $urls.Count)
        [void]$verified.Add(@{ Key = $k; AvgDelay = $avg })
    }
}

Write-Host ""
Write-Host ("==== {0} / 6 seeds verified ====" -f $verified.Count)
$verified | Sort-Object {$_.AvgDelay} | ForEach-Object {
    $shortKey = $_.Key.Substring(0, [Math]::Min(80, $_.Key.Length))
    Write-Host ("  {0,5}ms  {1}..." -f $_.AvgDelay, $shortKey)
}

# Save full URIs
if ($verified.Count -gt 0) {
    $outPath = "E:\Projects\VlessFilter\dev\verified-from-rus\manual-seeds-verified.txt"
    $verified | Sort-Object {$_.AvgDelay} | ForEach-Object { $_.Key } | Set-Content $outPath -Encoding UTF8
    Write-Host ""
    Write-Host "Saved verified seeds to: $outPath"
}
