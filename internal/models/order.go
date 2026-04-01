package models

import (
	"reflect"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

type Order struct {
	ClientOrderID  string    `firestore:"client_order_id,omitempty"`
	AssetID        string    `firestore:"asset_id,omitempty"`
	Symbol         string    `firestore:"symbol,omitempty"`
	Side           string    `firestore:"side,omitempty"`
	Type           string    `firestore:"type,omitempty"`
	OrderClass     string    `firestore:"order_class,omitempty"`
	TimeInForce    string    `firestore:"time_in_force,omitempty"`
	Status         string    `firestore:"status,omitempty"`
	Quantity       string    `firestore:"quantity,omitempty"`
	Notional       string    `firestore:"notional,omitempty"`
	FilledQty      string    `firestore:"filled_qty,omitempty"`
	FilledAvgPrice string    `firestore:"filled_avg_price,omitempty"`
	LimitPrice     string    `firestore:"limit_price,omitempty"`
	StopPrice      string    `firestore:"stop_price,omitempty"`
	CreatedAt      time.Time `firestore:"created_at,omitempty"`
	UpdatedAt      time.Time `firestore:"updated_at,omitempty"`
	SubmittedAt    time.Time `firestore:"submitted_at,omitempty"`
	FilledAt       time.Time `firestore:"filled_at,omitempty"`
	ExpiredAt      time.Time `firestore:"expired_at,omitempty"`
	CanceledAt     time.Time `firestore:"canceled_at,omitempty"`
	SysCreatedAt   time.Time `firestore:"sys_created_at,omitempty"`
	SysUpdatedAt   time.Time `firestore:"sys_updated_at,omitempty"`
}

func (m *Order) SetSysUpdate() { m.SysUpdatedAt = time.Now().UTC() }
func (m *Order) SetSysCreate() { m.SysCreatedAt = time.Now().UTC() }
func (m *Order) GetID() string { return m.ClientOrderID }
func (m *Order) SetID(id string) {
	m.ClientOrderID = id
}

func OrderFromAlpaca(o *alpaca.Order) *Order {
	if o == nil {
		return nil
	}

	return &Order{
		ClientOrderID:  o.ClientOrderID,
		AssetID:        o.AssetID,
		Symbol:         o.Symbol,
		Side:           string(o.Side),
		Type:           string(o.Type),
		OrderClass:     string(o.OrderClass),
		TimeInForce:    string(o.TimeInForce),
		Status:         o.Status,
		Quantity:       decimalPtrToString(o.Qty),
		Notional:       decimalPtrToString(o.Notional),
		FilledQty:      o.FilledQty.String(),
		FilledAvgPrice: decimalPtrToString(o.FilledAvgPrice),
		LimitPrice:     decimalPtrToString(o.LimitPrice),
		StopPrice:      decimalPtrToString(o.StopPrice),
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
		SubmittedAt:    o.SubmittedAt,
		FilledAt:       timePtrValue(o.FilledAt),
		ExpiredAt:      timePtrValue(o.ExpiredAt),
		CanceledAt:     timePtrValue(o.CanceledAt),
	}
}

func decimalPtrToString(d interface{ String() string }) string {
	if d == nil {
		return ""
	}
	val := reflect.ValueOf(d)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return ""
	}
	return d.String()
}

func timePtrValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
