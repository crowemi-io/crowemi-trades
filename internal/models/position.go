package models

import "time"

type Position struct {
	ID            string    `firestore:"-"`
	Symbol        string    `firestore:"symbol,omitempty"`
	Quantity      string    `firestore:"quantity,omitempty"`
	AvgEntryPrice string    `firestore:"avg_entry_price,omitempty"`
	MarketValue   float64   `firestore:"market_value,omitempty"`
	UnrealizedPL  float64   `firestore:"unrealized_pl,omitempty"`
	RecordedAt    time.Time `firestore:"recorded_at,omitempty"`
}

func (m *Position) GetID() string { return m.ID }
func (m *Position) SetID(id string) {
	m.ID = id
}
