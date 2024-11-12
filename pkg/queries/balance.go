package queries

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type BalanceQuery struct {
	Query
}

func NewBalanceQuery(address common.Address, fromBlock, toBlock *big.Int) *BalanceQuery {
	return &BalanceQuery{
		Query: Query{
			Type:      "SELECT",
			Address:   address,
			Method:    "BALANCE",
			FromBlock: fromBlock,
			ToBlock:   toBlock,
		},
	}
}
