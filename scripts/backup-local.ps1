# scripts/backup-local.ps1 - point-in-time snapshot of vlessfilter state.
#
# Backs up (in order of preciousness):
#   1. xray-knife.db       - the test-result pool. IRREPLACEABLE.
#   2. all-results.csv     - per-run summary, regenerable from DB.
#   3. subs/               - currently published top-3, replaceable.
#   4. sources.yaml        - source declarations.
#   5. .git via git bundle - full history, durable to GitHub deletion.
#
# Output: <BackupRoot>\<timestamp>\
# Usage : powershell -ExecutionPolicy Bypass -File scripts\backup-local.ps1

param(
    [string]$BackupRoot = "E:\Backups\vlessfilter",
    [string]$RepoRoot   = (Resolve-Path "$PSScriptRoot\..").Path,
    [string]$XrayDB     = "$env:USERPROFILE\.xray-knife\xray-knife.db"
)

$ErrorActionPreference = "Stop"
$ts = Get-Date -Format "yyyyMMdd-HHmmss"
$dest = Join-Path $BackupRoot $ts
New-Item -ItemType Directory -Path $dest -Force | Out-Null

Write-Host "Backup destination: $dest"
Write-Host ""

# 1. xray-knife.db
if (Test-Path $XrayDB) {
    $size = (Get-Item $XrayDB).Length
    $sizeMB = [Math]::Round($size / 1MB, 1)
    Write-Host "  copying xray-knife.db ($sizeMB MB)..."
    Copy-Item $XrayDB (Join-Path $dest "xray-knife.db")
} else {
    Write-Host "  WARN: $XrayDB not found - DB pool is on a different machine"
}

# 2. all-results.csv
$arc = Join-Path $RepoRoot "all-results.csv"
if (Test-Path $arc) {
    Copy-Item $arc (Join-Path $dest "all-results.csv")
    Write-Host "  copied all-results.csv"
}

# 3. subs/
$subs = Join-Path $RepoRoot "subs"
if (Test-Path $subs) {
    Compress-Archive -Path "$subs\*" -DestinationPath (Join-Path $dest "subs.zip") -Force
    Write-Host "  zipped subs/ -> subs.zip"
}

# 4. sources.yaml + sources.txt
foreach ($f in @("sources.yaml", "sources.txt")) {
    $p = Join-Path $RepoRoot $f
    if (Test-Path $p) {
        Copy-Item $p (Join-Path $dest $f)
        Write-Host "  copied $f"
    }
}

# 5. git bundle
Push-Location $RepoRoot
try {
    git bundle create (Join-Path $dest "vlessfilter.bundle") --all 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        $bsize = (Get-Item (Join-Path $dest "vlessfilter.bundle")).Length
        $bMB = [Math]::Round($bsize / 1MB, 1)
        Write-Host "  git bundle ($bMB MB) - restore: git clone vlessfilter.bundle vlessfilter"
    } else {
        Write-Host "  WARN: git bundle failed"
    }
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "Backup complete: $dest"
Get-ChildItem $dest | ForEach-Object {
    $mb = [Math]::Round($_.Length / 1MB, 2)
    $line = "  {0,-30} {1,8} MB" -f $_.Name, $mb
    Write-Host $line
}
