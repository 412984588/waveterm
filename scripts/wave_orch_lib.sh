#!/bin/bash
# Wave-Orch Shell Block Library
# 提供 ensure_shell_widget() 和 launch_shell_block() 函数

# === Resolve WSH ===
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

# === ensure_shell_widget ===
# 确保 widgets.json 中存在 wave-orch-shell widget
ensure_shell_widget() {
    local WSH="$1"
    [[ -z "$WSH" ]] && WSH=$(resolve_wsh)

    # 获取 config 目录
    local CONFIG_DIR=$($WSH wavepath config 2>/dev/null)
    if [[ -z "$CONFIG_DIR" ]]; then
        echo "❌ Cannot get Wave config path" >&2
        return 1
    fi

    # 尝试定位 widgets.json
    local WIDGETS_FILE=""
    if [[ -f "$CONFIG_DIR/widgets.json" ]]; then
        WIDGETS_FILE="$CONFIG_DIR/widgets.json"
    elif [[ -f "$CONFIG_DIR/config/widgets.json" ]]; then
        WIDGETS_FILE="$CONFIG_DIR/config/widgets.json"
    else
        # 创建目录和空 JSON
        mkdir -p "$CONFIG_DIR"
        WIDGETS_FILE="$CONFIG_DIR/widgets.json"
        echo '{}' > "$WIDGETS_FILE"
    fi

    # 定义 wave-orch-shell widget
    local WIDGET_DEF='{
        "wave-orch-shell": {
            "display:name": "Wave-Orch Shell",
            "blockdef": {
                "meta": {
                    "view": "term",
                    "controller": "shell"
                }
            }
        }
    }'

    # 用 jq merge（不覆盖已有 widgets）
    local MERGED=$(jq -s '.[0] * .[1]' "$WIDGETS_FILE" <(echo "$WIDGET_DEF") 2>/dev/null)
    if [[ -z "$MERGED" ]]; then
        echo "❌ jq merge failed" >&2
        return 1
    fi

    # 写回
    echo "$MERGED" > "$WIDGETS_FILE"
    echo "✅ Widget wave-orch-shell ensured in $WIDGETS_FILE"
    return 0
}

# === launch_shell_block ===
# 创建一个 shell block 并返回其 blockid
# 用法: BLOCK_ID=$(launch_shell_block "$WSH")
launch_shell_block() {
    local WSH="$1"
    [[ -z "$WSH" ]] && WSH=$(resolve_wsh)

    # 记录 before
    local BEFORE=$($WSH blocks list --view=term --json 2>/dev/null | jq -r '.[].blockid' | sort)

    # 启动 shell block
    $WSH launch wave-orch-shell &>/dev/null
    sleep 0.5

    # 记录 after
    local AFTER=$($WSH blocks list --view=term --json 2>/dev/null | jq -r '.[].blockid' | sort)

    # diff 找新增的 blockid
    local NEW_BLOCK=$(comm -13 <(echo "$BEFORE") <(echo "$AFTER") | head -1)

    if [[ -z "$NEW_BLOCK" ]]; then
        echo ""
        return 1
    fi

    echo "$NEW_BLOCK"
    return 0
}

# === check_cli_available ===
# 检查 CLI 是否可用
check_cli_available() {
    local CLI="$1"
    if command -v "$CLI" &>/dev/null; then
        return 0
    fi
    return 1
}
