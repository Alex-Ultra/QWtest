#!/bin/bash

echo "Starting Proxy Control System..."

# Start the server in the background
echo "Starting server..."
cd /workspace/proxy_control_system/server
go run main.go &
SERVER_PID=$!

# Wait a moment for the server to start
sleep 2

# Start the agent in the background
echo "Starting agent..."
cd /workspace/proxy_control_system/agent
go run main.go &
AGENT_PID=$!

echo "Server PID: $SERVER_PID"
echo "Agent PID: $AGENT_PID"
echo "Server is available at http://localhost:8080"
echo "Press Ctrl+C to stop both processes"

# Wait for both processes
wait $SERVER_PID $AGENT_PID