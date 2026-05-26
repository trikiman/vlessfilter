Add-Type -AssemblyName System.Web
$password = "8r<[9'l6hAO#8ZQi"
$sni = "Koma-YT.PAGeS.Dev"
$wsPath = "/trTelegram🇨🇳+@WangCai2"
$wsPathSpace = "/trTelegram🇨🇳 @WangCai2"
$encPwd  = [System.Uri]::EscapeDataString($password)
$encPath = [System.Uri]::EscapeDataString($wsPath).Replace("%2F","/")  # keep leading slash readable
$encPathSpace = [System.Uri]::EscapeDataString($wsPathSpace).Replace("%2F","/")

$seeds = @(
    @{ Host = "93.77.177.164";   Port = 443;  Path = $encPath;       Tag = "saleapp-1" },
    @{ Host = "150.241.74.98";   Port = 8443; Path = $encPathSpace;  Tag = "saleapp-2" },
    @{ Host = "212.113.112.236"; Port = 8443; Path = $encPath;       Tag = "saleapp-3" },
    @{ Host = "85.193.90.131";   Port = 2053; Path = $encPath;       Tag = "saleapp-4" },
    @{ Host = "85.193.91.193";   Port = 8443; Path = $encPath;       Tag = "saleapp-5" },
    @{ Host = "91.196.32.171";   Port = 8443; Path = $encPath;       Tag = "saleapp-6" }
)

$uris = $seeds | ForEach-Object {
    "trojan://$encPwd@$($_.Host):$($_.Port)?security=tls&type=ws&path=$($_.Path)&host=$sni&sni=$sni&fp=chrome&allowInsecure=0#$($_.Tag)"
}

# Append to verified-russia.txt
$existing = @(Get-Content "E:\Projects\VlessFilter\subs\verified-russia.txt")
$merged = $existing + $uris
$merged | Set-Content "E:\Projects\VlessFilter\subs\verified-russia.txt" -Encoding UTF8

Write-Host "Appended 6 saleapp manual seeds. Total: $($merged.Count)"
$uris | ForEach-Object { Write-Host "  $_" }
