package stream

import (
	"context"
	"fmt"
	"sync"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/shopspring/decimal"

	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
)

// PortfolioGetter fetches a portfolio by ID (used by LoadSymbols and test mocks).
type PortfolioGetter interface {
	GetPortfolio(ctx context.Context, portfolioID string) (*models.Portfolio, error)
}

// FirestorePortfolioGetter adapts *db.Firestore to PortfolioGetter.
type FirestorePortfolioGetter struct{ FS *db.Firestore }

func (g *FirestorePortfolioGetter) GetPortfolio(ctx context.Context, portfolioID string) (*models.Portfolio, error) {
	return db.Get[*models.Portfolio](ctx, g.FS, db.CollectionPortfolios, portfolioID)
}

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

// LoadSymbols loads portfolio symbols from the given getter and starts goroutines for each
func (m *SymbolManager) LoadSymbols(ctx context.Context, getter PortfolioGetter, portfolioID string, allocationCategory ...string) error {
	m.logger.Log("msg", "loading portfolio symbols", "portfolio_id", portfolioID, "category", fmt.Sprintf("%v", allocationCategory))

	portfolio, err := getter.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("failed to fetch portfolio: %w", err)
	}

	// Extract symbols from allocations
	symbolSet := make(map[string]bool)
	
	if len(allocationCategory) > 0 && allocationCategory[0] != "" {
		// Filter for specific allocation category (app node)
		if alloc, exists := portfolio.Allocations[allocationCategory[0]]; exists {
			for symbol := range alloc.Symbols {
				symbolSet[symbol] = true
			}
		} else {
			m.logger.Log("msg", "allocation category not found", "category", allocationCategory[0])
		}
	} else {
		// Load all symbols from all allocations (backward compatibility)
		for _, alloc := range portfolio.Allocations {
			for symbol := range alloc.Symbols {
				symbolSet[symbol] = true
			}
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
	symbol := tu.Order.Symbol
	m.mu.RLock()
	ch, exists := m.symbols[symbol]
	m.mu.RUnlock()

	if !exists {
		// Symbol not in portfolio, log globally for debugging
		level.Debug(m.logger).Log("msg", "trade update for unknown symbol", "symbol", symbol)
		return
	}

	// Non-blocking send with fallback to drop if channel is full
	select {
	case ch <- tu:
		// Message sent successfully
	default:
		// Channel full, log warning and drop message
		level.Warn(m.logger).Log("msg", "symbol channel full, dropping message", "symbol", symbol)
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
		level.Info(m.logger).Log("msg", "closed channel for symbol", "symbol", symbol)
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
			// Log using TradeUpdate fields: Order, Price, Qty (no Fill/Outstanding/Status structs in SDK)
			switch tu.Event {
			case "fill":
				if tu.Price != nil || tu.Qty != nil {
					level.Info(m.logger).Log(
						"component", "stream",
						"symbol", symbol,
						"event", tu.Event,
						"price", ptrDecimalStr(tu.Price),
						"qty", ptrDecimalStr(tu.Qty),
						"side", tu.Order.Side,
						"msg", "trade fill received",
					)
				}
			case "outstanding":
				level.Info(m.logger).Log(
					"component", "stream",
					"symbol", symbol,
					"event", tu.Event,
					"order_id", tu.Order.ID,
					"msg", "outstanding order update",
				)
			case "status":
				level.Info(m.logger).Log(
					"component", "stream",
					"symbol", symbol,
					"event", tu.Event,
					"order_id", tu.Order.ID,
					"order_status", tu.Order.Status,
					"msg", "order status update",
				)
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

func ptrDecimalStr(d *decimal.Decimal) string {
	if d == nil {
		return ""
	}
	return d.String()
}
