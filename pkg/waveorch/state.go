// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"fmt"
	"sync"
	"time"
)

// State 表示编排引擎的状态
type State string

const (
	StateIdle       State = "IDLE"
	StatePlanning   State = "PLANNING"
	StateExecuting  State = "EXECUTING"
	StateCollecting State = "COLLECTING"
	StateDeciding   State = "DECIDING"
	StateCompleted  State = "COMPLETED"
	StatePaused     State = "PAUSED"
	StateError      State = "ERROR"
	StateBlocked    State = "BLOCKED"
)

// TaskStatus 表示任务状态
type TaskStatus struct {
	TaskId      string    `json:"task_id"`
	ProjectPath string    `json:"project_path"`
	State       State     `json:"state"`
	Round       int       `json:"round"`
	StartTime   time.Time `json:"start_time"`
	LastUpdate  time.Time `json:"last_update"`
	Error       string    `json:"error,omitempty"`
}

// StateMachine 管理状态转换
type StateMachine struct {
	mu           sync.RWMutex
	currentState State
	taskStatus   *TaskStatus
	transitions  map[State][]State
}

// NewStateMachine 创建新的状态机
func NewStateMachine() *StateMachine {
	sm := &StateMachine{
		currentState: StateIdle,
		transitions:  make(map[State][]State),
	}
	sm.initTransitions()
	return sm
}

// initTransitions 初始化状态转换规则
func (sm *StateMachine) initTransitions() {
	sm.transitions[StateIdle] = []State{StatePlanning, StatePaused}
	sm.transitions[StatePlanning] = []State{StateExecuting, StatePaused, StateError}
	sm.transitions[StateExecuting] = []State{StateCollecting, StatePaused, StateError}
	sm.transitions[StateCollecting] = []State{StateDeciding, StatePaused, StateError}
	sm.transitions[StateDeciding] = []State{StateExecuting, StateCompleted, StateBlocked, StatePaused, StateError}
	sm.transitions[StatePaused] = []State{StateIdle, StatePlanning, StateExecuting, StateCollecting, StateDeciding}
	sm.transitions[StateBlocked] = []State{StateDeciding, StatePaused}
}

// CurrentState 返回当前状态
func (sm *StateMachine) CurrentState() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

// CanTransition 检查是否可以转换到目标状态
func (sm *StateMachine) CanTransition(target State) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	allowed, ok := sm.transitions[sm.currentState]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

// Transition 执行状态转换
func (sm *StateMachine) Transition(target State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	allowed, ok := sm.transitions[sm.currentState]
	if !ok {
		return fmt.Errorf("no transitions from state %s", sm.currentState)
	}
	for _, s := range allowed {
		if s == target {
			sm.currentState = target
			if sm.taskStatus != nil {
				sm.taskStatus.State = target
				sm.taskStatus.LastUpdate = time.Now()
			}
			return nil
		}
	}
	return fmt.Errorf("invalid transition from %s to %s", sm.currentState, target)
}
