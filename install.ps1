# Soverstack Launcher Installer for Windows
# Usage: irm https://raw.githubusercontent.com/soverstack/launcher/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "soverstack/launcher"
$asset = "soverstack-windows-amd64.exe"
$installDir = "$env:LOCALAPPDATA\Soverstack"
$binaryName = "soverstack.exe"

Write-Host "Installing Soverstack Launcher..." -ForegroundColor Cyan

# Create install directory
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# Download latest release
$downloadUrl = "https://github.com/$repo/releases/latest/download/$asset"
$destPath = Join-Path $installDir $binaryName

Write-Host "Downloading from $downloadUrl..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $destPath -UseBasicParsing

# Add to PATH if not already there
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
    Write-Host "Added $installDir to PATH" -ForegroundColor Green
}

Write-Host ""
Write-Host "Soverstack installed successfully!" -ForegroundColor Green
Write-Host ""

# Show version
& $destPath --version

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
