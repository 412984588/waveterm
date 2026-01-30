# Wave-Orch 状态追踪

> 最后更新: 2026-01-30 09:00

## 当前里程碑

### ✅ R1: 优化轮次1 基线验证 (2026-01-30 09:00)

**验证结果**:
```
go test ./pkg/waveorch/... -v           # ✅ 21/21 PASS
./scripts/wave_orch_e2e_smoke.sh        # ✅ PASS
./scripts/wave_orch_demo_3_agents.sh    # ✅ 3/3 PASS
./scripts/wave_orch_demo_multi_project.sh # ✅ 2/2 PASS
```

**Codex 审计**: 已注入提示，等待处理中
**wsh 路径**: `./dist/bin/wsh-0.13.2-alpha.0-darwin.arm64`

---

### ✅ M12: 幽灵任务计数修复 (2026-01-30 08:15)

**问题**: 会话恢复后显示 "[44 active, 44 pending] The boulder never stops..." 提示，但实际 TaskList 为空。

**根因分析**:

- 提示来自 oh-my-claudecode 插件的 `pre-tool-enforcer.sh` / `pre-tool-enforcer.mjs`
- 脚本从 `~/.claude/todos/*.json` 读取所有会话的任务缓存并累加计数
- 旧会话的任务文件未清理，导致计数累积

**源文件位置**:

```
~/.claude/plugins/cache/omc/oh-my-claudecode/*/scripts/pre-tool-enforcer.sh (line 24-29)
~/.claude/plugins/cache/omc/oh-my-claudecode/*/scripts/pre-tool-enforcer.mjs (line 35-88)
```

**缓存文件位置**:

```
~/.claude/todos/*.json          # 全局任务缓存（主要来源）
<project>/.omc/todos.json       # 项目本地缓存
<project>/.claude/todos.json    # 项目本地缓存（备选）
```

**修复策略**:

- 创建一键清理脚本：`~/.claude/hooks/bin/clear_task_cache.sh`
- 清理所有旧任务缓存文件
- 不修改 oh-my-claudecode 插件源码（避免升级覆盖）

**手动清理命令**:

```bash
~/.claude/hooks/bin/clear_task_cache.sh          # 执行清理
~/.claude/hooks/bin/clear_task_cache.sh --dry-run  # 预览（不删除）
```

**验证结果**:

```
# 清理前
cat ~/.claude/todos/*.json | grep -c '"pending"'     # 44
cat ~/.claude/todos/*.json | grep -c '"in_progress"' # 44

# 清理后
ls ~/.claude/todos/ | wc -l  # 0
# Hook 提示不再显示 "[44 active, 44 pending]" 前缀
```

---

### ✅ M11: Shell Block 修复 (2026-01-30 07:30)

**问题**: demo 脚本使用 `wsh run` 创建的是 `controller="cmd"` 块，不支持 inject。

**解决**: 改用 `wsh launch` + widgets.json 创建 `controller="shell"` 块。

**关键概念**:

- `controller="shell"`: 持久 shell 会话，支持 inject/output
- `controller="cmd"`: 一次性命令块，不支持 inject
- `wsh launch <widget-id>`: 从 widgets.json 创建 block
- `wsh` 需要 Wave 运行提供 access token

**修复内容**:

- [x] 新增 scripts/wave_orch_lib.sh (ensure_shell_widget, launch_shell_block)
- [x] 重写 wave_orch_demo_3_agents.sh 使用 shell block
- [x] 重写 wave_orch_demo_multi_project.sh 使用 shell block
- [x] 修复 JSON 提取（终端换行导致跨行）

**验证结果**:

```
go test ./pkg/waveorch/... -v  # ✅ 21/21 PASS
./scripts/wave_orch_e2e_smoke.sh  # ✅ PASS
./scripts/wave_orch_demo_3_agents.sh  # ✅ 3/3 PASS
./scripts/wave_orch_demo_multi_project.sh  # ✅ 2/2 PASS
```

### ✅ P0-P2 Bug 修复 (2026-01-30 05:45)

- [x] A. Kill switch fail-closed (isPaused 任何错误返回 true)
- [x] B. Log cleanup days>=1 验证 + YYYY-MM-DD 目录名解析
- [x] C. REPORT 严格验证 (project_id, round>0, summary, files_changed, commands_run)
- [x] D. 增强脱敏 (JWT, Bearer, GitLab, Slack, AWS secret, 国际电话)
- [x] E. wsh output 真实实现 (TermGetScrollbackLinesCommand)
- [x] F. wsh inject --wait 标志 (shell 就绪重试)
- [x] G. Engine 任务清理策略 (maxTasks=500, TTL=24h)

