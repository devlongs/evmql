package parser

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/devlongs/evmql/pkg/queries"
	"github.com/ethereum/go-ethereum/common"
)

// Parser struct to handle parsing logic
type Parser struct{}

// NewParser creates a new instance of Parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseQuery parses the EVMQL query string and returns a Query object
func (p *Parser) ParseQuery(queryStr string) (*queries.Query, error) {
	queryStr = SanitizeInput(queryStr)

	if len(queryStr) == 0 {
		return nil, errors.New("query cannot be empty")
	}

	if len(queryStr) > 10000 {
		return nil, errors.New("query too long: maximum 10000 characters")
	}

	parts := strings.Fields(queryStr)
	if len(parts) < 4 || strings.ToUpper(parts[0]) != "SELECT" || strings.ToUpper(parts[2]) != "FROM" {
		return nil, errors.New("invalid query format; expected SELECT <method> FROM <address>")
	}

	method := strings.ToUpper(parts[1])
	validMethods := map[string]bool{
		"BALANCE":      true,
		"LOGS":         true,
		"TRANSACTIONS": true,
	}
	if !validMethods[method] {
		return nil, fmt.Errorf("unsupported method: %s (supported: BALANCE, LOGS, TRANSACTIONS)", method)
	}

	// Initialize the query
	query := &queries.Query{
		Type:   "SELECT",
		Method: method,
	}

	// Parse and sanitize the address
	address := NormalizeAddress(parts[3])
	if !ValidateAddressFormat(address) {
		return nil, fmt.Errorf("invalid Ethereum address format: %s", TruncateForDisplay(parts[3], 50))
	}
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid Ethereum address: %s (must be 42 character hex starting with 0x)", TruncateForDisplay(address, 50))
	}
	query.Address = common.HexToAddress(address)

	// Parse optional block range
	if len(parts) > 4 && strings.ToUpper(parts[4]) == "BLOCK" {
		if len(parts) < 7 {
			return nil, errors.New("BLOCK keyword requires both from and to block numbers")
		}

		fromBlockStr := strings.TrimSpace(parts[5])
		toBlockStr := strings.TrimSpace(parts[6])

		fromBlock, ok := new(big.Int).SetString(fromBlockStr, 10)
		if !ok || fromBlock.Sign() < 0 {
			return nil, fmt.Errorf("invalid from block: %s (must be non-negative integer)", TruncateForDisplay(fromBlockStr, 20))
		}

		toBlock, ok := new(big.Int).SetString(toBlockStr, 10)
		if !ok || toBlock.Sign() < 0 {
			return nil, fmt.Errorf("invalid to block: %s (must be non-negative integer)", TruncateForDisplay(toBlockStr, 20))
		}

		if fromBlock.Cmp(toBlock) > 0 {
			return nil, fmt.Errorf("from block cannot be greater than to block")
		}

		blockRange := new(big.Int).Sub(toBlock, fromBlock)
		if blockRange.Cmp(big.NewInt(10000)) > 0 {
			return nil, fmt.Errorf("block range too large: %d blocks (maximum: 10000)", blockRange.Int64())
		}

		query.FromBlock = fromBlock
		query.ToBlock = toBlock
	}

	return query, nil
}
