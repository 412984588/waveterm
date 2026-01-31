#!/bin/bash
# Wave-Orch E2E Smoke Test
# 必须先打开 Wave Terminal 并保持运行

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/wave_orch_lib.sh"

# 期望的 cwd（优先使用调用方环境变量）
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
EXPECTED_CWD="${EXPECTED_CWD:-$REPO_ROOT}"

# Block 缓存文件（防止长跑时无限创建新 block）
CACHE_FILE="/tmp/wave_orch_smoke_blockid"

# 强制新建开关（默认关闭）
FORCE_NEW_BLOCK="${WAVE_ORCH_FORCE_NEW_BLOCK:-0}"

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

# === Block Selection ===
# 选择最后一个匹配 cwd 的 term block（不再要求唯一）
select_block_by_cwd() {
    local blocks_json="$1"
    local cwd="$2"
    echo "$blocks_json" | jq -r --arg cwd "$cwd" '
      map(select(.view=="term" and .meta["cmd:cwd"]==$cwd))
      | .[-1].blockid // empty
    '
}

# 检查缓存的 blockid 是否仍然有效
check_cached_block() {
    local cached_id="$1"
    local blocks_json="$2"
    local cwd="$3"
    local match
    match=$(echo "$blocks_json" | jq -r --arg id "$cached_id" --arg cwd "$cwd" '
      map(select(.blockid==$id and .view=="term" and .meta["cmd:cwd"]==$cwd))
      | length
    ')
    [[ "$match" -ge 1 ]]
}

# 尝试注入并验证
try_inject_and_verify() {
    local block_id="$1"
    local cmd="$2"
    local pattern="$3"

    if ! $WSH inject --wait "$block_id" "$cmd" &>/dev/null; then
        return 1
    fi
    sleep 1
    if ! $WSH wait "$block_id" --pattern="$pattern" --timeout=15000 &>/dev/null; then
        return 1
    fi
    return 0
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

# 策略：缓存 > cwd匹配 > 新建
BLOCK_ID=""
NEED_FRESH=0

# 1) 检查强制新建开关
if [[ "$FORCE_NEW_BLOCK" == "1" ]]; then
    echo "WAVE_ORCH_FORCE_NEW_BLOCK=1, forcing fresh block"
    NEED_FRESH=1
fi

# 2) 尝试复用缓存的 blockid
if [[ "$NEED_FRESH" -eq 0 ]] && [[ -f "$CACHE_FILE" ]]; then
    CACHED_ID=$(cat "$CACHE_FILE" 2>/dev/null || echo "")
    if [[ -n "$CACHED_ID" ]] && check_cached_block "$CACHED_ID" "$BLOCKS_JSON" "$EXPECTED_CWD"; then
        echo "Reusing cached block: $CACHED_ID"
        BLOCK_ID="$CACHED_ID"
    fi
fi

# 3) 按 cwd 匹配选择最后一个
if [[ -z "$BLOCK_ID" ]] && [[ "$NEED_FRESH" -eq 0 ]]; then
    BLOCK_ID=$(select_block_by_cwd "$BLOCKS_JSON" "$EXPECTED_CWD")
    if [[ -n "$BLOCK_ID" ]]; then
        echo "Reusing block by cwd match: $BLOCK_ID"
    fi
fi

# 4) 尝试注入验证，失败则标记需要新建
INJECT_CMD="cd \"$EXPECTED_CWD\" && echo \"wave-orch-smoke-test-ok\""
PATTERN="wave-orch-smoke-test-ok"

if [[ -n "$BLOCK_ID" ]]; then
    echo "--- Injecting command ---"
    if ! try_inject_and_verify "$BLOCK_ID" "$INJECT_CMD" "$PATTERN"; then
        echo "⚠️ Inject/verify failed on $BLOCK_ID, will create fresh block"
        BLOCK_ID=""
        NEED_FRESH=1
    fi
fi

# 5) 需要新建时才创建
if [[ -z "$BLOCK_ID" ]]; then
    echo "Creating fresh shell block..."
    ensure_shell_widget "$WSH" >/dev/null 2>&1 || true
    BLOCK_ID=$(launch_shell_block "$WSH" || echo "")
    sleep 1
    if [[ -z "$BLOCK_ID" ]]; then
        BLOCKS_JSON=$($WSH blocks list --view=term --json 2>/dev/null || echo "[]")
        BLOCK_ID=$(echo "$BLOCKS_JSON" | jq -r '.[-1].blockid // empty')
    fi
    # 新建后需要注入
    if [[ -n "$BLOCK_ID" ]]; then
        echo "--- Injecting command to fresh block ---"
        $WSH inject --wait "$BLOCK_ID" "$INJECT_CMD" &>/dev/null || true
        sleep 1
        $WSH wait "$BLOCK_ID" --pattern="$PATTERN" --timeout=15000 &>/dev/null || true
    fi
fi

if [[ -z "$BLOCK_ID" ]]; then
    echo "❌ Failed to get block ID"
    exit 3
fi

# 更新缓存
echo "$BLOCK_ID" > "$CACHE_FILE"
echo "✅ Block ID: $BLOCK_ID (cached to $CACHE_FILE)"

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
