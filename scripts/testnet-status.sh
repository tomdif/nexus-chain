#!/bin/bash
# Check status of testnet nodes

BASE_DIR="$HOME/.nexus-testnet"
LOG_DIR="$BASE_DIR/logs"

echo "=============================================="
echo "  NEXUS Testnet Status"
echo "=============================================="
echo ""

# Check each node
for i in 1 2 3; do
    PID_FILE="$LOG_DIR/node$i.pid"
    RPC_PORT=$((26657 + (i-1)*10))
    
    echo "Node $i:"
    
    # Check if process is running
    if [ -f "$PID_FILE" ]; then
        PID=$(cat $PID_FILE)
        if kill -0 $PID 2>/dev/null; then
            echo "  Status: ✓ Running (PID: $PID)"
            
            # Try to get block height from RPC
            HEIGHT=$(curl -s http://127.0.0.1:$RPC_PORT/status 2>/dev/null | jq -r '.result.sync_info.latest_block_height // "N/A"')
            echo "  Height: $HEIGHT"
            
            # Get number of peers
            PEERS=$(curl -s http://127.0.0.1:$RPC_PORT/net_info 2>/dev/null | jq -r '.result.n_peers // "N/A"')
            echo "  Peers: $PEERS"
        else
            echo "  Status: ✗ Not running (stale PID)"
        fi
    else
        echo "  Status: ✗ Not running"
    fi
    
    echo "  RPC: http://127.0.0.1:$RPC_PORT"
    echo ""
done

echo "To view logs: tail -f $LOG_DIR/node1.log"
echo ""
