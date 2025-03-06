package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devlongs/evmql/internal/executor"
	"github.com/devlongs/evmql/internal/parser"
	"github.com/devlongs/evmql/internal/repl"
	"github.com/devlongs/evmql/pkg/config"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	// Command line flags
	configPath      = flag.String("config", "", "Path to configuration file")
	generateConfig  = flag.Bool("generate-config", false, "Generate a default configuration file")
	initConfigPath  = flag.String("init", "", "Initialize a default configuration file at the specified path")
	showVersion     = flag.Bool("version", false, "Show version information and exit")
	networkName     = flag.String("network", "", "Network to connect to (mainnet, sepolia, etc.)")
	nodeURL         = flag.String("node", "", "Ethereum node URL (overrides config)")
	interactiveMode = flag.Bool("interactive", true, "Run in interactive mode")
)

const (
	Version = "0.1.0"
)

func main() {
	flag.Parse()

	// Show version if requested
	if *showVersion {
		fmt.Printf("EVMQL Version %s\n", Version)
		return
	}

	// Generate default config if requested
	if *generateConfig {
		if *initConfigPath == "" {
			*initConfigPath = config.ConfigFilePath()
		}
		if err := config.InitConfigFile(*initConfigPath); err != nil {
			log.Fatalf("Failed to generate config file: %v", err)
		}
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath, os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override with command line flags
	if *networkName != "" {
		network, exists := cfg.Networks[*networkName]
		if !exists {
			log.Fatalf("Unknown network: %s", *networkName)
		}
		cfg.DefaultChainID = network.ChainID
		if network.NodeURL != "" {
			cfg.Node.URL = network.NodeURL
		}
	}

	if *nodeURL != "" {
		cfg.Node.URL = *nodeURL
	}

	if err := config.ValidateConfig(cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupts
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
		cancel()
		// Give a little time for graceful shutdown then force exit
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()

	// Connect to Ethereum node
	fmt.Printf("Connecting to Ethereum node at %s...\n", cfg.Node.URL)

	clientCtx, clientCancel := context.WithTimeout(ctx, cfg.Node.Timeout)
	defer clientCancel()

	client, err := ethclient.DialContext(clientCtx, cfg.Node.URL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}
	defer client.Close()

	// Get chain ID to confirm connection
	chainID, err := client.ChainID(clientCtx)
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}
	fmt.Printf("Connected to chain with ID: %d\n", chainID)

	// Check if connected to the expected network
	if chainID.Int64() != cfg.DefaultChainID {
		networkInfo, found := cfg.GetNetworkByChainID(chainID.Int64())
		if found {
			fmt.Printf("Warning: Connected to %s instead of the configured default network\n", networkInfo.Name)
		} else {
			fmt.Printf("Warning: Connected to chain ID %d, which doesn't match any configured network\n", chainID)
		}
	}

	// Initialize parser and executor
	queryParser := parser.NewParser()
	queryExecutor := executor.NewQueryExecutor(client)

	// Set timeout for query execution
	queryExecutor.SetTimeout(time.Duration(cfg.Query.TimeoutSeconds) * time.Second)

	// If in interactive mode, start REPL
	if *interactiveMode {
		replConfig := repl.Config{
			HistoryFile:   cfg.REPL.HistoryFile,
			MaxHistoryLen: cfg.REPL.MaxHistoryLen,
			ColorOutput:   cfg.REPL.ColorOutput,
			ShowTimings:   cfg.REPL.ShowTimings,
		}
		repl.Start(queryParser, queryExecutor, replConfig)
	} else {
		// Execute any provided queries from command line args
		if flag.NArg() > 0 {
			queryStr := flag.Arg(0)
			query, err := queryParser.ParseQuery(queryStr)
			if err != nil {
				log.Fatalf("Error parsing query: %v", err)
			}

			queryCtx, queryCancel := context.WithTimeout(ctx, time.Duration(cfg.Query.TimeoutSeconds)*time.Second)
			defer queryCancel()

			result, err := queryExecutor.Execute(queryCtx, query)
			if err != nil {
				log.Fatalf("Error executing query: %v", err)
			}
			fmt.Printf("Result: %v\n", result)
		} else {
			fmt.Println("No query provided. Run with -interactive flag for REPL mode or provide a query as an argument.")
		}
	}
}
