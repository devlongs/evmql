package parser

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestParseQuery_ValidQueries(t *testing.T) {
	tests := []struct {
		name           string
		queryStr       string
		expectedType   string
		expectedMethod string
		expectedAddr   string
		expectBlocks   bool
		fromBlock      string
		toBlock        string
	}{
		{
			name:           "Balance query without block",
			queryStr:       "SELECT BALANCE FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			expectedType:   "SELECT",
			expectedMethod: "BALANCE",
			expectedAddr:   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			expectBlocks:   false,
		},
		{
			name:           "Balance query with block",
			queryStr:       "SELECT BALANCE FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000 1000000",
			expectedType:   "SELECT",
			expectedMethod: "BALANCE",
			expectedAddr:   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			expectBlocks:   true,
			fromBlock:      "1000000",
			toBlock:        "1000000",
		},
		{
			name:           "Logs query with block range",
			queryStr:       "SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000 1001000",
			expectedType:   "SELECT",
			expectedMethod: "LOGS",
			expectedAddr:   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			expectBlocks:   true,
			fromBlock:      "1000000",
			toBlock:        "1001000",
		},
		{
			name:           "Transactions query with block range",
			queryStr:       "SELECT TRANSACTIONS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000 1000100",
			expectedType:   "SELECT",
			expectedMethod: "TRANSACTIONS",
			expectedAddr:   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			expectBlocks:   true,
			fromBlock:      "1000000",
			toBlock:        "1000100",
		},
		{
			name:           "Lowercase query",
			queryStr:       "select balance from 0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			expectedType:   "SELECT",
			expectedMethod: "BALANCE",
			expectedAddr:   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			expectBlocks:   false,
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := parser.ParseQuery(tt.queryStr)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if query.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, query.Type)
			}

			if query.Method != tt.expectedMethod {
				t.Errorf("Expected method %s, got %s", tt.expectedMethod, query.Method)
			}

			expectedAddr := common.HexToAddress(tt.expectedAddr)
			if query.Address != expectedAddr {
				t.Errorf("Expected address %s, got %s", expectedAddr.Hex(), query.Address.Hex())
			}

			if tt.expectBlocks {
				if query.FromBlock == nil || query.ToBlock == nil {
					t.Error("Expected block range to be set")
				} else {
					expectedFrom := new(big.Int)
					expectedFrom.SetString(tt.fromBlock, 10)
					expectedTo := new(big.Int)
					expectedTo.SetString(tt.toBlock, 10)

					if query.FromBlock.Cmp(expectedFrom) != 0 {
						t.Errorf("Expected from block %s, got %s", expectedFrom.String(), query.FromBlock.String())
					}
					if query.ToBlock.Cmp(expectedTo) != 0 {
						t.Errorf("Expected to block %s, got %s", expectedTo.String(), query.ToBlock.String())
					}
				}
			} else {
				if query.FromBlock != nil || query.ToBlock != nil {
					t.Error("Expected no block range to be set")
				}
			}
		})
	}
}

func TestParseQuery_InvalidQueries(t *testing.T) {
	tests := []struct {
		name        string
		queryStr    string
		expectedErr string
	}{
		{
			name:        "Empty query",
			queryStr:    "",
			expectedErr: "query cannot be empty",
		},
		{
			name:        "Missing FROM keyword",
			queryStr:    "SELECT BALANCE 0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			expectedErr: "invalid query format",
		},
		{
			name:        "Missing address",
			queryStr:    "SELECT BALANCE FROM",
			expectedErr: "invalid query format",
		},
		{
			name:        "Invalid address format",
			queryStr:    "SELECT BALANCE FROM 0xinvalidaddress",
			expectedErr: "invalid Ethereum address",
		},
		{
			name:        "Unsupported method",
			queryStr:    "SELECT CONTRACT FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			expectedErr: "unsupported method",
		},
		{
			name:        "Invalid from block",
			queryStr:    "SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK abc 1000000",
			expectedErr: "invalid from block",
		},
		{
			name:        "Invalid to block",
			queryStr:    "SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000 xyz",
			expectedErr: "invalid to block",
		},
		{
			name:        "Negative from block",
			queryStr:    "SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK -100 1000000",
			expectedErr: "invalid from block",
		},
		{
			name:        "From block greater than to block",
			queryStr:    "SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 2000000 1000000",
			expectedErr: "from block",
		},
		{
			name:        "Block range too large",
			queryStr:    "SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000 2000001",
			expectedErr: "block range too large",
		},
		{
			name:        "Query too long",
			queryStr:    "SELECT BALANCE FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e " + strings.Repeat("X", 10000),
			expectedErr: "query too long",
		},
		{
			name:        "Missing to block",
			queryStr:    "SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000",
			expectedErr: "BLOCK keyword requires both from and to block numbers",
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := parser.ParseQuery(tt.queryStr)
			if err == nil {
				t.Fatalf("Expected error containing '%s', got no error. Query: %+v", tt.expectedErr, query)
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestParseQuery_EdgeCases(t *testing.T) {
	parser := NewParser()

	t.Run("Extra whitespace", func(t *testing.T) {
		queryStr := "  SELECT   BALANCE   FROM   0x742d35Cc6634C0532925a3b844Bc454e4438f44e  "
		query, err := parser.ParseQuery(queryStr)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if query.Method != "BALANCE" {
			t.Errorf("Expected BALANCE method, got %s", query.Method)
		}
	})

	t.Run("Block range of 0", func(t *testing.T) {
		queryStr := "SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000 1000000"
		query, err := parser.ParseQuery(queryStr)
		if err != nil {
			t.Fatalf("Expected no error for zero block range, got: %v", err)
		}
		if query.FromBlock.Cmp(query.ToBlock) != 0 {
			t.Error("Expected from and to blocks to be equal")
		}
	})

	t.Run("Maximum valid block range", func(t *testing.T) {
		queryStr := "SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000 1010000"
		query, err := parser.ParseQuery(queryStr)
		if err != nil {
			t.Fatalf("Expected no error for max block range, got: %v", err)
		}
		blockRange := new(big.Int).Sub(query.ToBlock, query.FromBlock)
		if blockRange.Cmp(big.NewInt(10000)) != 0 {
			t.Errorf("Expected block range of 10000, got %s", blockRange.String())
		}
	})
}
