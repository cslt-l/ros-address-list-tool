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
New-Item -ItemType Directory -Force -Path (Join-Path $PackageDir "web\src") | Out-Null

Write-Host ""
Write-Host "[1/7] Run tests..."
go test ./...
if ($LASTEXITCODE -ne 0) {
    throw "go test failed. Build stopped."
}

Write-Host ""
Write-Host "[2/7] Build Linux amd64 binary..."
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
Write-Host "[3/7] Copy config and docs..."
Copy-Item (Join-Path $ProjectRoot "configs\config.json") (Join-Path $PackageDir "config.json") -Force
Copy-Item (Join-Path $ProjectRoot "README.md") (Join-Path $PackageDir "README.md") -Force
Copy-Item (Join-Path $ProjectRoot "LICENSE") (Join-Path $PackageDir "LICENSE") -Force

Write-Host ""
Write-Host "[4/7] Copy web static files..."
$WebSrc = Join-Path $ProjectRoot "web\src"
$WebDst = Join-Path $PackageDir "web\src"
if (-not (Test-Path $WebSrc)) {
    throw "web/src was not found."
}
Copy-Item (Join-Path $WebSrc "*") $WebDst -Recurse -Force

Write-Host ""
Write-Host "[5/7] Ensure runtime directories..."
New-Item -ItemType Directory -Force -Path (Join-Path $PackageDir "logs") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $PackageDir "output") | Out-Null

Write-Host ""
Write-Host "[6/7] Create RUN.txt..."
$RunTxt = @"
ros-address-list-tool Linux amd64

Before running on Linux:
chmod +x ./ros-address-list-tool

1) Start HTTP service
   ./ros-address-list-tool -config ./config.json -serve -listen 127.0.0.1:8090

2) Run one render only
   ./ros-address-list-tool -config ./config.json

3) Login page
   http://127.0.0.1:8090/login.html

4) Web console
   http://127.0.0.1:8090/

5) Machine-to-machine RSC download endpoint
   http://127.0.0.1:8090/api/v1/render/latest.rsc?access_token=YOUR_TOKEN&mode=replace_all

Notes
- Run the program inside this release directory.
- config.json defaults to web_dir = ./web/src, so web files are packaged into web/src.
- output is used for generated .rsc files.
- logs is used for log files.
- For API automation, prefer auth_token / access_token rather than browser login.
"@
Set-Content -Path (Join-Path $PackageDir "RUN.txt") -Value $RunTxt -Encoding UTF8

Write-Host ""
Write-Host "[7/7] Create ZIP package..."
Compress-Archive -Path $PackageDir -DestinationPath $ZipPath -Force

Write-Host ""
Write-Host "Done."
Write-Host "Folder: $PackageDir"
Write-Host "ZIP: $ZipPath"
Write-Host ""
