package task

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
)

func TestAccountSyncTask_Run(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	configPath := filepath.Join("..", "..", ".secret", "config-local.json")
	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skipf("skipping integration test: config file not found at %s", configPath)
		}
		t.Fatalf("failed to stat config file: %v", err)
	}

	c, err := config.Bootstrap(configPath)
	if err != nil {
		t.Fatalf("bootstrap config: %v", err)
	}
	t.Cleanup(func() {
		if c.Firestore != nil {
			_ = c.Firestore.Close()
		}
	})

	alpacaClient := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:     c.Alpaca.APIKey,
		APISecret:  c.Alpaca.APISecretKey,
		BaseURL:    c.Alpaca.APIBaseURL,
		HTTPClient: &http.Client{Timeout: 20 * time.Second},
	})
	firestoreDB := db.NewFirestore(c.Firestore)

	account, err := alpacaClient.GetAccount()
	if err != nil {
		t.Fatalf("get alpaca account: %v", err)
	}

	task := &AccountSyncTask{
		AlpacaClient: alpacaClient,
		FirestoreDB:  firestoreDB,
		CronSchedule: "* * * * *",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_ = db.Delete(cleanupCtx, firestoreDB, db.CollectionAccounts, account.ID)
	})

	if err := task.Run(ctx); err != nil {
		t.Fatalf("run account sync task: %v", err)
	}

	saved, err := db.Get[*models.Account](ctx, firestoreDB, db.CollectionAccounts, account.ID)
	if err != nil {
		t.Fatalf("get saved account document: %v", err)
	}

	if saved == nil {
		t.Fatalf("expected saved account document, got nil")
	}
	if saved.ID == "" {
		t.Fatalf("expected saved account id to be set")
	}
	if saved.AccountNumber == "" {
		t.Fatalf("expected saved account number to be populated")
	}
	if saved.Status == "" {
		t.Fatalf("expected saved account status to be populated")
	}
	if saved.Currency == "" {
		t.Fatalf("expected saved account currency to be populated")
	}
}
