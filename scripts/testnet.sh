#!/bin/bash
# Set up a 3-node multi-validator testnet for NEXUS
# Note: This creates a network where node1 is the initial validator
# Additional validators can join via staking after launch
set -e

CHAIN_ID="nexus-localnet"
BASE_DIR="$HOME/.nexus-testnet"
BINARY="./nexusd"

rm -rf $BASE_DIR
mkdir -p $BASE_DIR/logs

echo "=============================================="
echo "  NEXUS Multi-Node Testnet Setup"
echo "=============================================="
echo ""

# Step 1: Initialize node1 (this will be the initial validator)
echo "Step 1: Initializing primary validator node..."
$BINARY init "validator-1" --home $BASE_DIR/node1 > /dev/null 2>&1
echo "  ✓ Node 1 initialized (primary validator)"

# Step 2: Initialize nodes 2 and 3 as non-validator nodes
echo ""
echo "Step 2: Initializing additional nodes..."
for i in 2 3; do
    NODE_DIR="$BASE_DIR/node$i"
    mkdir -p $NODE_DIR/config $NODE_DIR/data
    
    # Copy genesis from node1
    cp $BASE_DIR/node1/config/genesis.json $NODE_DIR/config/genesis.json
    
    # Generate node key
    $BINARY tendermint show-node-id --home $BASE_DIR/node1 > /dev/null 2>&1 || true
    
    # Create a fresh priv_validator_key and node_key for this node
    # We'll use a temporary init and copy the keys
    TEMP_DIR=$(mktemp -d)
    $BINARY init "temp" --home $TEMP_DIR > /dev/null 2>&1
    cp $TEMP_DIR/config/priv_validator_key.json $NODE_DIR/config/
    cp $TEMP_DIR/config/node_key.json $NODE_DIR/config/
    rm -rf $TEMP_DIR
    
    # Create empty priv_validator_state
    echo '{"height": "0", "round": 0, "step": 0}' > $NODE_DIR/data/priv_validator_state.json
    
    # Copy config.toml from node1 as base
    cp $BASE_DIR/node1/config/config.toml $NODE_DIR/config/config.toml
    
    echo "  ✓ Node $i initialized (full node)"
done

# Update chain ID in genesis
jq --arg chain_id "$CHAIN_ID" '.chain_id = $chain_id' $BASE_DIR/node1/config/genesis.json > $BASE_DIR/node1/config/genesis.tmp
mv $BASE_DIR/node1/config/genesis.tmp $BASE_DIR/node1/config/genesis.json

# Copy updated genesis to other nodes
cp $BASE_DIR/node1/config/genesis.json $BASE_DIR/node2/config/genesis.json
cp $BASE_DIR/node1/config/genesis.json $BASE_DIR/node3/config/genesis.json

# Step 3: Configure P2P networking
echo ""
echo "Step 3: Configuring P2P networking..."

NODE1_ID=$(cat $BASE_DIR/node1/config/node_key.json | jq -r '.id')
NODE2_ID=$(cat $BASE_DIR/node2/config/node_key.json | jq -r '.id')
NODE3_ID=$(cat $BASE_DIR/node3/config/node_key.json | jq -r '.id')

for i in 1 2 3; do
    NODE_DIR="$BASE_DIR/node$i"
    CONFIG="$NODE_DIR/config/config.toml"
    
    P2P_PORT=$((26656 + (i-1)*10))
    RPC_PORT=$((26657 + (i-1)*10))
    
    # Update P2P port (line ~206)
    sed -i.bak "s|laddr = \"tcp://0.0.0.0:26656\"|laddr = \"tcp://0.0.0.0:$P2P_PORT\"|g" $CONFIG
    
    # Update RPC port (line ~90) - be more specific to avoid matching other laddr lines
    sed -i.bak "s|laddr = \"tcp://127.0.0.1:26657\"|laddr = \"tcp://127.0.0.1:$RPC_PORT\"|g" $CONFIG
    
    # Verify the changes
    RPC_CHECK=$(grep "^laddr = \"tcp://127.0.0.1:" $CONFIG | head -1)
    P2P_CHECK=$(grep "^laddr = \"tcp://0.0.0.0:" $CONFIG | head -1)
    
    # Build peer list (connect to all other nodes)
    PEERS=""
    [ $i -ne 1 ] && PEERS="${PEERS}${NODE1_ID}@127.0.0.1:26656,"
    [ $i -ne 2 ] && PEERS="${PEERS}${NODE2_ID}@127.0.0.1:26666,"
    [ $i -ne 3 ] && PEERS="${PEERS}${NODE3_ID}@127.0.0.1:26676,"
    PEERS=${PEERS%,}
    
    sed -i.bak "s|persistent_peers = \"\"|persistent_peers = \"$PEERS\"|g" $CONFIG
    
    echo "  ✓ Node $i: P2P=$P2P_PORT, RPC=$RPC_PORT"
    echo "    RPC config: $RPC_CHECK"
    echo "    P2P config: $P2P_CHECK"
done

echo ""
echo "=============================================="
echo "  NEXUS Testnet Ready!"
echo "=============================================="
echo ""
echo "Network Configuration:"
echo "  Node 1: Validator (produces blocks)"
echo "  Node 2: Full node (syncs blocks)"
echo "  Node 3: Full node (syncs blocks)"
echo ""
echo "Start: ./scripts/start-testnet.sh"
echo "Stop:  ./scripts/stop-testnet.sh"
