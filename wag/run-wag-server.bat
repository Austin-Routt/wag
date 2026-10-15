@echo off
rem Start the WAG prototype in your browser (Windows).
rem Needs Go 1.24 or later (https://go.dev/dl/) and the whole repository,
rem because WAG builds against ..\carneades-4.
rem Extra flags pass through, for example:
rem   run-wag-server.bat --addr 127.0.0.1:9090
rem   run-wag-server.bat --kialo-dir C:\path\to\kialo_visualizer\data
rem   run-wag-server.bat --store D:\research\WAG
rem Models and annotation are in the browser (Models and Annotate tabs); your
rem data (players, games with models, annotation runs) goes to %USERPROFILE%\WAG
rem unless you give --store.
cd /d "%~dp0"
where go >nul 2>nul
if errorlevel 1 (
  echo Go was not found. Install it from https://go.dev/dl/ and run this again,
  echo or use the ready-made wag-server.exe instead.
  pause
  exit /b 1
)
go run ./cmd/wag-server %*
pause
