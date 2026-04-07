param(
    [string]$Version = "v1.0.0"
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "================================================"
Write-Host "ros-address-list-tool Linux amd64 build script"
Write-Host "================================================"
Write-Host "Version: $Version"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
Set-Location $ProjectRoot

Write-Host "Project Root: $ProjectRoot"

$DistRoot = Join-Path $ProjectRoot "dist"
$PackageName = "ros-address-list-tool_${Version}_linux_amd64"
$PackageDir = Join-Path $DistRoot $PackageName
$ZipPath = Join-Path $DistRoot "${PackageName}.zip"

if (Test-Path $PackageDir) {
    Remove-Item $PackageDir -Recurse -Force
}
if (Test-Path $ZipPath) {
    Remove-Item $ZipPath -Force
}

New-Item -ItemType Directory -Force -Path $PackageDir | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $PackageDir "logs") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $PackageDir "output") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $PackageDir "web") | Out-Null

Write-Host ""
Write-Host "[1/6] Run tests..."
go test ./...
if ($LASTEXITCODE -ne 0) {
    throw "go test failed. Build stopped."
}

Write-Host ""
Write-Host "[2/6] Build Linux amd64 binary..."
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

$BinaryPath = Join-Path $PackageDir "ros-address-list-tool"

go build `
    -trimpath `
    -ldflags="-s -w" `
    -o $BinaryPath `
    ./cmd/server

if ($LASTEXITCODE -ne 0) {
    throw "go build failed. Build stopped."
}

Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "[3/6] Copy config and docs..."

Copy-Item (Join-Path $ProjectRoot "configs\config.json") (Join-Path $PackageDir "config.json") -Force
Copy-Item (Join-Path $ProjectRoot "README.md") (Join-Path $PackageDir "README.md") -Force
Copy-Item (Join-Path $ProjectRoot "LICENSE") (Join-Path $PackageDir "LICENSE") -Force

Write-Host ""
Write-Host "[4/6] Copy web static files..."

$WebDistSrc = Join-Path $ProjectRoot "web\src"
$WebDistDst = Join-Path $PackageDir "web\dist"

if (-not (Test-Path $WebDistSrc)) {
    throw "web/dist was not found."
}

Copy-Item $WebDistSrc $WebDistDst -Recurse -Force

Write-Host ""
Write-Host "[5/6] Create RUN.txt..."

$RunTxt = @"
ros-address-list-tool Linux amd64

Before running on Linux:
  chmod +x ./ros-address-list-tool

Start HTTP service:
  ./ros-address-list-tool -config ./config.json -serve -listen 127.0.0.1:8090

Run one render only:
  ./ros-address-list-tool -config ./config.json

Web UI:
  http://127.0.0.1:8090/

Notes:
1. Run the program inside this release directory.
2. output is used for generated .rsc files.
3. logs is used for log files.
4. For remote access, configure auth_token or ROS_LIST_API_TOKEN.
"@

Set-Content -Path (Join-Path $PackageDir "RUN.txt") -Value $RunTxt -Encoding UTF8

Write-Host ""
Write-Host "[6/6] Create ZIP package..."

Compress-Archive -Path $PackageDir -DestinationPath $ZipPath -Force

Write-Host ""
Write-Host "Done."
Write-Host "Folder: $PackageDir"
Write-Host "ZIP:    $ZipPath"
Write-Host ""
