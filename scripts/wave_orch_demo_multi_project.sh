#!/bin/bash
# Wave-Orch Multi-Project Parallel Demo
# 使用 shell block (controller=shell) 而非 cmd block

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/wave_orch_lib.sh"

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
    git commit -q -m "Initial commit"
    mkdir -p .wave-orch
    echo "{\"project\": \"$name\"}" > .wave-orch/config.json
    echo "✅ Created $name"
}

setup_demo_project "$PROJECT_A" "project-alpha"
setup_demo_project "$PROJECT_B" "project-beta"

# === Ensure Shell Widget ===
echo "--- Ensuring shell widget ---"
ensure_shell_widget "$WSH" || exit 3

# === Launch Shell Blocks for Projects ===
echo "--- Launching project blocks ---"
declare -a PROJECT_BLOCKS
declare -a PROJECT_NAMES
declare -a PROJECT_PATHS

for proj_path in "$PROJECT_A" "$PROJECT_B"; do
    proj_name=$(basename "$proj_path")
    BLOCK_ID=$(launch_shell_block "$WSH")
    if [[ -z "$BLOCK_ID" ]]; then
        echo "⚠️ Failed to launch block for $proj_name"
        continue
    fi
    PROJECT_BLOCKS+=("$BLOCK_ID")
    PROJECT_NAMES+=("$proj_name")
    PROJECT_PATHS+=("$proj_path")
    echo "✅ $proj_name block: $BLOCK_ID"
done

echo "--- Created ${#PROJECT_BLOCKS[@]} project blocks ---"

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

    sleep 1
    OUTPUT=$($WSH output "$block_id" --lines=20 2>/dev/null || echo "")

    if echo "$OUTPUT" | grep -q "<<<REPORT>>>"; then
        JSON=$(echo "$OUTPUT" | grep -o '<<<REPORT>>>.*<<<END_REPORT>>>' | sed 's/<<<REPORT>>>//;s/<<<END_REPORT>>>//')
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
