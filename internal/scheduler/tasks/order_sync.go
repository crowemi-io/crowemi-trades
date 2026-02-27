package task

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const maxOrdersPerPage = 500

type OrderSyncTask struct {
	AlpacaClient *alpaca.Client
	FirestoreDB  *db.Firestore
	Logger       kitlog.Logger
	CronSchedule string
}

func (t *OrderSyncTask) Name() string {
	return "order_sync"
}

func (t *OrderSyncTask) Schedule() string {
	return t.CronSchedule
}

func (t *OrderSyncTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "order sync start")
	}

	latest, err := db.GetLatest[*models.Order](ctx, t.FirestoreDB, db.CollectionOrders, "created_at")
	if err != nil {
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
	if latest != nil && !latest.CreatedAt.IsZero() {
		req.After = latest.CreatedAt
	}

	var total int
	for {
		orders, err := t.AlpacaClient.GetOrders(req)
		if err != nil {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "fetch orders failed", "err", err)
			}
			return err
		}

		for _, o := range orders {
			doc := models.OrderFromAlpaca(&o)
			if _, err := db.Create(ctx, t.FirestoreDB, db.CollectionOrders, doc); err != nil {
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
