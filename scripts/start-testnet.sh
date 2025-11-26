#!/bin/bash
# Start all 3 testnet nodes in the background with logging

set -e

BASE_DIR="$HOME/.nexus-testnet"
BINARY="./nexusd"
LOG_DIR="$BASE_DIR/logs"

# Check if testnet is initialized
if [ ! -d "$BASE_DIR/node1" ]; then
    echo "Error: Testnet not initialized. Run ./scripts/testnet.sh first"
    exit 1
fi

# Create log directory
mkdir -p $LOG_DIR

echo "=============================================="
echo "  Starting NEXUS Multi-Validator Testnet"
echo "=============================================="
echo ""

# Start each node in the background
for i in 1 2 3; do
    NODE_DIR="$BASE_DIR/node$i"
    LOG_FILE="$LOG_DIR/node$i.log"
    
    echo "Starting node $i..."
    $BINARY start --home $NODE_DIR > $LOG_FILE 2>&1 &
    NODE_PID=$!
    echo $NODE_PID > $LOG_DIR/node$i.pid
    echo "  ✓ Node $i started (PID: $NODE_PID)"
    echo "    Log: $LOG_FILE"
    
    # Give each node a moment to start
    sleep 2
done

echo ""
echo "=============================================="
echo "  All nodes started!"
echo "=============================================="
echo ""
echo "RPC Endpoints:"
echo "  Node 1: http://127.0.0.1:26657"
echo "  Node 2: http://127.0.0.1:26667"
echo "  Node 3: http://127.0.0.1:26677"
echo ""
echo "To view logs:"
echo "  tail -f $LOG_DIR/node1.log"
echo "  tail -f $LOG_DIR/node2.log"
echo "  tail -f $LOG_DIR/node3.log"
echo ""
echo "To stop the testnet:"
echo "  ./scripts/stop-testnet.sh"
echo ""
