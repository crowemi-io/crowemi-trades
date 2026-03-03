package task

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type PositionSyncTask struct {
	AlpacaClient *alpaca.Client
	FirestoreDB  *db.Firestore
	Logger       kitlog.Logger
	CronSchedule string
}

func (t *PositionSyncTask) Name() string {
	return "position_sync"
}

func (t *PositionSyncTask) Schedule() string {
	return t.CronSchedule
}

func (t *PositionSyncTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "position sync start")
	}

	current, err := db.ListWhere[*models.Position](ctx, t.FirestoreDB, db.CollectionPositions, "is_current", "!=", false)
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "list current positions failed", "err", err)
		}
		return err
	}

	currentIDs := make(map[string]struct{}, len(current))
	for _, p := range current {
		currentIDs[p.GetID()] = struct{}{}
	}

	positions, err := t.AlpacaClient.GetPositions()
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "fetch positions failed", "err", err)
		}
		return err
	}

	var synced int
	for _, p := range positions {
		doc := models.PositionFromAlpaca(&p)
		if err := db.Upsert(ctx, t.FirestoreDB, db.CollectionPositions, doc); err != nil {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "upsert position failed", "asset_id", doc.GetID(), "err", err)
			}
			return err
		}

		delete(currentIDs, doc.GetID())
		synced++
	}

	var stale int
	for id := range currentIDs {
		if err := db.SetFields(ctx, t.FirestoreDB, db.CollectionPositions, id, map[string]interface{}{"is_current": false}); err != nil {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "mark position stale failed", "asset_id", id, "err", err)
			}
			return err
		}
		stale++
	}

	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "position sync complete", "synced", synced, "stale_marked", stale)
	}

	return nil
}
