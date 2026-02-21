package scheduler

import (
	"context"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/sync/errgroup"
)

type Task interface {
	Name() string
	Interval() time.Duration
	Run(context.Context) error
}

type Runner struct {
	Logger      kitlog.Logger
	Tasks       []Task
	TaskTimeout time.Duration
}

func (r *Runner) Run(ctx context.Context) error {
	if len(r.Tasks) == 0 {
		<-ctx.Done()
		return nil
	}

	group, groupCtx := errgroup.WithContext(ctx)
	for _, task := range r.Tasks {
		task := task
		if task.Interval() <= 0 {
			_ = level.Warn(r.Logger).Log("component", "scheduler", "task", task.Name(), "msg", "skip task with non-positive interval")
			continue
		}

		group.Go(func() error {
			return r.runTaskLoop(groupCtx, task)
		})
	}

	return group.Wait()
}

func (r *Runner) runTaskLoop(ctx context.Context, task Task) error {
	ticker := time.NewTicker(task.Interval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			runCtx := ctx
			cancel := func() {}
			if r.TaskTimeout > 0 {
				runCtx, cancel = context.WithTimeout(ctx, r.TaskTimeout)
			}

			start := time.Now()
			err := task.Run(runCtx)
			cancel()

			duration := time.Since(start).Milliseconds()
			if err != nil {
				_ = level.Error(r.Logger).Log(
					"component", "scheduler",
					"task", task.Name(),
					"duration_ms", duration,
					"err", err,
				)
				continue
			}

			_ = level.Info(r.Logger).Log(
				"component", "scheduler",
				"task", task.Name(),
				"duration_ms", duration,
				"msg", "task completed",
			)
		}
	}
}
