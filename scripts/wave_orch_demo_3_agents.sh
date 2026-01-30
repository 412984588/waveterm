#!/bin/bash
# Wave-Orch 3-Agent Parallel Demo
# 演示同时启动 Claude/Codex/Gemini 三个 Agent

set -e

# === Resolve WSH ===
# 优先使用本地构建的 wsh（包含 inject/output/wait 命令）
resolve_wsh() {
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

# === Detect Available Agents ===
echo ""
echo "--- Detecting available agents ---"

CLAUDE_CMD=""
CODEX_CMD=""
GEMINI_CMD=""

if command -v claude &>/dev/null; then
    CLAUDE_CMD="claude"
    echo "✅ Claude CLI: $CLAUDE_CMD"
else
    echo "⚠️ Claude CLI not found"
fi

if command -v codex &>/dev/null; then
    CODEX_CMD="codex"
    echo "✅ Codex CLI: $CODEX_CMD"
else
    echo "⚠️ Codex CLI not found"
fi

if command -v gemini &>/dev/null; then
    GEMINI_CMD="gemini"
    echo "✅ Gemini CLI: $GEMINI_CMD"
else
    echo "⚠️ Gemini CLI not found"
fi

AGENT_COUNT=0
[[ -n "$CLAUDE_CMD" ]] && ((AGENT_COUNT++))
[[ -n "$CODEX_CMD" ]] && ((AGENT_COUNT++))
[[ -n "$GEMINI_CMD" ]] && ((AGENT_COUNT++))

if [[ $AGENT_COUNT -eq 0 ]]; then
    echo "❌ No agent CLI found"
    exit 3
fi
echo ""
echo "✅ Found $AGENT_COUNT agent(s)"

# === Create Agent Blocks ===
echo ""
echo "--- Creating agent blocks ---"

declare -a BLOCK_IDS
declare -a AGENT_NAMES

# Function to create agent block
create_agent_block() {
    local name=$1
    local cmd=$2

    echo "Creating block for $name..."
    # Use wsh run to create a new terminal block
    $WSH run bash -c "echo '=== $name Agent Ready ===' && sleep 1" &>/dev/null || true
    sleep 1

    # Get the latest block
    local blocks=$($WSH blocks list --view=term --json 2>/dev/null || echo "[]")
    local block_id=$(echo "$blocks" | jq -r '.[0].blockid // empty')

    if [[ -n "$block_id" ]]; then
        echo "✅ $name block: $block_id"
        BLOCK_IDS+=("$block_id")
        AGENT_NAMES+=("$name")
    else
        echo "⚠️ Failed to create $name block"
    fi
}

# Create blocks for available agents
if [[ -n "$CLAUDE_CMD" ]]; then
    create_agent_block "Claude" "$CLAUDE_CMD"
fi

if [[ -n "$CODEX_CMD" ]]; then
    create_agent_block "Codex" "$CODEX_CMD"
fi

if [[ -n "$GEMINI_CMD" ]]; then
    create_agent_block "Gemini" "$GEMINI_CMD"
fi

echo ""
echo "--- Created ${#BLOCK_IDS[@]} agent blocks ---"

# === Inject Test Commands ===
echo ""
echo "--- Injecting test commands ---"

for i in "${!BLOCK_IDS[@]}"; do
    block_id="${BLOCK_IDS[$i]}"
    agent_name="${AGENT_NAMES[$i]}"

    echo "Injecting to $agent_name ($block_id)..."
    $WSH inject --wait "$block_id" "echo '<<<REPORT>>>{\"agent\":\"$agent_name\",\"status\":\"SUCCESS\",\"round\":1}<<<END_REPORT>>>'"
done

# === Wait and Collect ===
echo ""
echo "--- Waiting for outputs ---"
sleep 2

for i in "${!BLOCK_IDS[@]}"; do
    block_id="${BLOCK_IDS[$i]}"
    agent_name="${AGENT_NAMES[$i]}"

    echo ""
    echo "=== $agent_name Output ==="
    $WSH output "$block_id" --lines=5 2>/dev/null || echo "(no output)"
done

# === Summary ===
echo ""
echo "=== 3-AGENT DEMO COMPLETE ==="
echo "Blocks created: ${#BLOCK_IDS[@]}"
for i in "${!BLOCK_IDS[@]}"; do
    echo "  - ${AGENT_NAMES[$i]}: ${BLOCK_IDS[$i]}"
done
