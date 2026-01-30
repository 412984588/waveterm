// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ProjectTracker 项目追踪器
type ProjectTracker struct {
	mu       sync.RWMutex
	projects map[string]*ProjectState
}

// ProjectState 项目状态
type ProjectState struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Branch      string    `json:"branch"`
	LastRound   int       `json:"last_round"`
	TotalRounds int       `json:"total_rounds"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"started_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Reports     []Report  `json:"reports,omitempty"`
}

// NewProjectTracker 创建项目追踪器
func NewProjectTracker() *ProjectTracker {
	return &ProjectTracker{
		projects: make(map[string]*ProjectState),
	}
}

// RegisterProject 注册项目
func (pt *ProjectTracker) RegisterProject(projectPath string) *ProjectState {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if state, ok := pt.projects[projectPath]; ok {
		return state
	}

	state := &ProjectState{
		Path:      projectPath,
		Name:      filepath.Base(projectPath),
		Status:    "registered",
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	pt.projects[projectPath] = state
	return state
}

// GetProject 获取项目状态
func (pt *ProjectTracker) GetProject(projectPath string) *ProjectState {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.projects[projectPath]
}

// UpdateRound 更新轮次
func (pt *ProjectTracker) UpdateRound(projectPath string, round int, report *Report) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	state, ok := pt.projects[projectPath]
	if !ok {
		return
	}
	state.LastRound = round
	state.TotalRounds++
	state.UpdatedAt = time.Now()
	if report != nil {
		state.Reports = append(state.Reports, *report)
	}
}

// SetStatus 设置项目状态
func (pt *ProjectTracker) SetStatus(projectPath, status string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if state, ok := pt.projects[projectPath]; ok {
		state.Status = status
		state.UpdatedAt = time.Now()
	}
}

// ListProjects 列出所有项目
func (pt *ProjectTracker) ListProjects() []*ProjectState {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	result := make([]*ProjectState, 0, len(pt.projects))
	for _, p := range pt.projects {
		result = append(result, p)
	}
	return result
}

// SaveToFile 保存项目状态到文件
func (pt *ProjectTracker) SaveToFile(projectPath string) error {
	pt.mu.RLock()
	state, ok := pt.projects[projectPath]
	pt.mu.RUnlock()
	if !ok {
		return nil
	}

	orchDir := GetProjectOrchDir(projectPath)
	if err := os.MkdirAll(orchDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(orchDir, "state.json"), data, 0644)
}

// SetBranch 设置项目分支
func (pt *ProjectTracker) SetBranch(projectPath, branchName string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if state, ok := pt.projects[projectPath]; ok {
		state.Branch = branchName
		state.UpdatedAt = time.Now()
	}
}
