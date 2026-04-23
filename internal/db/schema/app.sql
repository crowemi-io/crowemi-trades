CREATE SCHEMA IF NOT EXISTS app;

-- account
CREATE TABLE IF NOT EXISTS app.account (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,

    account_number TEXT UNIQUE,
    alpaca_id TEXT,
    status TEXT,
    crypto_status TEXT,
    currency TEXT,
    buying_power NUMERIC,
    regt_buying_power NUMERIC,
    daytrading_buying_power NUMERIC,
    effective_buying_power NUMERIC,
    non_marginable_buying_power NUMERIC,
    bod_dtbp NUMERIC,
    cash NUMERIC,
    accrued_fees NUMERIC,
    portfolio_value NUMERIC,
    pattern_day_trader BOOLEAN,
    trading_blocked BOOLEAN,
    transfers_blocked BOOLEAN,
    account_blocked BOOLEAN,
    shorting_enabled BOOLEAN,
    trade_suspended_by_user BOOLEAN,
    multiplier NUMERIC,
    equity NUMERIC,
    last_equity NUMERIC,
    long_market_value NUMERIC,
    short_market_value NUMERIC,
    position_market_value NUMERIC,
    initial_margin NUMERIC,
    maintenance_margin NUMERIC,
    last_maintenance_margin NUMERIC,
    sma NUMERIC,
    daytrade_count BIGINT,
    crypto_tier INTEGER,
    alpaca_created_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    updated_by TEXT NOT NULL
);

-- activity
CREATE TABLE IF NOT EXISTS app.activity (
    id BIGSERIAL PRIMARY KEY,
    account_id INT NOT NULL,
    activity_id TEXT NOT NULL,
    activity_type TEXT NOT NULL,
    transaction_time TIMESTAMP NOT NULL,
    type TEXT,
    price NUMERIC,
    qty NUMERIC,
    side TEXT,
    symbol TEXT,
    leaves_qty NUMERIC,
    cum_qty NUMERIC,
    date DATE,
    net_amount NUMERIC,
    description TEXT,
    per_share_amount NUMERIC,
    order_id TEXT,
    order_status TEXT,
    status TEXT,
    is_option BOOLEAN DEFAULT FALSE,
    options_income NUMERIC,
    symbol_derived TEXT,
    created_at TIMESTAMP NOT NULL,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    updated_by TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES app.account(id) ON DELETE CASCADE,
    CONSTRAINT uidx_activity_id UNIQUE (account_id, activity_id)
);

-- corporate_action
-- order
-- portfolio
CREATE TABLE IF NOT EXISTS app.portfolio (
    id BIGSERIAL PRIMARY KEY,
    account_id INT NOT NULL,
    portfolio_key TEXT NULL,
    name TEXT NOT NULL,
    weight DOUBLE PRECISION NOT NULL,

    created_at TIMESTAMP NOT NULL,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    updated_by TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES app.account(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX uidx_account_name
ON app.portfolio (account_id, name);

CREATE TABLE IF NOT EXISTS app.portfolio_symbol (
    id BIGSERIAL PRIMARY KEY,
    portfolio_id BIGINT NOT NULL,
    symbol TEXT NOT NULL,
    weight DOUBLE PRECISION NOT NULL,

    created_at TIMESTAMP NOT NULL,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    updated_by TEXT NOT NULL,
    FOREIGN KEY (portfolio_id) REFERENCES app.portfolio(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX uidx_portfolio_symbol
ON app.portfolio_symbol (portfolio_id, symbol);

-- order
CREATE TABLE IF NOT EXISTS app.order (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL,
    client_order_id TEXT UNIQUE NOT NULL,
    asset_id TEXT,
    symbol TEXT,
    side TEXT,
    type TEXT,
    order_class TEXT,
    time_in_force TEXT,
    status TEXT,
    quantity NUMERIC,
    notional NUMERIC,
    filled_qty NUMERIC,
    filled_avg_price NUMERIC,
    limit_price NUMERIC,
    stop_price NUMERIC,
    alpaca_created_at TIMESTAMPTZ,
    alpaca_updated_at TIMESTAMPTZ,
    submitted_at TIMESTAMPTZ,
    filled_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    created_at TIMESTAMP NOT NULL,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    updated_by TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES app.account(id) ON DELETE CASCADE
);

-- position
CREATE TABLE IF NOT EXISTS app.position (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL,
    symbol TEXT NOT NULL,
    quantity NUMERIC,
    avg_entry_price NUMERIC,
    market_value NUMERIC,
    unrealized_pl NUMERIC,
    is_current BOOLEAN DEFAULT TRUE,
    recorded_at TIMESTAMPTZ,
    created_at TIMESTAMP NOT NULL,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    updated_by TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES app.account(id) ON DELETE CASCADE,
    UNIQUE (account_id, symbol)
);

-- corporate_action
CREATE TABLE IF NOT EXISTS app.corporate_action (
    id BIGSERIAL PRIMARY KEY,
    account_id INT NOT NULL,
    corporate_actions_id TEXT UNIQUE,
    ca_type TEXT,
    ca_sub_type TEXT,
    initiating_symbol TEXT,
    initiating_original_cusip TEXT,
    target_symbol TEXT,
    target_original_cusip TEXT,
    declaration_date TEXT,
    expiration_date TEXT,
    record_date TEXT,
    payable_date TEXT,
    cash TEXT,
    old_rate TEXT,
    new_rate TEXT,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMP NOT NULL,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    updated_by TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES app.account(id) ON DELETE CASCADE
);

-- withholding
CREATE TABLE IF NOT EXISTS app.withholding (
    id BIGSERIAL PRIMARY KEY,
    account_id INT NOT NULL,
    symbol TEXT,
    tax_type TEXT,
    amount NUMERIC,
    currency TEXT,
    occurred_at TIMESTAMPTZ,
    description TEXT,
    created_at TIMESTAMP NOT NULL,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    updated_by TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES app.account(id) ON DELETE CASCADE
);
