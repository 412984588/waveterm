// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// AgentConfig 定义 Agent 配置
type AgentConfig struct {
	Name              string        `json:"name"`
	ExecCmd           string        `json:"exec_cmd"`
	ConfigDir         string        `json:"config_dir"`
	PromptTemplate    string        `json:"prompt_template"`
	ReportInstruction string        `json:"report_instruction"`
	Timeout           time.Duration `json:"timeout"`
	Capabilities      []string      `json:"capabilities"`
	Available         bool          `json:"available"`
}

// AgentRegistry 管理所有 Agent 配置
type AgentRegistry struct {
	agents map[string]*AgentConfig
}

// NewAgentRegistry 创建新的 Agent 注册表
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*AgentConfig),
	}
}

// DefaultReportInstruction 默认的 REPORT 格式指令
const DefaultReportInstruction = `你必须在完成每轮任务后，输出以下格式的 JSON 报告：
<<<REPORT>>>
{
  "project_id": "项目标识",
  "agent": "你的名称",
  "round": 轮次数字,
  "status": "SUCCESS|FAIL|BLOCKED|PARTIAL",
  "summary": "一句话总结",
  "actions": ["执行的操作列表"],
  "files_changed": [{"path": "文件路径", "action": "create|modify|delete", "summary": "简述"}],
  "commands_run": ["执行的命令"],
  "tests": {"passed": 0, "failed": 0, "skipped": 0},
  "risks": [],
  "needs_human": false,
  "needs_human_reason": null,
  "next_actions": ["下一步建议"]
}
<<<END_REPORT>>>`

// InitDefaultAgents 初始化默认支持的 Agent
func (r *AgentRegistry) InitDefaultAgents() {
	home, _ := os.UserHomeDir()

	// Claude Code
	r.agents["claude-code"] = &AgentConfig{
		Name:              "claude-code",
		ExecCmd:           "claude",
		ConfigDir:         filepath.Join(home, ".claude"),
		Timeout:           7 * time.Minute,
		ReportInstruction: DefaultReportInstruction,
		Capabilities:      []string{"code", "test", "review"},
	}

	// OpenAI Codex CLI
	r.agents["codex"] = &AgentConfig{
		Name:              "codex",
		ExecCmd:           "codex",
		ConfigDir:         filepath.Join(home, ".codex"),
		Timeout:           7 * time.Minute,
		ReportInstruction: DefaultReportInstruction,
		Capabilities:      []string{"code", "test"},
	}

	// Gemini CLI
	r.agents["gemini"] = &AgentConfig{
		Name:              "gemini",
		ExecCmd:           "gemini",
		ConfigDir:         filepath.Join(home, ".gemini"),
		Timeout:           7 * time.Minute,
		ReportInstruction: DefaultReportInstruction,
		Capabilities:      []string{"code", "review"},
	}
}

// DetectAvailableAgents 探测本机可用的 Agent
func (r *AgentRegistry) DetectAvailableAgents() {
	for name, agent := range r.agents {
		_, err := exec.LookPath(agent.ExecCmd)
		agent.Available = err == nil
		r.agents[name] = agent
	}
}

// GetAgent 获取指定 Agent 配置
func (r *AgentRegistry) GetAgent(name string) *AgentConfig {
	return r.agents[name]
}

// GetAvailableAgents 获取所有可用的 Agent
func (r *AgentRegistry) GetAvailableAgents() []*AgentConfig {
	var result []*AgentConfig
	for _, agent := range r.agents {
		if agent.Available {
			result = append(result, agent)
		}
	}
	return result
}
