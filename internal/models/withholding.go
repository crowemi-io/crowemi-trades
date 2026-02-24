package models

import "time"

type Withholding struct {
	ID          string    `firestore:"-"`
	Symbol      string    `firestore:"symbol,omitempty"`
	TaxType     string    `firestore:"tax_type,omitempty"`
	Amount      float64   `firestore:"amount,omitempty"`
	Currency    string    `firestore:"currency,omitempty"`
	OccurredAt  time.Time `firestore:"occurred_at,omitempty"`
	Description string    `firestore:"description,omitempty"`
}

func (m *Withholding) GetID() string { return m.ID }
func (m *Withholding) SetID(id string) {
	m.ID = id
}
