-- name: UpsertAccount :one
INSERT INTO app.account (
    name, account_number, alpaca_id, status, crypto_status, currency,
    buying_power, regt_buying_power, daytrading_buying_power, effective_buying_power,
    non_marginable_buying_power, bod_dtbp, cash, accrued_fees, portfolio_value,
    pattern_day_trader, trading_blocked, transfers_blocked, account_blocked,
    shorting_enabled, trade_suspended_by_user, multiplier, equity, last_equity,
    long_market_value, short_market_value, position_market_value, initial_margin,
    maintenance_margin, last_maintenance_margin, sma, daytrade_count, crypto_tier,
    alpaca_created_at, created_at, created_by, updated_at, updated_by
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17,
    $18,
    $19,
    $20,
    $21,
    $22,
    $23,
    $24,
    $25,
    $26,
    $27,
    $28,
    $29,
    $30,
    $31,
    $32,
    $33,
    $34,
    NOW(),
    'system',
    NOW(),
    'system'
)
ON CONFLICT (account_number) DO UPDATE SET
    name = EXCLUDED.name,
    account_number = EXCLUDED.account_number,
    status = EXCLUDED.status,
    crypto_status = EXCLUDED.crypto_status,
    currency = EXCLUDED.currency,
    buying_power = EXCLUDED.buying_power,
    regt_buying_power = EXCLUDED.regt_buying_power,
    daytrading_buying_power = EXCLUDED.daytrading_buying_power,
    effective_buying_power = EXCLUDED.effective_buying_power,
    non_marginable_buying_power = EXCLUDED.non_marginable_buying_power,
    bod_dtbp = EXCLUDED.bod_dtbp,
    cash = EXCLUDED.cash,
    accrued_fees = EXCLUDED.accrued_fees,
    portfolio_value = EXCLUDED.portfolio_value,
    pattern_day_trader = EXCLUDED.pattern_day_trader,
    trading_blocked = EXCLUDED.trading_blocked,
    transfers_blocked = EXCLUDED.transfers_blocked,
    account_blocked = EXCLUDED.account_blocked,
    shorting_enabled = EXCLUDED.shorting_enabled,
    trade_suspended_by_user = EXCLUDED.trade_suspended_by_user,
    multiplier = EXCLUDED.multiplier,
    equity = EXCLUDED.equity,
    last_equity = EXCLUDED.last_equity,
    long_market_value = EXCLUDED.long_market_value,
    short_market_value = EXCLUDED.short_market_value,
    position_market_value = EXCLUDED.position_market_value,
    initial_margin = EXCLUDED.initial_margin,
    maintenance_margin = EXCLUDED.maintenance_margin,
    last_maintenance_margin = EXCLUDED.last_maintenance_margin,
    sma = EXCLUDED.sma,
    daytrade_count = EXCLUDED.daytrade_count,
    crypto_tier = EXCLUDED.crypto_tier,
    alpaca_created_at = EXCLUDED.alpaca_created_at,
    updated_at = NOW(),
    updated_by = 'system'
RETURNING *;

-- name: GetAccountByAlpacaID :one
SELECT * FROM app.account WHERE alpaca_id = $1;

-- name: GetAccountByNumber :one
SELECT * FROM app.account WHERE account_number = $1;

-- name: ListAccounts :many
SELECT * FROM app.account ORDER BY id DESC;

-- name: GetAccountByID :one
SELECT * FROM app.account WHERE id = $1;
