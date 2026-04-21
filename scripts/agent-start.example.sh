#!/usr/bin/env bash
# Run cmdproxy with an explicit env allowlist; same for the agent (no token in agent env).
# Expects `cmdproxy` on PATH.
set -euo pipefail

LOG=/tmp/cmdproxy-serve.log

echo "Running cmdproxy..."
env -i \
  GITHUB_TOKEN="$GITHUB_TOKEN" \
  cmdproxy >"$LOG" 2>&1 &
SERVE_PID=$!
cleanup() { echo "Terminating cmdproxy..." && kill "$SERVE_PID" 2>/dev/null || true; }
trap cleanup EXIT

cmdproxy wait-socket -log "$LOG"

SHIM="$(cmdproxy shim-path)"
# If the agent ignores this PATH (Claude Code’s Bash tool often does), run once:
#   sudo cmdproxy link-shims
# so /usr/local/bin/gh points at the shim; default macOS PATH order then prefers it over Homebrew.
PATH="$SHIM:/opt/homebrew/bin:/opt/homebrew/sbin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"

echo "Running agent with limited ENV vars..."
exec env -i \
  HOME="$HOME" \
  USER="${USER:-$(id -un)}" \
  PATH="$PATH" \
  TERM="${TERM:-dumb}" \
  "$HOME/.local/bin/agent" "$@"
