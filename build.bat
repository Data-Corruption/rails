@echo off
SET "DIST_DIR=dist\windows"
SET "SRC_DIR=src"
SET "RUN_ARG=%1"

if exist "%DIST_DIR%" (
  rmdir /s /q "%DIST_DIR%"
)
mkdir "%DIST_DIR%"

echo Building for Windows...
SET GOOS=windows
SET GOARCH=amd64
go build -o %DIST_DIR%\rails.exe .\%SRC_DIR%

if %ERRORLEVEL% == 0 (
  echo Build successful.
  if "%RUN_ARG%" == "-run" (
    echo Running...
    %DIST_DIR%\rails.exe
  )
) else (
  echo Build failed.
  exit /b 1
)