# Wave-Orch 架构设计文档

> 版本: 0.1.0 (MVP)
> 最后更新: 2026-01-29

## 1. 系统概述

Wave-Orch 是构建在 Wave Terminal 上的多 Agent CLI 工具自动编排系统。通过魔改 Wave Terminal，实现：

- **多项目并行**：同时管理多个项目的编排任务
- **多 Agent 并行**：Claude Code / OpenAI Codex CLI / Gemini CLI 在 Wave 内多窗口并行
- **指挥官自动化**：状态机驱动，默认无需人工批准
- **结构化回合制**：JSON REPORT 协议，支持任务转交

## 2. 架构总览

```
┌─────────────────────────────────────────────────────────────────┐
│                        Wave Terminal (魔改版)                     │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ Agent Block │  │ Agent Block │  │ Agent Block │   ...        │
│  │ (Claude)    │  │ (Codex)     │  │ (Gemini)    │              │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘              │
│         │                │                │                      │
│         └────────────────┼────────────────┘                      │
│                          │                                       │
│  ┌───────────────────────▼───────────────────────┐              │
│  │            Wave Bridge Layer (wsh 扩展)         │              │
│  │  - 列出/创建/删除 blocks                        │              │
│  │  - 注入文本 + Enter (ControllerInputCommand)   │              │
│  │  - 获取输出 (TermGetScrollbackLinesCommand)    │              │
│  └───────────────────────┬───────────────────────┘              │
├──────────────────────────┼──────────────────────────────────────┤
│                          │                                       │
│  ┌───────────────────────▼───────────────────────┐              │
│  │           Orchestration Engine (指挥官)         │              │
│  │  - 状态机驱动                                   │              │
│  │  - 任务分解与指派                               │              │
│  │  - REPORT 收集与仲裁                           │              │
│  │  - 超时/错误处理                               │              │
│  └───────────────────────┬───────────────────────┘              │
│                          │                                       │
│  ┌───────────────────────▼───────────────────────┐              │
│  │              支撑模块                           │              │
│  │  ┌─────────────┐  ┌─────────────┐             │              │
│  │  │Agent Registry│  │Config Inspector│           │              │
│  │  └─────────────┘  └─────────────┘             │              │
│  │  ┌─────────────┐  ┌─────────────┐             │              │
│  │  │Project Tracker│ │Output Reporter│           │              │
│  │  └─────────────┘  └─────────────┘             │              │
│  └───────────────────────────────────────────────┘              │
└─────────────────────────────────────────────────────────────────┘
```

## 3. Wave Bridge Layer 实现方案

### 3.1 设计原则

**最小侵入**：优先复用 Wave 现有的 wsh 体系，仅在必要时扩展

### 3.2 现有 wsh 能力分析

通过源码分析，Wave 的 wsh 已提供以下能力：

| 能力               | wsh 命令/API                          | 状态    |
| ------------------ | ------------------------------------- | ------- |
| 列出 blocks        | `wsh blocks list --json`              | ✅ 已有 |
| 创建终端 block     | `wsh term [dir]`                      | ✅ 已有 |
| 创建运行命令 block | `wsh run -- cmd args`                 | ✅ 已有 |
| 删除 block         | `wsh deleteblock`                     | ✅ 已有 |
| 注入输入           | `ControllerInputCommand` (RPC)        | ✅ 已有 |
| 获取输出           | `TermGetScrollbackLinesCommand` (RPC) | ✅ 已有 |

### 3.3 需要新增的 wsh 子命令

为了让 Orchestration Engine 更方便地调用，新增以下 wsh 子命令：

```bash
# 1. 向指定 block 注入文本并发送 Enter
wsh inject <blockid> "text to inject"
wsh inject <blockid> --file=prompt.txt

# 2. 获取指定 block 的最近 N 行输出
wsh output <blockid> --lines=100
wsh output <blockid> --last-command  # 获取最后一条命令的输出

# 3. 等待 block 输出包含特定模式（用于等待 REPORT）
wsh wait <blockid> --pattern="<<<END_REPORT>>>" --timeout=420000

# 4. Wave-Orch 控制命令
wsh wave-orch status          # 查看编排状态
wsh wave-orch pause           # 暂停所有自动注入
wsh wave-orch resume          # 恢复自动注入
wsh wave-orch kill            # 立即停止所有 Agent
```

### 3.4 实现文件规划

```
cmd/wsh/cmd/
├── wshcmd-inject.go      # 新增：注入文本命令
├── wshcmd-output.go      # 新增：获取输出命令
├── wshcmd-wait.go        # 新增：等待模式命令
└── wshcmd-waveorch.go    # 新增：编排控制命令

pkg/waveorch/             # 新增：Wave-Orch 核心包
├── engine.go             # Orchestration Engine
├── registry.go           # Agent Registry
├── config.go             # Config Inspector
├── tracker.go            # Project Tracker
├── reporter.go           # Output Reporter
├── state.go              # 状态机定义
└── redact.go             # 脱敏处理
```

