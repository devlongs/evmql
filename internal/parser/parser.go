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
	parts := strings.Fields(strings.TrimSpace(queryStr))
	if len(parts) < 4 || strings.ToUpper(parts[0]) != "SELECT" || strings.ToUpper(parts[2]) != "FROM" {
		return nil, errors.New("invalid query format; expected SELECT <method> FROM <address>")
	}

	// Initialize the query
	query := &queries.Query{
		Type:   "SELECT",
		Method: strings.ToUpper(parts[1]),
	}

	// Parse the address
	address := parts[3]
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid address: %s", address)
	}
	query.Address = common.HexToAddress(address)

	// Parse optional block range
	if len(parts) > 4 && strings.ToUpper(parts[4]) == "BLOCK" && len(parts) >= 7 {
		fromBlock, ok := new(big.Int).SetString(parts[5], 10)
		if !ok {
			return nil, fmt.Errorf("invalid from block: %s", parts[5])
		}
		toBlock, ok := new(big.Int).SetString(parts[6], 10)
		if !ok {
			return nil, fmt.Errorf("invalid to block: %s", parts[6])
		}
		query.FromBlock = fromBlock
		query.ToBlock = toBlock
	}

	return query, nil
}
