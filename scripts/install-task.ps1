# scripts/install-task.ps1
#
# Registers a Windows Scheduled Task that runs scripts/refresh-local.ps1
# every 6 hours. No admin privileges required.
#
# After install, the pipeline runs at 04:00, 10:00, 16:00, 22:00 local time
# and pushes fresh subs/<CC>.txt to GitHub.
#
# Usage:
#   powershell -ExecutionPolicy Bypass -File scripts\install-task.ps1
#
# Uninstall:
#   Unregister-ScheduledTask -TaskName "VlessFilter Refresh" -Confirm:$false

$ErrorActionPreference = "Stop"
$repoDir = (Resolve-Path "$PSScriptRoot\..").Path
$scriptPath = Join-Path $repoDir "scripts\refresh-local.ps1"

if (-not (Test-Path $scriptPath)) {
    throw "refresh-local.ps1 not found at $scriptPath"
}

$taskName = "VlessFilter Refresh"

# Remove existing task if present
$existing = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
if ($existing) {
    Write-Host "Removing existing task..."
    Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
}

# Action: run powershell with the refresh script
$action = New-ScheduledTaskAction `
    -Execute "powershell.exe" `
    -Argument "-ExecutionPolicy Bypass -WindowStyle Hidden -File `"$scriptPath`"" `
    -WorkingDirectory $repoDir

# Trigger: 4 daily runs at 04:00, 10:00, 16:00, 22:00 local time.
# (Cleaner than -RepetitionInterval/-RepetitionDuration which has XML
# schema limits on the duration value.)
$triggers = @(
    New-ScheduledTaskTrigger -Daily -At "04:00"
    New-ScheduledTaskTrigger -Daily -At "10:00"
    New-ScheduledTaskTrigger -Daily -At "16:00"
    New-ScheduledTaskTrigger -Daily -At "22:00"
)
$startTime = (Get-Date).Date.AddHours(4)
if ($startTime -lt (Get-Date)) { $startTime = $startTime.AddHours(6) }

# Settings: don't run if on battery, allow start on demand, etc.
$settings = New-ScheduledTaskSettingsSet `
    -AllowStartIfOnBatteries `
    -DontStopIfGoingOnBatteries `
    -StartWhenAvailable `
    -DontStopOnIdleEnd `
    -ExecutionTimeLimit (New-TimeSpan -Hours 1) `
    -MultipleInstances IgnoreNew

# Principal: current user, no elevated privileges
$principal = New-ScheduledTaskPrincipal `
    -UserId "$env:USERDOMAIN\$env:USERNAME" `
    -LogonType Interactive `
    -RunLevel Limited

Register-ScheduledTask `
    -TaskName $taskName `
    -Description "VlessFilter -- scheduled refresh of VLESS proxy subscription files. Runs at 04:00, 10:00, 16:00, 22:00. Pushes to github.com/trikiman/vlessfilter." `
    -Action $action `
    -Trigger $triggers `
    -Settings $settings `
    -Principal $principal | Out-Null

Write-Host ""
Write-Host "=== Scheduled Task installed ==="
Write-Host ("Name:  " + $taskName)
Write-Host ("Runs:  every 6 hours starting " + $startTime.ToString("yyyy-MM-dd HH:mm"))
Write-Host ("Logs:  " + (Join-Path $repoDir "scripts\logs"))
Write-Host ""
Write-Host "Run once manually now:"
Write-Host "  Start-ScheduledTask -TaskName 'VlessFilter Refresh'"
Write-Host ""
Write-Host "Uninstall later:"
$confirmZero = [char]36 + "false"
Write-Host ("  Unregister-ScheduledTask -TaskName 'VlessFilter Refresh' -Confirm:" + $confirmZero)
