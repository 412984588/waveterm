# Wave-Orch 实现计划

> 版本: 0.1.0 (MVP)
> 最后更新: 2026-01-29

## Phase 1: 基础设施（预计 30 分钟）

### 1.1 创建 Wave-Orch 包结构

- **文件**: `pkg/waveorch/` 目录
- **验证**: `ls pkg/waveorch/`
- **完成标准**: 目录存在

### 1.2 实现脱敏模块

- **文件**: `pkg/waveorch/redact.go`
- **验证**: `go test ./pkg/waveorch/... -run TestRedact`
- **完成标准**: 测试通过

### 1.3 实现状态机定义

- **文件**: `pkg/waveorch/state.go`
- **验证**: `go build ./pkg/waveorch/...`
- **完成标准**: 编译通过

## Phase 2: Wave Bridge Layer（预计 45 分钟）

### 2.1 实现 wsh inject 命令

- **文件**: `cmd/wsh/cmd/wshcmd-inject.go`
- **验证**: `go build ./cmd/wsh && ./bin/wsh inject --help`
- **完成标准**: 帮助信息显示正确

### 2.2 实现 wsh output 命令

- **文件**: `cmd/wsh/cmd/wshcmd-output.go`
- **验证**: `go build ./cmd/wsh && ./bin/wsh output --help`
- **完成标准**: 帮助信息显示正确

### 2.3 实现 wsh wait 命令

- **文件**: `cmd/wsh/cmd/wshcmd-wait.go`
- **验证**: `go build ./cmd/wsh && ./bin/wsh wait --help`
- **完成标准**: 帮助信息显示正确

### 2.4 实现 wsh wave-orch 命令

- **文件**: `cmd/wsh/cmd/wshcmd-waveorch.go`
- **验证**: `go build ./cmd/wsh && ./bin/wsh wave-orch --help`
- **完成标准**: 帮助信息显示正确

## Phase 3: Agent Registry（预计 20 分钟）

### 3.1 实现 Agent 配置结构

- **文件**: `pkg/waveorch/registry.go`
- **验证**: `go build ./pkg/waveorch/...`
- **完成标准**: 编译通过

### 3.2 实现 Agent 自动探测

- **文件**: `pkg/waveorch/registry.go`
- **验证**: `go test ./pkg/waveorch/... -run TestAgentDetect`
- **完成标准**: 能检测到本机已安装的 CLI

## Phase 4: Config Inspector（预计 20 分钟）

### 4.1 实现配置扫描

- **文件**: `pkg/waveorch/config.go`
- **验证**: `go build ./pkg/waveorch/...`
- **完成标准**: 能读取 ~/.claude/ 等目录

### 4.2 实现诊断快照生成

- **文件**: `pkg/waveorch/config.go`
- **验证**: `go test ./pkg/waveorch/... -run TestDiagnostic`
- **完成标准**: 生成脱敏后的 JSON

## Phase 5: Orchestration Engine（预计 45 分钟）

### 5.1 实现 REPORT 解析器

- **文件**: `pkg/waveorch/reporter.go`
- **验证**: `go test ./pkg/waveorch/... -run TestReportParse`
- **完成标准**: 能从文本中提取 JSON

### 5.2 实现状态机核心

- **文件**: `pkg/waveorch/engine.go`
- **验证**: `go test ./pkg/waveorch/... -run TestStateMachine`
- **完成标准**: 状态转换正确

### 5.3 实现任务分解逻辑

- **文件**: `pkg/waveorch/engine.go`
- **验证**: `go build ./pkg/waveorch/...`
- **完成标准**: 编译通过

## Phase 6: Project Tracker（预计 15 分钟）

### 6.1 实现项目状态管理

- **文件**: `pkg/waveorch/tracker.go`
- **验证**: `go build ./pkg/waveorch/...`
- **完成标准**: 编译通过

### 6.2 实现 Git 分支管理

- **文件**: `pkg/waveorch/tracker.go`
- **验证**: `go test ./pkg/waveorch/... -run TestGitBranch`
- **完成标准**: 能创建/删除分支

## Phase 7: 集成测试（预计 30 分钟）

### 7.1 构建 Wave Terminal

- **命令**: `task build:backend`
- **验证**: `ls bin/`
- **完成标准**: wsh 二进制存在

### 7.2 端到端测试

- **命令**: `task dev`
- **验证**: 手动在 Wave 内测试 wsh 命令
- **完成标准**: inject/output/wait 命令可用

### 7.3 多 Agent 并行测试

- **验证**: 同时启动 3 个 Agent block
- **完成标准**: 能并行运行

## Phase 8: 日志与清理（预计 15 分钟）

### 8.1 实现日志落盘

- **文件**: `pkg/waveorch/reporter.go`
- **验证**: 检查 `~/.wave-orch/logs/` 目录
- **完成标准**: 日志文件生成

### 8.2 实现自动清理

- **文件**: `pkg/waveorch/reporter.go`
- **验证**: `wsh wave-orch cleanup --days=7`
- **完成标准**: 过期日志被删除

---

## 执行顺序

1. Phase 1 → Phase 2 → Phase 3 → Phase 4
2. Phase 5 → Phase 6
3. Phase 7 → Phase 8

每个 Phase 完成后 commit 一次
