package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func TestVWAPCalculator_EdgeCases(t *testing.T) {
	t.Run("EmptyCalculator", func(t *testing.T) {
		calc := NewVWAPCalculator()
		if result := calc.Calculate(); result != "0" {
			t.Errorf("Expected 0, got %s", result)
		}
	})

	t.Run("SingleTrade", func(t *testing.T) {
		calc := NewVWAPCalculator()
		if err := calc.Update("100", "2"); err != nil {
			t.Errorf("Update returned error: %v", err)
		}
		expected := "100.0000"
		if result := calc.Calculate(); result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("FullWindow", func(t *testing.T) {
		calc := NewVWAPCalculator()
		totalPV := 0.0
		totalVolume := 0.0

		for i := 1; i <= windowSize; i++ {
			price := float64(i)
			size := 1.0
			if err := calc.Update(fmt.Sprintf("%g", price), fmt.Sprintf("%g", size)); err != nil {
				t.Errorf("Update returned error: %v", err)
			}
			totalPV += price * size
			totalVolume += size
		}

		expectedFloat := totalPV / totalVolume
		expected := fmt.Sprintf("%.4f", expectedFloat)
		if result := calc.Calculate(); result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("WindowOverflow", func(t *testing.T) {
		calc := NewVWAPCalculator()
		var expectedPV float64

		// Add windowSize+1 trades
		for i := 1; i <= windowSize+1; i++ {
			price := float64(i)
			size := 1.0
			if err := calc.Update(fmt.Sprintf("%g", price), fmt.Sprintf("%g", size)); err != nil {
				t.Errorf("Update returned error: %v", err)
			}
			if i > 1 {
				expectedPV += price * size
			}
		}

		expectedFloat := expectedPV / float64(windowSize)
		expected := fmt.Sprintf("%.4f", expectedFloat)
		if result := calc.Calculate(); result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
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
			err := calc.Update(fmt.Sprintf("%g", tc.price), fmt.Sprintf("%g", tc.size))
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
				price := rand.Float64()*1000 + 0.0001
				size := rand.Float64()*10 + 0.0001
				if err := calc.Update(fmt.Sprintf("%g", price), fmt.Sprintf("%g", size)); err != nil {
					t.Errorf("Update returned error: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	resultStr := calc.Calculate()
	result, err := strconv.ParseFloat(resultStr, 64)
	if err != nil {
		t.Errorf("Failed to parse VWAP result: %v", err)
	}
	if result < 0 {
		t.Errorf("Invalid VWAP result: %f", result)
	}
}
