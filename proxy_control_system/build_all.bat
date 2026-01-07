@echo off
echo Building Proxy Control System for Windows...
mkdir bin 2>nul

echo Building server...
cd server
go build -o ../bin/server.exe main.go
if %errorlevel% neq 0 (
    echo Failed to build server
    pause
    exit /b %errorlevel%
)
echo Server built successfully!

echo.
echo Building agent...
cd ../agent
go build -o ../bin/agent.exe main.go
if %errorlevel% neq 0 (
    echo Failed to build agent
    pause
    exit /b %errorlevel%
)
echo Agent built successfully!

echo.
echo All components built successfully!
echo Binaries are in the bin directory.
pause