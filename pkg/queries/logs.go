package queries

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type LogsQuery struct {
	Query
}

func NewLogsQuery(address common.Address, fromBlock, toBlock *big.Int) *LogsQuery {
	return &LogsQuery{
		Query: Query{
			Type:      "SELECT",
			Address:   address,
			Method:    "LOGS",
			FromBlock: fromBlock,
			ToBlock:   toBlock,
		},
	}
}
