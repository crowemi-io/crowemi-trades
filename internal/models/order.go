package models

import "time"

type Order struct {
	ID          string    `firestore:"-"`
	Symbol      string    `firestore:"symbol,omitempty"`
	Side        string    `firestore:"side,omitempty"`
	Type        string    `firestore:"type,omitempty"`
	TimeInForce string    `firestore:"time_in_force,omitempty"`
	Quantity    string    `firestore:"quantity,omitempty"`
	Status      string    `firestore:"status,omitempty"`
	SubmittedAt time.Time `firestore:"submitted_at,omitempty"`
	FilledAt    time.Time `firestore:"filled_at,omitempty"`
}

func (m *Order) GetID() string { return m.ID }
func (m *Order) SetID(id string) {
	m.ID = id
}
