package trader

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

type Alpaca struct {
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
func (a *Alpaca) GetActivities(token string) ([]Activities, error) {
	var ret []Activities
	for {
		request := alpaca.GetAccountActivitiesRequest{
			PageToken: token,
			Direction: "ASC",
		}
		alpacaActivities, err := a.Client.GetAccountActivities(request)
		if err != nil {
			return nil, err
		}

		if len(alpacaActivities) == 0 {
			break
		}

		for _, a := range alpacaActivities {
			activity := Activities{
				ID:              a.ID,
				ActivityType:    a.ActivityType,
				TransactionTime: a.TransactionTime,
				Type:            a.Type,
				Price:           a.Price.InexactFloat64(),
				Qty:             a.Qty.InexactFloat64(),
				Side:            a.Side,
				Symbol:          a.Symbol,
				LeavesQty:       a.LeavesQty.InexactFloat64(),
				CumQty:          a.CumQty.InexactFloat64(),
				Date:            a.Date,
				NetAmount:       a.NetAmount.InexactFloat64(),
				Description:     a.Description,
				PerShareAmount:  a.PerShareAmount.InexactFloat64(),
				OrderID:         a.OrderID,
				OrderStatus:     a.OrderStatus,
				Status:          a.Status,
			}
			ret = append(ret, activity)
		}
		token = alpacaActivities[len(alpacaActivities)-1].ID
	}

	return ret, nil
}
func (alpaca *Alpaca) GetCash() (float64, error) {
	alpacaAccount, err := alpaca.Client.GetAccount()
	if err != nil {
		return 0, err
	}
	cashValue, exact := alpacaAccount.Cash.Float64()
	if !exact {
		// return cashValue!
		return cashValue, nil
	}
	return cashValue, nil
}
