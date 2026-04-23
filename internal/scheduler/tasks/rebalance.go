package task

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"

	ct "github.com/crowemi-io/crowemi-trades"
	cfg "github.com/crowemi-io/crowemi-trades/internal/config"
)

type RebalanceTask struct {
	Config       *cfg.Config
	Alpaca       *ct.Alpaca
	Queries      *sqlc.Queries
	Logger       kitlog.Logger
	CronSchedule string
	Options      cfg.TaskOptions
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

	err := Compute(ctx, t.Alpaca.Client, t.Queries)
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
// config from PostgreSQL, then returns the rebalance plan without placing orders.
func Compute(ctx context.Context, client *alpaca.Client, queries *sqlc.Queries) error {
	// Get account
	account, err := queries.GetAccountByAlpacaID(ctx, nil)
	if err != nil {
		return err
	}

	// Get positions
	positions, err := queries.ListPositionsByAccountID(ctx, int64(account.ID))
	if err != nil {
		return err
	}

	// Get portfolio symbols
	symbols, err := queries.ListPortfolioSymbolsByAccountID(ctx, account.ID)
	if err != nil {
		return err
	}

	// TODO: Implement rebalance logic
	// This is a stub - the actual rebalance logic would:
	// 1. Get current positions and their values
	// 2. Get target allocations from portfolio
	// 3. Calculate deltas
	// 4. Return SymbolAction slice with rebalance plan

	_ = positions
	_ = symbols

	return nil
}

// 	rebalance, err := compute(totalCapital, category, symbols, positionValues)
// 	if err != nil {
// 		return err
// 	}
// 	print(rebalance)
// 	// update result
// }

// // compute is the pure computation extracted for testability.
// func compute(totalCapital float64, category *models.Category, symbols []*models.Symbol, positionValues map[string]float64) ([]SymbolAction, error) {

// 	ret := make([]SymbolAction, 0, len(symbols))

// 	categoryTarget := totalCapital * category.Percentage
// 	for _, symbol := range symbols {
// 		symbolTarget := categoryTarget * symbol.Weight
// 		current := positionValues[symbol.ID]

// 		ret = append(ret, SymbolAction{
// 			Category:     category.ID,
// 			Symbol:       symbol.ID,
// 			TargetValue:  symbolTarget,
// 			CurrentValue: current,
// 			Delta:        symbolTarget - current,
// 		})
// 	}

// 	return ret, nil
// }
