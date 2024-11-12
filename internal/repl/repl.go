package repl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/devlongs/evmql/internal/executor"
	"github.com/devlongs/evmql/internal/parser"
)

// Start initializes and runs the REPL loop for querying
func Start(parser *parser.Parser, executor *executor.QueryExecutor) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Entering EVMQL interactive mode. Type your query, or type 'exit' to quit.")

	for {
		fmt.Print("evmql> ")
		scanned := scanner.Scan()
		if !scanned {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "exit" {
			fmt.Println("Exiting EVMQL interactive mode.")
			break
		}

		query, err := parser.ParseQuery(input)
		if err != nil {
			fmt.Printf("Error parsing query: %v\n", err)
			continue
		}

		result, err := executor.Execute(context.Background(), query)
		if err != nil {
			fmt.Printf("Error executing query: %v\n", err)
			continue
		}

		fmt.Printf("Result: %v\n\n", result)
	}
}
