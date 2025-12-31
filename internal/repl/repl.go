package repl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/devlongs/evmql/internal/executor"
	"github.com/devlongs/evmql/internal/logger"
	"github.com/devlongs/evmql/internal/parser"
)

// holds REPL configuration
type Config struct {
	HistoryFile   string
	MaxHistoryLen int
	ColorOutput   bool
	ShowTimings   bool
}

// Start initializes and runs the REPL loop for querying
func Start(parser *parser.Parser, executor *executor.QueryExecutor, config Config) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Entering EVMQL interactive mode. Type your query, or type 'exit' to quit.")
	fmt.Println("Type 'help' for available commands.")

	for {
		fmt.Print("evmql> ")
		scanned := scanner.Scan()
		if !scanned {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Exiting EVMQL interactive mode.")
			break
		}

		if input == "help" {
			showHelp()
			continue
		}

		// Handle timing if enabled
		var startTime time.Time
		if config.ShowTimings {
			startTime = time.Now()
		}

		// Process the query
		query, err := parser.ParseQuery(input)
		if err != nil {
			logger.Error("query parsing failed", "error", err, "input", input)
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Execute the query with a timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		result, err := executor.Execute(ctx, query)
		cancel()

		if err != nil {
			logger.Error("query execution failed", "error", err, "query", query.Method)
			fmt.Printf("Error: %v\n", err)
			continue
		}

		logger.Info("query executed", "method", query.Method, "address", query.Address.Hex())
		fmt.Printf("Result: %v\n", result)

		// Show execution time if enabled
		if config.ShowTimings {
			duration := time.Since(startTime)
			logger.Debug("execution time", "duration", duration)
			fmt.Printf("Executed in %v\n", duration)
		}

		fmt.Println() // Empty line for better readability
	}
}

// showHelp displays available commands
func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  SELECT BALANCE FROM <address> [BLOCK <number>] - Get account balance")
	fmt.Println("  SELECT LOGS FROM <address> BLOCK <from> <to> - Get logs within block range")
	fmt.Println("  SELECT TRANSACTIONS FROM <address> [BLOCK <from> <to>] - Get transactions")
	fmt.Println("  exit, quit - Exit the program")
	fmt.Println("  help - Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  SELECT BALANCE FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
	fmt.Println("  SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000 1100000")
	fmt.Println()
}
