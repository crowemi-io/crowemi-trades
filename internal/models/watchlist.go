package models

import "time"

type Watchlist struct {
	ID          string    `firestore:"-"`
	Name        string    `firestore:"name,omitempty"`
	Description string    `firestore:"description,omitempty"`
	Symbols     []string  `firestore:"symbols,omitempty"`
	CreatedAt   time.Time `firestore:"created_at,omitempty"`
	UpdatedAt   time.Time `firestore:"updated_at,omitempty"`
}

func (m *Watchlist) GetID() string { return m.ID }
func (m *Watchlist) SetID(id string) {
	m.ID = id
}
