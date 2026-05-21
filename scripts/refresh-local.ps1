# scripts/refresh-local.ps1
#
# Runs the VlessFilter pipeline locally via WSL, then commits + pushes
# fresh subs/<CC>.txt and subs/all.txt to GitHub. No GitHub Actions
# involvement, no billing required. Designed for Windows Task Scheduler.
#
# Logs each run to scripts/logs/refresh-YYYYMMDD-HHmmss.log
#
# Usage (manual): powershell -File scripts\refresh-local.ps1
# Usage (scheduled): see scripts/install-task.ps1

param(
    [int]$BudgetMin = 30,
    [int]$Threads1 = 1000,
    [string]$RepoDir = "E:\Projects\VlessFilter"
)

$ErrorActionPreference = "Continue"
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$logDir = Join-Path $RepoDir "scripts\logs"
New-Item -ItemType Directory -Path $logDir -Force | Out-Null
$logFile = Join-Path $logDir "refresh-$timestamp.log"

function Log {
    param([string]$msg)
    $line = "[{0}] {1}" -f (Get-Date -Format "HH:mm:ss"), $msg
    Add-Content -Path $logFile -Value $line
    Write-Host $line
}

Log "=== refresh-local START ==="
Log "RepoDir: $RepoDir"
Log "Budget: ${BudgetMin}min, Threads1: $Threads1"

Set-Location $RepoDir

# 1. Run the pipeline via WSL
Log "Running pipeline via WSL..."
$bash = @"
#!/bin/bash
export PATH=`$HOME/.local/go/bin:`$HOME/go/bin:`$PATH
export VLESSFILTER_QUIET=1
cd /mnt/e/Projects/VlessFilter
./bin/vlessfilter run --threads1 $Threads1 --budget-min $BudgetMin 2>&1 | tail -40
"@
$bashFile = "$env:TEMP\vf-refresh-$timestamp.sh"
$bash -replace "`r`n", "`n" | Out-File -Encoding ascii -FilePath $bashFile
$wslOut = wsl -- bash $bashFile 2>&1
$wslOut | ForEach-Object { Add-Content -Path $logFile -Value $_ }
Remove-Item $bashFile -ErrorAction SilentlyContinue

# 2. Force-add outputs (gitignore lists them but publish path needs them)
Log "Staging outputs..."
git add -f subs/ README.md all-results.csv raw/dead.txt 2>&1 | Out-Null

# 3. Detect changes and commit
$changes = git status --porcelain
if (-not $changes) {
    Log "No changes to commit. Done."
    Log "=== refresh-local END (no-op) ==="
    exit 0
}

$count = ($changes | Measure-Object).Count
Log "Changed files: $count"
$msg = "results: scheduled refresh $timestamp"
git commit -m $msg 2>&1 | Out-Null
Log "Committed."

# 4. Push using cached credentials
Log "Pushing to origin/main..."
$pushOut = git push origin main 2>&1
$pushOut | ForEach-Object { Add-Content -Path $logFile -Value $_ }
$exit = $LASTEXITCODE
if ($exit -ne 0) {
    Log "Push failed (exit $exit) — keeping commit local. Will retry on next run."
} else {
    Log "Pushed successfully."
}

Log "=== refresh-local END ==="
exit $exit
