#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$root"

targets=(pkg cmd scripts)
if [ -d internal ]; then
  targets+=(internal)
fi

pattern='(?i)(xox[bap]-[A-Za-z0-9-]+|github_pat_[A-Za-z0-9_]+|ghp_[A-Za-z0-9]{20,}|sk-[A-Za-z0-9]{20,}|AIzaSy[A-Za-z0-9_-]{10,}|authorization:\\s*bearer\\s+[A-Za-z0-9._-]{10,}|bearer\\s+[A-Za-z0-9._-]{10,}|eyJ[A-Za-z0-9_-]+\\.[A-Za-z0-9_-]+\\.[A-Za-z0-9_-]+)'

tmpfile="$(mktemp /tmp/r5_token_scan.XXXXXX)"
trap 'rm -f "$tmpfile"' EXIT

rg -n --no-heading --color never "$pattern" "${targets[@]}" \
  --glob '!**/docs/**' \
  --glob '!**/dist/**' \
  --glob '!**/node_modules/**' \
  --glob '!**/.git/**' \
  --glob '!pkg/waveorch/redact.go' \
  --glob '!pkg/waveorch/redact_test.go' \
  --glob '!pkg/wavejwt/**' \
  --glob '!pkg/waveorch/**/redact*.go' \
  > "$tmpfile" || true

if [ -s "$tmpfile" ]; then
  cut -d: -f1,2 "$tmpfile" | sort -u
  exit 2
fi

echo "OK"
