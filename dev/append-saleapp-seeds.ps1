# Build saleapp seed URIs with PROPER UTF-8 URL encoding for the emoji.
# The Chinese flag 🇨🇳 in saleapp's manual_seeds.py needs to be URL-encoded byte-level
# as %F0%9F%87%A8%F0%9F%87%B3 (UTF-8). PowerShell's EscapeDataString defaults to
# UTF-8 in PS Core but Windows PowerShell may re-encode incorrectly.

# The correct paths (verified from manual_seeds.py source):
$pathPlus  = "/trTelegram%F0%9F%87%A8%F0%9F%87%B3%2B%40WangCai2"
$pathSpace = "/trTelegram%F0%9F%87%A8%F0%9F%87%B3%20%40WangCai2"

# Password url-encoded
# raw: 8r<[9'l6hAO#8ZQi
# After URI-component-encoding: 8r%3C%5B9'l6hAO%238ZQi
# (apostrophe is safe per RFC 3986, only <, [, # need escaping)
$pwdEnc = "8r%3C%5B9'l6hAO%238ZQi"

$sni = "Koma-YT.PAGeS.Dev"

$seeds = @(
    @{ Host = "93.77.177.164";   Port = 443;  Path = $pathPlus;  Tag = "saleapp-1" },
    @{ Host = "150.241.74.98";   Port = 8443; Path = $pathSpace; Tag = "saleapp-2" },
    @{ Host = "212.113.112.236"; Port = 8443; Path = $pathPlus;  Tag = "saleapp-3" },
    @{ Host = "85.193.90.131";   Port = 2053; Path = $pathPlus;  Tag = "saleapp-4" },
    @{ Host = "85.193.91.193";   Port = 8443; Path = $pathPlus;  Tag = "saleapp-5" },
    @{ Host = "91.196.32.171";   Port = 8443; Path = $pathPlus;  Tag = "saleapp-6" }
)

$uris = $seeds | ForEach-Object {
    "trojan://$pwdEnc@$($_.Host):$($_.Port)?security=tls&type=ws&path=$($_.Path)&host=$sni&sni=$sni&fp=chrome&allowInsecure=0#$($_.Tag)"
}

# Strip the bad ones from previous attempt and re-append clean ones
$verifiedFile = "E:\Projects\VlessFilter\subs\verified-russia.txt"
$existing = Get-Content $verifiedFile | Where-Object { $_ -notmatch "saleapp-" }
$merged = @($existing) + @($uris)
$merged | Set-Content $verifiedFile -Encoding UTF8

Write-Host "Cleaned + appended 6 saleapp seeds. Total: $($merged.Count)"
$uris | ForEach-Object { Write-Host "  $_" }
