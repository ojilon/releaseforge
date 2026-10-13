@echo off
REM build-installer.bat [version] — lane C packaging for ReleaseForge itself.
REM
REM   1. builds version-stamped releaseforge.exe  (release variant)
REM   2. builds installer.exe from ./installer
REM   3. packs both into dist\  +  dist\releaseforge-<version>.zip
REM
REM Version: %1 when given, else the VERSION file. Mirrors docs/10-bootstrap-v001.md section 9.
REM Run from the repo root, or from the ReleaseForge TUI:  bat build-installer.bat 0.0.4
setlocal EnableExtensions
cd /d "%~dp0"
if not exist VERSION (
  echo ERROR: VERSION file not found. Run from the repo root.
  exit /b 1
)

set "VER=%~1"
if "%VER%"=="" (
  for /f "usebackq delims=" %%v in ("VERSION") do set "VER=%%v"
)
REM trim spaces
for /f "tokens=* delims= " %%a in ("%VER%") do set "VER=%%a"
if "%VER%"=="" (
  echo ERROR: empty version.
  exit /b 1
)

echo ==^> version %VER%
if not exist dist mkdir dist

echo ==^> build releaseforge.exe ^(stamped %VER%^)
go build -ldflags "-X github.com/ojilon/releaseforge/internal/version.ToolVersion=%VER%" -o dist\releaseforge.exe .
if errorlevel 1 (
  echo ERROR: releaseforge build failed.
  exit /b 1
)

echo ==^> build installer.exe ^(stamped %VER%^)
go build -ldflags "-X github.com/ojilon/releaseforge/internal/version.ToolVersion=%VER%" -o dist\installer.exe ./installer
if errorlevel 1 (
  echo ERROR: installer build failed.
  exit /b 1
)

echo ==^> zip dist\releaseforge-%VER%.zip
powershell -NoProfile -Command "Compress-Archive -Force dist\releaseforge.exe, dist\installer.exe dist\releaseforge-%VER%.zip"
if errorlevel 1 (
  echo ERROR: zip failed.
  exit /b 1
)

echo.
echo DONE %VER%
echo   dist\releaseforge.exe
echo   dist\installer.exe
echo   dist\releaseforge-%VER%.zip
