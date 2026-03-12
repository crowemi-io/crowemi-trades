# Plan: Implement WebSocket Streaming for Alpaca Trading (Issue #3)

## Issue Reference
- **Repository**: crowemi-io/crowemi-trades
- **Issue**: #3 - feat: implement websocket streaming
- **Branch**: Working from `dev` branch, creating feature branch off dev

## Objective (from issue)
> For each symbol in the app node in the portfolio we want to add to the alpaca web socket handler. When a message is received the symbol should be routed to its own goroutine to perform the task. For now, the task is to log the current price and symbol.

## Code Review Summary

### Existing Architecture Found:
1. **Portfolio Model** (`internal/models/portfolio.go`):
   ```go
   type Portfolio struct {
       Allocations map[string]Allocation `firestore:"allocations"`
   }
   // Allocation has: Percentage + Symbols (map[symbol]weight)
   ```

2. **Existing Streaming Infrastructure** (`internal/stream/trade_updates.go`):
   - `TradeUpdatesConsumer` already handles Alpaca websocket connections
   - Has reconnection logic with exponential backoff
   - Currently logs ALL messages globally without per-symbol routing

3. **Firestore Database Layer** (`internal/db/firestore.go`):
   - `Get[T]()` and `List[T]()`, functions for fetching documents
   - Collection constants: `CollectionPortfolios = "portfolios"`

4. **Rebalance Module** (`internal/rebalance/rebalance.go`):
   - Already fetches portfolio from Firestore using `db.Get[*models.Portfolio]()`
   - Demonstrates how to extract symbols from allocations

### Current Implementation Gap
The existing `TradeUpdatesConsumer.Run()` method has this handler:
```go
func(tu alpaca.TradeUpdate) {
    // Just logs globally, no per-symbol routing
    _ = level.Info(c.Logger).Log(...)
}
```

## Updated Implementation Plan

### Phase 1: Symbol Manager Module (NEW)
**File**: `internal/stream/symbol_manager.go`

Create a manager to track symbols and route messages:

```go
type SymbolManager struct {
    logger   kitlog.Logger
    symbols  map[string]chan alpaca.TradeUpdate // symbol -> channel
    wg       sync.WaitGroup                      // for graceful shutdown
}

func NewSymbolManager(logger kitlog.Logger) *SymbolManager
    
// LoadSymbols loads portfolio symbols from Firestore
func (m *SymbolManager) LoadSymbols(ctx context.Context, fs *db.Firestore, portfolioID string) error
    
// OnMessage routes incoming TradeUpdate to the correct symbol's goroutine
func (m *SymbolManager) OnMessage(tu alpaca.TradeUpdate)
    
// Shutdown gracefully stops all symbol goroutines
func (m *SymbolManager) Shutdown()
```

### Phase 2: Extend TradeUpdatesConsumer
**File**: `internal/stream/trade_updates.go` (modify existing)

Add SymbolManager field and integrate into Run():

```go
type TradeUpdatesConsumer struct {
    Logger       kitlog.Logger
    Streamer     TradeUpdateStreamer
    ReconnectMin time.Duration
    ReconnectMax time.Duration
    SymbolMgr    *SymbolManager  // NEW: manage symbol routing
}

func (c *TradeUpdatesConsumer) Run(ctx context.Context) error {
    // ... existing reconnection logic ...
    
    err := c.Streamer.StreamTradeUpdates(ctx, func(tu alpaca.TradeUpdate) {
        // Route to SymbolManager instead of logging directly
        if c.SymbolMgr != nil {
            c.SymbolMgr.OnMessage(tu)
        } else {
            // Fallback: log globally for backwards compatibility
            _ = level.Info(c.Logger).Log("component", "stream", "msg", "trade update received")
        }
    }, req)
    
    return err
}
```

### Phase 3: Symbol Goroutine Handler (NEW)
**File**: `internal/stream/symbol_handler.go` (or inside symbol_manager.go)

Each symbol gets its own goroutine that logs price + symbol:

