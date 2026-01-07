#!/bin/bash
echo "Building Proxy Control System for Linux..."
mkdir -p bin

echo "Building server..."
cd server
GOOS=linux go build -o ../bin/server main.go
if [ $? -ne 0 ]; then
    echo "Failed to build server"
    exit 1
fi
echo "Server built successfully!"

echo
echo "Building agent..."
cd ../agent
GOOS=linux go build -o ../bin/agent main.go
if [ $? -ne 0 ]; then
    echo "Failed to build agent"
    exit 1
fi
echo "Agent built successfully!"

echo
echo "All components built successfully!"
echo "Binaries are in the bin directory."