#!/bin/bash
echo "Building Proxy Control System..."
mkdir -p bin

echo "Building server..."
cd server
GOOS=linux go build -o ../bin/server main.go auth.go
if [ $? -ne 0 ]; then
    echo "Failed to build server"
    exit 1
fi
echo "Server built successfully!"

echo
echo "Building Linux agent..."
cd ../agent
GOOS=linux go build -o ../bin/agent_linux .
if [ $? -ne 0 ]; then
    echo "Failed to build Linux agent"
    exit 1
fi
echo "Linux agent built successfully!"

echo
echo "Building Windows agent..."
GOOS=windows go build -o ../bin/agent_windows.exe .
if [ $? -ne 0 ]; then
    echo "Failed to build Windows agent"
    exit 1
fi
echo "Windows agent built successfully!"

echo
echo "All components built successfully!"
echo "Binaries are in the bin directory."