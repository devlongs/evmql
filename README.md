# EVMQL

> ⚠️ **Note: This project is under active development.** Features may change, and the API is not yet stable.

EVMQL is a SQL-like query language interface for interacting with EVM (Ethereum Virtual Machine) chains. It provides a familiar, easy-to-use syntax for blockchain data queries.

## 🚧 Development Status

This project is in active development. Current focus areas:
- [x] Basic query structure and parser
- [x] Balance queries
- [x] Log queries
- [ ] Transaction queries
- [ ] Smart contract interaction
- [ ] Query result formatting
- [ ] Configuration management
- [ ] Advanced filtering options

## 🚀 Quick Start

### Prerequisites
- Go 1.20 or higher
- Access to an Ethereum node (e.g., Infura API key)

### Installation
```bash
# Clone the repository
git clone https://github.com/devlongs/evmql.git
cd evmql

# Install dependencies
go mod tidy

# Build the project
go build -o evmql cmd/evmql/main.go
```

### Usage

Start the interactive shell:
```bash
./evmql
```

Example queries:
```sql
-- Get account balance
SELECT BALANCE FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e

-- Get balance at specific block
SELECT BALANCE FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000

-- Get logs within block range
SELECT LOGS FROM 0x742d35Cc6634C0532925a3b844Bc454e4438f44e BLOCK 1000000 1100000
```

## 📖 Documentation

### Supported Queries

#### SELECT BALANCE
```sql
SELECT BALANCE FROM <address> [BLOCK <number>]
```
- `address`: Ethereum address (hex format)
- `BLOCK`: (Optional) Specific block number for historical balance

#### SELECT LOGS
```sql
SELECT LOGS FROM <address> BLOCK <from> <to>
```
- `address`: Contract address
- `from`: Starting block number
- `to`: Ending block number

## 🏗️ Project Structure
```
evmql/
├── cmd/evmql/          # Entry point for the application
├── internal/           # Internal packages
│   ├── executor/       # Query execution logic
│   ├── parser/         # Query parsing
│   └── repl/          # Interactive shell
├── pkg/               # Public packages
│   ├── queries/       # Query implementations
│   └── utils/         # Utility functions
```

## 🤝 Contributing

Since this project is under active development, I welcome contributions! Here's how you can help:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Development Guidelines
- Write tests for new features
- Follow Go best practices and conventions
- Update documentation as needed
- Add comments for complex logic

## ⚙️ Configuration

Currently, the application requires:
1. EVM node endpoint (e.g., Infura URL)
2. (Optional) Block range limits

Configuration support is under development.

## 🔬 Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/parser
go test ./internal/executor
```

