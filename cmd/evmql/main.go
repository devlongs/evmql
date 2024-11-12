package main

import (
	"log"

	"github.com/devlongs/evmql/internal/executor"
	"github.com/devlongs/evmql/internal/parser"
	"github.com/devlongs/evmql/internal/repl"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/")
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}
	defer client.Close()

	// Initialize parser and executor
	queryParser := parser.NewParser()
	queryExecutor := executor.NewQueryExecutor(client)

	// Start REPL
	repl.Start(queryParser, queryExecutor)
}
