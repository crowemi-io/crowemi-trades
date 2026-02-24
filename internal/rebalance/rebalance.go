package rebalance

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
)

type SymbolAction struct {
	Category     string
	Symbol       string
	TargetValue  float64
	CurrentValue float64
	Delta        float64
}

type Result struct {
	TotalCapital float64
	Actions      []SymbolAction
}

// Compute fetches live account/position data from Alpaca and the portfolio
// config from Firestore, then returns the rebalance plan without placing orders.
func Compute(ctx context.Context, client *alpaca.Client, fs *db.Firestore, portfolioID string) (*Result, error) {
	account, err := client.GetAccount()
	if err != nil {
		return nil, err
	}

	positions, err := client.GetPositions()
	if err != nil {
		return nil, err
	}

	portfolio, err := db.Get[*models.Portfolio](ctx, fs, db.CollectionPortfolios, portfolioID)
	if err != nil {
		return nil, err
	}

	totalCapital := account.PortfolioValue.InexactFloat64()

	positionValues := make(map[string]float64, len(positions))
	for _, p := range positions {
		if p.MarketValue != nil {
			positionValues[p.Symbol] = p.MarketValue.InexactFloat64()
		}
	}

	return compute(totalCapital, portfolio.Allocations, positionValues), nil
}

// compute is the pure computation extracted for testability.
func compute(totalCapital float64, allocations map[string]models.Allocation, positionValues map[string]float64) *Result {
	result := &Result{TotalCapital: totalCapital}

	allocated := make(map[string]bool)

	for category, alloc := range allocations {
		if len(alloc.Symbols) == 0 {
			continue
		}
		categoryTarget := totalCapital * alloc.Percentage
		for symbol, pct := range alloc.Symbols {
			symbolTarget := categoryTarget * pct
			current := positionValues[symbol]
			result.Actions = append(result.Actions, SymbolAction{
				Category:     category,
				Symbol:       symbol,
				TargetValue:  symbolTarget,
				CurrentValue: current,
				Delta:        symbolTarget - current,
			})
			allocated[symbol] = true
		}
	}

	for symbol, current := range positionValues {
		if allocated[symbol] {
			continue
		}
		result.Actions = append(result.Actions, SymbolAction{
			Category:     "",
			Symbol:       symbol,
			TargetValue:  0,
			CurrentValue: current,
			Delta:        -current,
		})
	}

	return result
}
