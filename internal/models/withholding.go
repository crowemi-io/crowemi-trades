package models

import "time"

type Withholding struct {
	ID           string    `firestore:"-"`
	Symbol       string    `firestore:"symbol,omitempty"`
	TaxType      string    `firestore:"tax_type,omitempty"`
	Amount       float64   `firestore:"amount,omitempty"`
	Currency     string    `firestore:"currency,omitempty"`
	OccurredAt   time.Time `firestore:"occurred_at,omitempty"`
	Description  string    `firestore:"description,omitempty"`
	SysCreatedAt time.Time `firestore:"sys_created_at,omitempty"`
	SysUpdatedAt time.Time `firestore:"sys_updated_at,omitempty"`
}

func (m *Withholding) SetSysUpdate() { m.SysUpdatedAt = time.Now().UTC() }
func (m *Withholding) SetSysCreate() { m.SysCreatedAt = time.Now().UTC() }
func (m *Withholding) GetID() string { return m.ID }
func (m *Withholding) SetID(id string) {
	m.ID = id
}
