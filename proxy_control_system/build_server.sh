#!/bin/bash
echo "Building server for Linux..."
cd server
GOOS=linux go build -o ../bin/server main.go
echo "Server built successfully!"