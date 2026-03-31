package models

import (
	"regexp"
	"strconv"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

var optionSymbolRe = regexp.MustCompile(`^[A-Z]+[0-3][0-9][0-1][0-9][0-3][0-9][CP][0-9]+$`)
var optionDateRe = regexp.MustCompile(`[0-3][0-9][0-1][0-9][0-3][0-9]`)

type Activity struct {
	ID             string    `firestore:"-"`
	ActivityType   string    `firestore:"activity_type,omitempty"`
	Type           string    `firestore:"type,omitempty"`
	Symbol         string    `firestore:"symbol,omitempty"`
	Side           string    `firestore:"side,omitempty"`
	Quantity       string    `firestore:"quantity,omitempty"`
	Price          string    `firestore:"price,omitempty"`
	NetAmount      string    `firestore:"net_amount,omitempty"`
	PerShareAmount string    `firestore:"per_share_amount,omitempty"`
	OrderID        string    `firestore:"order_id,omitempty"`
	OrderStatus    string    `firestore:"order_status,omitempty"`
	Status         string    `firestore:"status,omitempty"`
	Description    string    `firestore:"description,omitempty"`
	OccurredAt     time.Time `firestore:"occurred_at,omitempty"`
	IsOption       bool      `firestore:"is_option"`
	OptionsIncome  float64   `firestore:"options_income"`
	SymbolDerived  string    `firestore:"symbol_derived,omitempty"`
	SysCreatedAt   time.Time `firestore:"sys_created_at,omitempty"`
	SysUpdatedAt   time.Time `firestore:"sys_updated_at,omitempty"`
}

func (m *Activity) SetSysUpdate() { m.SysUpdatedAt = time.Now().UTC() }
func (m *Activity) SetSysCreate() { m.SysCreatedAt = time.Now().UTC() }
func (m *Activity) GetID() string { return m.ID }
func (m *Activity) SetID(id string) {
	m.ID = id
}

func (m *Activity) ComputeDerivedFields() {
	m.IsOption = m.Symbol != "" && optionSymbolRe.MatchString(m.Symbol)

	m.OptionsIncome = 0
	if m.IsOption && m.ActivityType == "FILL" {
		price, _ := strconv.ParseFloat(m.Price, 64)
		qty, _ := strconv.ParseFloat(m.Quantity, 64)
		base := price * 100 * qty
		if m.Side == "buy" {
			base = -base
		}
		m.OptionsIncome = base
	}

	m.SymbolDerived = m.getSymbolDerived()
}

func (m *Activity) getSymbolDerived() string {
	if m.Symbol == "" {
		return ""
	}
	if optionSymbolRe.MatchString(m.Symbol) {
		loc := optionDateRe.FindStringIndex(m.Symbol)
		if loc != nil {
			return m.Symbol[:loc[0]]
		}
	}
	return m.Symbol
}

func ActivityFromAlpaca(a *alpaca.AccountActivity) *Activity {
	if a == nil {
		return nil
	}

	act := &Activity{
		ID:             a.ID,
		ActivityType:   a.ActivityType,
		Type:           a.Type,
		Symbol:         a.Symbol,
		Side:           a.Side,
		Quantity:       a.Qty.String(),
		Price:          a.Price.String(),
		NetAmount:      a.NetAmount.String(),
		PerShareAmount: a.PerShareAmount.String(),
		OrderID:        a.OrderID,
		OrderStatus:    a.OrderStatus,
		Status:         a.Status,
		Description:    a.Description,
		OccurredAt:     a.TransactionTime,
	}
	act.ComputeDerivedFields()
	return act
}
