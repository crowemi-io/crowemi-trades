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
	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
)

func TestActivityTask_Name(t *testing.T) {
	task := &ActivityTask{}
	if got, want := task.Name(), "activity_sync"; got != want {
		t.Fatalf("ActivityTask.Name() = %q, want %q", got, want)
	}
}

func TestActivityTask_Schedule(t *testing.T) {
	t.Run("uses configured schedule", func(t *testing.T) {
		task := &ActivityTask{CronSchedule: "*/10 * * * *"}
		if got, want := task.Schedule(), "*/10 * * * *"; got != want {
			t.Fatalf("ActivityTask.Schedule() = %q, want %q", got, want)
		}
	})

	t.Run("uses default schedule when empty", func(t *testing.T) {
		task := &ActivityTask{}
		if got, want := task.Schedule(), "0/30 * * * *"; got != want {
			t.Fatalf("ActivityTask.Schedule() = %q, want %q", got, want)
		}
	})
}

func TestActivityTask_Run(t *testing.T) {
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
		t.Fatalf("failed to fetch live account: %v", err)
	}
	cfg.Alpaca.AccountID = liveAccount.ID

	if _, err := postgres.Queries.UpsertAccount(context.Background(), db.AccountParamsFromAlpaca(liveAccount)); err != nil {
		t.Fatalf("failed to ensure account exists before activity sync: %v", err)
	}

	apiActivities, err := client.GetAccountActivities(alpaca.GetAccountActivitiesRequest{Direction: "asc"})
	if err != nil {
		t.Fatalf("failed to fetch live activities for assertion baseline: %v", err)
	}

	task := &ActivityTask{
		Config:  cfg,
		Alpaca:  &ct.Alpaca{Client: client},
		Queries: postgres.Queries,
		Logger:  cfg.Logger,
	}

	if err := task.Run(context.Background()); err != nil {
		t.Fatalf("ActivityTask.Run() error = %v", err)
	}

	if len(apiActivities) == 0 {
		return
	}

	storedAccount, err := postgres.Queries.GetAccountByAlpacaID(context.Background(), &liveAccount.ID)
	if err != nil {
		t.Fatalf("failed to fetch stored account for activity assertion: %v", err)
	}

	stored, err := postgres.Queries.ListActivitiesByAccountID(context.Background(), sqlc.ListActivitiesByAccountIDParams{
		AccountID: storedAccount.ID,
		Limit:     1,
		Offset:    0,
	})
	if err != nil {
		t.Fatalf("failed to query stored activities: %v", err)
	}
	if len(stored) == 0 {
		t.Fatal("expected at least one stored activity after sync")
	}
}
