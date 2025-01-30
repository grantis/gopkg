## Real-time VWAP Calculator

A Go implementation of a real-time Volume-Weighted Average Price (VWAP) calculator using Coinbase's WebSocket feed.

### Features

- Real-time VWAP calculation for BTC-USD, ETH-USD, and ETH-BTC
- Sliding window of 200 trades
- Concurrent WebSocket handling
- Production-grade error handling and retries
- Comprehensive unit tests

### Design

- **Ring Buffer**: Efficient O(1) sliding window implementation
- **Interface-based Design**: `Calculator` interface allows different implementations
- **Concurrency**: Goroutines for WebSocket handling and message processing
- **Error Handling**: Automatic reconnection with retry limits
- **Logging**: Structured logging with different levels

### Requirements

- Go 1.21+
- Gorilla WebSocket (`github.com/gorilla/websocket`)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/vwap-calculator.git
cd vwap-calculator
```
2. Install dependencies:
```bash
go mod download
```

### Usage
Run the application:

```bash
go run main.go
```
Run tests:

```bash
go test -v
```
Run linter:
```bash
golangci-lint run
Example Output
[VWAP] INFO: 2023/09/15 10:00:00 Connected to wss://ws-feed.exchange.coinbase.com
[VWAP] INFO: 2023/09/15 10:00:01 Subscribed to matches channel
BTC-USD VWAP: 45000.1234
ETH-USD VWAP: 3000.5678
ETH-BTC VWAP: 0.06789
```
### Configuration
Adjust windowSize in main.go to change the number of trades considered

Modify retryDelay and maxRetries for connection handling

### Testing
The test suite covers:
- Basic VWAP calculations
- Window sliding behavior
- Concurrent updates
- Error conditions

