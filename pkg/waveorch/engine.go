// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"fmt"
	"sync"
	"time"
)

// Task 编排任务
type Task struct {
	ID          string    `json:"id"`
	ProjectPath string    `json:"project_path"`
	Prompt      string    `json:"prompt"`
	Agent       string    `json:"agent"`
	Status      string    `json:"status"` // pending, running, completed, failed
	Round       int       `json:"round"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Report      *Report   `json:"report,omitempty"`
	BlockID     string    `json:"block_id,omitempty"`
}

// Engine 编排引擎
type Engine struct {
	mu           sync.RWMutex
	registry     *AgentRegistry
	stateMachine *StateMachine
	tasks        map[string]*Task
	taskQueue    chan *Task
	maxParallel  int
	paused       bool
}

// NewEngine 创建编排引擎
func NewEngine(registry *AgentRegistry, maxParallel int) *Engine {
	if maxParallel <= 0 {
		maxParallel = 3
	}
	return &Engine{
		registry:     registry,
		stateMachine: NewStateMachine(),
		tasks:        make(map[string]*Task),
		taskQueue:    make(chan *Task, 100),
		maxParallel:  maxParallel,
		paused:       false,
	}
}

// SubmitTask 提交任务
func (e *Engine) SubmitTask(projectPath, prompt, preferredAgent string) (*Task, error) {
	e.mu.Lock()

	if e.paused {
		e.mu.Unlock()
		return nil, fmt.Errorf("engine is paused")
	}

	task := &Task{
		ID:          fmt.Sprintf("task-%d", time.Now().UnixNano()),
		ProjectPath: projectPath,
		Prompt:      prompt,
		Agent:       preferredAgent,
		Status:      "pending",
		Round:       1,
		CreatedAt:   time.Now(),
	}

	e.tasks[task.ID] = task
	e.mu.Unlock()

	select {
	case e.taskQueue <- task:
		return task, nil
	default:
		e.mu.Lock()
		delete(e.tasks, task.ID)
		e.mu.Unlock()
		return nil, fmt.Errorf("task queue full")
	}
}

// Pause 暂停引擎
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.paused = true
	e.stateMachine.Transition(StatePaused)
}

// Resume 恢复引擎
func (e *Engine) Resume() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.paused = false
	e.stateMachine.Transition(StateIdle)
}

// IsPaused 检查是否暂停
func (e *Engine) IsPaused() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.paused
}

// GetTask 获取任务
func (e *Engine) GetTask(taskID string) *Task {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.tasks[taskID]
}

// GetState 获取当前状态
func (e *Engine) GetState() State {
	return e.stateMachine.CurrentState()
}

// ListTasks 列出所有任务
func (e *Engine) ListTasks() []*Task {
	e.mu.RLock()
	defer e.mu.RUnlock()
	tasks := make([]*Task, 0, len(e.tasks))
	for _, t := range e.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// UpdateTaskStatus 更新任务状态
func (e *Engine) UpdateTaskStatus(taskID, status string, report *Report) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if task, ok := e.tasks[taskID]; ok {
		task.Status = status
		task.Report = report
		if status == "completed" || status == "failed" {
			task.CompletedAt = time.Now()
		}
	}
}
