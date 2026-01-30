# Wave-Orch 状态追踪

> 最后更新: 2026-01-29 23:30

## 当前里程碑

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
