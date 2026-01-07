@echo off
echo Building agent for Windows...
cd agent
go build -o ../bin/agent.exe main.go
echo Agent built successfully!
pause