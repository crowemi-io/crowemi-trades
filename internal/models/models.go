package models

import (
	"encoding/json"
	"regexp"
	"strconv"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

type Account struct {
	ID                    string    `firestore:"-"`
	AccountNumber         string    `firestore:"account_number,omitempty"`
	Status                string    `firestore:"status,omitempty"`
	CryptoStatus          string    `firestore:"crypto_status,omitempty"`
	Currency              string    `firestore:"currency,omitempty"`
	BuyingPower           string    `firestore:"buying_power,omitempty"`
	RegTBuyingPower       string    `firestore:"regt_buying_power,omitempty"`
	DaytradingBuyingPower string    `firestore:"daytrading_buying_power,omitempty"`
	EffectiveBuyingPower  string    `firestore:"effective_buying_power,omitempty"`
	NonMarginBuyingPower  string    `firestore:"non_marginable_buying_power,omitempty"`
	BodDtbp               string    `firestore:"bod_dtbp,omitempty"`
	Cash                  string    `firestore:"cash,omitempty"`
	AccruedFees           string    `firestore:"accrued_fees,omitempty"`
	PortfolioValue        string    `firestore:"portfolio_value,omitempty"`
	PatternDayTrader      bool      `firestore:"pattern_day_trader,omitempty"`
	TradingBlocked        bool      `firestore:"trading_blocked,omitempty"`
	TransfersBlocked      bool      `firestore:"transfers_blocked,omitempty"`
	AccountBlocked        bool      `firestore:"account_blocked,omitempty"`
	ShortingEnabled       bool      `firestore:"shorting_enabled,omitempty"`
	TradeSuspendedByUser  bool      `firestore:"trade_suspended_by_user,omitempty"`
	CreatedAt             time.Time `firestore:"created_at,omitempty"`
	Multiplier            string    `firestore:"multiplier,omitempty"`
	Equity                string    `firestore:"equity,omitempty"`
	LastEquity            string    `firestore:"last_equity,omitempty"`
	LongMarketValue       string    `firestore:"long_market_value,omitempty"`
	ShortMarketValue      string    `firestore:"short_market_value,omitempty"`
	PositionMarketValue   string    `firestore:"position_market_value,omitempty"`
	InitialMargin         string    `firestore:"initial_margin,omitempty"`
	MaintenanceMargin     string    `firestore:"maintenance_margin,omitempty"`
	LastMaintenanceMargin string    `firestore:"last_maintenance_margin,omitempty"`
	SMA                   string    `firestore:"sma,omitempty"`
	DaytradeCount         int64     `firestore:"daytrade_count,omitempty"`
	CryptoTier            int       `firestore:"crypto_tier,omitempty"`
}

func (m *Account) GetID() string { return m.ID }
func (m *Account) SetID(id string) {
	m.ID = id
}

func AccountFromAlpaca(acct *alpaca.Account) *Account {
	if acct == nil {
		return nil
	}

	return &Account{
		ID:                    acct.ID,
		AccountNumber:         acct.AccountNumber,
		Status:                acct.Status,
		CryptoStatus:          acct.CryptoStatus,
		Currency:              acct.Currency,
		BuyingPower:           acct.BuyingPower.String(),
		RegTBuyingPower:       acct.RegTBuyingPower.String(),
		DaytradingBuyingPower: acct.DaytradingBuyingPower.String(),
		EffectiveBuyingPower:  acct.EffectiveBuyingPower.String(),
		NonMarginBuyingPower:  acct.NonMarginBuyingPower.String(),
		BodDtbp:               acct.BodDtbp.String(),
		Cash:                  acct.Cash.String(),
		AccruedFees:           acct.AccruedFees.String(),
		PortfolioValue:        acct.PortfolioValue.String(),
		PatternDayTrader:      acct.PatternDayTrader,
		TradingBlocked:        acct.TradingBlocked,
		TransfersBlocked:      acct.TransfersBlocked,
		AccountBlocked:        acct.AccountBlocked,
		ShortingEnabled:       acct.ShortingEnabled,
		TradeSuspendedByUser:  acct.TradeSuspendedByUser,
		CreatedAt:             acct.CreatedAt,
		Multiplier:            acct.Multiplier.String(),
		Equity:                acct.Equity.String(),
		LastEquity:            acct.LastEquity.String(),
		LongMarketValue:       acct.LongMarketValue.String(),
		ShortMarketValue:      acct.ShortMarketValue.String(),
		PositionMarketValue:   acct.PositionMarketValue.String(),
		InitialMargin:         acct.InitialMargin.String(),
		MaintenanceMargin:     acct.MaintenanceMargin.String(),
		LastMaintenanceMargin: acct.LastMaintenanceMargin.String(),
		SMA:                   acct.SMA.String(),
		DaytradeCount:         acct.DaytradeCount,
		CryptoTier:            acct.CryptoTier,
	}
}

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
}

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

type CorporateAction struct {
	ID             string    `firestore:"-"`
	ActionType     string    `firestore:"action_type,omitempty"`
	Symbol         string    `firestore:"symbol,omitempty"`
	EffectiveDate  time.Time `firestore:"effective_date,omitempty"`
	DeclaredDate   time.Time `firestore:"declared_date,omitempty"`
	PayableDate    time.Time `firestore:"payable_date,omitempty"`
	RecordDate     time.Time `firestore:"record_date,omitempty"`
	CashAmount     float64   `firestore:"cash_amount,omitempty"`
	Description    string    `firestore:"description,omitempty"`
	LastSyncedAt   time.Time `firestore:"last_synced_at,omitempty"`
}

func (m *CorporateAction) GetID() string { return m.ID }
func (m *CorporateAction) SetID(id string) {
	m.ID = id
}

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

type Allocation struct {
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

type Position struct {
	ID             string    `firestore:"-"`
	Symbol         string    `firestore:"symbol,omitempty"`
	Quantity       string    `firestore:"quantity,omitempty"`
	AvgEntryPrice  string    `firestore:"avg_entry_price,omitempty"`
	MarketValue    float64   `firestore:"market_value,omitempty"`
	UnrealizedPL   float64   `firestore:"unrealized_pl,omitempty"`
	RecordedAt     time.Time `firestore:"recorded_at,omitempty"`
}

func (m *Position) GetID() string { return m.ID }
func (m *Position) SetID(id string) {
	m.ID = id
}

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
