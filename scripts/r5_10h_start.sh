#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
OUT="/tmp/r5_10h_runner.out"
PID_FILE="/tmp/r5_10h_runner.pid"

if [ -f "$PID_FILE" ]; then
  if kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
    echo "runner already running (pid $(cat "$PID_FILE"))"
    exit 0
  fi
fi

PATH="$HOME/.local/bin:$PATH"
export PATH

if command -v setsid >/dev/null 2>&1; then
  nohup setsid "$ROOT/scripts/r5_10h_runner.sh" >>"$OUT" 2>&1 &
else
  nohup "$ROOT/scripts/r5_10h_runner.sh" >>"$OUT" 2>&1 &
fi

echo $! > "$PID_FILE"
disown || true
echo "runner started (pid $!)"
