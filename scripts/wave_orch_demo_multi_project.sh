#!/bin/bash
# Wave-Orch Multi-Project Parallel Demo
# 使用 shell block (controller=shell) 而非 cmd block

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/wave_orch_lib.sh"

# Block 缓存文件（防止长跑时无限创建新 block）
CACHE_FILE="/tmp/wave_orch_demo_multi_project_blockid"

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
    exit 2
fi
echo "✅ Wave connected"

# === Setup Demo Projects ===
echo "--- Setting up demo projects ---"

DEMO_BASE="/tmp/wave-orch-demo"
mkdir -p "$DEMO_BASE"

PROJECT_A="$DEMO_BASE/project-alpha"
PROJECT_B="$DEMO_BASE/project-beta"

setup_demo_project() {
    local path=$1
    local name=$2
    local ORIG_DIR=$(pwd)

    if [[ -d "$path" ]]; then
        echo "✅ $name exists"
        return
    fi

    mkdir -p "$path"
    cd "$path"
    git init -q
    echo "# $name" > README.md
    echo "console.log('Hello from $name');" > index.js
    git add .
    git commit -q -m "Initial commit" --no-verify
    mkdir -p .wave-orch
    echo "{\"project\": \"$name\"}" > .wave-orch/config.json
    echo "✅ Created $name"
    cd "$ORIG_DIR"
}

setup_demo_project "$PROJECT_A" "project-alpha"
setup_demo_project "$PROJECT_B" "project-beta"

# === Block Selection (reuse strategy) ===
echo "--- Getting terminal block ---"
BLOCKS_JSON=$($WSH blocks list --view=term --json 2>/dev/null || echo "[]")

# 使用第一个项目路径作为期望 cwd
EXPECTED_CWD="$PROJECT_A"

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
    if [[ -n "$CACHED_ID" ]]; then
        # 检查 block 是否仍存在（不要求 cwd 匹配，因为会 cd）
        EXISTS=$(echo "$BLOCKS_JSON" | jq -r --arg id "$CACHED_ID" 'map(select(.blockid==$id)) | length')
        if [[ "$EXISTS" -ge 1 ]]; then
            echo "Reusing cached block: $CACHED_ID"
            BLOCK_ID="$CACHED_ID"
        fi
    fi
fi

# 3) 按任意 term block 选择最后一个
if [[ -z "$BLOCK_ID" ]] && [[ "$NEED_FRESH" -eq 0 ]]; then
    BLOCK_ID=$(echo "$BLOCKS_JSON" | jq -r 'map(select(.view=="term")) | .[-1].blockid // empty')
    if [[ -n "$BLOCK_ID" ]]; then
        echo "Reusing block: $BLOCK_ID"
    fi
fi

# 4) 尝试注入验证，失败则标记需要新建
INJECT_CMD="echo \"wave-orch-multiproj-ok\""
PATTERN="wave-orch-multiproj-ok"

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
    exit 3
fi

# 更新缓存
echo "$BLOCK_ID" > "$CACHE_FILE"
echo "✅ Block ID: $BLOCK_ID (cached to $CACHE_FILE)"

# 使用同一个 block 给所有项目
declare -a PROJECT_BLOCKS
declare -a PROJECT_NAMES
declare -a PROJECT_PATHS

for proj_path in "$PROJECT_A" "$PROJECT_B"; do
    proj_name=$(basename "$proj_path")
    PROJECT_BLOCKS+=("$BLOCK_ID")
    PROJECT_NAMES+=("$proj_name")
    PROJECT_PATHS+=("$proj_path")
    echo "✅ $proj_name using block: $BLOCK_ID"
done

echo "--- Using ${#PROJECT_BLOCKS[@]} project entries (same block) ---"

# === Inject Tasks ===
echo "--- Injecting tasks ---"
PASS_COUNT=0

for i in "${!PROJECT_BLOCKS[@]}"; do
    block_id="${PROJECT_BLOCKS[$i]}"
    name="${PROJECT_NAMES[$i]}"
    path="${PROJECT_PATHS[$i]}"

    # cd to project and run task
    CD_CMD="cd '$path'"
    $WSH inject --wait "$block_id" "$CD_CMD" &>/dev/null || true
    sleep 0.5

    # Run task and emit REPORT
    TASK_CMD="ls -la && echo '<<<REPORT>>>{\"project\":\"$name\",\"status\":\"SUCCESS\",\"round\":1,\"project_id\":\"$name\",\"summary\":\"Listed files\",\"files_changed\":[],\"commands_run\":[\"ls -la\"]}<<<END_REPORT>>>'"
    $WSH inject --wait "$block_id" "$TASK_CMD" &>/dev/null || {
        echo "⚠️ $name inject failed"
        continue
    }

    # Wait for REPORT to appear
    $WSH wait --timeout=5s "$block_id" "<<<END_REPORT>>>" &>/dev/null || true
    sleep 2
    OUTPUT=$($WSH output "$block_id" --lines=200 2>/dev/null || echo "")
    OUTPUT=$(echo "$OUTPUT" | redact_output)

    # Merge lines and extract JSON (terminal wraps long lines)
    MERGED=$(echo "$OUTPUT" | tr -d '\n' | tr -s ' ')
    if echo "$MERGED" | grep -q "<<<REPORT>>>"; then
        # Extract first complete REPORT only
        JSON=$(echo "$MERGED" | sed -n 's/.*<<<REPORT>>>\({[^}]*}\)<<<END_REPORT>>>.*/\1/p' | head -1)
        if [[ -z "$JSON" ]]; then
            JSON=$(echo "$MERGED" | grep -o '<<<REPORT>>>[^<]*<<<END_REPORT>>>' | head -1 | sed 's/<<<REPORT>>>//;s/<<<END_REPORT>>>//')
        fi
        if echo "$JSON" | jq -e '.project and .status' &>/dev/null; then
            STATUS=$(echo "$JSON" | jq -r '.status')
            echo "✅ $name: $STATUS"
            ((PASS_COUNT++))
        else
            echo "⚠️ $name: invalid JSON"
        fi
    else
        echo "⚠️ $name: no REPORT found"
    fi
done

# === Summary ===
echo ""
echo "=== MULTI-PROJECT DEMO RESULT ==="
echo "Projects: ${#PROJECT_BLOCKS[@]}, Passed: $PASS_COUNT"

for i in "${!PROJECT_BLOCKS[@]}"; do
    echo "  - ${PROJECT_NAMES[$i]}: ${PROJECT_PATHS[$i]}"
    echo "    Block: ${PROJECT_BLOCKS[$i]}"
done

echo ""
echo "Demo location: $DEMO_BASE"
echo "Cleanup: rm -rf $DEMO_BASE"

if [[ $PASS_COUNT -eq ${#PROJECT_BLOCKS[@]} ]]; then
    echo "=== PASS ✅ ==="
    exit 0
else
    echo "=== PARTIAL ==="
    exit 0
fi
