package stream

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
)

// MockFirestore for testing
type MockFirestore struct {
	portfolios map[string]*models.Portfolio
}

func (m *MockFirestore) GetPortfolio(id string) (*models.Portfolio, error) {
	if m.portfolios == nil {
		return nil, db.ErrDocumentNotFound
	}
	p, exists := m.portfolios[id]
	if !exists {
		return nil, db.ErrDocumentNotFound
	}
	return p, nil
}

// MockPortfolioDB implements the portfolio fetching for testing
func (m *MockFirestore) Get[T any](ctx context.Context, fs *db.Firestore, collection string, id string) (T, error) {
	var zero T
	if collection != db.CollectionPortfolios {
		return zero, db.ErrDocumentNotFound
	}

	if m.portfolios == nil {
		return zero, db.ErrDocumentNotFound
	}

	p, exists := m.portfolios[id]
	if !exists {
		return zero, db.ErrDocumentNotFound
	}

	// Type assertion and return
	if portfolio, ok := interface{}(p).(T); ok {
		return portfolio, nil
	}

	return zero, db.ErrInvalidType
}

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
	err := mgr.LoadSymbols(ctx, &db.Firestore{Client: mockFS}, "default")
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
	if err := mgr.LoadSymbols(ctx, &db.Firestore{Client: mockFS}, "default"); err != nil {
		t.Fatalf("Failed to load symbols: %v", err)
	}

	// Send a trade update for AAPL
	tu := alpaca.TradeUpdate{
		Symbol: "AAPL",
		Event:  "fill",
		Fill: &alpaca.Fill{
			Price: 150.25,
			Qty:   10,
		},
	}

	mgr.OnMessage(tu)

	// Give goroutine time to process
	time.Sleep(100 * time.Millisecond)

	// Send message for unknown symbol (should not panic)
	tu.Symbol = "UNKNOWN"
	mgr.OnMessage(tu)

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
	if err := mgr.LoadSymbols(ctx, &db.Firestore{Client: mockFS}, "default"); err != nil {
		t.Fatalf("Failed to load symbols: %v", err)
	}

	// Send some messages
	for i := 0; i < 5; i++ {
		mgr.OnMessage(alpaca.TradeUpdate{Symbol: "AAPL", Event: "fill"})
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
	if err := mgr.LoadSymbols(ctx, &db.Firestore{Client: mockFS}, "default"); err != nil {
		t.Fatalf("Failed to load symbols: %v", err)
	}

	// Send more messages than channel can handle (will drop some)
	droppedCount := int32(0)
	for i := 0; i < 100; i++ {
		mgr.OnMessage(alpaca.TradeUpdate{Symbol: "AAPL", Event: "fill"})
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
	if err := mgr.LoadSymbols(ctx, &db.Firestore{Client: mockFS}, "default"); err != nil {
		t.Fatalf("Failed to load symbols: %v", err)
	}

	symbols = mgr.GetSymbols()
	if len(symbols) != 2 {
		t.Errorf("Expected 2 symbols after loading, got %d", len(symbols))
	}
}
