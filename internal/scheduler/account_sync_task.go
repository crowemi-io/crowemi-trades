package scheduler

import (
	"context"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
)

type AccountSyncTask struct {
	AlpacaClient *alpaca.Client
	FirestoreDB  *db.Firestore
	Every        time.Duration
}

func (t *AccountSyncTask) Name() string {
	return "account_sync"
}

func (t *AccountSyncTask) Interval() time.Duration {
	return t.Every
}

func (t *AccountSyncTask) Run(ctx context.Context) error {
	account, err := t.AlpacaClient.GetAccount()
	if err != nil {
		return err
	}

	accountDoc := models.AccountFromAlpaca(account)
	_, err = db.Create[*models.Account](ctx, t.FirestoreDB, db.CollectionAccounts, accountDoc)
	return err
}
