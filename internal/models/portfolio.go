package models

import "encoding/json"

type Allocation struct {
	Rebalance  bool               `firestore:"rebalance" json:"rebalance"`
	Percentage float64            `firestore:"percentage" json:"percentage"`
	Symbols    map[string]float64 `firestore:"symbols,omitempty" json:"symbols"`
}

type Portfolio struct {
	ID          string                `firestore:"-"`
	Allocations map[string]Allocation `firestore:"allocations"`
}

func (m *Portfolio) GetID() string { return m.ID }
func (m *Portfolio) SetID(id string) {
	m.ID = id
}

func (m *Portfolio) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.Allocations)
}
