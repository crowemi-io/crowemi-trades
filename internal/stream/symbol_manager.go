package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
)

// SymbolManager manages per-symbol routing of trade updates to dedicated goroutines
type SymbolManager struct {
	logger  kitlog.Logger
	symbols map[string]chan alpaca.TradeUpdate // symbol -> channel
	mu      sync.RWMutex                       // protect symbols map
	wg      sync.WaitGroup                     // for graceful shutdown
}

// NewSymbolManager creates a new SymbolManager instance
func NewSymbolManager(logger kitlog.Logger) *SymbolManager {
	return &SymbolManager{
		logger:  logger,
		symbols: make(map[string]chan alpaca.TradeUpdate),
	}
}

// LoadSymbols loads portfolio symbols from Firestore and starts goroutines for each
func (m *SymbolManager) LoadSymbols(ctx context.Context, fs *db.Firestore, portfolioID string) error {
	m.logger.Log("msg", "loading portfolio symbols", "portfolio_id", portfolioID)

	portfolio, err := db.Get[*models.Portfolio](ctx, fs, db.CollectionPortfolios, portfolioID)
	if err != nil {
		return fmt.Errorf("failed to fetch portfolio: %w", err)
	}

	// Extract all symbols from allocations
	symbolSet := make(map[string]bool)
	for _, alloc := range portfolio.Allocations {
		for symbol := range alloc.Symbols {
			symbolSet[symbol] = true
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create channel and start goroutine for each symbol
	for symbol := range symbolSet {
		ch := make(chan alpaca.TradeUpdate, 100) // buffered to prevent message loss
		m.symbols[symbol] = ch
		m.startSymbolGoroutine(context.Background(), symbol, ch)
	}

	level.Info(m.logger).Log("msg", "loaded symbols", "count", len(symbolSet), "symbols", fmt.Sprintf("%v", symbolSet))
	return nil
}

// OnMessage routes incoming TradeUpdate to the correct symbol's goroutine
func (m *SymbolManager) OnMessage(tu alpaca.TradeUpdate) {
	m.mu.RLock()
	ch, exists := m.symbols[tu.Symbol]
	m.mu.RUnlock()

	if !exists {
		// Symbol not in portfolio, log globally for debugging
		level.Debug(m.logger).Log("msg", "trade update for unknown symbol", "symbol", tu.Symbol)
		return
	}

	// Non-blocking send with fallback to drop if channel is full
	select {
	case ch <- tu:
		// Message sent successfully
	default:
		// Channel full, log warning and drop message
		level.Warn(m.logger).Log("msg", "symbol channel full, dropping message", "symbol", tu.Symbol)
	}
}

// Shutdown gracefully stops all symbol goroutines
func (m *SymbolManager) Shutdown() {
	m.logger.Log("msg", "shutting down symbol manager")

	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all channels to signal goroutines to stop
	for symbol, ch := range m.symbols {
		close(ch)
		level.Debug(m.logger).Log("msg", "closed channel for symbol", "symbol", symbol)
	}

	// Wait for all goroutines to finish
	m.wg.Wait()

	m.symbols = make(map[string]chan alpaca.TradeUpdate)
	level.Info(m.logger).Log("msg", "symbol manager shut down")
}

// startSymbolGoroutine creates a dedicated goroutine for processing messages for a single symbol
func (m *SymbolManager) startSymbolGoroutine(ctx context.Context, symbol string, ch <-chan alpaca.TradeUpdate) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		for tu := range ch {
			// Log current price and symbol as per issue requirement
			switch tu.Event {
			case "fill":
				if tu.Fill != nil {
					level.Info(m.logger).Log(
						"component", "stream",
						"symbol", symbol,
						"event", tu.Event,
						"price", tu.Fill.Price,
						"qty", tu.Fill.Qty,
						"side", tu.Fill.Side,
						"msg", "trade fill received",
					)
				}
			case "outstanding":
				if tu.Outstanding != nil {
					level.Info(m.logger).Log(
						"component", "stream",
						"symbol", symbol,
						"event", tu.Event,
						"bid_price", tu.Outstanding.BidPrice,
						"ask_price", tu.Outstanding.AskPrice,
						"msg", "quote update received",
					)
				}
			case "status":
				if tu.Status != nil {
					level.Info(m.logger).Log(
						"component", "stream",
						"symbol", symbol,
						"event", tu.Event,
						"order_id", tu.Status.Order.ID,
						"msg", "order status update",
					)
				}
			default:
				level.Debug(m.logger).Log(
					"component", "stream",
					"symbol", symbol,
					"event", tu.Event,
					"msg", "trade update received",
				)
			}
		}

		level.Info(m.logger).Log("msg", "symbol goroutine stopped", "symbol", symbol)
	}()
}

// GetSymbols returns the current list of symbols being tracked (for testing/debugging)
func (m *SymbolManager) GetSymbols() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	symbols := make([]string, 0, len(m.symbols))
	for symbol := range m.symbols {
		symbols = append(symbols, symbol)
	}
	return symbols
}
