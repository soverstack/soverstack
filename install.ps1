# Soverstack Installer for Windows
# Usage: irm https://raw.githubusercontent.com/soverstack/soverstack/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "soverstack/soverstack"
$installDir = "$env:LOCALAPPDATA\Soverstack"
$binaryName = "soverstack.exe"

Write-Host "Installing Soverstack..." -ForegroundColor Cyan

# Get latest stable release
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -UseBasicParsing
$tag = $release.tag_name
$version = $tag.TrimStart("v")

Write-Host "Latest version: $version"

# Create install directory
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

$destPath = Join-Path $installDir $binaryName

# Try zip archive first, then fall back to raw binary
$zipAsset = $release.assets | Where-Object { $_.name -match "windows-amd64\.zip$" } | Select-Object -First 1
$exeAsset = $release.assets | Where-Object { $_.name -match "windows-amd64\.exe$" } | Select-Object -First 1

if ($zipAsset) {
    # Download and extract zip
    $zipPath = "$env:TEMP\soverstack.zip"
    Write-Host "Downloading $($zipAsset.name)..."
    Invoke-WebRequest -Uri $zipAsset.browser_download_url -OutFile $zipPath -UseBasicParsing

    Expand-Archive -Path $zipPath -DestinationPath "$env:TEMP\soverstack-extract" -Force
    Copy-Item "$env:TEMP\soverstack-extract\soverstack.exe" $destPath -Force

    Remove-Item $zipPath -Force -ErrorAction SilentlyContinue
    Remove-Item "$env:TEMP\soverstack-extract" -Recurse -Force -ErrorAction SilentlyContinue

} elseif ($exeAsset) {
    # Download raw binary
    Write-Host "Downloading $($exeAsset.name)..."
    Invoke-WebRequest -Uri $exeAsset.browser_download_url -OutFile $destPath -UseBasicParsing

} else {
    Write-Host "Error: Could not find Windows binary in release $tag" -ForegroundColor Red
    exit 1
}

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
