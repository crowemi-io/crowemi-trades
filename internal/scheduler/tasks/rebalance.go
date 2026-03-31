package task

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"

	ct "github.com/crowemi-io/crowemi-trades"
	cfg "github.com/crowemi-io/crowemi-trades/internal/config"
)

type RebalanceTask struct {
	Config       *cfg.Config
	Alpaca       *ct.Alpaca
	FirestoreDB  *db.Firestore
	Logger       kitlog.Logger
	CronSchedule string
	Options      config.TaskOptions
}

func (t *RebalanceTask) Name() string {
	return "rebalance"
}

func (t *RebalanceTask) Schedule() string {
	if t.Logger != nil {
		_ = level.Debug(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "Schedule() called", "CronSchedule", t.CronSchedule)
	}
	if t.CronSchedule != "" {
		return t.CronSchedule
	}
	return "0/30 * * * *"
}

func (t *RebalanceTask) Run(ctx context.Context) error {
	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "rebalance start")
	}

	err := Compute(ctx, t.Alpaca.Client, t.FirestoreDB)
	if err != nil {
		if t.Logger != nil {
			_ = level.Error(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "rebalance compute failed", "AccountID", t.Config.Alpaca.AccountID, "err", err)
		}
		return err
	}

	if t.Logger != nil {
		_ = level.Info(t.Logger).Log("component", "scheduler", "task", t.Name(), "msg", "rebalance complete")
	}
	return nil
}

type SymbolAction struct {
	Category     string
	Symbol       string
	TargetValue  float64
	CurrentValue float64
	Delta        float64
}

// Compute fetches live account/position data from Alpaca and the portfolio
// config from Firestore, then returns the rebalance plan without placing orders.
func Compute(ctx context.Context, client *alpaca.Client, fs *db.Firestore) error {
	account, err := client.GetAccount()
	if err != nil {
		return err
	}

	positions, err := client.GetPositions()
	if err != nil {
		return err
	}

	totalCash := account.Cash.InexactFloat64()
	positionValues := make(map[string]float64, len(positions))
	var costBasis float64
	for _, p := range positions {
		if p.MarketValue != nil {
			// cost basis of the position used to calculate total capital
			costBasis += p.CostBasis.InexactFloat64()
			positionValues[p.Symbol] = p.CostBasis.InexactFloat64()
		}
	}

	totalCapital := costBasis + totalCash

	categories, err := db.List[*models.Category](ctx, fs, db.CollectionAccounts+"/"+account.ID+"/"+db.CollectionAllocations)
	if err != nil {
		return err
	}
	for _, category := range categories {
		symbols, err := db.List[*models.Symbol](ctx, fs, db.CollectionPortfolios+"/"+account.ID+"/"+db.CollectionAllocations+"/"+category.ID+"/symbols")
		if err != nil {
			return err
		}
		rebalance, err := compute(totalCapital, category, symbols, positionValues)
		if err != nil {
			return err
		}
		print(rebalance)
		// update result
	}

	return nil
}

// compute is the pure computation extracted for testability.
func compute(totalCapital float64, category *models.Category, symbols []*models.Symbol, positionValues map[string]float64) ([]SymbolAction, error) {

	ret := make([]SymbolAction, 0, len(symbols))

	categoryTarget := totalCapital * category.Percentage
	for _, symbol := range symbols {
		symbolTarget := categoryTarget * symbol.Weight
		current := positionValues[symbol.ID]

		ret = append(ret, SymbolAction{
			Category:     category.ID,
			Symbol:       symbol.ID,
			TargetValue:  symbolTarget,
			CurrentValue: current,
			Delta:        symbolTarget - current,
		})
	}

	return ret, nil
}
