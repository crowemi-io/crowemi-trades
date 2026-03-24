package task

import (
	"context"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	ct "github.com/crowemi-io/crowemi-trades"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type CorporateActionTask struct {
	Alpaca       *ct.Alpaca
	FirestoreDB  *db.Firestore
	Logger       kitlog.Logger
	CronSchedule string
}

func (t *CorporateActionTask) Name() string {
	return "corporate_action_sync"
}

func (t *CorporateActionTask) Schedule() string {
	if t.CronSchedule != "" {
		return t.CronSchedule
	}
	return "0/30 * * * *"
}

func (t *CorporateActionTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "corporate action sync start")
	}

	portfolios, err := db.List[*models.Portfolio](ctx, t.FirestoreDB, db.CollectionPortfolios)
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "list portfolios failed", "err", err)
		}
		return err
	}

	symbols := extractSymbols(portfolios)
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "symbols extracted", "count", len(symbols))
	}

	now := time.Now()
	caTypes := []string{"dividend", "merger", "spinoff", "split"}
	var total int

	for _, symbol := range symbols {
		since := now.AddDate(0, 0, -60)

		latest, err := db.GetLatestWhere[*models.CorporateAction](
			ctx, t.FirestoreDB, db.CollectionCorporateActions,
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
			if _, err := db.Create(ctx, t.FirestoreDB, db.CollectionCorporateActions, doc); err != nil {
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

func extractSymbols(portfolios []*models.Portfolio) []string {
	seen := make(map[string]bool)
	var symbols []string
	for _, p := range portfolios {
		for _, alloc := range p.Allocations {
			for symbol := range alloc.Symbols {
				if !seen[symbol] {
					seen[symbol] = true
					symbols = append(symbols, symbol)
				}
			}
		}
	}
	return symbols
}
