package task

import (
	"context"

	ct "github.com/crowemi-io/crowemi-trades"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type AccountTask struct {
	Alpaca       *ct.Alpaca
	FirestoreDB  *db.Firestore
	Logger       kitlog.Logger
	CronSchedule string
}

func (t *AccountTask) DefaultSchedule() string {
	return "0/30 * * * *"
}

func (t *AccountTask) Name() string {
	return "account_sync"
}

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

	account, err := t.Alpaca.Client.GetAccount()
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "fetch account failed", "err", err)
		}
		return err
	}

	accountDoc := models.AccountFromAlpaca(account)
	_, err = db.Create(ctx, t.FirestoreDB, db.CollectionAccounts, accountDoc)
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
