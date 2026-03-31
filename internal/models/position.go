package models

import (
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

type Position struct {
	ID            string    `firestore:"-"`
	Symbol        string    `firestore:"symbol,omitempty"`
	Quantity      string    `firestore:"quantity,omitempty"`
	AvgEntryPrice string    `firestore:"avg_entry_price,omitempty"`
	MarketValue   float64   `firestore:"market_value,omitempty"`
	UnrealizedPL  float64   `firestore:"unrealized_pl,omitempty"`
	IsCurrent     bool      `firestore:"is_current,omitempty"`
	RecordedAt    time.Time `firestore:"recorded_at,omitempty"`
	SysCreatedAt  time.Time `firestore:"sys_created_at,omitempty"`
	SysUpdatedAt  time.Time `firestore:"sys_updated_at,omitempty"`
}

func (m *Position) SetSysUpdate() { m.SysUpdatedAt = time.Now().UTC() }
func (m *Position) SetSysCreate() { m.SysCreatedAt = time.Now().UTC() }
func (m *Position) GetID() string { return m.ID }
func (m *Position) SetID(id string) {
	m.ID = id
}

func PositionFromAlpaca(p *alpaca.Position) *Position {
	if p == nil {
		return nil
	}

	id := p.AssetID
	if id == "" {
		id = p.Symbol
	}

	return &Position{
		ID:            id,
		Symbol:        p.Symbol,
		Quantity:      p.Qty.String(),
		AvgEntryPrice: p.AvgEntryPrice.String(),
		MarketValue:   decimalPtrToFloat64(p.MarketValue),
		UnrealizedPL:  decimalPtrToFloat64(p.UnrealizedPL),
		IsCurrent:     true,
		RecordedAt:    time.Now(),
	}
}

func decimalPtrToFloat64(d interface{ InexactFloat64() float64 }) float64 {
	if d == nil {
		return 0
	}
	return d.InexactFloat64()
}
