#!/bin/bash

echo "Starting Proxy Control System Example"
echo "====================================="

# Start the server in the background
echo "Starting server..."
cd server
go run main.go &
SERVER_PID=$!
cd ..

# Wait a moment for the server to start
sleep 2

# Start an agent in the background
echo "Starting agent..."
cd agent
go run main.go &
AGENT_PID=$!
cd ..

echo "Server PID: $SERVER_PID"
echo "Agent PID: $AGENT_PID"

echo ""
echo "Server is running on http://localhost:8080"
echo "WebSocket endpoint: ws://localhost:8080/ws"
echo ""
echo "To test the API:"
echo "  - View all agents: curl http://localhost:8080/api/agents"
echo "  - View all proxies: curl http://localhost:8080/api/proxies"
echo ""
echo "Press Ctrl+C to stop the system"

# Wait for both processes
wait $SERVER_PID $AGENT_PID