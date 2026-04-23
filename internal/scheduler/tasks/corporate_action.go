package task

import (
	"context"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	ct "github.com/crowemi-io/crowemi-trades"
	cfg "github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type CorporateActionTask struct {
	Config       *cfg.Config
	Alpaca       *ct.Alpaca
	Queries      *sqlc.Queries
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

	// Get account ID from config
	account, err := t.Queries.GetAccountByAlpacaID(ctx, &t.Config.Alpaca.AccountID)
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "get account failed", "err", err)
		}
		return err
	}

	// Get portfolio symbols for this account
	symbols, err := t.Queries.ListPortfolioSymbolsByAccountID(ctx, account.ID)
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "get portfolio symbols failed", "err", err)
		}
		return err
	}

	symbolList := make([]string, len(symbols))
	copy(symbolList, symbols)

	now := time.Now()
	caTypes := []string{"dividend", "merger", "spinoff", "split"}
	var total int

	for _, symbol := range symbolList {
		since := now.AddDate(0, 0, -60)

		latest, err := t.Queries.GetLatestCorporateActionByAccountID(ctx, account.ID)
		if err != nil && err.Error() != "sql: no rows in result set" {
			if t.Logger != nil {
				_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "get latest corporate action failed", "symbol", symbol, "err", err)
			}
			return err
		}
		if err == nil && !latest.LastSyncedAt.Time.IsZero() && now.Sub(latest.LastSyncedAt.Time).Hours() < 60*24 {
			since = latest.LastSyncedAt.Time
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
			params := db.CorporateActionParamsFromAlpaca(account.ID, &a)
			_, err = t.Queries.UpsertCorporateAction(ctx, params)
			if err != nil {
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
