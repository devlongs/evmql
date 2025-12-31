package executor

import (
"math/big"
"testing"
"time"
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
{"Valid positive number", 10, 10},
{"Zero workers", 0, 5},
{"Negative workers", -5, 5},
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

func TestBlockRangeValidation(t *testing.T) {
tests := []struct {
name        string
fromBlock   int64
toBlock     int64
maxRange    int64
shouldError bool
}{
{"Within limit", 1000000, 1000999, 1000, false},
{"At limit", 1000000, 1001000, 1000, false},
{"Exceeds limit", 1000000, 1001001, 1000, true},
{"Single block", 1000000, 1000000, 1000, false},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
from := big.NewInt(tt.fromBlock)
to := big.NewInt(tt.toBlock)
blockRange := new(big.Int).Sub(to, from)
exceedsLimit := blockRange.Cmp(big.NewInt(tt.maxRange)) > 0
if exceedsLimit != tt.shouldError {
t.Errorf("Expected shouldError=%v, got=%v", tt.shouldError, exceedsLimit)
}
})
}
}
