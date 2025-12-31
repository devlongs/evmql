# EVMQL

![License](https://img.shields.io/badge/license-MIT-blue)
![Status](https://img.shields.io/badge/status-active%20development-green)

> A SQL-like query language interface for Ethereum and other EVM-compatible blockchains

EVMQL provides a familiar, SQL-inspired syntax for interacting with blockchain data. It aims to simplify blockchain queries and data retrieval for developers, analysts, and users familiar with SQL.

## Overview

EVMQL bridges the gap between traditional database querying and blockchain data access by providing:

- Simple, SQL-like syntax for blockchain queries
- Support for multiple Ethereum networks (Mainnet, Sepolia, etc.)
- Interactive shell with history and auto-completion
- Easily extensible architecture for custom query types

## Development Status

This project is under active development with the following status:

| Feature                    | Status       |
|----------------------------|--------------|
| Basic query structure      | Complete     |
| Balance queries            | Complete     |
| Log queries                | Complete     |
| Transaction queries        | Complete     |
| Configuration management   | Complete     |
| Query result formatting    | In Progress  |
| Smart contract interaction | Planned      |
| Comprehensive testing      | Planned      |
| Advanced filtering         | Planned      |

## Installation

### Prerequisites

- Go 1.20 or higher
- Access to an Ethereum node (e.g., Infura API key, or local node)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/devlongs/evmql.git
cd evmql

# Install dependencies
go mod tidy

# Build the project
go build -o evmql cmd/evmql/main.go

# Initialize configuration (optional)
./evmql --generate-config
```

## Usage

### Interactive Mode

Start the interactive shell:

```bash
./evmql
```
This launches the EVMQL interactive shell where you can type commands directly:

```bash
evmql> SELECT BALANCE FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e
Result: 1500000000000000000 (1.5 ETH)

evmql> SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000 1100000
...
```

### Command Line Mode

For one-off queries:

```bash
./evmql --interactive=false "SELECT BALANCE FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
```

### Network Selection

Specify a network to connect to:

```bash
../evmql --network sepolia
```

### Configuration
EVMQL uses a layered configuration system:

1. Default values built into the application
2. Configuration file (typically ~/.evmql/config.json)
3. Environment variables (prefixed with EVMQL_)
4. Command-line flags (take highest precedence)

### Configuration File
Generate a default configuration file:

```bash
./evmql --generate-config
```

This creates a configuration file at ~/.evmql/config.json with default settings.