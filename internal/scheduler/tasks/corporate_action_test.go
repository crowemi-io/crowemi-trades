package task

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	ct "github.com/crowemi-io/crowemi-trades"
	"github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
)

func TestCorporateActionTask_Name(t *testing.T) {
	task := &CorporateActionTask{}
	if got, want := task.Name(), "corporate_action_sync"; got != want {
		t.Fatalf("CorporateActionTask.Name() = %q, want %q", got, want)
	}
}

func TestCorporateActionTask_Schedule(t *testing.T) {
	t.Run("uses configured schedule", func(t *testing.T) {
		task := &CorporateActionTask{CronSchedule: "0 * * * *"}
		if got, want := task.Schedule(), "0 * * * *"; got != want {
			t.Fatalf("CorporateActionTask.Schedule() = %q, want %q", got, want)
		}
	})

	t.Run("uses default schedule when empty", func(t *testing.T) {
		task := &CorporateActionTask{}
		if got, want := task.Schedule(), "0/30 * * * *"; got != want {
			t.Fatalf("CorporateActionTask.Schedule() = %q, want %q", got, want)
		}
	})
}

func TestCorporateActionTask_Run(t *testing.T) {
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

	var corporateActionTable *string
	if err := postgres.Pool.QueryRow(context.Background(), "SELECT to_regclass('app.corporate_action')::text").Scan(&corporateActionTable); err != nil {
		t.Fatalf("failed to check for corporate_action table: %v", err)
	}
	if corporateActionTable == nil || *corporateActionTable == "" {
		t.Skip("configured database is missing app.corporate_action table")
	}

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

	storedAccount, err := postgres.Queries.UpsertAccount(context.Background(), db.AccountParamsFromAlpaca(liveAccount))
	if err != nil {
		t.Fatalf("failed to ensure account exists before corporate action sync: %v", err)
	}

	positions, err := client.GetPositions()
	if err != nil {
		t.Fatalf("failed to fetch live positions for test setup: %v", err)
	}
	if len(positions) == 0 {
		t.Skip("no live positions available to seed portfolio symbols for corporate action test")
	}

	symbol := positions[0].Symbol
	portfolio, err := postgres.Queries.CreatePortfolio(context.Background(), sqlc.CreatePortfolioParams{
		AccountID: storedAccount.ID,
		Name:      fmt.Sprintf("corporate-action-test-%d", time.Now().UnixNano()),
		Weight:    1,
	})
	if err != nil {
		t.Fatalf("failed to create portfolio for corporate action test: %v", err)
	}

	_, err = postgres.Queries.CreatePortfolioSymbol(context.Background(), sqlc.CreatePortfolioSymbolParams{
		PortfolioID: int64(portfolio.ID),
		Symbol:      symbol,
		Weight:      1,
	})
	if err != nil {
		t.Fatalf("failed to create portfolio symbol for corporate action test: %v", err)
	}

	now := time.Now()
	baselineAnnouncements, err := client.GetAnnouncements(alpaca.GetAnnouncementsRequest{
		CATypes: []string{"dividend", "merger", "spinoff", "split"},
		Symbol:  symbol,
		Since:   now.AddDate(0, 0, -60),
		Until:   now.AddDate(0, 0, 30),
	})
	if err != nil {
		t.Fatalf("failed to fetch live announcements for assertion baseline: %v", err)
	}

	task := &CorporateActionTask{
		Config:  cfg,
		Alpaca:  &ct.Alpaca{Client: client},
		Queries: postgres.Queries,
		Logger:  cfg.Logger,
	}

	if err := task.Run(context.Background()); err != nil {
		t.Fatalf("CorporateActionTask.Run() error = %v", err)
	}

	if len(baselineAnnouncements) == 0 {
		return
	}

	storedAction, err := postgres.Queries.GetCorporateActionByID(context.Background(), &baselineAnnouncements[0].CorporateActionsID)
	if err != nil {
		t.Fatalf("failed to fetch stored corporate action: %v", err)
	}
	if storedAction.CorporateActionsID == nil || *storedAction.CorporateActionsID != baselineAnnouncements[0].CorporateActionsID {
		t.Fatalf("stored corporate action ID mismatch: got=%v want=%q", storedAction.CorporateActionsID, baselineAnnouncements[0].CorporateActionsID)
	}
}
