package executor

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/devlongs/evmql/pkg/queries"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// QueryExecutor is responsible for executing queries
type QueryExecutor struct {
	client  *ethclient.Client
	timeout time.Duration
}

// NewQueryExecutor creates a new QueryExecutor instance
func NewQueryExecutor(client *ethclient.Client) *QueryExecutor {
	return &QueryExecutor{
		client:  client,
		timeout: 30 * time.Second, // Default timeout
	}
}

// SetTimeout sets the query execution timeout
func (qe *QueryExecutor) SetTimeout(timeout time.Duration) {
	qe.timeout = timeout
}

// Execute runs the query and returns the result
func (qe *QueryExecutor) Execute(ctx context.Context, query *queries.Query) (interface{}, error) {
	// Create a context with timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, qe.timeout)
		defer cancel()
	}

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
	var blockNumber *big.Int
	if query.FromBlock != nil {
		blockNumber = query.FromBlock
	}

	balance, err := qe.client.BalanceAt(ctx, query.Address, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

func (qe *QueryExecutor) getLogs(ctx context.Context, query *queries.Query) ([]types.Log, error) {
	// Ensure block range is specified
	if query.FromBlock == nil || query.ToBlock == nil {
		return nil, fmt.Errorf("both from and to block numbers must be specified for logs query")
	}

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
		// If no block range is specified, use the latest block
		latestBlock, err := qe.client.BlockByNumber(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest block: %w", err)
		}
		query.FromBlock = latestBlock.Number()
		query.ToBlock = query.FromBlock
	} else if query.ToBlock == nil {
		// If only from block is specified, use it as the only block
		query.ToBlock = query.FromBlock
	}

	// Ensure the block range is not too large to prevent heavy queries
	blockRange := new(big.Int).Sub(query.ToBlock, query.FromBlock)
	if blockRange.Cmp(big.NewInt(1000)) > 0 {
		return nil, fmt.Errorf("block range too large: max allowed is 1000 blocks")
	}

	// Process each block in the range
	for blockNum := new(big.Int).Set(query.FromBlock); blockNum.Cmp(query.ToBlock) <= 0; blockNum = new(big.Int).Add(blockNum, big.NewInt(1)) {
		// Check for context cancellation
		if ctx.Err() != nil {
			return transactions, ctx.Err()
		}

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
