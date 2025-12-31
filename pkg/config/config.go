package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Node           NodeConfig     `json:"node" mapstructure:"node"`
	Query          QueryConfig    `json:"query" mapstructure:"query"`
	Cache          CacheConfig    `json:"cache" mapstructure:"cache"`
	REPL           REPLConfig     `json:"repl" mapstructure:"repl"`
	Networks       NetworksConfig `json:"networks" mapstructure:"networks"`
	DefaultChainID int64          `json:"default_chain_id" mapstructure:"default_chain_id"`
}

// Ethereum node connection settings
type NodeConfig struct {
	URL               string        `json:"url" mapstructure:"url"`
	Timeout           time.Duration `json:"timeout" mapstructure:"timeout"`
	MaxConcurrentReqs int           `json:"max_concurrent_requests" mapstructure:"max_concurrent_requests"`
	RetryCount        int           `json:"retry_count" mapstructure:"retry_count"`
	RetryDelay        time.Duration `json:"retry_delay" mapstructure:"retry_delay"`
}

// settings for query execution
type QueryConfig struct {
	DefaultBlockRange  int64 `json:"default_block_range" mapstructure:"default_block_range"`
	MaxBlockRange      int64 `json:"max_block_range" mapstructure:"max_block_range"`
	DefaultGasLimit    int64 `json:"default_gas_limit" mapstructure:"default_gas_limit"`
	ResultSizeLimit    int   `json:"result_size_limit" mapstructure:"result_size_limit"`
	TimeoutSeconds     int   `json:"timeout_seconds" mapstructure:"timeout_seconds"`
	ShowGasEstimates   bool  `json:"show_gas_estimates" mapstructure:"show_gas_estimates"`
	PrettyPrintResults bool  `json:"pretty_print_results" mapstructure:"pretty_print_results"`
	IncludeRawData     bool  `json:"include_raw_data" mapstructure:"include_raw_data"`
}

// query caching settings
type CacheConfig struct {
	Enabled      bool          `json:"enabled" mapstructure:"enabled"`
	MaxItems     int           `json:"max_items" mapstructure:"max_items"`
	DefaultTTL   time.Duration `json:"default_ttl" mapstructure:"default_ttl"`
	CleanupEvery time.Duration `json:"cleanup_every" mapstructure:"cleanup_every"`
}

// settings for the interactive REPL
type REPLConfig struct {
	HistoryFile   string `json:"history_file" mapstructure:"history_file"`
	MaxHistoryLen int    `json:"max_history_len" mapstructure:"max_history_len"`
	ColorOutput   bool   `json:"color_output" mapstructure:"color_output"`
	ShowTimings   bool   `json:"show_timings" mapstructure:"show_timings"`
}

// NetworkConfig holds settings for a specific network
type NetworkConfig struct {
	ChainID   int64             `json:"chain_id" mapstructure:"chain_id"`
	Name      string            `json:"name" mapstructure:"name"`
	NodeURL   string            `json:"node_url" mapstructure:"node_url"`
	Explorer  string            `json:"explorer" mapstructure:"explorer"`
	Contracts map[string]string `json:"contracts" mapstructure:"contracts"`
}

// a map of network configs by name
type NetworksConfig map[string]NetworkConfig

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Node: NodeConfig{
			URL:               "http://localhost:8545",
			Timeout:           10 * time.Second,
			MaxConcurrentReqs: 5,
			RetryCount:        3,
			RetryDelay:        2 * time.Second,
		},
		Query: QueryConfig{
			DefaultBlockRange:  1000,
			MaxBlockRange:      10000,
			DefaultGasLimit:    3000000,
			ResultSizeLimit:    1000,
			TimeoutSeconds:     30,
			ShowGasEstimates:   true,
			PrettyPrintResults: true,
		},
		Cache: CacheConfig{
			Enabled:      true,
			MaxItems:     1000,
			DefaultTTL:   5 * time.Minute,
			CleanupEvery: 10 * time.Minute,
		},
		REPL: REPLConfig{
			HistoryFile:   filepath.Join(os.Getenv("HOME"), ".evmql_history"),
			MaxHistoryLen: 1000,
			ColorOutput:   true,
			ShowTimings:   true,
		},
		Networks: NetworksConfig{
			"mainnet": {
				ChainID:  1,
				Name:     "Ethereum Mainnet",
				NodeURL:  "https://mainnet.infura.io/v3/YOUR_KEY",
				Explorer: "https://etherscan.io",
			},
			"sepolia": {
				ChainID:  11155111,
				Name:     "Sepolia Testnet",
				NodeURL:  "https://sepolia.infura.io/v3/YOUR_KEY",
				Explorer: "https://sepolia.etherscan.io",
			},
		},
		DefaultChainID: 1, // Mainnet by default
	}
}

