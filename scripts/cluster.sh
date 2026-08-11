#!/bin/bash
# Bash script to start a 3-node blockchain cluster
# Usage: ./scripts/cluster.sh

set -e

echo "========================================"
echo "  Starting Blockchain Cluster (3 nodes)"
echo "========================================"

# Build the node binary
echo ""
echo "Building node binary..."
go build -o node ./cmd/node/
echo "Build successful!"

# Cleanup function
cleanup() {
    echo ""
    echo "Stopping cluster..."
    kill $NODE1_PID $NODE2_PID $NODE3_PID 2>/dev/null
    wait $NODE1_PID $NODE2_PID $NODE3_PID 2>/dev/null
    echo "Cluster stopped."
    exit 0
}

trap cleanup SIGINT SIGTERM

# Start Node 1
echo ""
echo "Starting Node 1 on :3001..."
./node --addr localhost:3001 --peers "localhost:3002,localhost:3003" --difficulty 4 --mine > node1.log 2>&1 &
NODE1_PID=$!

sleep 1

# Start Node 2
echo "Starting Node 2 on :3002..."
./node --addr localhost:3002 --peers "localhost:3001,localhost:3003" --difficulty 4 --mine > node2.log 2>&1 &
NODE2_PID=$!

sleep 1

# Start Node 3
echo "Starting Node 3 on :3003..."
./node --addr localhost:3003 --peers "localhost:3001,localhost:3002" --difficulty 4 --mine > node3.log 2>&1 &
NODE3_PID=$!

echo ""
echo "========================================"
echo "  Cluster is running!"
echo "========================================"
echo ""
echo "Nodes:"
echo "  Node 1: http://localhost:3001/status"
echo "  Node 2: http://localhost:3002/status"
echo "  Node 3: http://localhost:3003/status"
echo ""
echo "Useful commands:"
echo "  curl http://localhost:3001/status     # Node status"
echo "  curl http://localhost:3001/chain       # Full chain"
echo "  curl http://localhost:3001/balances    # Account balances"
echo "  curl http://localhost:3001/mempool     # Pending pool"
echo "  curl http://localhost:3001/peers       # Known peers"
echo ""
echo "Press Ctrl+C to stop the cluster..."

# Wait
wait
