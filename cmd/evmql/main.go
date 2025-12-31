package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devlongs/evmql/internal/cache"
	"github.com/devlongs/evmql/internal/executor"
	"github.com/devlongs/evmql/internal/logger"
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
		logger.Info("version", "version", Version)
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
		logger.Info("shutdown initiated", "signal", sig)
		cancel()
		// Give a little time for graceful shutdown then force exit
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()

	// Connect to Ethereum node
	logger.Info("connecting to ethereum node", "url", cfg.Node.URL)

	clientCtx, clientCancel := context.WithTimeout(ctx, cfg.Node.Timeout)
	defer clientCancel()

	client, err := ethclient.DialContext(clientCtx, cfg.Node.URL)
	if err != nil {
		logger.Error("failed to connect to ethereum node", "error", err)
		log.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	// Verify connection and get chain ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		logger.Error("failed to get chain ID", "error", err)
		log.Fatalf("Failed to get chain ID: %v", err)
	}
	logger.Info("connected to ethereum", "chain_id", chainID)

	// Check if connected to the expected network
	if chainID.Int64() != cfg.DefaultChainID {
		networkInfo, found := cfg.GetNetworkByChainID(chainID.Int64())
		if found {
			logger.Warn("network mismatch", "connected_to", networkInfo.Name, "expected", "default network")
		} else {
			logger.Warn("unknown network", "chain_id", chainID)
		}
	}

	// Initialize parser and executor
	queryParser := parser.NewParser()
	queryExecutor := executor.NewQueryExecutor(client)

	// Set timeout for query execution
	queryExecutor.SetTimeout(time.Duration(cfg.Query.TimeoutSeconds) * time.Second)

	// Initialize cache if enabled
	if cfg.Cache.Enabled {
		queryCache := cache.NewInMemoryCache(
			cfg.Cache.MaxItems,
			cfg.Cache.DefaultTTL,
			cfg.Cache.CleanupEvery,
		)
		queryExecutor.SetCache(queryCache)
		logger.Info("cache enabled", "max_items", cfg.Cache.MaxItems, "ttl", cfg.Cache.DefaultTTL)
	} else {
		logger.Info("cache disabled")
	}

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
				logger.Error("query execution failed", "error", err)
				log.Fatalf("Error executing query: %v", err)
			}
			logger.Info("query result", "result", result)
		} else {
			logger.Info("no query provided", "hint", "use -interactive flag for REPL mode or provide a query as an argument")
		}
	}
}
