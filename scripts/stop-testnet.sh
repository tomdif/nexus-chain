#!/bin/bash
# Stop all testnet nodes

set -e

BASE_DIR="$HOME/.nexus-testnet"
LOG_DIR="$BASE_DIR/logs"

echo "=============================================="
echo "  Stopping NEXUS Testnet"
echo "=============================================="
echo ""

# Stop each node
for i in 1 2 3; do
    PID_FILE="$LOG_DIR/node$i.pid"
    
    if [ -f "$PID_FILE" ]; then
        PID=$(cat $PID_FILE)
        if kill -0 $PID 2>/dev/null; then
            echo "Stopping node $i (PID: $PID)..."
            kill $PID
            rm $PID_FILE
            echo "  ✓ Node $i stopped"
        else
            echo "  ⚠ Node $i not running (stale PID file)"
            rm $PID_FILE
        fi
    else
        echo "  ⚠ Node $i PID file not found"
    fi
done

echo ""
echo "All nodes stopped."
echo ""
