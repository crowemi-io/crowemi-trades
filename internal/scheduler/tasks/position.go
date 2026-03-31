package task

import (
	"context"

	ct "github.com/crowemi-io/crowemi-trades"
	cfg "github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type PositionTask struct {
	Config       *cfg.Config
	Alpaca       *ct.Alpaca
	FirestoreDB  *db.Firestore
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

	current, err := db.ListWhere[*models.Position](ctx, t.FirestoreDB, t.Config.RootCollection()+db.CollectionPositions, "is_current", "!=", false)
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

	positions, err := t.Alpaca.Client.GetPositions()
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "fetch positions failed", "err", err)
		}
		return err
	}

	var synced int
	for _, p := range positions {
		doc := models.PositionFromAlpaca(&p)
		if err := db.Upsert(ctx, t.FirestoreDB, t.Config.RootCollection()+db.CollectionPositions, doc); err != nil {
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
		if err := db.SetFields(ctx, t.FirestoreDB, t.Config.RootCollection()+db.CollectionPositions, id, map[string]interface{}{"is_current": false}); err != nil {
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
