package crowemi_trades

import (
	"context"
	"time"

	"github.com/crowemi-io/crowemi-go-utils/db"
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

func GetOrders(mongoClient *db.MongoClient, filters []db.MongoFilter) ([]Order, error) {
	// Implement the logic to get allowable investment
	res, err := db.GetMany[Order](context.TODO(), mongoClient, "orders", filters)
	if err != nil {
		return nil, err
	}
	return res, nil
}
