package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Calculator interface defines the VWAP operations
type Calculator interface {
	Update(price, size float64) error
	Calculate() float64
}

const (
	windowSize    = 200
	websocketURL  = "wss://ws-feed.exchange.coinbase.com"
	retryDelay    = 3 * time.Second
	maxRetries    = 5
)

type Trade struct {
	Type      string  `json:"type"`
	ProductID string  `json:"product_id"`
	Price     float64 `json:"price,string"`
	Size      float64 `json:"size,string"`
}

type RingBuffer struct {
	data  [windowSize * 2]float64
	start int
	count int
}

func (rb *RingBuffer) Add(price, size float64) (oldPrice, oldSize float64, removed bool) {
	if rb.count == windowSize {
		oldPrice = rb.data[rb.start]
		oldSize = rb.data[rb.start+1]
		rb.start = (rb.start + 2) % len(rb.data)
		removed = true
	} else {
		rb.count++
	}
	pos := (rb.start + (rb.count-1)*2) % len(rb.data)
	rb.data[pos] = price
	rb.data[pos+1] = size
	return
}

type VWAPCalculator struct {
	mu          sync.Mutex
	buffer      RingBuffer
	totalPV     float64
	totalVolume float64
}

func NewVWAPCalculator() *VWAPCalculator {
	return &VWAPCalculator{}
}

func (v *VWAPCalculator) Update(price, size float64) error {
	if price <= 0 || size <= 0 {
		return errors.New("invalid trade data: price and size must be positive")
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	oldPrice, oldSize, removed := v.buffer.Add(price, size)
	if removed {
		v.totalPV -= oldPrice * oldSize
		v.totalVolume -= oldSize
	}
	v.totalPV += price * size
	v.totalVolume += size
	return nil
}

func (v *VWAPCalculator) Calculate() float64 {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.totalVolume == 0 {
		return 0
	}
	return v.totalPV / v.totalVolume
}

// Logger interface for dependency injection
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type DefaultLogger struct {
	*log.Logger
}

func NewLogger() *DefaultLogger {
	return &DefaultLogger{
		Logger: log.New(os.Stdout, "[VWAP] ", log.LstdFlags|log.Lmsgprefix),
	}
}

func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	l.Printf("INFO: "+format, args...)
}

func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	l.Printf("ERROR: "+format, args...)
}

func main() {
	logger := NewLogger()
	calculators := map[string]Calculator{
		"BTC-USD": NewVWAPCalculator(),
		"ETH-USD": NewVWAPCalculator(),
		"ETH-BTC": NewVWAPCalculator(),
	}

	retryCount := 0
	for {
		conn, err := connectWebSocket(logger)
		if err != nil {
			if retryCount++; retryCount > maxRetries {
				logger.Errorf("Max connection retries (%d) reached", maxRetries)
				return
			}
			time.Sleep(retryDelay)
			continue
		}
		retryCount = 0

		if err := handleConnection(conn, calculators, logger); err != nil {
			logger.Errorf("Connection handling failed: %v", err)
		}
		conn.Close()
		time.Sleep(retryDelay)
	}
}

func connectWebSocket(logger Logger) (*websocket.Conn, error) {
	logger.Infof("Connecting to %s", websocketURL)
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(websocketURL, nil)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	return conn, nil
}

func handleConnection(conn *websocket.Conn, calculators map[string]Calculator, logger Logger) error {
	if err := subscribe(conn, logger); err != nil {
		return err
	}

	messageChan := make(chan []byte)
	errChan := make(chan error)

	go readMessages(conn, messageChan, errChan, logger)

	for {
		select {
		case message := <-messageChan:
			processMessage(message, calculators, logger)
		case err := <-errChan:
			return err
		}
	}
}

func readMessages(conn *websocket.Conn, messageChan chan<- []byte, errChan chan<- error, logger Logger) {
	defer close(messageChan)
	defer close(errChan)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			errChan <- fmt.Errorf("read error: %w", err)
			return
		}
		messageChan <- message
	}
}

func processMessage(message []byte, calculators map[string]Calculator, logger Logger) {
	var trade Trade
	if err := json.Unmarshal(message, &trade); err != nil {
		logger.Errorf("JSON decode error: %v", err)
		return
	}

	if trade.Type != "match" {
		return
	}

	logger.Infof("Received trade: %s %.4f @ %.2f", 
		trade.ProductID, trade.Size, trade.Price)

	calculator, exists := calculators[trade.ProductID]
	if !exists {
		logger.Errorf("Received trade for unknown product: %s", trade.ProductID)
		return
	}

	if err := calculator.Update(trade.Price, trade.Size); err != nil {
		logger.Errorf("Update failed: %v", err)
		return
	}

	vwap := calculator.Calculate()
	fmt.Printf("%s VWAP: %.4f\n", trade.ProductID, vwap)
}

func subscribe(conn *websocket.Conn, logger Logger) error {
	subMsg := map[string]interface{}{
		"type":        "subscribe",
		"product_ids": []string{"BTC-USD", "ETH-USD", "ETH-BTC"},
		"channels":    []string{"matches"},
	}
	if err := conn.WriteJSON(subMsg); err != nil {
		return fmt.Errorf("subscribe failed: %w", err)
	}
	logger.Infof("Subscribed to matches channel")
	return nil
}