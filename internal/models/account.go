package models

import (
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

type Account struct {
	ID                    string    `firestore:"ID" json:"id"`
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
	SysCreatedAt          time.Time `firestore:"sys_created_at,omitempty"`
	SysUpdatedAt          time.Time `firestore:"sys_updated_at,omitempty"`
}

func (m *Account) SetSysUpdate() { m.SysUpdatedAt = time.Now().UTC() }
func (m *Account) SetSysCreate() { m.SysCreatedAt = time.Now().UTC() }
func (m *Account) GetID() string { return m.AccountNumber }
func (m *Account) SetID(id string) {
	m.AccountNumber = id
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

type Symbol struct {
	ID           string    `firestore:"id" json:"id"`
	Weight       float64   `firestore:"weight" json:"weight"`
	SysCreatedAt time.Time `firestore:"sys_created_at,omitempty"`
	SysUpdatedAt time.Time `firestore:"sys_updated_at,omitempty"`
}

func (m *Symbol) SetSysUpdate()   { m.SysUpdatedAt = time.Now().UTC() }
func (m *Symbol) SetSysCreate()   { m.SysCreatedAt = time.Now().UTC() }
func (m *Symbol) GetID() string   { return m.ID }
func (m *Symbol) SetID(id string) { m.ID = id }
func NewSymbol(name string, weight float64) Symbol {
	return Symbol{
		ID:     name,
		Weight: weight,
	}
}

type Category struct {
	ID           string    `firestore:"id" json:"id"`
	Rebalance    bool      `firestore:"rebalance" json:"rebalance"`
	Percentage   float64   `firestore:"percentage" json:"percentage"`
	SysCreatedAt time.Time `firestore:"sys_created_at,omitempty"`
	SysUpdatedAt time.Time `firestore:"sys_updated_at,omitempty"`
}

func (m *Category) SetSysUpdate()   { m.SysUpdatedAt = time.Now().UTC() }
func (m *Category) SetSysCreate()   { m.SysCreatedAt = time.Now().UTC() }
func (m *Category) GetID() string   { return m.ID }
func (m *Category) SetID(id string) { m.ID = id }
func NewCategory(id string, rebalance bool, percentage float64) Category {
	return Category{
		ID:         id,
		Rebalance:  rebalance,
		Percentage: percentage,
	}
}

type Allocation struct {
	Rebalance    bool              `firestore:"rebalance" json:"rebalance"`
	Percentage   float64           `firestore:"percentage" json:"percentage"`
	Symbols      map[string]Symbol `firestore:"symbols,omitempty" json:"symbols"`
	SysCreatedAt time.Time         `firestore:"sys_created_at,omitempty"`
	SysUpdatedAt time.Time         `firestore:"sys_updated_at,omitempty"`
}

func (m *Allocation) SetSysUpdate() { m.SysUpdatedAt = time.Now().UTC() }
func (m *Allocation) SetSysCreate() { m.SysCreatedAt = time.Now().UTC() }
