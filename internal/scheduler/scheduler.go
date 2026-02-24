package scheduler

import (
	"context"
	"fmt"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/sync/errgroup"
)

type Task interface {
	Name() string
	Schedule() string
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

		expr, err := ParseCron(task.Schedule())
		if err != nil {
			return fmt.Errorf("scheduler: task %q: %w", task.Name(), err)
		}

		_ = level.Info(r.Logger).Log(
			"component", "scheduler",
			"task", task.Name(),
			"schedule", task.Schedule(),
			"msg", "registered task",
		)

		group.Go(func() error {
			return r.runTaskLoop(groupCtx, task, expr)
		})
	}

	return group.Wait()
}

func (r *Runner) runTaskLoop(ctx context.Context, task Task, expr CronExpr) error {
	for {
		next := expr.Next(time.Now())
		if next.IsZero() {
			_ = level.Error(r.Logger).Log("component", "scheduler", "task", task.Name(), "msg", "no next fire time found")
			return nil
		}

		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}

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
