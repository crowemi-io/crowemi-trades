# Plan: Implement WebSocket Streaming for Alpaca Trading

## Issue Reference
- **Issue**: #3 in crowemi-io/crowemi-trades
- **Title**: feat: implement websocket streaming

## Objective
For each symbol in the app node in the portfolio, add an Alpaca WebSocket handler. When a message is received, route it to its own goroutine to perform the task (currently just log the current price and symbol).

## Implementation Plan

### 1. Create WebSocket Handler Module
**File**: `internal/alpaca/websocket.go`

- Initialize Alpaca WebSocket connection using existing `alpaca.go` as reference
- Set up message handlers for different event types (trade updates, quotes, etc.)
- Implement automatic reconnection logic with exponential backoff

### 2. Portfolio Symbol Management
**File**: `internal/portfolio/symbols.go`

- Add method to retrieve all symbols from portfolio configuration
- Maintain a map of active WebSocket connections per symbol
- Handle dynamic symbol additions/removals as portfolio updates

### 3. Goroutine Routing System
**File**: `internal/alpaca/router.go`

- Create message dispatcher that routes each symbol's messages to separate goroutines
- Implement concurrent processing with proper synchronization
- Add graceful shutdown for all goroutines

### 4. Logging Implementation
**Current Task**: Log current price and symbol per message

```go
func handleSymbolMessage(symbol string, msg alpaca.Message) {
    go func() {
        // Process symbol-specific logic here
        log.Printf("Symbol: %s, Price: $%f", symbol, msg.Price)
    }()
}
```

### 5. Integration with Existing Code
**Files to modify**:
- `cmd/*.go` - Initialize WebSocket handler on startup
- Update existing `alpaca.go` to use new routing system if needed

## Testing Strategy
1. Unit tests for message routing logic
2. Integration test with Alpaca paper trading API
3. Load test with multiple concurrent symbol handlers

## Dependencies
- Existing `alpaca.go` module (for reference)
- Go standard library `sync` package for goroutine management
- Logging infrastructure already in place

## Next Steps After Plan Approval
1. Implement WebSocket handler module
2. Add portfolio symbol retrieval
3. Create message router with goroutine per symbol
4. Implement logging functionality
5. Test end-to-end with Alpaca API
6. Update issue with progress, create PR when complete

---
**Plan created by AI assistant** | **Review required before implementation**
