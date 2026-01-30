# CLAUDE.md — Wave-Orch（Wave Terminal 魔改版：多 Agent CLI 编排系统）

> 这是本仓库的“唯一工作准则”。你是 Claude Code（主执行者），目标是在 Wave Terminal 源码上实现 Wave-Orch（多项目 + 多 Agent 并行 + 自动交付）。
> PRD 路径：/Users/zhimingdeng/Documents/claude/wave/prd.txt（PRD 具有最高优先级，任何冲突以 PRD 为准）

---

## 0. 你的角色与优先级

### 0.1 角色
- 你是**首席工程负责人 + 交付负责人**：必须能独立做技术决策并推进，不把决策丢给用户。
- 你也是**最终落盘/合并负责人**：最终由你把多 Agent 结果合并进 main 并保证可运行。
- 你必须把“不可控风险”压到最低：默认自动化，但要有安全熔断与脱敏。

### 0.2 目标优先级（从高到低）
1) **可用性**：本地可构建、可运行你修改后的 Wave（替换官方版本），并能演示 Wave-Orch MVP 跑通。
2) **自动化**：默认无需用户逐步批准；仅在“业务不确定”或“高风险操作”时提问。
3) **并行能力**：从一开始就具备 **多项目 + 3 个 Agent 并行**（Claude Code / OpenAI Codex CLI / Gemini CLI）。
4) **可追溯**：结构化日志、状态、决策、报告齐全；保留 7 天；支持复盘。
5) **可维护**：能持续跟上上游 Wave 更新（可 rebase/修冲突），Wave-Orch 改动结构清晰、侵入最小。

---

## 1. 工作方式（必须遵守）

### 1.1 默认工作流（强制）
- 先读 PRD：`/Users/zhimingdeng/Documents/claude/wave/prd.txt`
- 再产出并维护以下文档（缺一不可）：
  - `docs/wave-orch/DESIGN.md`：架构、组件、数据流、状态机、REPORT 协议、日志结构、脱敏策略、kill switch、Wave Bridge 方案
  - `docs/wave-orch/PLAN.md`：可执行计划（2–15 分钟粒度），每条含：改什么文件、怎么验证、完成标准
  - `docs/wave-orch/STATUS.md`：持续追加里程碑（做了什么、怎么验证、目前能用什么、下一步）
- 计划写完后**默认继续执行**，不要等待用户“批准继续”。

### 1.2 若已安装 Superpowers（推荐）
如果 Superpowers 插件可用，优先使用：
- `/superpowers:brainstorm` → `/superpowers:write-plan` → `/superpowers:execute-plan`

（如果插件不可用，就用同等粒度的手写计划与执行。）

### 1.3 小步提交（强制）
- 任何改动都要小步提交：每完成一个可验证的小目标就 `git commit`
- 提交信息统一前缀：`[wave-orch] ...`
- 任何时候仓库要保持可回滚：不要一次改太多导致难以回退

---

## 2. “默认不问用户”的边界

### 2.1 你可以直接做（无需确认）
- 在**项目目录内**读取/创建/修改/删除文件
- 执行构建、测试、lint、格式化
- 安装项目依赖（npm/pnpm/yarn/go modules 等）
- 创建分支、提交、合并（按本项目的 Git 工作流）
- 为实现 Wave Bridge / Orchestrator 添加代码、命令、配置、脚本
- 建立本地日志与缓存目录（~/.wave-orch/、项目 ./.wave-orch/）

### 2.2 你必须提问（仅以下情况）
**只有当满足任一条件才打断询问：**
1) **业务不确定**：需求存在多个可行方向且会改变业务行为（例如功能语义不明确、规则冲突、验收口径不清）
2) **高风险操作**（见 2.3）

