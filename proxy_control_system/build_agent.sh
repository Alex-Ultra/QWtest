#!/bin/bash
echo "Building agent for Linux..."
cd agent
GOOS=linux go build -o ../bin/agent main.go
echo "Agent built successfully!"