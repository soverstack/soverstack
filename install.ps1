# Soverstack Launcher Installer for Windows
# Usage: irm https://raw.githubusercontent.com/soverstack/cli-launcher/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "soverstack/cli-launcher"
$installDir = "$env:LOCALAPPDATA\Soverstack"
$binaryName = "soverstack.exe"

Write-Host "Installing Soverstack Launcher..." -ForegroundColor Cyan

# Get latest release (including prereleases)
$releases = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases?per_page=1" -UseBasicParsing
$release = $releases[0]
$tag = $release.tag_name
$version = $tag.TrimStart("v")

Write-Host "Latest version: $version"

# Find the windows zip asset
$asset = $release.assets | Where-Object { $_.name -match "windows-amd64\.zip$" } | Select-Object -First 1

if (-not $asset) {
    Write-Host "Error: Could not find Windows binary in release $tag" -ForegroundColor Red
    exit 1
}

# Create install directory
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# Download and extract
$zipPath = "$env:TEMP\soverstack.zip"
Write-Host "Downloading $($asset.name)..."
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $zipPath -UseBasicParsing

Expand-Archive -Path $zipPath -DestinationPath "$env:TEMP\soverstack-extract" -Force
Copy-Item "$env:TEMP\soverstack-extract\soverstack.exe" "$installDir\$binaryName" -Force

# Cleanup
Remove-Item $zipPath -Force -ErrorAction SilentlyContinue
Remove-Item "$env:TEMP\soverstack-extract" -Recurse -Force -ErrorAction SilentlyContinue

# Add to PATH if not already there
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
    Write-Host "Added $installDir to PATH" -ForegroundColor Green
}

Write-Host ""
Write-Host "Soverstack $version installed successfully!" -ForegroundColor Green
Write-Host ""

& "$installDir\$binaryName" --version

Write-Host ""

# Check Docker
if (Get-Command docker -ErrorAction SilentlyContinue) {
    Write-Host "Docker: found" -ForegroundColor Green
} else {
    Write-Host "WARNING: Docker is not installed." -ForegroundColor Yellow
    Write-Host "  Install it from: https://www.docker.com/products/docker-desktop/"
}

Write-Host ""
Write-Host "Open a new terminal and run: soverstack --help" -ForegroundColor Cyan
