// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// DiagnosticSnapshot 诊断快照
type DiagnosticSnapshot struct {
	Timestamp time.Time              `json:"timestamp"`
	Agents    map[string]AgentDiag   `json:"agents"`
	Project   *ProjectDiag           `json:"project,omitempty"`
}

// AgentDiag Agent 诊断信息
type AgentDiag struct {
	Available   bool     `json:"available"`
	Version     string   `json:"version,omitempty"`
	ConfigFound bool     `json:"config_found"`
	Plugins     []string `json:"plugins,omitempty"`
	RulesCount  int      `json:"rules_count,omitempty"`
}

// ProjectDiag 项目诊断信息
type ProjectDiag struct {
	Path              string `json:"path"`
	GitBranch         string `json:"git_branch,omitempty"`
	HasWaveOrchConfig bool   `json:"has_wave_orch_config"`
}

// ConfigInspector 配置检查器
type ConfigInspector struct {
	registry *AgentRegistry
}

// NewConfigInspector 创建配置检查器
func NewConfigInspector(registry *AgentRegistry) *ConfigInspector {
	return &ConfigInspector{registry: registry}
}

// ScanAgentConfig 扫描单个 Agent 的配置
func (ci *ConfigInspector) ScanAgentConfig(agentName string) AgentDiag {
	agent := ci.registry.GetAgent(agentName)
	if agent == nil {
		return AgentDiag{Available: false, ConfigFound: false}
	}

	diag := AgentDiag{
		Available:   agent.Available,
		ConfigFound: false,
	}

	// 检查配置目录是否存在
	if _, err := os.Stat(agent.ConfigDir); err == nil {
		diag.ConfigFound = true
		// 扫描规则文件数量
		diag.RulesCount = ci.countRulesFiles(agent.ConfigDir)
		// 扫描插件
		diag.Plugins = ci.scanPlugins(agent.ConfigDir)
	}

	return diag
}

// countRulesFiles 统计规则文件数量
func (ci *ConfigInspector) countRulesFiles(configDir string) int {
	count := 0
	rulesDir := filepath.Join(configDir, "rules")
	if entries, err := os.ReadDir(rulesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
				count++
			}
		}
	}
	// 检查根目录的 CLAUDE.md 等文件
	rootFiles := []string{"CLAUDE.md", "instructions.md", "config.json"}
	for _, f := range rootFiles {
		if _, err := os.Stat(filepath.Join(configDir, f)); err == nil {
			count++
		}
	}
	return count
}

// scanPlugins 扫描插件目录
func (ci *ConfigInspector) scanPlugins(configDir string) []string {
	var plugins []string
	pluginsDir := filepath.Join(configDir, "plugins")
	if entries, err := os.ReadDir(pluginsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				plugins = append(plugins, entry.Name())
			}
		}
	}
	return plugins
}

// GenerateDiagnostic 生成诊断快照
func (ci *ConfigInspector) GenerateDiagnostic(projectPath string) *DiagnosticSnapshot {
	snapshot := &DiagnosticSnapshot{
		Timestamp: time.Now(),
		Agents:    make(map[string]AgentDiag),
	}

	// 扫描所有已注册的 Agent
	for _, agentName := range []string{"claude-code", "codex", "gemini"} {
		snapshot.Agents[agentName] = ci.ScanAgentConfig(agentName)
	}

	// 扫描项目信息
	if projectPath != "" {
		snapshot.Project = ci.scanProject(projectPath)
	}

	return snapshot
}

// scanProject 扫描项目信息
func (ci *ConfigInspector) scanProject(projectPath string) *ProjectDiag {
	diag := &ProjectDiag{
		Path: projectPath,
	}

	// 检查是否有 .wave-orch 目录
	waveOrchDir := filepath.Join(projectPath, ".wave-orch")
	if _, err := os.Stat(waveOrchDir); err == nil {
		diag.HasWaveOrchConfig = true
	}

	// 获取 Git 分支（简单实现）
	gitHeadPath := filepath.Join(projectPath, ".git", "HEAD")
	if data, err := os.ReadFile(gitHeadPath); err == nil {
		content := string(data)
		if len(content) > 16 && content[:16] == "ref: refs/heads/" {
			diag.GitBranch = content[16 : len(content)-1] // 去掉换行符
		}
	}

	return diag
}

// SaveDiagnostic 保存诊断快照到文件
func (ci *ConfigInspector) SaveDiagnostic(snapshot *DiagnosticSnapshot, outputPath string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0600)
}

// GetGlobalLogDir 获取全局日志目录
func GetGlobalLogDir() string {
	home, _ := os.UserHomeDir()
	today := time.Now().Format("2006-01-02")
	return filepath.Join(home, ".wave-orch", "logs", today)
}

// GetProjectOrchDir 获取项目编排目录
func GetProjectOrchDir(projectPath string) string {
	return filepath.Join(projectPath, ".wave-orch")
}
