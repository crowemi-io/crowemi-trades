package runtime

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/crowemi-io/crowemi-trades/internal/scheduler"
	"github.com/crowemi-io/crowemi-trades/internal/stream"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/sync/errgroup"
)

type App struct {
	Logger    kitlog.Logger
	Server    *http.Server
	Scheduler *scheduler.Runner
	Watcher   *stream.Watcher
	Updater   *stream.Updater
}

func (a *App) Run(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)

	if a.Server != nil {
		group.Go(func() error {
			_ = level.Info(a.Logger).Log("component", "http", "msg", "server start", "addr", a.Server.Addr)
			err := a.Server.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return err
		})

		group.Go(func() error {
			<-groupCtx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = level.Info(a.Logger).Log("component", "http", "msg", "server shutdown")
			return a.Server.Shutdown(shutdownCtx)
		})
	}

	if a.Scheduler != nil {
		group.Go(func() error {
			_ = level.Info(a.Logger).Log("component", "scheduler", "msg", "scheduler start")
			return a.Scheduler.Run(groupCtx)
		})
	}

	if a.Watcher != nil {
		group.Go(func() error {
			_ = level.Info(a.Logger).Log("component", "watcher", "msg", "consumer start")
			return a.Watcher.Run(groupCtx)
		})
	}

	if a.Updater != nil {
		group.Go(func() error {
			_ = level.Info(a.Logger).Log("component", "updater", "msg", "consumer start")
			return a.Updater.Run(groupCtx)
		})
	}

	return group.Wait()
}