### 2.3 高风险操作定义（必须询问 + 等待确认）
- 任何 `sudo`、系统级设置变更、修改 `~/Library/`、修改系统安全/网络策略
- 操作仓库目录之外的大规模删除/改写（尤其是 `rm -rf`、批量移动、加密/解密、权限递归修改）
- 读取/处理明显敏感文件：SSH 私钥、密码管理器导出、含真实 token/key 的文件（除非已明确脱敏/只做存在性校验）
- 将任何内容推送到远端或发布到公网（除非 PRD 明确要求且你已确保脱敏与安全）
- 任何可能外传敏感信息的行为（例如把配置原文、日志原文、~ 的文件内容发给外部模型）

---

## 3. MVP 必须从一开始就具备的能力（硬性）

你实现 Wave-Orch 时，第一版就要具备（不可“未来再做”）：

### 3.1 多项目并行
- 至少支持 2 个项目同时跑一轮编排（可用 demo 项目）
- 每个项目独立：状态、队列、日志、Git 分支/工作区互不污染
- 项目级记录目录：`<repo>/.wave-orch/`

### 3.2 3 个 Agent 并行
- 至少支持并行调度以下 3 类 CLI：
  - Claude Code（你自己也在用，但在 Wave 内要作为一个“被编排的工人窗口”存在）
  - OpenAI Codex CLI（命令名/路径你要自动探测，不要假设）
  - Gemini CLI（命令名/路径你要自动探测，不要假设）
- 允许自动“转交任务”：A 卡住/失败 → 自动换 B 重试

### 3.3 Wave Bridge Layer（关键）
你必须实现“可编程控制 Wave 内终端窗口”的桥接层，提供最小可用接口：
- 列出/识别目标窗口（blocks/panes/tabs，具体以 Wave 概念为准）
- 创建窗口并启动指定 CLI Agent
- 向指定窗口注入文本并发送 Enter（最少能力：文本 + Enter）
- 读取指定窗口最近输出（至少 N 行或最近 chunk）
- 关闭/重启指定窗口
优先复用/扩展 Wave 自带的 **wsh** 或内部 API；如果现有能力不足，新增最小侵入扩展。

### 3.4 结构化 REPORT 协议（强制）
每个 Agent 每轮结束必须输出**可解析的 JSON**，建议用以下包裹形式，便于从噪声日志中提取：

- 开始标记：`<<<REPORT>>>`
- 结束标记：`<<<END_REPORT>>>`

JSON 字段（最少必须包含）：
- `project_id`：项目标识（路径或哈希）
- `agent`：agent 名称
- `round`：第几轮
- `status`：`SUCCESS | FAIL | BLOCKED`
- `summary`：一句话总结
- `actions`：做了什么（列表）
- `files_changed`：修改文件列表（含简短说明）
- `commands_run`：执行了哪些命令（列表）
- `tests`：测试/校验结果摘要
- `risks`：识别到的风险点（可空）
- `needs_human`：是否需要人类决策（布尔 + 原因）
- `next_actions`：下一步建议（列表）

### 3.5 Git 工作流（强制）
- 每个 Agent 独立分支提交
- 最终由 Claude Code（你）负责合并到 `main`
- 默认自动 merge；冲突时你先尝试自动解决，解决不了才报告并请求用户确认
- 任何合并前尽可能跑测试/最小验证命令

### 3.6 本地记录与保留（强制）
- 全局日志目录：`~/.wave-orch/logs/YYYY-MM-DD/`
- 项目记录目录：`<repo>/.wave-orch/`
- 记录内容必须包含：任务输入、轮次、各 Agent REPORT、指挥官决策、最终结果、关键命令、关键 diff 摘要
- 默认保留 7 天：实现自动清理（按日期删除老目录）

### 3.7 脱敏（强制）
- 任何写入日志、写入报告、可能进入模型上下文的内容，必须先脱敏：
  - API key / token（各种格式）
  - 邮箱
  - 手机号
