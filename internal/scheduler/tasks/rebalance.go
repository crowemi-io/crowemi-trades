package task

import (
	"context"
	"encoding/json"
	"math"
	"os"

	"github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	"github.com/crowemi-io/crowemi-trades/internal/rebalance"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"

	ct "github.com/crowemi-io/crowemi-trades"
)

type RebalanceTask struct {
	Alpaca       *ct.Alpaca
	FirestoreDB  *db.Firestore
	Logger       kitlog.Logger
	CronSchedule string
	Options      config.TaskOptions
}

func (t *RebalanceTask) Name() string {
	return "rebalance"
}

func (t *RebalanceTask) Schedule() string {
	if t.Logger != nil {
		_ = level.Debug(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "Schedule() called", "CronSchedule", t.CronSchedule)
	}
	if t.CronSchedule != "" {
		return t.CronSchedule
	}
	return "0/30 * * * *"
}

func (t *RebalanceTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "rebalance start")
	}

	portfolios, err := db.List[*models.Portfolio](ctx, t.FirestoreDB, db.CollectionPortfolios)
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "list portfolios failed", "err", err)
		}
		return err
	}

	for _, portfolio := range portfolios {
		result, err := rebalance.Compute(ctx, t.Alpaca.Client, t.FirestoreDB, portfolio.ID)
		if err != nil {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "rebalance compute failed", "portfolio_id", portfolio.ID, "err", err)
			}
			return err
		}
		if t.Options.WriteFile {
			// write result to json
			jsonData, _ := json.Marshal(result)
			err := os.WriteFile("rebalance_result.json", jsonData, 0644)
			if err != nil {
				if t.Logger != nil {
					_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "write result to file failed", "err", err)
				}
				return err
			}
		}

		for _, action := range result.Actions {
			if action.Delta == 0 {
				continue
			}
			side := "buy"
			if action.Delta < 0 {
				side = "sell"
			}
			amount := math.Abs(action.Delta)
			t.Alpaca.SetOrder(ctx, action.Symbol, side, amount)
		}
	}

	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "rebalance complete")
	}
	return nil
}
