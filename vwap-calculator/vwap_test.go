package main

import (
	"math/rand"
	"sync"
	"testing"
)

func TestVWAPCalculator_EdgeCases(t *testing.T) {
	t.Run("EmptyCalculator", func(t *testing.T) {
		calc := NewVWAPCalculator()
		if result := calc.Calculate(); result != 0 {
			t.Errorf("Expected 0, got %.2f", result)
		}
	})

	t.Run("SingleTrade", func(t *testing.T) {
		calc := NewVWAPCalculator()
		calc.Update(100, 2)
		expected := 100.0
		if result := calc.Calculate(); result != expected {
			t.Errorf("Expected %.2f, got %.2f", expected, result)
		}
	})

	t.Run("FullWindow", func(t *testing.T) {
		calc := NewVWAPCalculator()
		totalPV := 0.0
		totalVolume := 0.0

		for i := 1; i <= windowSize; i++ {
			price := float64(i)
			size := 1.0
			calc.Update(price, size)
			totalPV += price * size
			totalVolume += size
		}

		expected := totalPV / totalVolume
		if result := calc.Calculate(); result != expected {
			t.Errorf("Expected %.2f, got %.2f", expected, result)
		}
	})

	t.Run("WindowOverflow", func(t *testing.T) {
		calc := NewVWAPCalculator()
		var expectedPV float64

		// Add windowSize+1 trades
		for i := 1; i <= windowSize+1; i++ {
			price := float64(i)
			size := 1.0
			calc.Update(price, size)
			
			if i > 1 { // First trade will be pushed out at i=201
				expectedPV += price * size
			}
		}

		expected := expectedPV / float64(windowSize)
		if result := calc.Calculate(); result != expected {
			t.Errorf("Expected %.2f, got %.2f", expected, result)
		}
	})

	t.Run("InvalidInputs", func(t *testing.T) {
		calc := NewVWAPCalculator()
		cases := []struct {
			price, size float64
		}{
			{-100, 1},
			{100, -1},
			{-50, -2},
		}

		for _, tc := range cases {
			err := calc.Update(tc.price, tc.size)
			if err == nil {
				t.Errorf("Expected error for price=%.2f size=%.2f", tc.price, tc.size)
			}
		}
	})
}

func TestConcurrentUpdates(t *testing.T) {
	calc := NewVWAPCalculator()
	var wg sync.WaitGroup
	workers := 100
	updatesPerWorker := 100

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < updatesPerWorker; j++ {
				price := rand.Float64() * 1000
				size := rand.Float64() * 10
				calc.Update(price, size)
			}
		}()
	}

	wg.Wait()
	
	// Final calculation shouldn't panic
	result := calc.Calculate()
	if result < 0 {
		t.Errorf("Invalid VWAP result: %.2f", result)
	}
}