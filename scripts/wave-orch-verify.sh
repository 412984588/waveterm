#!/bin/bash
# Wave-Orch 功能验证脚本
# 用于验证核心模块是否正常工作

set -e

echo "=== Wave-Orch 功能验证 ==="
echo ""

# 1. 检查 wsh 命令
WSH="./dist/bin/wsh-0.13.2-alpha.0-darwin.arm64"
if [[ ! -f "$WSH" ]]; then
    echo "❌ wsh 未找到，请先运行 task build:backend"
    exit 1
fi
echo "✅ wsh 已构建"

# 2. 验证新命令存在
echo ""
echo "--- 验证新命令 ---"
$WSH inject --help > /dev/null 2>&1 && echo "✅ inject 命令可用"
$WSH output --help > /dev/null 2>&1 && echo "✅ output 命令可用"
$WSH wait --help > /dev/null 2>&1 && echo "✅ wait 命令可用"
$WSH wave-orch --help > /dev/null 2>&1 && echo "✅ wave-orch 命令可用"

# 3. 验证 wave-orch 子命令
echo ""
echo "--- 验证 wave-orch 子命令 ---"
$WSH wave-orch status > /dev/null 2>&1 && echo "✅ wave-orch status 可用"
$WSH wave-orch pause --help > /dev/null 2>&1 && echo "✅ wave-orch pause 可用"
$WSH wave-orch resume --help > /dev/null 2>&1 && echo "✅ wave-orch resume 可用"

# 4. 运行单元测试
echo ""
echo "--- 运行单元测试 ---"
go test ./pkg/waveorch/... -v 2>&1 | tail -5

echo ""
echo "=== 验证完成 ==="
