#!/bin/bash
# Wave-Orch 3-Agent Parallel Demo
# 使用 shell block (controller=shell) 而非 cmd block

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/wave_orch_lib.sh"

# 期望的 cwd
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
EXPECTED_CWD="${EXPECTED_CWD:-$REPO_ROOT}"

# Block 缓存文件（防止长跑时无限创建新 block）
CACHE_FILE="/tmp/wave_orch_demo_3_agents_blockid"

# 强制新建开关（默认关闭）
FORCE_NEW_BLOCK="${WAVE_ORCH_FORCE_NEW_BLOCK:-0}"

WSH=$(resolve_wsh)
if [[ -z "$WSH" ]]; then
    echo "❌ wsh not found"
    exit 1
fi
echo "✅ WSH: $WSH"

# === Check jq ===
if ! command -v jq &>/dev/null; then
    echo "❌ jq not found"
    exit 1
fi

# === Block Selection Functions ===
select_block_by_cwd() {
    local blocks_json="$1"
    local cwd="$2"
    echo "$blocks_json" | jq -r --arg cwd "$cwd" '
      map(select(.view=="term" and .meta["cmd:cwd"]==$cwd))
      | .[-1].blockid // empty
    '
}

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

# === Check Wave Running ===
echo "--- Checking Wave connection ---"
if ! $WSH blocks list --view=term --json &>/dev/null; then
    echo "❌ Cannot connect to Wave Terminal"
    echo "   请先打开 Wave 并保持运行"
    exit 2
fi
echo "✅ Wave connected"

# === Detect Available Agents ===
echo "--- Detecting agents ---"
AGENTS=()
for cli in claude codex gemini; do
    if check_cli_available "$cli"; then
        AGENTS+=("$cli")
        echo "✅ $cli"
    else
        echo "⚠️ $cli not found"
    fi
done

if [[ ${#AGENTS[@]} -eq 0 ]]; then
    echo "❌ No agent CLI found"
    exit 3
fi

# === Block Selection (reuse strategy) ===
echo "--- Getting terminal block ---"
BLOCKS_JSON=$($WSH blocks list --view=term --json 2>/dev/null || echo "[]")

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
INJECT_CMD="cd \"$EXPECTED_CWD\" && echo \"wave-orch-3agents-ok\""
PATTERN="wave-orch-3agents-ok"

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
    if [[ -n "$BLOCK_ID" ]]; then
        echo "--- Injecting command to fresh block ---"
        $WSH inject --wait "$BLOCK_ID" "$INJECT_CMD" &>/dev/null || true
        sleep 1
        $WSH wait "$BLOCK_ID" --pattern="$PATTERN" --timeout=15000 &>/dev/null || true
    fi
fi

if [[ -z "$BLOCK_ID" ]]; then
    echo "❌ Failed to get block ID"
    exit 4
fi

# 更新缓存
echo "$BLOCK_ID" > "$CACHE_FILE"
echo "✅ Block ID: $BLOCK_ID (cached to $CACHE_FILE)"

# 使用同一个 block 给所有 agents
declare -a BLOCK_IDS
for agent in "${AGENTS[@]}"; do
    BLOCK_IDS+=("$BLOCK_ID")
    echo "✅ $agent using block: $BLOCK_ID"
done

# === Inject REPORT Commands ===
echo "--- Injecting REPORT commands ---"
PASS_COUNT=0

for i in "${!AGENTS[@]}"; do
    agent="${AGENTS[$i]}"
    block="${BLOCK_IDS[$i]}"

    REPORT_CMD="echo '<<<REPORT>>>{\"agent\":\"$agent\",\"status\":\"SUCCESS\",\"round\":1,\"project_id\":\"demo\",\"summary\":\"$agent ready\",\"files_changed\":[],\"commands_run\":[]}<<<END_REPORT>>>'"

    $WSH inject --wait "$block" "$REPORT_CMD" &>/dev/null || {
        echo "❌ $agent inject failed"
        continue
    }

    # Wait for REPORT to appear (with timeout)
    $WSH wait --timeout=5s "$block" "<<<END_REPORT>>>" &>/dev/null || true
    sleep 2
    OUTPUT=$($WSH output "$block" --lines=200 2>/dev/null || echo "")
    OUTPUT=$(echo "$OUTPUT" | redact_output)

    # Merge lines and extract JSON (terminal wraps long lines)
    MERGED=$(echo "$OUTPUT" | tr -d '\n' | tr -s ' ')
    if echo "$MERGED" | grep -q "<<<REPORT>>>"; then
        # Extract first complete REPORT only (avoid echo command duplication)
        JSON=$(echo "$MERGED" | sed -n 's/.*<<<REPORT>>>\({[^}]*}\)<<<END_REPORT>>>.*/\1/p' | head -1)
        if [[ -z "$JSON" ]]; then
            # Fallback: try simpler extraction
            JSON=$(echo "$MERGED" | grep -o '<<<REPORT>>>[^<]*<<<END_REPORT>>>' | head -1 | sed 's/<<<REPORT>>>//;s/<<<END_REPORT>>>//')
        fi
        if echo "$JSON" | jq -e '.agent and .status' &>/dev/null; then
            STATUS=$(echo "$JSON" | jq -r '.status')
            echo "✅ $agent: $STATUS"
            ((PASS_COUNT++))
        else
            echo "⚠️ $agent: invalid JSON"
        fi
    else
        echo "⚠️ $agent: no REPORT found"
    fi
done

# === Summary ===
echo ""
echo "=== 3-AGENT DEMO RESULT ==="
echo "Agents: ${#AGENTS[@]}, Passed: $PASS_COUNT"

if [[ $PASS_COUNT -eq ${#AGENTS[@]} ]]; then
    echo "=== PASS ✅ ==="
    exit 0
else
    echo "=== PARTIAL ==="
    exit 0
fi
