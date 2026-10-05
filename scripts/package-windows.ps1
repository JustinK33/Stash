param(
  [string]$ZipPath = "Stash-Windows.zip"
)
$ErrorActionPreference = "Stop"

$Version = if ($env:VERSION) { $env:VERSION } else { "1.0.0" }
$BuildNumber = if ($env:BUILD_NUMBER) { $env:BUILD_NUMBER } else { "1" }
$Root = Split-Path -Parent $PSScriptRoot
$Exe = Join-Path $Root "cmd\stash\Stash.exe"

Remove-Item -Force -ErrorAction SilentlyContinue $Exe, $ZipPath

# fyne package builds a GUI-subsystem exe (no console window) with the icon
# and version info embedded.
go run fyne.io/tools/cmd/fyne@v1.7.3 package `
  --target windows `
  --release `
  --src (Join-Path $Root "cmd\stash") `
  --name Stash `
  --app-id com.justink33.stash `
  --icon (Join-Path $Root "resources\icons\stash.png") `
  --app-version $Version `
  --app-build $BuildNumber
if ($LASTEXITCODE -ne 0) { throw "fyne package failed" }

Compress-Archive -Path $Exe -DestinationPath $ZipPath
Remove-Item $Exe
