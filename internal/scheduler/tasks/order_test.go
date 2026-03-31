package task

import (
	"context"
	"net/http"
	"testing"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	ct "github.com/crowemi-io/crowemi-trades"
	"github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
)

func TestOrderTask_Run(t *testing.T) {
	cfg, err := config.Bootstrap("../../../.secret/config-local.json")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	defer cfg.Firestore.Close()

	httpClient := &http.Client{}
	alpaca := &ct.Alpaca{
		Client: alpaca.NewClient(alpaca.ClientOpts{
			APIKey:     cfg.Alpaca.APIKey,
			APISecret:  cfg.Alpaca.APISecretKey,
			BaseURL:    cfg.Alpaca.APIBaseURL,
			HTTPClient: httpClient,
		}),
		Notifier: nil,
	}

	task := &OrderTask{
		Config:      cfg,
		Alpaca:      alpaca,
		FirestoreDB: db.NewFirestore(cfg.Firestore),
		Logger:      cfg.Logger,
	}
	err = task.Run(context.Background())
	if err != nil {
		t.Fatalf("AccountTask.Run() error = %v", err)
	}
}
