#!/bin/bash
# Wave-Orch 3-Agent Parallel Demo
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

# === Ensure Shell Widget ===
echo "--- Ensuring shell widget ---"
ensure_shell_widget "$WSH" || exit 4

# === Launch Shell Blocks ===
echo "--- Launching shell blocks ---"
declare -a BLOCK_IDS

for agent in "${AGENTS[@]}"; do
    BLOCK_ID=$(launch_shell_block "$WSH")
    if [[ -z "$BLOCK_ID" ]]; then
        echo "❌ Failed to launch block for $agent"
        exit 5
    fi
    BLOCK_IDS+=("$BLOCK_ID")
    echo "✅ $agent block: $BLOCK_ID"
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

    sleep 1
    OUTPUT=$($WSH output "$block" --lines=20 2>/dev/null || echo "")

    if echo "$OUTPUT" | grep -q "<<<REPORT>>>"; then
        JSON=$(echo "$OUTPUT" | grep -o '<<<REPORT>>>.*<<<END_REPORT>>>' | sed 's/<<<REPORT>>>//;s/<<<END_REPORT>>>//')
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
