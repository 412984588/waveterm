# Wave-Orch 实现计划

> 版本: 0.1.0 (MVP)
> 最后更新: 2026-01-29 22:30

## 已完成任务 (17/19)

| Phase | 任务 | 状态 | Commit |
|-------|------|------|--------|
| 1 | 创建包结构 | ✅ | aa3b6394 |
| 1 | 脱敏模块 | ✅ | aa3b6394 |
| 1 | 状态机定义 | ✅ | aa3b6394 |
| 2 | wsh inject | ✅ | aa3b6394 |
| 2 | wsh output | ✅ | aa3b6394 |
| 2 | wsh wait | ✅ | aa3b6394 |
| 2 | wsh wave-orch | ✅ | aa3b6394 |
| 3 | Agent 配置 | ✅ | aa3b6394 |
| 3 | Agent 探测 | ✅ | aa3b6394 |
| 4 | 配置扫描 | ✅ | aa3b6394 |
| 4 | 诊断快照 | ✅ | aa3b6394 |
| 5 | REPORT 解析 | ✅ | aa3b6394 |
| 5 | 状态机核心 | ✅ | aa3b6394 |
| 5 | 任务分解 | ✅ | aa3b6394 |
| 6 | 项目状态 | ✅ | aa3b6394 |
| 6 | Git 分支 | ✅ | 50f59015 |
| 7 | 构建 Wave | ✅ | task build:backend |
| 8 | 日志落盘 | ✅ | aa3b6394 |
| 8 | 自动清理 | ✅ | 50f59015 |

---

## 剩余 TODO (优先级排序)

### P0: MVP 可用性

1. **E2E 验证**: 在 Wave 内运行 inject/output/wait 命令
   - 验证: `wsh inject <blockid> 'echo test'`
   - 阻塞: 需要 Wave Terminal 运行

2. **多 Agent 并行测试**: 同时启动 3 个 Agent block
   - 验证: 并行运行 claude/codex/gemini
   - 阻塞: 需要 Wave Terminal 运行

### P1: 增强功能 (MVP 后)

3. **wsh wave-orch run**: 一键启动编排任务
4. **wsh wave-orch diag**: 输出诊断快照
5. **Agent 自动切换**: 超时后切换到备用 Agent

---

## 验证命令

```bash
# 构建
task build:backend

# 单元测试
go test ./pkg/waveorch/... -v

# E2E 验证脚本
./scripts/wave-orch-verify.sh
```
