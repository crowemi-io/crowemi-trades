package task

import (
	"context"
	"errors"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	ct "github.com/crowemi-io/crowemi-trades"
	cfg "github.com/crowemi-io/crowemi-trades/internal/config"

	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type AccountTask struct {
	Config       *cfg.Config
	Alpaca       *ct.Alpaca
	Queries      *sqlc.Queries
	Logger       kitlog.Logger
	CronSchedule string

	getAccountFn    func() (*alpaca.Account, error)
	upsertAccountFn func(context.Context, sqlc.UpsertAccountParams) (sqlc.AppAccount, error)
}

func (t *AccountTask) DefaultSchedule() string { return "0/30 * * * *" }
func (t *AccountTask) Name() string            { return "account_sync" }
func (t *AccountTask) Schedule() string {
	if t.Logger != nil {
		_ = level.Debug(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "Schedule() called", "CronSchedule", t.CronSchedule)
	}
	if t.CronSchedule != "" {
		return t.CronSchedule
	}
	return "0/30 * * * *"
}

func (t *AccountTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "account sync start")
	}

	var (
		account *alpaca.Account
		err     error
	)
	if t.getAccountFn != nil {
		account, err = t.getAccountFn()
	} else {
		if t.Alpaca == nil || t.Alpaca.Client == nil {
			return errors.New("alpaca client is required")
		}
		account, err = t.Alpaca.Client.GetAccount()
	}
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "fetch account failed", "err", err)
		}
		return err
	}

	params := db.AccountParamsFromAlpaca(account)
	if t.upsertAccountFn != nil {
		_, err = t.upsertAccountFn(ctx, params)
	} else {
		if t.Queries == nil {
			return errors.New("queries is required")
		}
		_, err = t.Queries.UpsertAccount(ctx, params)
	}

	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "persist account failed", "account_id", account.ID, "err", err)
		}
		return err
	}

	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "account sync complete", "account_id", account.ID)
	}

	return err
}
