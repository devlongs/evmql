#!/bin/bash

# This script initializes the EVMQL configuration with default settings

echo "Initializing EVMQL..."

# Create the config directory if it doesn't exist
mkdir -p ~/.evmql

# Build the application if it doesn't exist
if [ ! -f "./evmql" ]; then
    echo "Building EVMQL..."
    go mod tidy
    go build -o evmql cmd/evmql/main.go
    if [ $? -ne 0 ]; then
        echo "Failed to build EVMQL. Please check for errors above."
        exit 1
    fi
    echo "Build successful!"
fi

# Run the application with the generate-config flag
echo "Generating configuration..."
./evmql --generate-config

if [ -f "$HOME/.evmql/config.json" ]; then
    echo "Configuration successfully initialized!"
    echo "You can now edit ~/.evmql/config.json to customize your settings."
    echo "To use a specific network, you can set your Infura API key:"
    echo "export INFURA_API_KEY=your_key_here"
    echo ""
    echo "Start EVMQL with: ./evmql"
    echo "Or specify a network: ./evmql --network sepolia"
else
    echo "Configuration file wasn't created. Please check for errors above."
fi