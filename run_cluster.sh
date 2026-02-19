#!/bin/bash
trap "kill 0" EXIT

echo "Starting cluster..."

# Create data directories
mkdir -p data/node1 data/node2 data/node3

# Start Node 1
echo "Starting Node 1 on port 8001"
PORT=8001 NODE_ID=node1 PEERS=localhost:8002,localhost:8003 DATA_DIR=./data/node1 go run cmd/node/main.go &

# Start Node 2
echo "Starting Node 2 on port 8002"
PORT=8002 NODE_ID=node2 PEERS=localhost:8001,localhost:8003 DATA_DIR=./data/node2 go run cmd/node/main.go &

# Start Node 3
echo "Starting Node 3 on port 8003"
PORT=8003 NODE_ID=node3 PEERS=localhost:8001,localhost:8002 DATA_DIR=./data/node3 go run cmd/node/main.go &

echo "Cluster started. Press Ctrl+C to stop."
wait
