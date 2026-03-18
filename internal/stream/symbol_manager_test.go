package stream

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	kitlog "github.com/go-kit/log"

	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
)

// MockFirestore for testing
type MockFirestore struct {
	portfolios map[string]*models.Portfolio
}

func (m *MockFirestore) GetPortfolio(ctx context.Context, portfolioID string) (*models.Portfolio, error) {
	if m.portfolios == nil {
		return nil, errDocumentNotFound
	}
	p, exists := m.portfolios[portfolioID]
	if !exists {
		return nil, errDocumentNotFound
	}
	return p, nil
}

// GetFromMock is a generic helper for tests that need to fetch a document from MockFirestore.
// Go disallows type parameters on methods, so this is a package-level function.
func GetFromMock[T any](m *MockFirestore, ctx context.Context, collection, id string) (T, error) {
	var zero T
	if collection != db.CollectionPortfolios {
		return zero, errDocumentNotFound
	}
	if m.portfolios == nil {
		return zero, errDocumentNotFound
	}
	p, exists := m.portfolios[id]
	if !exists {
		return zero, errDocumentNotFound
	}
	if portfolio, ok := any(p).(T); ok {
		return portfolio, nil
	}
	return zero, errInvalidType
}

var (
	errDocumentNotFound = errors.New("document not found")
	errInvalidType      = errors.New("invalid type")
)

func newTestLogger(t *testing.T) kitlog.Logger {
	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(&testWriter{t}))
	return kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.DefaultCaller)
}

type testWriter struct {
	t *testing.T
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

func TestSymbolManager_New(t *testing.T) {
	logger := newTestLogger(t)
	mgr := NewSymbolManager(logger)

	if mgr.logger != logger {
		t.Error("Expected logger to be set")
	}

	if mgr.symbols == nil {
		t.Error("Expected symbols map to be initialized")
	}
}

func TestSymbolManager_LoadSymbols(t *testing.T) {
	logger := newTestLogger(t)
	mgr := NewSymbolManager(logger)

	// Create mock portfolio with test symbols
	mockFS := &MockFirestore{
		portfolios: map[string]*models.Portfolio{
			"default": {
				ID: "default",
				Allocations: map[string]models.Allocation{
					"tech": {
						Percentage: 0.6,
						Symbols: map[string]float64{
							"AAPL": 0.5,
							"GOOGL": 0.3,
							"MSFT": 0.2,
						},
					},
				},
			},
		},
	}

	ctx := context.Background()
	err := mgr.LoadSymbols(ctx, mockFS, "default")
	if err != nil {
		t.Fatalf("Failed to load symbols: %v", err)
	}

	symbols := mgr.GetSymbols()
	if len(symbols) != 3 {
		t.Errorf("Expected 3 symbols, got %d: %v", len(symbols), symbols)
	}

	expectedSymbols := map[string]bool{
		"AAPL":  true,
		"GOOGL": true,
		"MSFT":  true,
	}

	for _, sym := range symbols {
		if !expectedSymbols[sym] {
			t.Errorf("Unexpected symbol: %s", sym)
		}
	}
}

func TestSymbolManager_OnMessage(t *testing.T) {
	logger := newTestLogger(t)
	mgr := NewSymbolManager(logger)

	// Load symbols first
	mockFS := &MockFirestore{
		portfolios: map[string]*models.Portfolio{
			"default": {
				ID: "default",
				Allocations: map[string]models.Allocation{
					"tech": {
						Symbols: map[string]float64{"AAPL": 1.0},
					},
				},
			},
		},
	}

	ctx := context.Background()
	if err := mgr.LoadSymbols(ctx, mockFS, "default"); err != nil {
		t.Fatalf("Failed to load symbols: %v", err)
	}

	// Send a trade update for AAPL (symbol lives on Order in alpaca.TradeUpdate)
	tu := alpaca.TradeUpdate{
		Event: "fill",
		Order: alpaca.Order{Symbol: "AAPL"},
	}

	mgr.OnMessage(tu)

	// Give goroutine time to process
	time.Sleep(100 * time.Millisecond)

	// Send message for unknown symbol (should not panic)
	tuUnknown := alpaca.TradeUpdate{Event: "fill", Order: alpaca.Order{Symbol: "UNKNOWN"}}
	mgr.OnMessage(tuUnknown)

	time.Sleep(50 * time.Millisecond)
}

func TestSymbolManager_Shutdown(t *testing.T) {
	logger := newTestLogger(t)
	mgr := NewSymbolManager(logger)

	// Load symbols
	mockFS := &MockFirestore{
		portfolios: map[string]*models.Portfolio{
			"default": {
				ID: "default",
				Allocations: map[string]models.Allocation{
					"test": {
						Symbols: map[string]float64{"AAPL": 1.0, "GOOGL": 1.0},
					},
				},
			},
		},
	}

	ctx := context.Background()
	if err := mgr.LoadSymbols(ctx, mockFS, "default"); err != nil {
		t.Fatalf("Failed to load symbols: %v", err)
	}

	// Send some messages
	for i := 0; i < 5; i++ {
		mgr.OnMessage(alpaca.TradeUpdate{Event: "fill", Order: alpaca.Order{Symbol: "AAPL"}})
	}

	// Shutdown should not block and should stop goroutines
	done := make(chan bool)
	go func() {
		mgr.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		// Success - shutdown completed
	case <-time.After(2 * time.Second):
		t.Error("Shutdown timed out")
	}
}

func TestSymbolManager_ChannelBackpressure(t *testing.T) {
	logger := newTestLogger(t)
	mgr := NewSymbolManager(logger)

	// Load symbols with small channel (buffered to 1)
	mockFS := &MockFirestore{
		portfolios: map[string]*models.Portfolio{
			"default": {
				ID: "default",
				Allocations: map[string]models.Allocation{
					"test": {
						Symbols: map[string]float64{"AAPL": 1.0},
					},
				},
			},
		},
	}

	ctx := context.Background()
	if err := mgr.LoadSymbols(ctx, mockFS, "default"); err != nil {
		t.Fatalf("Failed to load symbols: %v", err)
	}

	// Send more messages than channel can handle (will drop some)
	for i := 0; i < 100; i++ {
		mgr.OnMessage(alpaca.TradeUpdate{Event: "fill", Order: alpaca.Order{Symbol: "AAPL"}})
	}

	time.Sleep(100 * time.Millisecond)
	mgr.Shutdown()

	t.Log("Test completed without hanging (backpressure handling works)")
}

func TestSymbolManager_GetSymbols(t *testing.T) {
	logger := newTestLogger(t)
	mgr := NewSymbolManager(logger)

	// Initially should be empty
	symbols := mgr.GetSymbols()
	if len(symbols) != 0 {
		t.Errorf("Expected 0 symbols initially, got %d", len(symbols))
	}

	// Load symbols
	mockFS := &MockFirestore{
		portfolios: map[string]*models.Portfolio{
			"default": {
				ID: "default",
				Allocations: map[string]models.Allocation{
					"test": {
						Symbols: map[string]float64{"AAPL": 1.0, "GOOGL": 1.0},
					},
				},
			},
		},
	}

	ctx := context.Background()
	if err := mgr.LoadSymbols(ctx, mockFS, "default"); err != nil {
		t.Fatalf("Failed to load symbols: %v", err)
	}

	symbols = mgr.GetSymbols()
	if len(symbols) != 2 {
		t.Errorf("Expected 2 symbols after loading, got %d", len(symbols))
	}
}
