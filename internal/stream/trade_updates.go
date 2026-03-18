package stream

import (
	"context"
	"errors"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type TradeUpdateStreamer interface {
	StreamTradeUpdates(ctx context.Context, handler func(alpaca.TradeUpdate), req alpaca.StreamTradeUpdatesRequest) error
}

type TradeUpdatesConsumer struct {
	Logger       kitlog.Logger
	Streamer     TradeUpdateStreamer
	ReconnectMin time.Duration
	ReconnectMax time.Duration
	SymbolMgr    *SymbolManager // optional: for per-symbol routing
}

func (c *TradeUpdatesConsumer) Run(ctx context.Context) error {
	if c.Streamer == nil {
		<-ctx.Done()
		return nil
	}

	minBackoff := c.ReconnectMin
	if minBackoff <= 0 {
		minBackoff = time.Second
	}
	maxBackoff := c.ReconnectMax
	if maxBackoff < minBackoff {
		maxBackoff = 30 * time.Second
	}

	req := alpaca.StreamTradeUpdatesRequest{}
	backoff := minBackoff
	for {
		err := c.Streamer.StreamTradeUpdates(ctx, func(tu alpaca.TradeUpdate) {
			req.Since = tu.At.Add(time.Nanosecond)

			// Route to SymbolManager if available for per-symbol processing
			if c.SymbolMgr != nil {
				c.SymbolMgr.OnMessage(tu)
			} else {
				// Fallback: log globally for backwards compatibility
				_ = level.Info(c.Logger).Log(
					"component", "stream",
					"event", tu.Event,
					"event_id", tu.EventID,
					"order_id", tu.Order.ID,
					"symbol", tu.Order.Symbol,
					"msg", "trade update received",
				)
			}
		}, req)

		if err == nil || errors.Is(err, context.Canceled) {
			return nil
		}

		_ = level.Error(c.Logger).Log("component", "stream", "msg", "stream error", "err", err, "retry_in", backoff.String())
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}
