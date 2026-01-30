#!/bin/bash
# Wave-Orch E2E Smoke Test
# 必须先打开 Wave Terminal 并保持运行

set -e

# === Resolve WSH ===
resolve_wsh() {
    if command -v wsh &>/dev/null; then
        echo "wsh"
    elif [[ -f "./dist/bin/wsh-0.13.2-alpha.0-darwin.arm64" ]]; then
        echo "./dist/bin/wsh-0.13.2-alpha.0-darwin.arm64"
    else
        local latest=$(ls -t ./dist/bin/wsh-*darwin.arm64 2>/dev/null | head -1)
        if [[ -n "$latest" ]]; then
            echo "$latest"
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
echo "--- Getting terminal block ---"
BLOCKS_JSON=$($WSH blocks list --view=term --json 2>/dev/null || echo "[]")
BLOCK_ID=$(echo "$BLOCKS_JSON" | jq -r '.[0].blockid // empty')

if [[ -z "$BLOCK_ID" ]]; then
    echo "No terminal block found, creating one..."
    $WSH run bash -c 'echo "wave-orch-smoke-ready"' &>/dev/null || true
    sleep 2
    BLOCKS_JSON=$($WSH blocks list --view=term --json 2>/dev/null || echo "[]")
    BLOCK_ID=$(echo "$BLOCKS_JSON" | jq -r '.[0].blockid // empty')
fi

if [[ -z "$BLOCK_ID" ]]; then
    echo "❌ Failed to get block ID"
    echo "Blocks: $BLOCKS_JSON"
    exit 3
fi
echo "✅ Block ID: $BLOCK_ID"

# === Inject ===
echo "--- Injecting command ---"
$WSH inject "$BLOCK_ID" 'echo "wave-orch-smoke-test-ok"'
echo "✅ Injected"

# === Wait ===
echo "--- Waiting for pattern ---"
sleep 1
if $WSH wait "$BLOCK_ID" --pattern="wave-orch-smoke-test-ok" --timeout=5000; then
    echo "✅ Pattern found"
else
    echo "⚠️ Wait timeout, checking output anyway..."
fi

# === Output ===
echo "--- Getting output ---"
OUTPUT=$($WSH output "$BLOCK_ID" --lines=10 2>/dev/null || echo "")
echo "$OUTPUT"

if echo "$OUTPUT" | grep -q "wave-orch-smoke-test-ok"; then
    echo ""
    echo "=== E2E SMOKE PASSED ==="
else
    echo ""
    echo "=== E2E SMOKE FAILED ==="
    exit 4
fi
