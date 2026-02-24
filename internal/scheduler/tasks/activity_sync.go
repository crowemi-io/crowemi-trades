package task

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type ActivitySyncTask struct {
	AlpacaClient *alpaca.Client
	FirestoreDB  *db.Firestore
	Logger       kitlog.Logger
	CronSchedule string
}

func (t *ActivitySyncTask) Name() string {
	return "activity_sync"
}

func (t *ActivitySyncTask) Schedule() string {
	return t.CronSchedule
}

func (t *ActivitySyncTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "activity sync start")
	}

	latest, err := db.GetLatest[*models.Activity](ctx, t.FirestoreDB, db.CollectionActivities, "occurred_at")
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "get latest activity failed", "err", err)
		}
		return err
	}

	req := alpaca.GetAccountActivitiesRequest{Direction: "asc"}
	if latest != nil {
		req.PageToken = latest.GetID()
	}

	var total int
	for {
		activities, err := t.AlpacaClient.GetAccountActivities(req)
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
			doc := models.ActivityFromAlpaca(&a)
			if _, err := db.Create(ctx, t.FirestoreDB, db.CollectionActivities, doc); err != nil {
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
