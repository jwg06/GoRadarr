package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/jwg06/goradarr/internal/events"
)

type TaskFunc func(context.Context) error

type Task struct {
	Name     string
	Interval time.Duration
	Timeout  time.Duration
	Run      TaskFunc
}

type Runner struct {
	logger *slog.Logger
	tasks  []Task
	wg     sync.WaitGroup
}

func NewRunner(logger *slog.Logger) *Runner {
	return &Runner{logger: logger}
}

func (r *Runner) Add(task Task) {
	if task.Interval <= 0 {
		task.Interval = time.Minute
	}
	if task.Timeout <= 0 {
		task.Timeout = task.Interval
	}
	r.tasks = append(r.tasks, task)
}

func (r *Runner) Start(ctx context.Context) {
	for _, task := range r.tasks {
		task := task
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			r.runTask(ctx, task)
			ticker := time.NewTicker(task.Interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					r.runTask(ctx, task)
				}
			}
		}()
	}
}

func (r *Runner) Wait() {
	r.wg.Wait()
}

func (r *Runner) runTask(ctx context.Context, task Task) {
	events.PublishDefault(events.Event{Type: events.EventTaskStarted, Data: map[string]any{"name": task.Name}})
	taskCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	err := task.Run(taskCtx)
	if err != nil && !errors.Is(err, context.Canceled) {
		if r.logger != nil {
			r.logger.Warn("scheduled task failed", "task", task.Name, "error", err)
		}
		events.PublishDefault(events.Event{Type: events.EventHealthChanged, Data: map[string]any{"task": task.Name, "error": err.Error()}})
		return
	}
	events.PublishDefault(events.Event{Type: events.EventTaskCompleted, Data: map[string]any{"name": task.Name}})
}
