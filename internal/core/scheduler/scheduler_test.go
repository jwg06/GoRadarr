package scheduler

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunnerExecutesTask(t *testing.T) {
	runner := NewRunner(slog.Default())
	var runs atomic.Int32
	runner.Add(Task{
		Name:     "test",
		Interval: 20 * time.Millisecond,
		Timeout:  20 * time.Millisecond,
		Run: func(ctx context.Context) error {
			runs.Add(1)
			return nil
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Millisecond)
	defer cancel()
	runner.Start(ctx)
	runner.Wait()

	if runs.Load() < 2 {
		t.Fatalf("expected task to run at least twice, ran %d times", runs.Load())
	}
}
