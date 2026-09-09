@echo off
REM ===========================================================================
REM  GLM-Free API - one-command launcher (Windows)
REM
REM  Double-click this file, or run it from a terminal:  start.bat
REM
REM  What it does:
REM    1. Creates .env from .env.example on first run
REM    2. Uses a prebuilt binary if present, otherwise builds with Go
REM    3. Starts the server and prints the URLs
REM ===========================================================================
setlocal enabledelayedexpansion
cd /d "%~dp0"

echo.

if not exist .env (
    copy .env.example .env >nul
    echo [!] First-run setup:
    echo     A starter .env was created. Open it in Notepad and paste your
    echo     Z.AI token^)s^):
    echo         ZAI_TOKENS=token1,token2,token3
    echo     ^(tokens: chat.z.ai -^> F12 -^> Console -^> localStorage.getItem('token'^)
    echo.
    echo     Guest mode works without tokens - just run again to continue.
    echo.
    pause
)

set BIN=
if exist zai-api.exe                 set BIN=zai-api.exe
if not defined BIN if exist zai-api-windows-amd64.exe set BIN=zai-api-windows-amd64.exe

if not defined BIN (
    where go >nul 2>nul
    if errorlevel 1 (
        echo [X] No binary and no Go toolchain found.
        echo.
        echo     Pick ONE:
        echo       A^) Download a ready binary from the Releases page:
        echo          https://github.com/Godde3s/glm-free-api/releases
        echo          unzip it next to start.bat, then run start.bat again.
        echo       B^) Install Go 1.25+ from https://go.dev/dl and run again.
        pause
        exit /b 1
    )
    echo [+] Building zai-api.exe ^(one-time, ~30s^)...
    go build -trimpath -ldflags="-s -w" -o zai-api.exe . || (pause ^& exit /b 1)
    set BIN=zai-api.exe
)

if not defined PORT set PORT=3001
echo [+] Starting GLM-Free API ...
echo     Dashboard:  http://localhost:%PORT%/health
echo     OpenAI:     http://localhost:%PORT%/v1/chat/completions
echo     Anthropic:  http://localhost:%PORT%/v1/messages
echo     Stop:       CTRL+C
echo.
"%BIN%" %*
pause
