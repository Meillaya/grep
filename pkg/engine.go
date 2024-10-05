package pkg

import (
	"context"
	"fmt"
	"sync"
)

// Engine represents the core processing engine
type Engine struct {
	mu      sync.Mutex
	running bool
	tasks   []Task
}

// Task represents a unit of work to be processed by the engine
type Task interface {
	Execute(ctx context.Context) error
}

// New creates a new instance of the Engine
func New() *Engine {
	return &Engine{
		tasks: make([]Task, 0),
	}
}

// AddTask adds a new task to the engine
func (e *Engine) AddTask(task Task) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tasks = append(e.tasks, task)
}

// Start begins processing tasks in the engine
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("engine is already running")
	}
	e.running = true
	e.mu.Unlock()

	for _, task := range e.tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := task.Execute(ctx); err != nil {
				return fmt.Errorf("task execution failed: %w", err)
			}
		}
	}

	e.mu.Lock()
	e.running = false
	e.mu.Unlock()

	return nil
}

// Stop halts the engine's processing
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running = false
}

// IsRunning returns the current running state of the engine
func (e *Engine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}
