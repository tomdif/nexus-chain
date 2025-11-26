#!/bin/bash
# Set up a 3-node local testnet for NEXUS

set -e

CHAIN_ID="nexus-localnet"
BASE_DIR="$HOME/.nexus-testnet"
BINARY="./nexusd"

# Clean up previous testnet
rm -rf $BASE_DIR
mkdir -p $BASE_DIR

echo "=== Setting up 3-node NEXUS testnet ==="

# Initialize 3 nodes
for i in 1 2 3; do
    NODE_DIR="$BASE_DIR/node$i"
    $BINARY init "validator-$i" --home $NODE_DIR
    echo "Initialized node $i at $NODE_DIR"
done

# Update chain-id in all genesis files
for i in 1 2 3; do
    NODE_DIR="$BASE_DIR/node$i"
    sed -i "s/nexus-testnet-1/$CHAIN_ID/g" $NODE_DIR/config/genesis.json
done

# Copy genesis from node1 to all nodes (use node1's genesis as base)
cp $BASE_DIR/node1/config/genesis.json $BASE_DIR/node2/config/genesis.json
cp $BASE_DIR/node1/config/genesis.json $BASE_DIR/node3/config/genesis.json

# Configure P2P ports
# Node 1: 26656, Node 2: 26666, Node 3: 26676
# RPC ports: 26657, 26667, 26677

for i in 1 2 3; do
    NODE_DIR="$BASE_DIR/node$i"
    CONFIG="$NODE_DIR/config/config.toml"
    
    # Calculate ports
    P2P_PORT=$((26656 + (i-1)*10))
    RPC_PORT=$((26657 + (i-1)*10))
    
    # Update config.toml with sed
    sed -i "s/laddr = \"tcp:\/\/127.0.0.1:26657\"/laddr = \"tcp:\/\/127.0.0.1:$RPC_PORT\"/" $CONFIG
    sed -i "s/laddr = \"tcp:\/\/0.0.0.0:26656\"/laddr = \"tcp:\/\/0.0.0.0:$P2P_PORT\"/" $CONFIG
    
    echo "Node $i: P2P=$P2P_PORT, RPC=$RPC_PORT"
done

echo ""
echo "=== Testnet Configuration Complete ==="
echo ""
echo "To start the testnet, run these in separate terminals:"
echo ""
echo "Terminal 1: $BINARY start --home $BASE_DIR/node1"
echo "Terminal 2: $BINARY start --home $BASE_DIR/node2"  
echo "Terminal 3: $BINARY start --home $BASE_DIR/node3"
echo ""
echo "Node directories:"
echo "  Node 1: $BASE_DIR/node1 (RPC: 26657)"
echo "  Node 2: $BASE_DIR/node2 (RPC: 26667)"
echo "  Node 3: $BASE_DIR/node3 (RPC: 26677)"
