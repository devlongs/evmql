package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigFilePath returns the path to the default configuration file
func ConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if we can't get home dir
		return "evmql.json"
	}
	return filepath.Join(homeDir, ".evmql", "config.json")
}

// creates a default configuration file if it doesn't exist
func InitConfigFile(path string) error {
	if path == "" {
		path = ConfigFilePath()
	}

	// Check if the file already exists
	if _, err := os.Stat(path); err == nil {
		// File exists, don't overwrite
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking for config file: %w", err)
	}

	// File doesn't exist, create it with default config
	config := DefaultConfig()

	if err := SaveConfig(config, path); err != nil {
		return err
	}

	fmt.Printf("Created default configuration file at %s\n", path)
	return nil
}

// AddNetwork adds a new network configuration
func (c *Config) AddNetwork(name string, chainID int64, nodeURL, explorer string) {
	network := NetworkConfig{
		ChainID:   chainID,
		Name:      name,
		NodeURL:   nodeURL,
		Explorer:  explorer,
		Contracts: make(map[string]string),
	}

	if c.Networks == nil {
		c.Networks = make(NetworksConfig)
	}

	c.Networks[name] = network
}

// GetConfigFilePaths returns a list of potential config file paths in order of precedence
func GetConfigFilePaths() []string {
	paths := []string{}

	// Current directory
	paths = append(paths, "./evmql.json")

	// User's home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".evmql", "config.json"))
	}

	// System directory
	paths = append(paths, "/etc/evmql/config.json")

	return paths
}

// FindConfigFile locates the first available config file from the standard locations
func FindConfigFile() (string, error) {
	for _, path := range GetConfigFilePaths() {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no config file found in standard locations")
}
