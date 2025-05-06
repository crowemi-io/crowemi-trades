package crowemi_trades

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-go-utils/config"
)

type Alpaca struct {
	Config config.Alpaca
	Client *alpaca.Client
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
