package task

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type AccountSyncTask struct {
	AlpacaClient *alpaca.Client
	FirestoreDB  *db.Firestore
	Logger       kitlog.Logger
	CronSchedule string
}

func (t *AccountSyncTask) Name() string {
	return "account_sync"
}

func (t *AccountSyncTask) Schedule() string {
	return t.CronSchedule
}

func (t *AccountSyncTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "account sync start")
	}

	account, err := t.AlpacaClient.GetAccount()
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
