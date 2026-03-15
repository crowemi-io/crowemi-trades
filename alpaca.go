package crowemi_trades

import (
	"context"
	"fmt"
	"log"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/notifier"
)

type Alpaca struct {
	Client   *alpaca.Client
	Notifier notifier.Notifier
}

func (alpaca *Alpaca) GetPositions() ([]alpaca.Position, error) {
	positions, err := alpaca.Client.GetPositions()
	if err != nil {
		return nil, err
	}
	return positions, nil
}
func (alpaca *Alpaca) GetPosition(symbol string) (*alpaca.Position, error) {
	position, err := alpaca.Client.GetPosition(symbol)
	if err != nil {
		return nil, err
	}
	return position, nil
}
func (alpaca *Alpaca) GetActivities() {}
func (alpaca *Alpaca) GetCash() (float64, error) {
	alpacaAccount, err := alpaca.Client.GetAccount()
	if err != nil {
		return 0, err
	}
	cashValue, exact := alpacaAccount.Cash.Float64()
	if !exact {
		// log warning
		return cashValue, nil
	}
	return cashValue, nil
}

// SetOrder is a shared stub that logs and optionally notifies for an order.
// It does not place an order with the broker.
func (a *Alpaca) SetOrder(ctx context.Context, symbol, side string, amount float64) {
	log.Printf("rebalance order: %s %s %.2f", side, symbol, amount)
	if a.Notifier != nil {
		_ = a.Notifier.Notify(ctx, "Rebalance order: "+side+" "+symbol+" amount "+fmt.Sprintf("%.2f", amount))
	}
}
