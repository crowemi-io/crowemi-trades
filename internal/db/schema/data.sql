-- Market data time-series table
CREATE TABLE IF NOT EXISTS data.market_data (
    time TIMESTAMPTZ NOT NULL,
    symbol TEXT NOT NULL,
    open DOUBLE PRECISION NOT NULL,
    high DOUBLE PRECISION NOT NULL,
    low DOUBLE PRECISION NOT NULL,
    close DOUBLE PRECISION NOT NULL,
    volume BIGINT NOT NULL,
    PRIMARY KEY (time, symbol)
);
-- Create hypertable for time-series data
SELECT create_hypertable('data.market_data', 'time')
WHERE NOT EXISTS (
        SELECT 1
        FROM pg_tables
        WHERE tablename = 'data.market_data'
    );