## 4. Orchestration Engine（编排引擎）

### 4.1 状态机定义

```
                    ┌─────────────┐
                    │    IDLE     │
                    └──────┬──────┘
                           │ start_task()
                           ▼
                    ┌─────────────┐
                    │  PLANNING   │ ← 任务分解
                    └──────┬──────┘
                           │
                           ▼
                    ┌─────────────┐
              ┌────►│  EXECUTING  │◄────┐
              │     └──────┬──────┘     │
              │            │            │
              │            ▼            │
              │     ┌─────────────┐     │
              │     │ COLLECTING  │     │ 等待 REPORT
              │     └──────┬──────┘     │
              │            │            │
              │            ▼            │
              │     ┌─────────────┐     │
              └─────┤  DECIDING   ├─────┘ 需要下一轮
                    └──────┬──────┘
                           │ 任务完成
                           ▼
                    ┌─────────────┐
                    │  COMPLETED  │
                    └─────────────┘

特殊状态：
- PAUSED: Kill Switch 触发时进入
- ERROR: 不可恢复错误
- BLOCKED: 需要人工决策
```

### 4.2 核心接口

```go
type OrchestrationEngine interface {
    // 启动新任务
    StartTask(projectPath string, userRequest string) (taskId string, err error)

    // 获取任务状态
    GetTaskStatus(taskId string) (*TaskStatus, error)

    // 暂停/恢复
    Pause() error
    Resume() error

    // Kill Switch
    EmergencyStop() error

    // 内部方法
    advanceState() error
    handleReport(agentName string, report *AgentReport) error
    assignTask(agent *AgentConfig, subtask *Subtask) error
}
```

## 5. REPORT 协议

### 5.1 格式规范

每个 Agent 完成一轮任务后，必须输出以下格式的 JSON：

```
<<<REPORT>>>
{
  "project_id": "hash_of_project_path",
  "agent": "claude-code",
  "round": 1,
  "status": "SUCCESS",
  "summary": "完成了用户认证模块的实现",
  "actions": [
    "创建 src/auth/login.ts",
    "修改 src/routes/index.ts 添加路由"
  ],
  "files_changed": [
    {"path": "src/auth/login.ts", "action": "create", "summary": "登录逻辑"},
    {"path": "src/routes/index.ts", "action": "modify", "summary": "添加 /login 路由"}
  ],
  "commands_run": [
    "npm install bcrypt",
    "npm test"
  ],
  "tests": {
    "passed": 12,
    "failed": 0,
    "skipped": 2
  },
  "risks": [],
  "needs_human": false,
  "needs_human_reason": null,
  "next_actions": [
    "实现密码重置功能",
    "添加 JWT token 刷新"
  ]
}
<<<END_REPORT>>>
```

### 5.2 状态值定义

| status  | 含义                     |
| ------- | ------------------------ |
| SUCCESS | 任务成功完成             |
| FAIL    | 任务失败，需要重试或转交 |
| BLOCKED | 遇到阻塞，需要人工决策   |
| PARTIAL | 部分完成，可继续         |

### 5.3 解析策略

Orchestration Engine 通过以下方式提取 REPORT：

1. 监听 block 输出（`TermGetScrollbackLinesCommand`）
2. 正则匹配 `<<<REPORT>>>` 和 `<<<END_REPORT>>>` 之间的内容
3. JSON 解析并验证必填字段
4. 失败时记录原始输出供调试

## 6. Agent Registry（Agent 注册表）

### 6.1 支持的 Agent

| Agent            | 命令     | 配置目录     |
| ---------------- | -------- | ------------ |
| Claude Code      | `claude` | `~/.claude/` |
| OpenAI Codex CLI | `codex`  | `~/.codex/`  |
| Gemini CLI       | `gemini` | `~/.gemini/` |

### 6.2 Agent 配置结构

```go
type AgentConfig struct {
    Name        string            `json:"name"`
    ExecCmd     string            `json:"exec_cmd"`
    ConfigDir   string            `json:"config_dir"`
    PromptTemplate string         `json:"prompt_template"`
    ReportInstruction string      `json:"report_instruction"`
    Timeout     time.Duration     `json:"timeout"`
    Capabilities []string         `json:"capabilities"`
}
```

### 6.3 REPORT 指令注入

每个 Agent 启动时，自动注入 REPORT 格式要求：

```
你必须在完成每轮任务后，输出以下格式的 JSON 报告：
<<<REPORT>>>
{ ... }
<<<END_REPORT>>>
```

