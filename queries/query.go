package queries

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Query represents a parsed query ready for execution
type Query struct {
	Type      string
	Address   common.Address
	Method    string
	FromBlock *big.Int
	ToBlock   *big.Int
}
