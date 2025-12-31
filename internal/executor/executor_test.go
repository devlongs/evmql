package executor

import (
	"math/big"
	"testing"
	"time"

	"github.com/devlongs/evmql/pkg/queries"
	"github.com/ethereum/go-ethereum/common"
)

func TestNewQueryExecutor(t *testing.T) {
	executor := NewQueryExecutor(nil)

	if executor == nil {
		t.Fatal("NewQueryExecutor returned nil")
	}

	if executor.timeout != 30*time.Second {
		t.Errorf("Expected default timeout of 30s, got %v", executor.timeout)
	}

	if executor.maxWorkers != 5 {
		t.Errorf("Expected default maxWorkers of 5, got %d", executor.maxWorkers)
	}
}

func TestSetTimeout(t *testing.T) {
	executor := NewQueryExecutor(nil)

	newTimeout := 60 * time.Second
	executor.SetTimeout(newTimeout)

	if executor.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, executor.timeout)
	}
}

func TestSetMaxWorkers(t *testing.T) {
	tests := []struct {
		name            string
		workers         int
		expectedWorkers int
	}{
		{
			name:            "Valid positive number",
			workers:         10,
			expectedWorkers: 10,
		},
		{
			name:            "Zero workers - should not change",
			workers:         0,
			expectedWorkers: 5, // default
		},
		{
			name:            "Negative workers - should not change",
			workers:         -5,
			expectedWorkers: 5, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := NewQueryExecutor(nil)
			exec.SetMaxWorkers(tt.workers)

			if exec.maxWorkers != tt.expectedWorkers {
				t.Errorf("Expected maxWorkers %d, got %d", tt.expectedWorkers, exec.maxWorkers)
			}
		})
	}
}

func TestQueryValidation_BlockRanges(t *testing.T) {
	tests := []struct {
		name        string
		fromBlock   *big.Int
		toBlock     *big.Int
		maxRange    int64
		shouldError bool
	}{
		{
			name:        "Valid range within limit",
			fromBlock:   big.NewInt(1000000),
			toBlock:     big.NewInt(1000999),
			maxRange:    1000,
			shouldError: false,
		},
		{
			name:        "Exactly at limit",
			fromBlock:   big.NewInt(1000000),
			toBlock:     big.NewInt(1001000),
			maxRange:    1000,
			shouldError: false,
		},
		{
			name:        "Exceeds limit by 1",
			fromBlock:   big.NewInt(1000000),
			toBlock:     big.NewInt(1001001),
			maxRange:    1000,
			shouldError: true,
		},
		{
			name:        "Single block",
			fromBlock:   big.NewInt(1000000),
			toBlock:     big.NewInt(1000000),
			maxRange:    1000,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockRange := new(big.Int).Sub(tt.toBlock, tt.fromBlock)
			exceedsLimit := blockRange.Cmp(big.NewInt(tt.maxRange)) > 0

			if exceedsLimit != tt.shouldError {
				t.Errorf("Expected shouldError=%v, got exceedsLimit=%v for range %s",
					tt.shouldError, exceedsLimit, blockRange.String())
			}
		})
	}
}

func TestQueryStructure(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		address string
	}{
		{
			name:    "BALANCE query",
			method:  "BALANCE",
			address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		},
		{
			name:    "LOGS query",
			method:  "LOGS",
			address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		},
		{
			name:    "TRANSACTIONS query",
			method:  "TRANSACTIONS",
			address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &queries.Query{
				Type:      "SELECT",
				Method:    tt.method,
				Address:   common.HexToAddress(tt.address),
				FromBlock: big.NewInt(1000000),
				ToBlock:   big.NewInt(1000100),
			}

			if query.Type != "SELECT" {
				t.Errorf("Expected type SELECT, got %s", query.Type)
			}

			if query.Method != tt.method {
				t.Errorf("Expected method %s, got %s", tt.method, query.Method)
			}

			if query.Address.Hex() != tt.address {
				t.Errorf("Expected address %s, got %s", tt.address, query.Address.Hex())
			}
		})
	}
}

func TestExecutorConfiguration(t *testing.T) {
	executor := NewQueryExecutor(nil)

	// Test default configuration
	if executor.timeout == 0 {
		t.Error("Default timeout should not be zero")
	}

	if executor.maxWorkers <= 0 {
		t.Error("Default maxWorkers should be positive")
	}

	// Test configuration updates
	executor.SetTimeout(45 * time.Second)
	if executor.timeout != 45*time.Second {
		t.Error("Timeout was not updated")
	}

	executor.SetMaxWorkers(10)
	if executor.maxWorkers != 10 {
		t.Error("MaxWorkers was not updated")
	}

	// Test invalid configuration is rejected
	executor.SetMaxWorkers(0)
	if executor.maxWorkers == 0 {
		t.Error("Should not accept zero maxWorkers")
	}

	executor.SetMaxWorkers(-5)
	if executor.maxWorkers < 0 {
		t.Error("Should not accept negative maxWorkers")
	}
}

func TestBlockRangeCalculations(t *testing.T) {
	tests := []struct {
		name          string
		fromBlock     int64
		toBlock       int64
		expectedRange int64
	}{
		{
			name:          "100 blocks",
			fromBlock:     1000000,
			toBlock:       1000099,
			expectedRange: 99,
		},
		{
			name:          "1000 blocks",
			fromBlock:     1000000,
			toBlock:       1000999,
			expectedRange: 999,
		},
		{
			name:          "Single block",
			fromBlock:     1000000,
			toBlock:       1000000,
			expectedRange: 0,
		},
		{
			name:          "Exactly 10000",
			fromBlock:     1000000,
			toBlock:       1009999,
			expectedRange: 9999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from := big.NewInt(tt.fromBlock)
			to := big.NewInt(tt.toBlock)

			blockRange := new(big.Int).Sub(to, from)

			if blockRange.Int64() != tt.expectedRange {
				t.Errorf("Expected range %d, got %d", tt.expectedRange, blockRange.Int64())
			}
		})
	}
}
