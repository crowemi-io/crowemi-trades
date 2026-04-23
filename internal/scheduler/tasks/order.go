package task

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	ct "github.com/crowemi-io/crowemi-trades"
	cfg "github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const maxOrdersPerPage = 500

type OrderTask struct {
	Config       *cfg.Config
	Alpaca       *ct.Alpaca
	Queries      *sqlc.Queries
	Logger       kitlog.Logger
	CronSchedule string
}

func (t *OrderTask) DefaultSchedule() string {
	return "0/30 * * * *"
}

func (t *OrderTask) Name() string {
	return "order_sync"
}

func (t *OrderTask) Schedule() string {
	if t.Logger != nil {
		_ = level.Debug(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "Schedule() called", "CronSchedule", t.CronSchedule)
	}
	if t.CronSchedule != "" {
		return t.CronSchedule
	}
	return "0/30 * * * *"
}
func (t *OrderTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "order sync start")
	}

	// Get account ID from config
	account, err := t.Queries.GetAccountByAlpacaID(ctx, &t.Config.Alpaca.AccountID)
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "get account failed", "err", err)
		}
		return err
	}

	// Get latest order for this account
	latest, err := t.Queries.GetLatestOrderByAccountID(ctx, int64(account.ID))
	if err != nil && err.Error() != "sql: no rows in result set" {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "get latest order failed", "err", err)
		}
		return err
	}

	req := alpaca.GetOrdersRequest{
		Status:    "all",
		Limit:     maxOrdersPerPage,
		Direction: "asc",
	}
	if latest.ID > 0 && !latest.AlpacaCreatedAt.Time.IsZero() {
		req.After = latest.AlpacaCreatedAt.Time
	}

	var total int
	for {
		orders, err := t.Alpaca.Client.GetOrders(req)
		if err != nil {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "fetch orders failed", "err", err)
			}
			return err
		}

		for _, o := range orders {
			params := db.OrderParamsFromAlpaca(int64(account.ID), &o)
			_, err = t.Queries.UpsertOrder(ctx, params)
			if err != nil {
				if t.Logger != nil {
					_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "persist order failed", "order_id", o.ID, "err", err)
				}
				return err
			}
		}
		total += len(orders)

		if len(orders) < maxOrdersPerPage {
			break
		}

		req.After = orders[len(orders)-1].CreatedAt
	}

	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "order sync complete", "total", total)
	}
	return nil
}
