package queries

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestNewBalanceQuery(t *testing.T) {
	addr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
	fromBlock := big.NewInt(1000000)
	toBlock := big.NewInt(1000100)

	query := NewBalanceQuery(addr, fromBlock, toBlock)

	if query.Type != "SELECT" {
		t.Errorf("Expected type SELECT, got %s", query.Type)
	}

	if query.Method != "BALANCE" {
		t.Errorf("Expected method BALANCE, got %s", query.Method)
	}

	if query.Address != addr {
		t.Errorf("Expected address %s, got %s", addr.Hex(), query.Address.Hex())
	}

	if query.FromBlock.Cmp(fromBlock) != 0 {
		t.Errorf("Expected from block %s, got %s", fromBlock.String(), query.FromBlock.String())
	}

	if query.ToBlock.Cmp(toBlock) != 0 {
		t.Errorf("Expected to block %s, got %s", toBlock.String(), query.ToBlock.String())
	}
}

func TestNewLogsQuery(t *testing.T) {
	addr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
	fromBlock := big.NewInt(1000000)
	toBlock := big.NewInt(1001000)

	query := NewLogsQuery(addr, fromBlock, toBlock)

	if query.Type != "SELECT" {
		t.Errorf("Expected type SELECT, got %s", query.Type)
	}

	if query.Method != "LOGS" {
		t.Errorf("Expected method LOGS, got %s", query.Method)
	}

	if query.Address != addr {
		t.Errorf("Expected address %s, got %s", addr.Hex(), query.Address.Hex())
	}
}

func TestNewTransactionsQuery(t *testing.T) {
	addr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
	fromBlock := big.NewInt(1000000)
	toBlock := big.NewInt(1000100)

	query := NewTransactionsQuery(addr, fromBlock, toBlock)

	if query.Type != "SELECT" {
		t.Errorf("Expected type SELECT, got %s", query.Type)
	}

	if query.Method != "TRANSACTIONS" {
		t.Errorf("Expected method TRANSACTIONS, got %s", query.Method)
	}

	if query.Address != addr {
		t.Errorf("Expected address %s, got %s", addr.Hex(), query.Address.Hex())
	}
}

func TestQueryWithNilBlocks(t *testing.T) {
	addr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")

	query := NewBalanceQuery(addr, nil, nil)

	if query.FromBlock != nil {
		t.Error("Expected nil from block")
	}

	if query.ToBlock != nil {
		t.Error("Expected nil to block")
	}
}