```go
func (m *SymbolManager) startSymbolGoroutine(ctx context.Context, symbol string) {
    m.wg.Add(1)
    go func() {
        defer m.wg.Done()
        
        for tu := range m.symbols[symbol] {
            // Task: log current price and symbol (as per issue requirement)
            if tu.Event == "fill" && tu.Fill != nil {
                _ = level.Info(m.logger).Log(
                    "component", "stream",
                    "symbol", symbol,
                    "event", tu.Event,
                    "price", tu.Fill.Price,
                    "qty", tu.Fill.Qty,
                    "msg", "symbol price update received",
                )
            } else if tu.Event == "outstanding" {
                // Handle quote updates (top of book)
                _ = level.Info(m.logger).Log(
                    "component", "stream",
                    "symbol", symbol,
                    "event", tu.Event,
                    "bid_price", tu.Outstanding.BidPrice,
                    "ask_price", tu.Outstanding.AskPrice,
                    "msg", "symbol quote update received",
                )
            }
        }
    }()
}
```

### Phase 4: Integration with Main App
**File**: `cmd/trader/main.go` (modify existing)

Initialize SymbolManager and load portfolio symbols on startup:

```go
// After creating streamConsumer, add symbol manager initialization:
portfolioID := "default" // or read from config
symbolMgr := &stream.SymbolManager{Logger: c.Logger}

if err := symbolMgr.LoadSymbols(ctx, firestoreDB.Client, portfolioID); err != nil {
    log.Printf("warning: failed to load portfolio symbols: %v", err)
}

streamConsumer := &stream.TradeUpdatesConsumer{
    Logger:       c.Logger,
    Streamer:     alpacaClient,
    ReconnectMin: streamReconnectMin,
    ReconnectMax: streamReconnectMax,
    SymbolMgr:    symbolMgr, // inject the manager
}

// Ensure cleanup on shutdown
defer symbolMgr.Shutdown()
```

### Phase 5: Firestore Portfolio Loading (NEW)
**File**: `internal/stream/symbol_manager.go`

Load symbols from portfolio allocations in Firestore:

```go
func (m *SymbolManager) LoadSymbols(ctx context.Context, fs *db.Firestore, portfolioID string) error {
    portfolio, err := db.Get[*models.Portfolio](ctx, fs, db.CollectionPortfolios, portfolioID)
    if err != nil {
        return fmt.Errorf("failed to fetch portfolio: %w", err)
    }
    
    // Extract all symbols from allocations
    symbols := make(map[string]bool)
    for _, alloc := range portfolio.Allocations {
        for symbol := range alloc.Symbols {
            symbols[symbol] = true
        }
    }
    
    m.symbols = make(map[string]chan alpaca.TradeUpdate, len(symbols))
    
    // Create channel and start goroutine for each symbol
    for symbol := range symbols {
        m.symbols[symbol] = make(chan alpaca.TradeUpdate, 100) // buffered
        m.startSymbolGoroutine(ctx, symbol)
    }
    
    return nil
}
```

## Testing Strategy

### Unit Tests
**File**: `internal/stream/symbol_manager_test.go`
- Test symbol loading from mock portfolio
- Test message routing to correct channels
- Test goroutine lifecycle (start/shutdown)

### Integration Tests
1. Start trader with test portfolio containing 2-3 symbols
2. Verify each symbol gets its own goroutine logged
3. Send mock TradeUpdate messages and verify per-symbol logging
4. Verify graceful shutdown stops all goroutines

## Dependencies
- Existing `internal/stream/trade_updates.go` (modify, not replace)
- Existing `internal/models/portfolio.go` (use Portfolio struct)
- Existing Firestore client from config/bootstrap
- Go standard library: `sync`, `context`

## Implementation Steps

1. ✅ Create `symbol_manager.go` with SymbolManager struct and methods
2. ✅ Add `startSymbolGoroutine()` for per-symbol logging
3. ✅ Add `LoadSymbols()` to fetch portfolio from Firestore
4. ✅ Modify `TradeUpdatesConsumer` to inject SymbolManager
5. ✅ Update `cmd/trader/main.go` to initialize and shutdown SymbolManager
6. ⏳ Write unit tests for symbol routing logic
7. ⏳ Integration test with Alpaca paper trading API

## Notes
- **Branch**: Creating feature branch from `dev`: `feature/issue-3-websocket-routing`
- **Backwards Compatible**: If SymbolManager is nil, falls back to global logging
- **Performance**: Using buffered channels (100 capacity) to prevent message loss during high-frequency updates
- **Scalability**: Can easily extend goroutine logic beyond just logging (e.g., trigger alerts, update UI, etc.)

---
**Plan created after code review on 2026-03-12** | **Review required before implementation**
