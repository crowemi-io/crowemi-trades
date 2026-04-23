package db

import (
	"regexp"
	"time"

	"cloud.google.com/go/civil"
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

var optionSymbolRegex = regexp.MustCompile(`^([A-Z]{1,6})\d{6}[CP]\d{8}$`)

func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrBool(v bool) *bool {
	return &v
}

func ptrInt64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func ptrInt32(v int32) *int32 {
	if v == 0 {
		return nil
	}
	return &v
}

func decimalToPgtype(d decimal.Decimal) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(d.String()); err != nil {
		return pgtype.Numeric{}
	}
	return n
}

func decimalPtrToPgtype(d *decimal.Decimal) pgtype.Numeric {
	if d == nil {
		return pgtype.Numeric{}
	}
	return decimalToPgtype(*d)
}

func timeToPgtype(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t.UTC(), Valid: true}
}

func timeToPgtypeTz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func timePtrToPgtypeTz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return timeToPgtypeTz(*t)
}

func civilDateToPgtype(d civil.Date) pgtype.Date {
	return pgtype.Date{
		Time:  time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func isOptionSymbol(symbol string) bool {
	return optionSymbolRegex.MatchString(symbol)
}

func getSymbolDerived(symbol string) *string {
	match := optionSymbolRegex.FindStringSubmatch(symbol)
	if len(match) != 2 {
		return nil
	}
	return ptrString(match[1])
}

func computeOptionsIncome(a *alpaca.AccountActivity) pgtype.Numeric {
	if a == nil || !isOptionSymbol(a.Symbol) {
		return pgtype.Numeric{}
	}
	income := a.PerShareAmount.Mul(a.Qty)
	return decimalToPgtype(income)
}

func AccountParamsFromAlpaca(account *alpaca.Account) sqlc.UpsertAccountParams {
	name := account.AccountNumber
	if name == "" {
		name = account.ID
	}

	return sqlc.UpsertAccountParams{
		Name:                     name,
		AccountNumber:            ptrString(account.AccountNumber),
		AlpacaID:                 ptrString(account.ID),
		Status:                   ptrString(account.Status),
		CryptoStatus:             ptrString(account.CryptoStatus),
		Currency:                 ptrString(account.Currency),
		BuyingPower:              decimalToPgtype(account.BuyingPower),
		RegtBuyingPower:          decimalToPgtype(account.RegTBuyingPower),
		DaytradingBuyingPower:    decimalToPgtype(account.DaytradingBuyingPower),
		EffectiveBuyingPower:     decimalToPgtype(account.EffectiveBuyingPower),
		NonMarginableBuyingPower: decimalToPgtype(account.NonMarginBuyingPower),
		BodDtbp:                  decimalToPgtype(account.BodDtbp),
		Cash:                     decimalToPgtype(account.Cash),
		AccruedFees:              decimalToPgtype(account.AccruedFees),
		PortfolioValue:           decimalToPgtype(account.PortfolioValue),
		PatternDayTrader:         ptrBool(account.PatternDayTrader),
		TradingBlocked:           ptrBool(account.TradingBlocked),
		TransfersBlocked:         ptrBool(account.TransfersBlocked),
		AccountBlocked:           ptrBool(account.AccountBlocked),
		ShortingEnabled:          ptrBool(account.ShortingEnabled),
		TradeSuspendedByUser:     ptrBool(account.TradeSuspendedByUser),
		Multiplier:               decimalToPgtype(account.Multiplier),
		Equity:                   decimalToPgtype(account.Equity),
		LastEquity:               decimalToPgtype(account.LastEquity),
		LongMarketValue:          decimalToPgtype(account.LongMarketValue),
		ShortMarketValue:         decimalToPgtype(account.ShortMarketValue),
		PositionMarketValue:      decimalToPgtype(account.PositionMarketValue),
		InitialMargin:            decimalToPgtype(account.InitialMargin),
		MaintenanceMargin:        decimalToPgtype(account.MaintenanceMargin),
		LastMaintenanceMargin:    decimalToPgtype(account.LastMaintenanceMargin),
		Sma:                      decimalToPgtype(account.SMA),
		DaytradeCount:            ptrInt64(account.DaytradeCount),
		CryptoTier:               ptrInt32(int32(account.CryptoTier)),
		AlpacaCreatedAt:          timeToPgtype(account.CreatedAt),
	}
}

func ActivityParamsFromAlpaca(accountID int32, a *alpaca.AccountActivity) sqlc.CreateActivityParams {
	return sqlc.CreateActivityParams{
		AccountID:       accountID,
		ActivityID:      a.ID,
		ActivityType:    a.ActivityType,
		TransactionTime: timeToPgtype(a.TransactionTime),
		Type:            ptrString(a.Type),
		Price:           decimalToPgtype(a.Price),
		Qty:             decimalToPgtype(a.Qty),
		Side:            ptrString(a.Side),
		Symbol:          ptrString(a.Symbol),
		LeavesQty:       decimalToPgtype(a.LeavesQty),
		CumQty:          decimalToPgtype(a.CumQty),
		Date:            civilDateToPgtype(a.Date),
		NetAmount:       decimalToPgtype(a.NetAmount),
		Description:     ptrString(a.Description),
		PerShareAmount:  decimalToPgtype(a.PerShareAmount),
		OrderID:         ptrString(a.OrderID),
		OrderStatus:     ptrString(a.OrderStatus),
		Status:          ptrString(a.Status),
		IsOption:        ptrBool(isOptionSymbol(a.Symbol)),
		OptionsIncome:   computeOptionsIncome(a),
		SymbolDerived:   getSymbolDerived(a.Symbol),
	}
}

func OrderParamsFromAlpaca(accountID int64, o *alpaca.Order) sqlc.UpsertOrderParams {
	return sqlc.UpsertOrderParams{
		AccountID:       accountID,
		ClientOrderID:   o.ClientOrderID,
		AssetID:         ptrString(o.AssetID),
		Symbol:          ptrString(o.Symbol),
		Side:            ptrString(string(o.Side)),
		Type:            ptrString(string(o.Type)),
		OrderClass:      ptrString(string(o.OrderClass)),
		TimeInForce:     ptrString(string(o.TimeInForce)),
		Status:          ptrString(o.Status),
		Quantity:        decimalPtrToPgtype(o.Qty),
		Notional:        decimalPtrToPgtype(o.Notional),
		FilledQty:       decimalToPgtype(o.FilledQty),
		FilledAvgPrice:  decimalPtrToPgtype(o.FilledAvgPrice),
		LimitPrice:      decimalPtrToPgtype(o.LimitPrice),
		StopPrice:       decimalPtrToPgtype(o.StopPrice),
		AlpacaCreatedAt: timeToPgtypeTz(o.CreatedAt),
		AlpacaUpdatedAt: timeToPgtypeTz(o.UpdatedAt),
		SubmittedAt:     timeToPgtypeTz(o.SubmittedAt),
		FilledAt:        timePtrToPgtypeTz(o.FilledAt),
		ExpiredAt:       timePtrToPgtypeTz(o.ExpiredAt),
		CanceledAt:      timePtrToPgtypeTz(o.CanceledAt),
	}
}

func PositionParamsFromAlpaca(accountID int64, p *alpaca.Position) sqlc.UpsertPositionParams {
	return sqlc.UpsertPositionParams{
		AccountID:     accountID,
		Symbol:        p.Symbol,
		Quantity:      decimalToPgtype(p.Qty),
		AvgEntryPrice: decimalToPgtype(p.AvgEntryPrice),
		MarketValue:   decimalPtrToPgtype(p.MarketValue),
		UnrealizedPl:  decimalPtrToPgtype(p.UnrealizedPL),
		IsCurrent:     ptrBool(true),
		RecordedAt:    timeToPgtypeTz(time.Now()),
	}
}

func CorporateActionParamsFromAlpaca(accountID int32, a *alpaca.Announcement) sqlc.UpsertCorporateActionParams {
	now := time.Now().UTC()
	return sqlc.UpsertCorporateActionParams{
		AccountID:               accountID,
		CorporateActionsID:      ptrString(a.CorporateActionsID),
		CaType:                  ptrString(a.CAType),
		CaSubType:               ptrString(a.CASubType),
		InitiatingSymbol:        ptrString(a.InitiatingSymbol),
		InitiatingOriginalCusip: ptrString(a.InitiatingOriginalCusip),
		TargetSymbol:            ptrString(a.TargetSymbol),
		TargetOriginalCusip:     ptrString(a.TargetOriginalCusip),
		DeclarationDate:         ptrString(a.DeclarationDate),
		ExpirationDate:          ptrString(a.ExpirationDate),
		RecordDate:              ptrString(a.RecordDate),
		PayableDate:             ptrString(a.PayableDate),
		Cash:                    ptrString(a.Cash),
		OldRate:                 ptrString(a.OldRate),
		NewRate:                 ptrString(a.NewRate),
		LastSyncedAt:            timeToPgtypeTz(now),
	}
}
