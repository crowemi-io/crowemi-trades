package trader

import (
	"context"
	"fmt"
	"time"

	alpaca "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-go-utils/db"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID               string    `bson:"_id"`
	CreatedAt        time.Time `bson:"created_at,omitempty"`
	CreatedAtSession string    `bson:"created_at_session,omitempty"`
	UpdatedAt        time.Time `bson:"updated_at,omitempty"`
	UpdatedAtSession string    `bson:"updated_at_session,omitempty"`
	Type             string    `bson:"type,omitempty"`
	SubType          string    `bson:"sub_type,omitempty"`
	Symbol           string    `bson:"symbol,omitempty"`
	Quantity         float64   `bson:"quantity,omitempty"`
	Notional         float64   `bson:"notional,omitempty"`
	Profit           float64   `bson:"profit,omitempty"`
	BuyOrderID       string    `bson:"buy_order_id,omitempty"`
	BuyStatus        string    `bson:"buy_status,omitempty"`
	BuyPrice         float64   `bson:"buy_price,omitempty"`
	BuyAtUTC         time.Time `bson:"buy_at_utc,omitempty"`
	BuySession       string    `bson:"buy_session,omitempty"`
	SellOrderID      string    `bson:"sell_order_id,omitempty"`
	SellStatus       string    `bson:"sell_status,omitempty"`
	SellPrice        float64   `bson:"sell_price,omitempty"`
	SellAtUTC        time.Time `bson:"sell_at_utc,omitempty"`
	SellSession      string    `bson:"sell_session,omitempty"`
}

func GetOrders(mongoClient *db.MongoClient, filters []db.MongoFilter) (*[]Order, error) {
	res, err := db.GetMany[Order](context.TODO(), mongoClient, "orders", filters)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Rebalance(client *alpaca.Client, freeCapital *float64, portfolio *Portfolio) {
	for _, group := range portfolio.AllocationGroup {
		for _, allocation := range group.Allocation {
			// PercentDiff * Total Cost Basis ?
			// this is the total expected free capital available for symbol allocation
			symbolOutstandingCapital := ((portfolio.CurrentCostBasis + *freeCapital) * allocation.Percent) - allocation.Current.CostBasis
			fmt.Printf("Outstanding capital %s: %f\n", allocation.Symbol, symbolOutstandingCapital)
			if symbolOutstandingCapital > 1.00 { //TODO: find a better approach here
				notional := decimal.NewFromFloat(symbolOutstandingCapital).Round(2)
				fmt.Printf("Purchasing %s $%s\n", allocation.Symbol, notional)

				req := alpaca.PlaceOrderRequest{
					Symbol:      allocation.Symbol,
					Notional:    &notional,
					Side:        alpaca.Buy,
					Type:        alpaca.Market,
					TimeInForce: alpaca.Day,
				}
				_, err := client.PlaceOrder(req)
				if err != nil {
					print(err)
				}
			}
		}
	}
}
