package executor

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/devlongs/evmql/internal/cache"
	"github.com/devlongs/evmql/internal/logger"
	"github.com/devlongs/evmql/queries"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// QueryExecutor is responsible for executing queries
type QueryExecutor struct {
	client     *ethclient.Client
	timeout    time.Duration
	maxWorkers int
	cache      cache.Cache
}

// NewQueryExecutor creates a new QueryExecutor instance
func NewQueryExecutor(client *ethclient.Client) *QueryExecutor {
	return &QueryExecutor{
		client:     client,
		timeout:    30 * time.Second,
		maxWorkers: 5,
		cache:      cache.NewNoOpCache(), // Default to no caching
	}
}

// SetCache sets the cache implementation
func (qe *QueryExecutor) SetCache(c cache.Cache) {
	qe.cache = c
}

// SetTimeout sets the query execution timeout
func (qe *QueryExecutor) SetTimeout(timeout time.Duration) {
	qe.timeout = timeout
}

// SetMaxWorkers sets the maximum number of concurrent workers
func (qe *QueryExecutor) SetMaxWorkers(maxWorkers int) {
	if maxWorkers > 0 {
		qe.maxWorkers = maxWorkers
	}
}

// Execute runs the query and returns the result
func (qe *QueryExecutor) Execute(ctx context.Context, query *queries.Query) (interface{}, error) {
	// Create a context with timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, qe.timeout)
		defer cancel()
	}

	logger.Info("executing query",
		"method", query.Method,
		"address", query.Address.Hex(),
		"from_block", query.FromBlock,
		"to_block", query.ToBlock)

	startTime := time.Now()
	var result interface{}
	var err error

	switch query.Method {
	case "BALANCE":
		result, err = qe.getBalance(ctx, query)
	case "LOGS":
		result, err = qe.getLogs(ctx, query)
	case "TRANSACTIONS":
		result, err = qe.getTransactionsConcurrent(ctx, query)
	default:
		err = fmt.Errorf("unsupported select method: %s", query.Method)
	}

	duration := time.Since(startTime)
	if err != nil {
		logger.Error("query execution failed",
			"method", query.Method,
			"duration", duration,
			"error", err)
	} else {
		logger.Info("query execution completed",
			"method", query.Method,
			"duration", duration)
	}

	return result, err
}

func (qe *QueryExecutor) getBalance(ctx context.Context, query *queries.Query) (*big.Int, error) {
	blockNumber := (*big.Int)(nil)
	if query.FromBlock != nil {
		blockNumber = query.FromBlock
	}

	// Generate cache key
	cacheKey := cache.GenerateKey("balance", query.Address.Hex(), blockNumber)

	// Check cache first
	if cached, found := qe.cache.Get(cacheKey); found {
		logger.Debug("cache hit", "key", cacheKey)
		if balance, ok := cached.(*big.Int); ok {
			return balance, nil
		}
	}

	balance, err := qe.client.BalanceAt(ctx, query.Address, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("error fetching balance: %w", err)
	}

	// Cache the result
	qe.cache.Set(cacheKey, balance, 0)
	logger.Debug("cached balance", "key", cacheKey)

	return balance, nil
}

func (qe *QueryExecutor) getLogs(ctx context.Context, query *queries.Query) ([]types.Log, error) {
	if query.FromBlock == nil || query.ToBlock == nil {
		return nil, fmt.Errorf("both from and to block numbers must be specified for logs query")
	}

	// Validate block range
	blockRange := new(big.Int).Sub(query.ToBlock, query.FromBlock)
	if blockRange.Cmp(big.NewInt(10000)) > 0 {
		return nil, fmt.Errorf("block range too large for logs query: %d blocks (maximum: 10000)", blockRange.Int64())
	}

	// Generate cache key
	cacheKey := cache.GenerateKey("logs", query.Address.Hex(), query.FromBlock, query.ToBlock)

	// Check cache first
	if cached, found := qe.cache.Get(cacheKey); found {
		logger.Debug("cache hit", "key", cacheKey)
		if logs, ok := cached.([]types.Log); ok {
			return logs, nil
		}
	}

	filterQuery := ethereum.FilterQuery{
		FromBlock: query.FromBlock,
		ToBlock:   query.ToBlock,
		Addresses: []common.Address{query.Address},
	}

	logs, err := qe.client.FilterLogs(ctx, filterQuery)
	if err != nil {
		return nil, fmt.Errorf("error fetching logs: %w", err)
	}

	// Enforce result size limit
	const maxLogs = 10000
	if len(logs) > maxLogs {
		return nil, fmt.Errorf("result too large: %d logs (maximum: %d)", len(logs), maxLogs)
	}

	// Cache the result
	qe.cache.Set(cacheKey, logs, 0)
	logger.Debug("cached logs", "key", cacheKey, "count", len(logs))

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
		return nil, fmt.Errorf("block range too large for transactions query: %s blocks (maximum: 1000)", blockRange.String())
	}

	// Process each block in the range
	for blockNum := new(big.Int).Set(query.FromBlock); blockNum.Cmp(query.ToBlock) <= 0; blockNum = new(big.Int).Add(blockNum, big.NewInt(1)) {
		// Check for context cancellation
		if ctx.Err() != nil {
			return nil, fmt.Errorf("query cancelled after processing %d transactions: %w", len(transactions), ctx.Err())
		}

		block, err := qe.client.BlockByNumber(ctx, blockNum)
		if err != nil {
			return nil, fmt.Errorf("failed to get block %s: %w", blockNum.String(), err)
		}

		for _, tx := range block.Transactions() {
			msg, err := core.TransactionToMessage(tx, types.NewLondonSigner(tx.ChainId()), nil)
			if err != nil {
				// Log error but continue processing
				continue
			}

			// Check if transaction is from or to the specified address
			if (tx.To() != nil && *tx.To() == query.Address) || msg.From == query.Address {
				transactions = append(transactions, tx)

				// Enforce result size limit
				if len(transactions) >= 10000 {
					return nil, fmt.Errorf("result set too large: 10000+ transactions found - please narrow your block range")
				}
			}
		}
	}

	return transactions, nil
}