## 7. Config Inspector（配置检查器）

### 7.1 扫描目录

| Agent       | 扫描路径                                                                 |
| ----------- | ------------------------------------------------------------------------ |
| Claude Code | `~/.claude/CLAUDE.md`, `~/.claude/rules/*.md`, `~/.claude/settings.json` |
| Codex CLI   | `~/.codex/config.yaml`, `~/.codex/instructions.md`                       |
| Gemini CLI  | `~/.gemini/config.json`                                                  |
| 项目级      | `.cursorrules`, `CLAUDE.md`, `AGENTS.md`, `.wave-orch/config.json`       |

### 7.2 诊断快照

启动时生成诊断快照，保存到：

- `~/.wave-orch/logs/YYYY-MM-DD/diagnostic.json`
- `<project>/.wave-orch/diagnostic.json`

## 8. 日志与记录结构

### 8.1 目录结构

```
~/.wave-orch/
├── config.json              # 全局配置
├── logs/
│   └── 2026-01-29/
│       ├── diagnostic.json  # 诊断快照
│       └── task-*.json      # 任务日志
└── state/
    └── paused               # Kill Switch 状态文件

<project>/.wave-orch/
├── config.json              # 项目配置
├── state.json               # 项目状态
├── diagnostic.json          # 项目诊断
└── history/
    └── task-*.json          # 历史任务
```

### 8.2 日志保留策略

- 默认保留 7 天
- 每日凌晨自动清理过期日志
- 清理命令：`wsh wave-orch cleanup --days=7`

## 9. 脱敏策略

### 9.1 敏感信息模式

| 类型           | 正则模式                                         | 替换为                  |
| -------------- | ------------------------------------------------ | ----------------------- |
| API Key (通用) | `[a-zA-Z0-9_-]{20,}` (上下文判断)                | `***REDACTED***`        |
| OpenAI Key     | `sk-[a-zA-Z0-9]{48}`                             | `sk-***REDACTED***`     |
| Anthropic Key  | `sk-ant-[a-zA-Z0-9-]{95}`                        | `sk-ant-***REDACTED***` |
| 邮箱           | `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}` | `***@***.***`           |
| 手机号         | `1[3-9]\d{9}`                                    | `1**********`           |

### 9.2 脱敏时机

1. **写入日志前**：所有日志内容必须经过脱敏
2. **生成 REPORT 前**：Agent 输出的 REPORT 脱敏后再存储
3. **诊断快照**：配置信息脱敏后再保存

## 10. Kill Switch（安全熔断）

### 10.1 触发方式

| 方式     | 命令/操作                        |
| -------- | -------------------------------- |
| CLI 命令 | `wsh wave-orch pause`            |
| 环境变量 | `WAVE_ORCH_PAUSED=1`             |
| 状态文件 | 创建 `~/.wave-orch/state/paused` |
| 快捷键   | Wave Terminal 内 `Ctrl+Shift+K`  |

### 10.2 暂停行为

触发 Kill Switch 后：

1. 立即停止所有待发送的注入操作
2. 不中断正在运行的 Agent（避免数据丢失）
3. 状态机进入 PAUSED 状态
4. 在 Wave Terminal 侧边栏显示暂停提示

### 10.3 恢复

```bash
wsh wave-orch resume
# 或删除状态文件
rm ~/.wave-orch/state/paused
```

## 11. Git 工作流

### 11.1 分支策略

```
main
 │
 ├── wave-orch/task-{taskId}/claude-code
 ├── wave-orch/task-{taskId}/codex
 └── wave-orch/task-{taskId}/gemini
```

### 11.2 合并流程

1. 每个 Agent 在独立分支提交
2. 任务完成后，Claude Code 负责合并
3. 自动尝试 merge，冲突时报告

### 11.3 清理

任务完成后自动删除临时分支（可配置保留）

## 12. Project Tracker（项目跟踪器）

### 12.1 多项目隔离

每个项目独立维护：

- 状态机实例
- Agent 会话组
- Git 分支
- 日志目录

### 12.2 项目配置

`<project>/.wave-orch/config.json`:

```json
{
  "preferred_agents": ["claude-code", "codex"],
  "excluded_paths": ["node_modules", ".git"],
  "timeout_ms": 420000,
  "auto_merge": true
}
```

## 13. 验收标准（MVP DoD）

1. ✅ 本地构建并运行魔改后的 Wave Terminal
2. ✅ 在 Wave 内能启动 3 个 Agent 窗口并行执行
3. ✅ 指挥官能注入文本、抓取输出、解析 REPORT JSON
4. ✅ 同时对至少 2 个项目跑通一轮编排
5. ✅ 自动分支提交与最终合并 main 跑通
6. ✅ Kill Switch 可立即暂停所有自动注入
