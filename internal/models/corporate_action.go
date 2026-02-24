package models

import (
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

type CorporateAction struct {
	ID                      string    `firestore:"-"`
	CorporateActionsID      string    `firestore:"corporate_actions_id,omitempty"`
	CAType                  string    `firestore:"ca_type,omitempty"`
	CASubType               string    `firestore:"ca_sub_type,omitempty"`
	InitiatingSymbol        string    `firestore:"initiating_symbol,omitempty"`
	InitiatingOriginalCusip string    `firestore:"initiating_original_cusip,omitempty"`
	TargetSymbol            string    `firestore:"target_symbol,omitempty"`
	TargetOriginalCusip     string    `firestore:"target_original_cusip,omitempty"`
	DeclarationDate         string    `firestore:"declaration_date,omitempty"`
	ExpirationDate          string    `firestore:"expiration_date,omitempty"`
	RecordDate              string    `firestore:"record_date,omitempty"`
	PayableDate             string    `firestore:"payable_date,omitempty"`
	Cash                    string    `firestore:"cash,omitempty"`
	OldRate                 string    `firestore:"old_rate,omitempty"`
	NewRate                 string    `firestore:"new_rate,omitempty"`
	LastSyncedAt            time.Time `firestore:"last_synced_at,omitempty"`
}

func (m *CorporateAction) GetID() string { return m.ID }
func (m *CorporateAction) SetID(id string) {
	m.ID = id
}

func CorporateActionFromAlpaca(a *alpaca.Announcement) *CorporateAction {
	if a == nil {
		return nil
	}

	return &CorporateAction{
		ID:                      a.ID,
		CorporateActionsID:      a.CorporateActionsID,
		CAType:                  a.CAType,
		CASubType:               a.CASubType,
		InitiatingSymbol:        a.InitiatingSymbol,
		InitiatingOriginalCusip: a.InitiatingOriginalCusip,
		TargetSymbol:            a.TargetSymbol,
		TargetOriginalCusip:     a.TargetOriginalCusip,
		DeclarationDate:         a.DeclarationDate,
		ExpirationDate:          a.ExpirationDate,
		RecordDate:              a.RecordDate,
		PayableDate:             a.PayableDate,
		Cash:                    a.Cash,
		OldRate:                 a.OldRate,
		NewRate:                 a.NewRate,
		LastSyncedAt:            time.Now(),
	}
}
