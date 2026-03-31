package stream

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	mdstream "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// MinuteBarsConsumer subscribes to Alpaca market data minute bars for the given symbols
// and runs until context is cancelled or the connection terminates.
type Watcher struct {
	Logger         kitlog.Logger
	Symbols        []string
	APIKey         string
	APISecret      string
	DataURL        string // optional; empty uses SDK default
	MarketDataFeed string // optional; "iex" (default) or "sip"
}

// Run connects to the market data stream, subscribes to minute bars for the consumer's symbols,
// and blocks until ctx is done or the connection terminates. Implements runtime stream runner.
func (c *Watcher) Run(ctx context.Context) error {
	if len(c.Symbols) == 0 {
		_ = level.Info(c.Logger).Log("component", "stream", "msg", "no symbols, minute bars consumer idle")
		<-ctx.Done()
		return nil
	}

	opts := []mdstream.StockOption{
		mdstream.WithCredentials(c.APIKey, c.APISecret),
		mdstream.WithBars(func(b mdstream.Bar) {
			_ = level.Info(c.Logger).Log(
				"component", "stream",
				"symbol", b.Symbol,
				"timestamp", b.Timestamp,
				"open", b.Open,
				"high", b.High,
				"low", b.Low,
				"close", b.Close,
				"volume", b.Volume,
				"msg", "minute bar",
			)
		}, c.Symbols...),
	}
	if c.DataURL != "" {
		opts = append(opts, mdstream.WithBaseURL(c.DataURL))
	}

	client := mdstream.NewStocksClient(feed(c.MarketDataFeed), opts...)
	return client.Connect(ctx)
}

func feed(m string) marketdata.Feed {
	switch m {
	case "sip":
		return marketdata.SIP
	default:
		return marketdata.IEX
	}
}
