package task

import (
	"context"
	"errors"
	"strings"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	ct "github.com/crowemi-io/crowemi-trades"
	cfg "github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jackc/pgx/v5"
)

type ActivityTask struct {
	Config       *cfg.Config
	Alpaca       *ct.Alpaca
	Queries      *sqlc.Queries
	Logger       kitlog.Logger
	CronSchedule string
}

func (t *ActivityTask) Name() string {
	return "activity_sync"
}

func (t *ActivityTask) Schedule() string {
	if t.Logger != nil {
		_ = level.Debug(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "Schedule() called", "CronSchedule", t.CronSchedule)
	}
	if t.CronSchedule != "" {
		return t.CronSchedule
	}
	return "0/30 * * * *"
}

func (t *ActivityTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "activity sync start")
	}

	// Get account ID from config
	account, err := t.Queries.GetAccountByAlpacaID(ctx, &t.Config.Alpaca.AccountID)
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "get account failed", "err", err)
		}
		return err
	}

	latest, err := t.Queries.GetLatestActivityByAccountID(ctx, account.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) && !strings.Contains(err.Error(), "no rows in result set") {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "get latest activity failed", "err", err)
		}
		return err
	}

	req := alpaca.GetAccountActivitiesRequest{Direction: "asc"}
	if latest.ID > 0 {
		req.PageToken = latest.ActivityID
	}

	var total int
	for {
		activities, err := t.Alpaca.Client.GetAccountActivities(req)
		if err != nil {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "fetch activities failed", "err", err)
			}
			return err
		}

		if t.Logger != nil {
			_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "activities page fetched", "count", len(activities))
		}

		for _, a := range activities {
			params := db.ActivityParamsFromAlpaca(account.ID, &a)
			_, err = t.Queries.CreateActivity(ctx, params)
			if err != nil {
				if t.Logger != nil {
					_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "persist activity failed", "activity_id", a.ID, "err", err)
				}
				return err
			}
		}

		total += len(activities)

		if len(activities) == 0 {
			break
		}
		req.PageToken = activities[len(activities)-1].ID
	}

	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "activity sync complete", "total", total)
	}
	return nil
}
