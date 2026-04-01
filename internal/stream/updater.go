package stream

import (
	"context"
	"errors"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	ct "github.com/crowemi-io/crowemi-trades"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type Updater struct {
	Logger       kitlog.Logger
	Alpaca       *ct.Alpaca
	ReconnectMin time.Duration
	ReconnectMax time.Duration
}

func (c *Updater) Run(ctx context.Context) error {
	if c.Alpaca.Client == nil {
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
		err := c.Alpaca.Client.StreamTradeUpdates(ctx, func(tu alpaca.TradeUpdate) {
			req.Since = tu.At.Add(time.Nanosecond)
			_ = level.Info(c.Logger).Log(
				"component", "stream",
				"event", tu.Event,
				"event_id", tu.EventID,
				"order_id", tu.Order.ID,
				"symbol", tu.Order.Symbol,
				"status", tu.Order.Status,
				"msg", "trade update received",
			)
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