### ✅ 安全/稳定性维护 (2026-01-30)

- [x] 日志/诊断/状态落盘权限收紧 (0700/0600)
- [x] 日志 data 字段递归脱敏 (RedactAny)
- [x] REPORT 解析边界修复
- [x] SubmitTask 队列满时避免阻塞
- [x] 新增相关单元测试

### ✅ M0: 设计文档完成

- [x] DESIGN.md 完成
- [x] PLAN.md 完成
- [x] STATUS.md 创建

### ✅ M1: 基础设施

- [x] 创建 pkg/waveorch/ 目录
- [x] 实现脱敏模块 (redact.go)
- [x] 实现状态机定义 (state.go)
- [x] 测试通过

### ✅ M2: Wave Bridge Layer

- [x] wsh inject 命令
- [x] wsh output 命令
- [x] wsh wait 命令
- [x] wsh wave-orch 控制命令

### ✅ M3: Agent Registry

- [x] AgentConfig 结构
- [x] 默认 Agent 初始化
- [x] Agent 可用性检测

### ✅ M4: Config Inspector

- [x] DiagnosticSnapshot 结构
- [x] 配置扫描方法
- [x] 诊断快照保存

### ✅ M5: 核心模块

- [x] Engine 编排引擎
- [x] ProjectTracker 项目追踪
- [x] Logger 日志记录
- [x] ReportParser REPORT 解析

### ✅ M6: 日志与清理

- [x] wsh wave-orch cleanup 命令
- [x] 7天保留策略
- [x] 分支追踪

### ✅ M7: 集成验证

- [x] Wave 后端构建通过
- [x] wsh 命令可用
- [x] 12/12 单元测试通过

### ✅ M8: Bootstrap 完成

- [x] ~/.wave-orch/config.json 默认模板创建
- [x] wsh wave-orch diagnostic 命令
- [x] 端到端诊断验证通过
- [x] 全局诊断: ~/.wave-orch/logs/YYYY-MM-DD/diagnostic-\*.json
- [x] 项目诊断: <project>/.wave-orch/diagnostic.json

### ✅ M9: E2E 演示脚本

- [x] scripts/wave_orch_e2e_smoke.sh - E2E 烟雾测试
- [x] scripts/wave_orch_demo_3_agents.sh - 3-Agent 并行演示
- [x] scripts/wave_orch_demo_multi_project.sh - 多项目并行演示
- [x] wsh wave-orch demo 命令

### ✅ M10: E2E 实测验证 (2026-01-30)

- [x] wsh inject: 成功 (injected 29 bytes)
- [x] wsh run: 成功创建 block
- [x] wsh wave-orch demo: 检测到 3 个 Agent (gemini, claude-code, codex)
- [x] wsh wave-orch status: 正常 (paused: false)
- [x] wsh output: 成功读取 scrollback

---

## 构建验证

```bash
task build:backend  # ✅ 通过
go test ./pkg/waveorch/... -v  # ✅ 12/12 通过
./scripts/wave-orch-verify.sh  # ✅ 全部通过
```

## 运行 Diagnostic

```bash
# 构建 wsh
go build -o ./dist/wsh ./cmd/wsh/

# 运行诊断（指定项目路径）
./dist/wsh wave-orch diagnostic --project /path/to/your/project

# 输出文件位置
# 全局: ~/.wave-orch/logs/YYYY-MM-DD/diagnostic-YYYYMMDD-HHMMSS.json
# 项目: <project>/.wave-orch/diagnostic.json
```

## 剩余 TODO (0 项)

✅ 所有 MVP 功能已完成！

---

## 演示脚本使用

```bash
# 1. E2E 烟雾测试（需要 Wave Terminal 运行）
./scripts/wave_orch_e2e_smoke.sh

# 2. 3-Agent 并行演示
./scripts/wave_orch_demo_3_agents.sh

# 3. 多项目并行演示
./scripts/wave_orch_demo_multi_project.sh

# 4. wsh 命令
wsh wave-orch demo      # 显示可用 Agent
wsh wave-orch status    # 查看状态
wsh wave-orch pause     # 暂停编排
wsh wave-orch resume    # 恢复编排
```

---

## 已提交

- `aa3b6394` [wave-orch] Implement core modules
- `c440ea08` [wave-orch] Add unit tests
- `7e37387e` [wave-orch] Add verification script
- `50f59015` [wave-orch] Add cleanup command
- `5636d035` [wave-orch] add e2e smoke script
- `2bad4222` [wave-orch] add 3-agent and multi-project demo scripts
