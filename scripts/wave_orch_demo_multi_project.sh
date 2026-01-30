#!/bin/bash
# Wave-Orch Multi-Project Parallel Demo
# 演示同时对多个项目进行编排

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

# === Setup Demo Projects ===
echo ""
echo "--- Setting up demo projects ---"

DEMO_BASE="/tmp/wave-orch-demo"
mkdir -p "$DEMO_BASE"

# Create two demo projects
PROJECT_A="$DEMO_BASE/project-alpha"
PROJECT_B="$DEMO_BASE/project-beta"

setup_demo_project() {
    local path=$1
    local name=$2

    if [[ -d "$path" ]]; then
        echo "✅ $name exists: $path"
        return
    fi

    mkdir -p "$path"
    cd "$path"
    git init -q
    echo "# $name" > README.md
    echo "console.log('Hello from $name');" > index.js
    git add .
    git commit -q -m "Initial commit"

    # Create .wave-orch directory
    mkdir -p .wave-orch
    echo "{\"project\": \"$name\"}" > .wave-orch/config.json

    echo "✅ Created $name: $path"
}

setup_demo_project "$PROJECT_A" "project-alpha"
setup_demo_project "$PROJECT_B" "project-beta"

# === Create Project Blocks ===
echo ""
echo "--- Creating project blocks ---"

declare -a PROJECT_BLOCKS
declare -a PROJECT_NAMES
declare -a PROJECT_PATHS

create_project_block() {
    local path=$1
    local name=$2

    echo "Creating block for $name..."
    $WSH run bash -c "cd '$path' && echo '=== $name Ready ===' && pwd" &>/dev/null || true
    sleep 1

    local blocks=$($WSH blocks list --view=term --json 2>/dev/null || echo "[]")
    local block_id=$(echo "$blocks" | jq -r '.[0].blockid // empty')

    if [[ -n "$block_id" ]]; then
        echo "✅ $name block: $block_id"
        PROJECT_BLOCKS+=("$block_id")
        PROJECT_NAMES+=("$name")
        PROJECT_PATHS+=("$path")
    else
        echo "⚠️ Failed to create $name block"
    fi
}

create_project_block "$PROJECT_A" "project-alpha"
create_project_block "$PROJECT_B" "project-beta"

echo ""
echo "--- Created ${#PROJECT_BLOCKS[@]} project blocks ---"

# === Inject Tasks ===
echo ""
echo "--- Injecting tasks to projects ---"

for i in "${!PROJECT_BLOCKS[@]}"; do
    block_id="${PROJECT_BLOCKS[$i]}"
    name="${PROJECT_NAMES[$i]}"
    path="${PROJECT_PATHS[$i]}"

    echo "Injecting task to $name..."

    # Simulate a simple task: list files and report
    $WSH inject "$block_id" "ls -la && echo '<<<REPORT>>>{\"project\":\"$name\",\"status\":\"SUCCESS\",\"files_changed\":[]}<<<END_REPORT>>>'"
done

# === Wait and Collect ===
echo ""
echo "--- Waiting for outputs ---"
sleep 2

for i in "${!PROJECT_BLOCKS[@]}"; do
    block_id="${PROJECT_BLOCKS[$i]}"
    name="${PROJECT_NAMES[$i]}"

    echo ""
    echo "=== $name Output ==="
    $WSH output "$block_id" --lines=10 2>/dev/null || echo "(no output)"
done

# === Summary ===
echo ""
echo "=== MULTI-PROJECT DEMO COMPLETE ==="
echo "Projects: ${#PROJECT_BLOCKS[@]}"
for i in "${!PROJECT_BLOCKS[@]}"; do
    echo "  - ${PROJECT_NAMES[$i]}: ${PROJECT_PATHS[$i]}"
    echo "    Block: ${PROJECT_BLOCKS[$i]}"
done

echo ""
echo "Demo projects location: $DEMO_BASE"
echo "To cleanup: rm -rf $DEMO_BASE"
