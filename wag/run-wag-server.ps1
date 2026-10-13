# Start the WAG prototype in your browser (Windows PowerShell).
# Needs Go 1.22 or later (https://go.dev/dl/) and the whole repository,
# because WAG builds against ..\carneades-4. Extra flags pass through:
#   .\run-wag-server.ps1 --kialo-dir C:\path\to\kialo_visualizer\data
Set-Location -Path $PSScriptRoot
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Go was not found. Install it from https://go.dev/dl/ and run this again,"
    Write-Host "or use the ready-made wag-server.exe instead."
    exit 1
}
go run ./cmd/wag-server @args
