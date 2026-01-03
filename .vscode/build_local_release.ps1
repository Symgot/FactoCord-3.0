# FactoCord3 Local Release Build Script
# This script mimics the GitHub Actions workflow for local testing
# Output will be placed in .vscode/release

Write-Host "Building FactoCord3 Release locally..." -ForegroundColor Cyan

# Get version info
$gitTag = git describe --tags 2>$null
if (-not $gitTag) {
    $gitTag = "v0.0.0-dev"
}
$gitHash = git rev-parse --short HEAD 2>$null
if (-not $gitHash) {
    $gitHash = "unknown"
}
$version = "$gitTag ($gitHash)"
Write-Host "Version: $version" -ForegroundColor Yellow

# Build executables
Write-Host "`nBuilding FactoCord3..." -ForegroundColor Cyan
go build -ldflags "-X 'github.com/Symgot/FactoCord-3.0/v3/support.FactoCordVersion=$version'" -o FactoCord3.exe -v .
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`nBuilding FactoCord3 without CGO..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
go build -ldflags "-X 'github.com/Symgot/FactoCord-3.0/v3/support.FactoCordVersion=$version'" -o FactoCord3-c.exe -v .
$env:CGO_ENABLED = ""
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build without CGO failed!" -ForegroundColor Red
    exit 1
}

# Create release directory structure
Write-Host "`nCreating release structure..." -ForegroundColor Cyan
$releaseDir = ".vscode\release"
$releaseFactoCordDir = "$releaseDir\FactoCord3"

# Clean and recreate directories
if (Test-Path $releaseDir) {
    Remove-Item -Recurse -Force $releaseDir
}
New-Item -ItemType Directory -Force -Path $releaseFactoCordDir | Out-Null

# Copy files to release directory
Write-Host "Copying files..." -ForegroundColor Cyan
Copy-Item -Path "config-example.json" -Destination $releaseFactoCordDir
Copy-Item -Path "control.lua" -Destination $releaseFactoCordDir
Copy-Item -Path "FactoCord3.exe" -Destination $releaseFactoCordDir\FactoCord3.exe
Copy-Item -Path "INSTALL.md" -Destination $releaseFactoCordDir
Copy-Item -Path "LICENSE" -Destination $releaseFactoCordDir
Copy-Item -Path "README.md" -Destination $releaseFactoCordDir
Copy-Item -Path "SECURITY.md" -Destination $releaseFactoCordDir
Copy-Item -Path "COMMANDS.md" -Destination $releaseFactoCordDir

# Copy CGO-disabled version to release root
Copy-Item -Path "FactoCord3-c.exe" -Destination $releaseDir\FactoCord3-c.exe

# Create archives
Write-Host "`nCreating archives..." -ForegroundColor Cyan

# Create ZIP archive
Push-Location $releaseDir
Compress-Archive -Path "FactoCord3" -DestinationPath "FactoCord3.zip" -Force
Write-Host "Created ZIP archive: $releaseDir\FactoCord3.zip" -ForegroundColor Green

# Create TAR.GZ archive (requires tar.exe available in Windows 10+)
if (Get-Command tar -ErrorAction SilentlyContinue) {
    tar -czf "FactoCord3.tar.gz" "FactoCord3"
    Write-Host "Created TAR.GZ archive: $releaseDir\FactoCord3.tar.gz" -ForegroundColor Green
} else {
    Write-Host "tar.exe not found - skipping .tar.gz creation" -ForegroundColor Yellow
}
Pop-Location

# Cleanup temporary build artifacts in root
Write-Host "`nCleaning up..." -ForegroundColor Cyan
Remove-Item -Force "FactoCord3.exe" -ErrorAction SilentlyContinue
Remove-Item -Force "FactoCord3-c.exe" -ErrorAction SilentlyContinue

Write-Host "`n=== Build Complete ===" -ForegroundColor Green
Write-Host "Release files are in: $releaseDir" -ForegroundColor Green
Write-Host "  - FactoCord3\ (directory with all files)" -ForegroundColor Gray
Write-Host "  - FactoCord3-c.exe (CGO-disabled executable)" -ForegroundColor Gray
Write-Host "  - FactoCord3.zip (archive)" -ForegroundColor Gray
if (Get-Command tar -ErrorAction SilentlyContinue) {
    Write-Host "  - FactoCord3.tar.gz (archive)" -ForegroundColor Gray
}
