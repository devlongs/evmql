package queries

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type TransactionsQuery struct {
	Query
}

func NewTransactionsQuery(address common.Address, fromBlock, toBlock *big.Int) *TransactionsQuery {
	return &TransactionsQuery{
		Query: Query{
			Type:      "SELECT",
			Address:   address,
			Method:    "TRANSACTIONS",
			FromBlock: fromBlock,
			ToBlock:   toBlock,
		},
	}
}