// getTransactionsConcurrent processes blocks concurrently for better performance
func (qe *QueryExecutor) getTransactionsConcurrent(ctx context.Context, query *queries.Query) ([]*types.Transaction, error) {
	var fromBlock, toBlock *big.Int

	if query.FromBlock == nil || query.ToBlock == nil {
		latestBlock, err := qe.client.BlockNumber(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting latest block: %w", err)
		}
		fromBlock = big.NewInt(int64(latestBlock - 100))
		toBlock = big.NewInt(int64(latestBlock))
	} else {
		fromBlock = query.FromBlock
		toBlock = query.ToBlock
	}

	// Validate block range
	blockRange := new(big.Int).Sub(toBlock, fromBlock)
	if blockRange.Cmp(big.NewInt(1000)) > 0 {
		return nil, fmt.Errorf("block range too large for transactions query: %d blocks (maximum: 1000)", blockRange.Int64())
	}

	// Generate cache key
	cacheKey := cache.GenerateKey("transactions", query.Address.Hex(), fromBlock, toBlock)

	// Check cache first
	if cached, found := qe.cache.Get(cacheKey); found {
		logger.Debug("cache hit", "key", cacheKey)
		if txs, ok := cached.([]*types.Transaction); ok {
			return txs, nil
		}
	}

	// Create worker pool
	type blockResult struct {
		blockNum     *big.Int
		transactions []*types.Transaction
		err          error
	}

	blockChan := make(chan *big.Int, qe.maxWorkers)
	resultChan := make(chan blockResult, qe.maxWorkers)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < qe.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for blockNum := range blockChan {
				if ctx.Err() != nil {
					return
				}

				block, err := qe.client.BlockByNumber(ctx, blockNum)
				if err != nil {
					resultChan <- blockResult{blockNum: blockNum, err: fmt.Errorf("failed to get block %s: %w", blockNum.String(), err)}
					continue
				}

				var blockTxs []*types.Transaction
				for _, tx := range block.Transactions() {
					msg, err := core.TransactionToMessage(tx, types.NewLondonSigner(tx.ChainId()), nil)
					if err != nil {
						continue
					}

					if (tx.To() != nil && *tx.To() == query.Address) || msg.From == query.Address {
						blockTxs = append(blockTxs, tx)
					}
				}

				resultChan <- blockResult{blockNum: blockNum, transactions: blockTxs}
			}
		}()
	}

	// Send blocks to workers
	go func() {
		for blockNum := new(big.Int).Set(query.FromBlock); blockNum.Cmp(query.ToBlock) <= 0; blockNum = new(big.Int).Add(blockNum, big.NewInt(1)) {
			select {
			case <-ctx.Done():
				close(blockChan)
				return
			case blockChan <- new(big.Int).Set(blockNum):
			}
		}
		close(blockChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var allTransactions []*types.Transaction
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		allTransactions = append(allTransactions, result.transactions...)

		// Enforce result size limit
		const maxTransactions = 10000
		if len(allTransactions) > maxTransactions {
			return nil, fmt.Errorf("result too large: %d transactions (maximum: %d)", len(allTransactions), maxTransactions)
		}
	}

	if ctx.Err() != nil {
		return nil, fmt.Errorf("query cancelled after processing %d transactions: %w", len(allTransactions), ctx.Err())
	}

	// Cache the result
	qe.cache.Set(cacheKey, allTransactions, 0)
	logger.Debug("cached transactions", "key", cacheKey, "count", len(allTransactions))

	return allTransactions, nil
}
