package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if cfg.Node.URL == "" {
		t.Error("Default node URL should not be empty")
	}

	if cfg.Node.Timeout == 0 {
		t.Error("Default timeout should not be zero")
	}

	if cfg.Query.MaxBlockRange <= 0 {
		t.Error("Default max block range should be positive")
	}

	if len(cfg.Networks) == 0 {
		t.Error("Default config should have at least one network")
	}

	if cfg.DefaultChainID == 0 {
		t.Error("Default chain ID should not be zero")
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	cfg := &Config{
		Node: NodeConfig{
			URL:               "http://localhost:8545",
			Timeout:           10 * time.Second,
			MaxConcurrentReqs: 5,
		},
		Query: QueryConfig{
			MaxBlockRange:  1000,
			TimeoutSeconds: 30,
		},
		Networks: NetworksConfig{
			"test": {
				ChainID: 1,
				Name:    "Test Network",
				NodeURL: "http://localhost:8545",
			},
		},
		DefaultChainID: 1,
	}

	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}
}

func TestValidateConfig_EmptyNodeURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Node.URL = ""

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("Expected error for empty node URL")
	}
}

func TestValidateConfig_PlaceholderAPIKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Node.URL = "https://mainnet.infura.io/v3/YOUR_KEY"

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("Expected error for placeholder API key in node URL")
	}
}

func TestValidateConfig_NetworkPlaceholderAPIKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Node.URL = "http://localhost:8545"
	cfg.Networks["mainnet"] = NetworkConfig{
		ChainID: 1,
		Name:    "Mainnet",
		NodeURL: "https://mainnet.infura.io/v3/YOUR_KEY",
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("Expected error for placeholder API key in network URL")
	}
}

func TestValidateConfig_MaxBlockRangeTooLarge(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Node.URL = "http://localhost:8545"
	cfg.Query.MaxBlockRange = 20000

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("Expected error for max block range exceeding 10000")
	}
}

func TestValidateConfig_TimeoutTooLarge(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Node.URL = "http://localhost:8545"
	cfg.Query.TimeoutSeconds = 500

	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("Expected error for timeout exceeding 300 seconds")
	}
}

func TestValidateConfig_NegativeValues(t *testing.T) {
	tests := []struct {
		name     string
		modifier func(*Config)
	}{
		{
			name: "Negative max block range",
			modifier: func(c *Config) {
				c.Node.URL = "http://localhost:8545"
				c.Query.MaxBlockRange = -1
			},
		},
		{
			name: "Negative timeout",
			modifier: func(c *Config) {
				c.Node.URL = "http://localhost:8545"
				c.Query.TimeoutSeconds = -1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modifier(cfg)

			err := ValidateConfig(cfg)
			if err == nil {
				t.Error("Expected error for negative value")
			}
		})
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	originalCfg := DefaultConfig()
	originalCfg.Node.URL = "http://testnode:8545"
	originalCfg.Query.MaxBlockRange = 5000

	err := SaveConfig(originalCfg, configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	loadedCfg, err := LoadConfig(configPath, []string{})
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.Node.URL != originalCfg.Node.URL {
		t.Errorf("Expected node URL %s, got %s", originalCfg.Node.URL, loadedCfg.Node.URL)
	}

	if loadedCfg.Query.MaxBlockRange != originalCfg.Query.MaxBlockRange {
		t.Errorf("Expected max block range %d, got %d", originalCfg.Query.MaxBlockRange, loadedCfg.Query.MaxBlockRange)
	}
}

func TestGetNetworkByChainID(t *testing.T) {
	cfg := DefaultConfig()

	network, found := cfg.GetNetworkByChainID(1)
	if !found {
		t.Error("Expected to find mainnet (chain ID 1)")
	}
	if network.ChainID != 1 {
		t.Errorf("Expected chain ID 1, got %d", network.ChainID)
	}

	_, found = cfg.GetNetworkByChainID(999999)
	if found {
		t.Error("Should not find network with non-existent chain ID")
	}
}

func TestGetDefaultNetwork(t *testing.T) {
	cfg := DefaultConfig()

	network, found := cfg.GetDefaultNetwork()
	if !found {
		t.Error("Expected to find default network")
	}
	if network.ChainID != cfg.DefaultChainID {
		t.Errorf("Expected chain ID %d, got %d", cfg.DefaultChainID, network.ChainID)
	}
}

func TestAddNetwork(t *testing.T) {
	cfg := DefaultConfig()
	initialCount := len(cfg.Networks)

	cfg.AddNetwork("custom", 1337, "http://custom:8545", "https://explorer.custom")

	if len(cfg.Networks) != initialCount+1 {
		t.Error("Network was not added")
	}

	network, exists := cfg.Networks["custom"]
	if !exists {
		t.Fatal("Custom network not found")
	}

	if network.ChainID != 1337 {
		t.Errorf("Expected chain ID 1337, got %d", network.ChainID)
	}
	if network.NodeURL != "http://custom:8545" {
		t.Errorf("Expected node URL http://custom:8545, got %s", network.NodeURL)
	}
}

func TestConfigFilePath(t *testing.T) {
	path := ConfigFilePath()
	if path == "" {
		t.Error("Config file path should not be empty")
	}

	if !filepath.IsAbs(path) && path != "evmql.json" {
		t.Errorf("Expected absolute path or evmql.json, got %s", path)
	}
}

func TestInitConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	err := InitConfigFile(configPath)
	if err != nil {
		t.Fatalf("Failed to init config file: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	err = InitConfigFile(configPath)
	if err != nil {
		t.Errorf("Second init should not error: %v", err)
	}
}

func TestGetContractAddress(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Networks["mainnet"] = NetworkConfig{
		ChainID: 1,
		Name:    "Mainnet",
		Contracts: map[string]string{
			"USDT": "0xdac17f958d2ee523a2206206994597c13d831ec7",
		},
	}
	cfg.DefaultChainID = 1

	addr, found := cfg.GetContractAddress("USDT")
	if !found {
		t.Error("Expected to find USDT contract")
	}
	if addr.Hex() != "0xdAC17F958D2ee523a2206206994597C13D831ec7" {
		t.Errorf("Unexpected contract address: %s", addr.Hex())
	}

	_, found = cfg.GetContractAddress("NONEXISTENT")
	if found {
		t.Error("Should not find non-existent contract")
	}
}
