#!/bin/bash

# This script initializes the EVMQL configuration with default settings

echo "Initializing EVMQL..."

# Create the config directory if it doesn't exist
mkdir -p ~/.evmql

# Run the application with the generate-config flag
evmql --generate-config

echo "Configuration initialized."
echo "You can now edit ~/.evmql/config.json to customize your settings."
echo "To use a specific network, you can set your Infura API key:"
echo "export INFURA_API_KEY=your_key_here"
echo ""
echo "Start EVMQL with: evmql"
echo "Or specify a network: evmql --network sepolia"