// loads the configuration from various sources
func LoadConfig(configPath string, args []string) (*Config, error) {
	// Start with default config
	config := DefaultConfig()

	// Initialize viper
	v := viper.New()
	v.SetConfigType("json")
	v.SetConfigName("evmql")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.evmql")
	v.AddConfigPath("/etc/evmql")

	// If config path is specified, add it
	if configPath != "" {
		v.SetConfigFile(configPath)
	}

	// Environment variables
	v.SetEnvPrefix("EVMQL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, but that's OK - we'll use defaults and other sources
	}

	flags := pflag.NewFlagSet("evmql", pflag.ContinueOnError)

	// Node flags
	flags.String("node.url", config.Node.URL, "Ethereum node URL")
	flags.Duration("node.timeout", config.Node.Timeout, "Node request timeout")

	// Query flags
	flags.Int64("query.max-block-range", config.Query.MaxBlockRange, "Maximum block range for queries")
	flags.Bool("query.pretty-print", config.Query.PrettyPrintResults, "Pretty-print query results")

	// Network flags
	flags.String("network", "", "Network to connect to (mainnet, sepolia, etc.)")
	flags.Int64("chain-id", config.DefaultChainID, "Chain ID to use")

	// Cache flags
	flags.Bool("cache.enabled", config.Cache.Enabled, "Enable query caching")

	// Parse the flags
	if err := flags.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	// Bind flags to viper
	if err := v.BindPFlags(flags); err != nil {
		return nil, fmt.Errorf("error binding flags: %w", err)
	}

	// Unmarshal config
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Handle network selection
	networkName := v.GetString("network")
	if networkName != "" {
		network, exists := config.Networks[networkName]
		if !exists {
			return nil, fmt.Errorf("unknown network: %s", networkName)
		}

		// Override node URL with the selected network's URL
		if network.NodeURL != "" {
			config.Node.URL = network.NodeURL
		}

		// Set chain ID
		config.DefaultChainID = network.ChainID
	}

	// Further customizations from env vars or other sources
	if yourKey := os.Getenv("YOUR_API_KEY"); yourKey != "" {
		// Replace placeholders in node URLs
		for name, network := range config.Networks {
			network.NodeURL = strings.Replace(network.NodeURL, "YOUR_KEY", yourKey, 1)
			config.Networks[name] = network
		}

		config.Node.URL = strings.Replace(config.Node.URL, "YOUR_KEY", yourKey, 1)
	}

	return config, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

func ValidateConfig(config *Config) error {
	if config.Node.URL == "" {
		return errors.New("node URL cannot be empty")
	}

	if strings.Contains(config.Node.URL, "YOUR_KEY") {
		return errors.New("node URL contains placeholder 'YOUR_KEY' - please set a valid API key")
	}

	if config.Query.MaxBlockRange <= 0 {
		return errors.New("max block range must be positive")
	}

	if config.Query.MaxBlockRange > 10000 {
		return errors.New("max block range cannot exceed 10000 to prevent resource exhaustion")
	}

	if config.Query.TimeoutSeconds <= 0 {
		return errors.New("query timeout must be positive")
	}

	if config.Query.TimeoutSeconds > 300 {
		return errors.New("query timeout cannot exceed 300 seconds")
	}

	if len(config.Networks) == 0 {
		return errors.New("at least one network must be defined")
	}

	for name, network := range config.Networks {
		if strings.Contains(network.NodeURL, "YOUR_KEY") {
			return fmt.Errorf("network '%s' contains placeholder 'YOUR_KEY' - please set a valid API key", name)
		}
	}

	return nil
}

func (config *Config) GetNetworkByChainID(chainID int64) (NetworkConfig, bool) {
	for _, network := range config.Networks {
		if network.ChainID == chainID {
			return network, true
		}
	}
	return NetworkConfig{}, false
}

func (config *Config) GetDefaultNetwork() (NetworkConfig, bool) {
	return config.GetNetworkByChainID(config.DefaultChainID)
}

func (config *Config) GetContractAddress(contractName string) (common.Address, bool) {
	network, found := config.GetDefaultNetwork()
	if !found {
		return common.Address{}, false
	}

	addressStr, found := network.Contracts[contractName]
	if !found {
		return common.Address{}, false
	}

	if !common.IsHexAddress(addressStr) {
		return common.Address{}, false
	}

	return common.HexToAddress(addressStr), true
}
