package stream

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	kitlog "github.com/go-kit/log"
)

type testStreamer struct {
	attempts atomic.Int32
}

func (s *testStreamer) StreamTradeUpdates(ctx context.Context, _ func(alpaca.TradeUpdate), _ alpaca.StreamTradeUpdatesRequest) error {
	attempt := s.attempts.Add(1)
	if attempt < 3 {
		return errors.New("temporary stream failure")
	}

	<-ctx.Done()
	return context.Canceled
}

func TestTradeUpdatesConsumerReconnects(t *testing.T) {
	streamer := &testStreamer{}
	consumer := &TradeUpdatesConsumer{
		Logger:       kitlog.NewNopLogger(),
		Streamer:     streamer,
		ReconnectMin: 1 * time.Millisecond,
		ReconnectMax: 2 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	if err := consumer.Run(ctx); err != nil {
		t.Fatalf("consumer returned error: %v", err)
	}

	if streamer.attempts.Load() < 3 {
		t.Fatalf("expected reconnect attempts, got %d", streamer.attempts.Load())
	}
}
