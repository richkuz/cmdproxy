#!/usr/bin/env bash
# Run cmdproxy with an explicit env allowlist; same for the agent (no token in agent env).
# Expects `cmdproxy` on PATH.
set -euo pipefail

LOG=/tmp/cmdproxy-serve.log

# Run cmdproxy daemon with only the ENV vars you want it and its allowed commands to see.
# For example, allow access to GITHUB_TOKEN.
env -i \
  HOME="$HOME" \
  GITHUB_TOKEN="$GITHUB_TOKEN" \
  cmdproxy >"$LOG" 2>&1 &
SERVE_PID=$!
cleanup() { kill "$SERVE_PID" 2>/dev/null || true; }
trap cleanup EXIT

cmdproxy wait-socket -log "$LOG"

# Agent: extend PATH with shims; no GITHUB_TOKEN.
exec env -i \
  HOME="$HOME" \
  USER="${USER:-$(id -un)}" \
  PATH="$(cmdproxy shim-path):/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin" \
  TERM="${TERM:-dumb}" \
  agent "$@"
