package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	kitlog "github.com/go-kit/log"
)

type testTask struct {
	name     string
	interval time.Duration
	runs     *atomic.Int32
}

func (t *testTask) Name() string {
	return t.name
}

func (t *testTask) Interval() time.Duration {
	return t.interval
}

func (t *testTask) Run(context.Context) error {
	t.runs.Add(1)
	return nil
}

func TestRunnerRunsTask(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Millisecond)
	defer cancel()

	var runs atomic.Int32
	runner := &Runner{
		Logger: kitlog.NewNopLogger(),
		Tasks: []Task{
			&testTask{name: "test", interval: 10 * time.Millisecond, runs: &runs},
		},
		TaskTimeout: 20 * time.Millisecond,
	}

	if err := runner.Run(ctx); err != nil {
		t.Fatalf("runner returned error: %v", err)
	}

	if runs.Load() == 0 {
		t.Fatalf("expected task to run at least once")
	}
}
