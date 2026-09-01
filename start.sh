#!/usr/bin/env bash
# ============================================================================
#  GLM-Free-API — one-command launcher (Linux / macOS)
#
#  First run:   chmod +x start.sh && ./start.sh
#  Next runs:   ./start.sh
#
#  What it does:
#    1. Creates .env from .env.example on first run (and asks for tokens)
#    2. Finds a prebuilt binary (zai-api), builds one if Go is installed,
#       or tells you exactly how to install Go
#    3. Starts the server and prints the URLs
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")"

GOLD='\033[1;33m'; GREEN='\033[1;32m'; RED='\033[1;31m'; DIM='\033[2m'; OFF='\033[0m'
say()  { printf "${GREEN}▸${OFF} %s\n" "$1"; }
warn() { printf "${GOLD}!${OFF} %s\n" "$1"; }
die()  { printf "${RED}✗ %s${OFF}\n" "$1"; exit 1; }

# ── 1. First-run setup ──────────────────────────────────────────────────────
if [ ! -f .env ]; then
    cp .env.example .env
    printf "\n${GOLD}⚠ ── First-run setup ──────────────────────────────────────${OFF}\n"
    say "A starter .env was created from .env.example."
    say "Open it and paste your Z.AI token(s):"
    printf "${DIM}       nano .env        # ZAI_TOKENS=token1,token2,token3${OFF}\n"
    printf "${DIM}       (tokens: chat.z.ai → F12 → Console → localStorage.getItem('token'))${OFF}\n"
    printf "${GOLD}───────────────────────────────────────────────────────────${OFF}\n\n"
    # non-interactive continue — guest mode still works without tokens
fi

# ── 2. Pick up PORT from .env for the URL preview below ───────────────────
if [ -f .env ]; then
    set -a; . ./.env 2>/dev/null; set +a
fi

# ── 3. Find or build the binary ────────────────────────────────────────────
BIN=""
if [ -x ./zai-api ]; then BIN=./zai-api;
elif [ -x ./zai-api-linux-amd64 ]; then BIN=./zai-api-linux-amd64;
elif [ -x ./zai-api-linux-arm64 ]; then BIN=./zai-api-linux-arm64;
elif [ -x ./zai-api-darwin-arm64 ]; then BIN=./zai-api-darwin-arm64;
elif [ -x ./zai-api-darwin-amd64 ]; then BIN=./zai-api-darwin-amd64; fi

if [ -z "$BIN" ]; then
    if command -v go >/dev/null 2>&1; then
        say "Building zai-api (one-time, ~30s)..."
        go build -trimpath -ldflags="-s -w" -o zai-api . || die "build failed"
        BIN=./zai-api
    else
        cat <<EOF
${RED}✗ No binary and no Go toolchain found.${OFF}

Pick ONE of these:

  A) Download a ready binary (no Go needed) — from the Releases page:
       https://github.com/Godde3s/glm-free-api/releases
     unzip it next to start.sh, then run ./start.sh again.

  B) Install Go 1.25+ (https://go.dev/dl) and run ./start.sh again.
       macOS:   brew install go
       Linux:   snap install go --classic   (or apt/dnf per distro)
EOF
        exit 1
    fi
fi

# ── 4. Launch ───────────────────────────────────────────────────────────────
say "Starting GLM-Free-API ..."
printf "${DIM}  Dashboard:  http://localhost:${PORT:-3001}/            ← بازش کن!${OFF}\n"
printf "${DIM}  OpenAI:     http://localhost:${PORT:-3001}/v1/chat/completions${OFF}\n"
printf "${DIM}  Anthropic:  http://localhost:${PORT:-3001}/v1/messages${OFF}\n"
printf "${DIM}  Stop:       CTRL+C${OFF}\n\n"
exec "$BIN" "$@"
