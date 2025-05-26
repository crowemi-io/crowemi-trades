package trader

import (
	"time"

	"cloud.google.com/go/civil"
)

type Activities struct {
	ID              string     `bson:"id"`
	ActivityType    string     `bson:"activity_type"`
	TransactionTime time.Time  `bson:"transaction_time"`
	Type            string     `bson:"type"`
	Price           float64    `bson:"price"`
	Qty             float64    `bson:"qty"`
	Side            string     `bson:"side"`
	Symbol          string     `bson:"symbol"`
	LeavesQty       float64    `bson:"leaves_qty"`
	CumQty          float64    `bson:"cum_qty"`
	Date            civil.Date `bson:"date"`
	NetAmount       float64    `bson:"net_amount"`
	Description     string     `bson:"description"`
	PerShareAmount  float64    `bson:"per_share_amount"`
	OrderID         string     `bson:"order_id"`
	OrderStatus     string     `bson:"order_status"`
	Status          string     `bson:"status"`
}

func ProcessActivities() {

}
