$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$backendDir = (Resolve-Path (Join-Path $root "..\backend")).Path
$frontendDir = (Resolve-Path (Join-Path $root "..\frontend")).Path

# Vite optimize cache can become inconsistent after repo layout/link changes.
# Start from a clean cache and force re-optimization for stable dev startup.
foreach ($cachePath in @(
    (Join-Path $frontendDir "node_modules/.vite"),
    (Join-Path $frontendDir ".vite")
)) {
    if (Test-Path $cachePath) {
        Remove-Item -Recurse -Force $cachePath
    }
}

$backend = Start-Process -FilePath "go" -ArgumentList @("run", "./cmd/server") -WorkingDirectory $backendDir -NoNewWindow -PassThru
$frontend = Start-Process -FilePath "npm.cmd" -ArgumentList @("run", "dev", "--", "--host", "0.0.0.0", "--port", "5173", "--force") -WorkingDirectory $frontendDir -NoNewWindow -PassThru

Write-Host "[dev] backend pid: $($backend.Id)"
Write-Host "[dev] frontend pid: $($frontend.Id)"
Write-Host "[dev] press Ctrl+C to stop both"

try {
    Wait-Process -Id @($backend.Id, $frontend.Id)
}
finally {
    foreach ($proc in @($backend, $frontend)) {
        if (-not $proc.HasExited) {
            Stop-Process -Id $proc.Id -Force
        }
    }
}
