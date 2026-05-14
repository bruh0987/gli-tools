param(
    [string]$InstallDir = "$HOME\.gli\bin"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$output = Join-Path $InstallDir "gli.exe"
go build -trimpath -ldflags="-s -w" -o $output (Join-Path $repoRoot "cmd\gli")

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$paths = @($userPath -split ';' | Where-Object { $_ -ne "" })

if ($paths -notcontains $InstallDir) {
    $newPath = if ($userPath) { "$userPath;$InstallDir" } else { $InstallDir }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "Installed gli to $output"
    Write-Host "Added $InstallDir to your user PATH. Open a new terminal before running gli."
} else {
    Write-Host "Installed gli to $output"
    Write-Host "$InstallDir is already on your user PATH."
}
