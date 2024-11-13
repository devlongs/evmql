package executor

import (
	"context"
	"fmt"
	"math/big"

	"github.com/devlongs/evmql/pkg/queries"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// QueryExecutor is responsible for executing queries
type QueryExecutor struct {
	client *ethclient.Client
}

// NewQueryExecutor creates a new QueryExecutor instance
func NewQueryExecutor(client *ethclient.Client) *QueryExecutor {
	return &QueryExecutor{client: client}
}

// Execute runs the query and returns the result
func (qe *QueryExecutor) Execute(ctx context.Context, query *queries.Query) (interface{}, error) {
	switch query.Method {
	case "BALANCE":
		return qe.getBalance(ctx, query)
	case "LOGS":
		return qe.getLogs(ctx, query)
	case "TRANSACTIONS":
		return qe.getTransactions(ctx, query)
	default:
		return nil, fmt.Errorf("unsupported select method: %s", query.Method)
	}
}

func (qe *QueryExecutor) getBalance(ctx context.Context, query *queries.Query) (*big.Int, error) {
	balance, err := qe.client.BalanceAt(ctx, query.Address, query.FromBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

func (qe *QueryExecutor) getLogs(ctx context.Context, query *queries.Query) ([]types.Log, error) {
	filterQuery := ethereum.FilterQuery{
		FromBlock: query.FromBlock,
		ToBlock:   query.ToBlock,
		Addresses: []common.Address{query.Address},
	}
	logs, err := qe.client.FilterLogs(ctx, filterQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	return logs, nil
}

func (qe *QueryExecutor) getTransactions(ctx context.Context, query *queries.Query) ([]*types.Transaction, error) {
	var transactions []*types.Transaction

	if query.FromBlock == nil {
		latestBlock, err := qe.client.BlockByNumber(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest block: %w", err)
		}
		query.FromBlock = latestBlock.Number()
		query.ToBlock = query.FromBlock
	}

	for blockNum := query.FromBlock; blockNum.Cmp(query.ToBlock) <= 0; blockNum = new(big.Int).Add(blockNum, big.NewInt(1)) {
		block, err := qe.client.BlockByNumber(ctx, blockNum)
		if err != nil {
			return nil, fmt.Errorf("failed to get block %s: %w", blockNum.String(), err)
		}

		for _, tx := range block.Transactions() {
			msg, err := core.TransactionToMessage(tx, types.NewLondonSigner(tx.ChainId()), nil)
			if err != nil {
				continue
			}

			// Check if transaction is from or to the specified address
			if (tx.To() != nil && *tx.To() == query.Address) || msg.From == query.Address {
				transactions = append(transactions, tx)
			}
		}
	}

	return transactions, nil
}
