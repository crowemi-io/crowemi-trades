package task

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	ct "github.com/crowemi-io/crowemi-trades"
	"github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
)

func TestAccountTask_Name(t *testing.T) {
	task := &AccountTask{}
	if got, want := task.Name(), "account_sync"; got != want {
		t.Fatalf("AccountTask.Name() = %q, want %q", got, want)
	}
}

func TestAccountTask_DefaultSchedule(t *testing.T) {
	task := &AccountTask{}
	if got, want := task.DefaultSchedule(), "0/30 * * * *"; got != want {
		t.Fatalf("AccountTask.DefaultSchedule() = %q, want %q", got, want)
	}
}

func TestAccountTask_Schedule(t *testing.T) {
	t.Run("uses configured schedule", func(t *testing.T) {
		task := &AccountTask{CronSchedule: "*/5 * * * *"}
		if got, want := task.Schedule(), "*/5 * * * *"; got != want {
			t.Fatalf("AccountTask.Schedule() = %q, want %q", got, want)
		}
	})

	t.Run("uses default schedule when empty", func(t *testing.T) {
		task := &AccountTask{}
		if got, want := task.Schedule(), "0/30 * * * *"; got != want {
			t.Fatalf("AccountTask.Schedule() = %q, want %q", got, want)
		}
	})
}

func TestAccountTask_Run(t *testing.T) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "../../../.secret/config-local.json"
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Skipf("integration config not found at %q: %v", configPath, err)
	}

	cfg, err := config.Bootstrap(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	postgres, err := db.NewPostgres(context.Background(), cfg.Crowemi.DatabaseURI)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer postgres.Close()

	httpClient := &http.Client{}
	client := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:     cfg.Alpaca.APIKey,
		APISecret:  cfg.Alpaca.APISecretKey,
		BaseURL:    cfg.Alpaca.APIBaseURL,
		HTTPClient: httpClient,
	})
	liveAccount, err := client.GetAccount()
	if err != nil {
		t.Fatalf("failed to fetch live account for assertion: %v", err)
	}

	task := &AccountTask{
		Config:  cfg,
		Alpaca:  &ct.Alpaca{Client: client},
		Queries: postgres.Queries,
		Logger:  cfg.Logger,
	}

	if err := task.Run(context.Background()); err != nil {
		t.Fatalf("AccountTask.Run() error = %v", err)
	}

	stored, err := postgres.Queries.GetAccountByAlpacaID(context.Background(), &liveAccount.ID)
	if err != nil {
		t.Fatalf("failed to fetch stored account: %v", err)
	}
	if stored.AlpacaID == nil || *stored.AlpacaID != liveAccount.ID {
		t.Fatalf("stored account alpaca_id mismatch: got=%v want=%q", stored.AlpacaID, liveAccount.ID)
	}
}
