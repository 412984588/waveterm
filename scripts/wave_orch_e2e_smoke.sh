#!/bin/bash
# Wave-Orch E2E Smoke Test
# 必须先打开 Wave Terminal 并保持运行

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/wave_orch_lib.sh"

# 期望的 cwd（优先使用调用方环境变量）
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
EXPECTED_CWD="${EXPECTED_CWD:-$REPO_ROOT}"

# === Resolve WSH ===
# 优先使用本地构建的 wsh（包含 inject/output/wait 命令）
resolve_wsh() {
    # 优先本地构建版本
    if [[ -f "./dist/bin/wsh-0.13.2-alpha.0-darwin.arm64" ]]; then
        echo "./dist/bin/wsh-0.13.2-alpha.0-darwin.arm64"
    else
        local latest=$(ls -t ./dist/bin/wsh-*darwin.arm64 2>/dev/null | head -1)
        if [[ -n "$latest" ]]; then
            echo "$latest"
        elif command -v wsh &>/dev/null; then
            echo "wsh"
        else
            echo ""
        fi
    fi
}

WSH=$(resolve_wsh)
if [[ -z "$WSH" ]]; then
    echo "❌ wsh not found. Build with: task build:backend"
    exit 1
fi
echo "✅ WSH: $WSH"

# === Check jq ===
if ! command -v jq &>/dev/null; then
    echo "❌ jq not found. Install with: brew install jq"
    exit 1
fi

# === Check Wave Running ===
echo "--- Checking Wave connection ---"
if ! $WSH blocks list --view=term --json 2>/dev/null; then
    echo ""
    echo "❌ Cannot connect to Wave Terminal"
    echo "   请先打开 Wave Terminal 并保持运行"
    exit 2
fi

# === Get or Create Block ===
select_block_by_cwd() {
    local blocks_json="$1"
    local cwd="$2"
    local matches
    matches=$(echo "$blocks_json" | jq -r --arg cwd "$cwd" '
      map(select(.view=="term" and .meta["cmd:cwd"]==$cwd))
    ')
    local count
    count=$(echo "$matches" | jq -r 'length')
    if [[ "$count" -eq 1 ]]; then
        echo "$matches" | jq -r '.[0].blockid // empty'
    else
        echo ""
    fi
}

redact_output() {
    sed -E \
        -e 's/(xox[baprs]-[A-Za-z0-9-]+)/REDACTED/g' \
        -e 's/(gh[pousr]_[A-Za-z0-9_]+)/REDACTED/g' \
        -e 's/(github_pat_[A-Za-z0-9_]+)/REDACTED/g' \
        -e 's/(sk-[A-Za-z0-9]{8,})/REDACTED/g' \
        -e 's/(AIzaSy[ A-Za-z0-9_-]{10,})/REDACTED/g' \
        -e 's/(Bearer[[:space:]]+)[A-Za-z0-9._-]+/\1REDACTED/Ig' \
        -e 's/(authorization:[[:space:]]*)([^[:space:]]+)/\1REDACTED/Ig' \
        -e 's/eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+/REDACTED/g'
}

echo "--- Getting terminal block ---"
BLOCKS_JSON=$($WSH blocks list --view=term --json 2>/dev/null || echo "[]")
BLOCK_ID=$(select_block_by_cwd "$BLOCKS_JSON" "$EXPECTED_CWD")

if [[ -z "$BLOCK_ID" ]]; then
    echo "No unique terminal block found for cwd: $EXPECTED_CWD"
    echo "Creating a fresh shell block..."
    ensure_shell_widget "$WSH" >/dev/null 2>&1 || true
    BLOCK_ID=$(launch_shell_block "$WSH" || echo "")
    sleep 1
    if [[ -z "$BLOCK_ID" ]]; then
        BLOCK_ID=$(echo "$BLOCKS_JSON" | jq -r '.[-1].blockid // empty')
    fi
    BLOCKS_JSON=$($WSH blocks list --view=term --json 2>/dev/null || echo "[]")
fi

if [[ -z "$BLOCK_ID" ]]; then
    echo "❌ Failed to get block ID"
    echo "Blocks: $BLOCKS_JSON"
    exit 3
fi
BLOCK_CWD=$(echo "$BLOCKS_JSON" | jq -r --arg id "$BLOCK_ID" 'map(select(.blockid==$id)) | .[0].meta["cmd:cwd"] // empty')
echo "✅ Block ID: $BLOCK_ID (cwd: $BLOCK_CWD)"
echo "Expected cwd: $EXPECTED_CWD"

# === Inject ===
echo "--- Injecting command ---"
INJECT_CMD="cd \"$EXPECTED_CWD\" && echo \"wave-orch-smoke-test-ok\""
$WSH inject --wait "$BLOCK_ID" "$INJECT_CMD" &>/dev/null
echo "✅ Injected"

# === Wait ===
echo "--- Waiting for pattern ---"
echo "Pattern: wave-orch-smoke-test-ok"
sleep 1
if $WSH wait "$BLOCK_ID" --pattern="wave-orch-smoke-test-ok" --timeout=15000 &>/dev/null; then
    echo "✅ Pattern found"
else
    echo "⚠️ Wait timeout, checking output anyway..."
fi

# === Output ===
echo "--- Getting output ---"
OUTPUT=$($WSH output "$BLOCK_ID" --lines=200 2>/dev/null || echo "")
echo "$OUTPUT" | redact_output

if echo "$OUTPUT" | grep -q "wave-orch-smoke-test-ok"; then
    echo ""
    echo "=== E2E SMOKE PASSED ==="
else
    echo ""
    echo "=== E2E SMOKE FAILED ==="
    exit 4
fi
