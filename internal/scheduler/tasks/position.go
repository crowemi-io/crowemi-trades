package task

import (
	"context"

	ct "github.com/crowemi-io/crowemi-trades"
	cfg "github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type PositionTask struct {
	Config       *cfg.Config
	Alpaca       *ct.Alpaca
	Queries      *sqlc.Queries
	Logger       kitlog.Logger
	CronSchedule string
}

func (t *PositionTask) Name() string {
	return "position_sync"
}

func (t *PositionTask) Schedule() string {
	if t.Logger != nil {
		_ = level.Debug(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "Schedule() called", "CronSchedule", t.CronSchedule)
	}
	if t.CronSchedule != "" {
		return t.CronSchedule
	}
	return "0/30 * * * *"
}

func (t *PositionTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "position sync start")
	}

	// Get account ID from config
	account, err := t.Queries.GetAccountByAlpacaID(ctx, &t.Config.Alpaca.AccountID)
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "get account failed", "err", err)
		}
		return err
	}

	// Get current positions from DB
	current, err := t.Queries.ListCurrentPositionsByAccountID(ctx, int64(account.ID))
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "list current positions failed", "err", err)
		}
		return err
	}

	currentIDs := make(map[string]struct{}, len(current))
	for _, p := range current {
		currentIDs[p.Symbol] = struct{}{}
	}

	positions, err := t.Alpaca.Client.GetPositions()
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "fetch positions failed", "err", err)
		}
		return err
	}

	var synced int
	for _, p := range positions {
		params := db.PositionParamsFromAlpaca(int64(account.ID), &p)
		_, err = t.Queries.UpsertPosition(ctx, params)
		if err != nil {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "upsert position failed", "symbol", p.Symbol, "err", err)
			}
			return err
		}

		delete(currentIDs, p.Symbol)
		synced++
	}

	// Mark remaining positions as stale
	if len(currentIDs) > 0 {
		err = t.Queries.MarkPositionsStale(ctx, int64(account.ID))
		if err != nil {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "mark positions stale failed", "err", err)
			}
			return err
		}
	}

	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "position sync complete", "synced", synced, "stale_marked", len(currentIDs))
	}

	return nil
}
