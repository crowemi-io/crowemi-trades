package task

import (
	"context"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	ct "github.com/crowemi-io/crowemi-trades"
	cfg "github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type CorporateActionTask struct {
	Config       *cfg.Config
	Alpaca       *ct.Alpaca
	FirestoreDB  *db.Firestore
	Logger       kitlog.Logger
	CronSchedule string
}

func (t *CorporateActionTask) Name() string {
	return "corporate_action_sync"
}

func (t *CorporateActionTask) Schedule() string {
	if t.Logger != nil {
		_ = level.Debug(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "Schedule() called", "CronSchedule", t.CronSchedule)
	}
	if t.CronSchedule != "" {
		return t.CronSchedule
	}
	return "0/30 * * * *"
}

func (t *CorporateActionTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "corporate action sync start")
	}

	var symbols []string = nil
	categories, err := db.List[*models.Category](ctx, t.FirestoreDB, t.Config.RootCollection()+db.CollectionAllocations)
	if err != nil {
		return err
	}
	for _, c := range categories {
		s, err := db.List[*models.Symbol](ctx, t.FirestoreDB, t.Config.RootCollection()+db.CollectionAllocations+"/"+c.ID+"/symbols")
		if err != nil {
			return err
		}
		for _, symbol := range s {
			symbols = append(symbols, symbol.ID)
		}

	}

	now := time.Now()
	caTypes := []string{"dividend", "merger", "spinoff", "split"}
	var total int

	for _, symbol := range symbols {
		since := now.AddDate(0, 0, -60)

		latest, err := db.GetLatestWhere[*models.CorporateAction](
			ctx, t.FirestoreDB, t.Config.RootCollection()+db.CollectionCorporateActions,
			"last_synced_at", "initiating_symbol", symbol,
		)
		if err != nil {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "get latest corporate action failed", "symbol", symbol, "err", err)
			}
			return err
		}
		if latest != nil && now.Sub(latest.LastSyncedAt).Hours() < 60*24 {
			since = latest.LastSyncedAt
		}

		until := now.AddDate(0, 0, 30)

		if t.Logger != nil {
			_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "fetching corporate actions", "symbol", symbol, "since", since.Format("2006-01-02"), "until", until.Format("2006-01-02"))
		}

		announcements, err := t.Alpaca.Client.GetAnnouncements(alpaca.GetAnnouncementsRequest{
			CATypes: caTypes,
			Symbol:  symbol,
			Since:   since,
			Until:   until,
		})
		if err != nil {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "fetch announcements failed", "symbol", symbol, "err", err)
			}
			return err
		}

		for _, a := range announcements {
			doc := models.CorporateActionFromAlpaca(&a)
			if _, err := db.Create(ctx, t.FirestoreDB, t.Config.RootCollection()+db.CollectionCorporateActions, doc); err != nil {
				if t.Logger != nil {
					_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "persist corporate action failed", "announcement_id", a.ID, "symbol", symbol, "err", err)
				}
				return err
			}
		}

		total += len(announcements)
	}

	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "corporate action sync complete", "total", total)
	}
	return nil
}
