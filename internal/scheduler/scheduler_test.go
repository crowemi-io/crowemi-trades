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
	schedule string
	runs     *atomic.Int32
}

func (t *testTask) Name() string {
	return t.name
}

func (t *testTask) Schedule() string {
	return t.schedule
}

func (t *testTask) Run(context.Context) error {
	t.runs.Add(1)
	return nil
}

func TestRunnerRunsTask(t *testing.T) {
	var runs atomic.Int32

	now := time.Now().Truncate(time.Minute).Add(time.Minute)
	waitDuration := time.Until(now) + 2*time.Second
	ctx, cancel := context.WithTimeout(context.Background(), waitDuration)
	defer cancel()

	runner := &Runner{
		Logger: kitlog.NewNopLogger(),
		Tasks: []Task{
			&testTask{name: "test", schedule: "* * * * *", runs: &runs},
		},
		TaskTimeout: 5 * time.Second,
	}

	if err := runner.Run(ctx); err != nil {
		t.Fatalf("runner returned error: %v", err)
	}

	if runs.Load() == 0 {
		t.Fatalf("expected task to run at least once")
	}
}

func TestRunnerInvalidSchedule(t *testing.T) {
	runner := &Runner{
		Logger: kitlog.NewNopLogger(),
		Tasks: []Task{
			&testTask{name: "bad", schedule: "not-a-cron"},
		},
	}

	err := runner.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid schedule")
	}
}
