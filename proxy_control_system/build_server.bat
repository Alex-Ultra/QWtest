@echo off
echo Building server for Windows...
cd server
go build -o ../bin/server.exe main.go
echo Server built successfully!
pause