#!/bin/bash

# NEXUS Chain Initialization Script
# Run this after building the binary with `make build`

set -e

CHAIN_ID="${CHAIN_ID:-nexus-testnet-1}"
MONIKER="${MONIKER:-myvalidator}"
BINARY="${BINARY:-./build/nexusd}"
HOME_DIR="${HOME_DIR:-$HOME/.nexus}"
DENOM="${DENOM:-unexus}"

echo "================================================"
echo "NEXUS Chain Initialization"
echo "================================================"
echo ""
echo "Chain ID:    $CHAIN_ID"
echo "Moniker:     $MONIKER"
echo "Home Dir:    $HOME_DIR"
echo "Denom:       $DENOM"
echo ""

# Remove existing data
echo "Cleaning previous data..."
rm -rf $HOME_DIR

# Initialize the chain
echo "Initializing chain..."
$BINARY init $MONIKER --chain-id $CHAIN_ID --home $HOME_DIR

# Create validator key
echo "Creating validator key..."
$BINARY keys add validator --keyring-backend test --home $HOME_DIR

# Get validator address
VALIDATOR_ADDR=$($BINARY keys show validator -a --keyring-backend test --home $HOME_DIR)
echo "Validator address: $VALIDATOR_ADDR"

# Add genesis account with tokens
echo "Adding genesis account..."
$BINARY genesis add-genesis-account $VALIDATOR_ADDR 1000000000000$DENOM --home $HOME_DIR

# Create genesis transaction
echo "Creating genesis transaction..."
# $BINARY genesis gentx validator 100000000000$DENOM \
    --chain-id $CHAIN_ID \
    --moniker $MONIKER \
    --commission-rate 0.1 \
    --commission-max-rate 0.2 \
    --commission-max-change-rate 0.01 \
    --min-self-delegation 1 \
    --keyring-backend test \
    --home $HOME_DIR

# Collect genesis transactions
echo "Collecting genesis transactions..."
# $BINARY genesis collect-gentxs --home $HOME_DIR

# Validate genesis
echo "Validating genesis..."
$BINARY genesis validate $HOME_DIR/config/genesis.json --home $HOME_DIR

# Update config for fast blocks (2 seconds)
echo "Configuring for 2-second blocks..."
sed -i 's/timeout_commit = "5s"/timeout_commit = "2s"/g' $HOME_DIR/config/config.toml
sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $HOME_DIR/config/config.toml

# Enable API
echo "Enabling API..."
sed -i 's/enable = false/enable = true/g' $HOME_DIR/config/app.toml
sed -i 's/swagger = false/swagger = true/g' $HOME_DIR/config/app.toml

# Set minimum gas price
echo "Setting minimum gas price..."
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001'$DENOM'"/g' $HOME_DIR/config/app.toml

echo ""
echo "================================================"
echo "Initialization complete!"
echo "================================================"
echo ""
echo "To start the chain:"
echo "  $BINARY start --home $HOME_DIR"
echo ""
echo "To check status:"
echo "  $BINARY status --home $HOME_DIR"
echo ""
echo "Validator address: $VALIDATOR_ADDR"
echo ""