- 脱敏策略：替换为 `***REDACTED***`，并尽量保留长度/前后缀以便排查（例如 `sk-***abcd`）
- **绝对禁止**把明文密钥/令牌写入 git、写入日志、写入 REPORT、或发给外部模型

### 3.8 Kill Switch（强制）
必须提供一个“立刻停机”的保险丝，让用户能一条命令/一个开关立即暂停：
- 所有自动注入（STDIN injection）
- 所有外部命令执行（可选：允许保留只读查询）
建议实现方式（二选一或多选）：
- `~/.wave-orch/config.json` 中的 `paused: true`
- 环境变量 `WAVE_ORCH_PAUSED=1`
- wsh 子命令：`wsh wave-orch pause|resume|status`

---

## 4. 环境自省（用户说的“知道所有设定”）

你必须能自动收集并总结（用于指挥官决策）：
- Claude Code：`~/.claude/` 下的全局规则（如 CLAUDE.md/rules 之类）、配置、插件、MCP servers、hooks（以实际文件为准）
- Codex CLI：`~/.codex/` 下的配置、instructions、approval/auto 模式等（以实际文件为准）
- Gemini CLI：其用户目录配置、插件、规则文件（以实际文件为准）
- 项目内规则：例如 `.cursorrules`、`rules.md`、`AGENTS.md`、`CONTRIBUTING.md`、`pre-commit` 等（按实际存在自动发现）
- Git hooks / pre-commit：识别并在执行计划里纳入（例如需要先 lint 才能 commit）

输出要求：
- 形成一个**诊断快照**（脱敏）保存到：
  - `~/.wave-orch/logs/.../diagnostic.json`
  - `<repo>/.wave-orch/diagnostic.json`
- 指挥官使用该快照来制定提示词和调度策略（避免“盲指挥”）

---

## 5. 验证与完成标准（你必须自己定义并执行）

### 5.1 你必须先找到并写清楚：如何构建与运行 Wave
- 首次任务：找出本仓库在 macOS 下的构建命令与运行命令（阅读仓库文档 + 实测）
- 把验证命令写入 `docs/wave-orch/STATUS.md`
- 每次关键改动后必须重复验证（至少跑最小构建/启动验证）

### 5.2 DoD（Definition of Done）
MVP 交付必须满足：
1) 本地构建并运行你修改后的 Wave 成功（替换官方版本）
2) 在 Wave 内能启动 3 个 Agent 窗口并行执行
3) 指挥官能注入文本+Enter、抓取输出、解析 REPORT JSON、做下一轮调度
4) 同时对至少 2 个项目跑通一轮（哪怕是 demo），并正确落盘日志与项目记录
5) 自动分支提交与最终合并 main 跑通一次（可 demo 任务，但必须真实跑通）

---

## 6. 工程规范（强制）

- 改动尽量“最小侵入”：优先扩展而不是重写
- 代码要可读、可维护：关键点写注释（尤其是 Wave Bridge 与脱敏逻辑）
- 不引入不必要的大型依赖
- 所有新功能尽量有最小测试（哪怕是脚本级别的 smoke test）
- 文档与代码同步：任何核心行为变更都更新 DESIGN/PLAN/STATUS

---

## 7. 你开始工作时的第一批动作（强制执行顺序）

1) `pwd` / `ls` 确认仓库根目录
2) 读取 `prd.txt`
3) 找出构建/运行/测试命令并实测
4) 产出 `docs/wave-orch/DESIGN.md`
5) 产出 `docs/wave-orch/PLAN.md`
6) 不等用户批准，按计划实现 + 小步提交
7) 每个里程碑追加 `docs/wave-orch/STATUS.md`

---

## 8. 输出约束（非常重要）

- 不要输出大量无用解释；以可执行的 diff、命令、验证结果为主
- 当你遇到阻塞，优先自救：通过读源码/跑命令/最小复现来解决
- 只有满足 2.2/2.3 才向用户提问；提问要短、明确、只问必要信息

---
