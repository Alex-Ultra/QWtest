# Proxy Control System

A comprehensive system for monitoring and controlling proxies through agents that run on Windows and Linux systems.

## Features

- **Persistent Agent Authentication**: Agents register with the server and receive persistent tokens that survive restarts
- **Real-time Web Dashboard**: Shows connected agents with live updates every 10 seconds
- **System Monitoring**: Collects CPU, memory, temperature, and hostname data every 30 seconds
- **Automatic Status Management**: Agents marked as offline after 3 minutes without stats
- **Cross-platform Support**: Server runs on Windows, agents work on both Windows and Linux
- **WebSocket Communication**: Real-time bidirectional communication between agents and server
- **REST API**: For proxy management and agent monitoring

## Architecture

### Server (Windows)
- Located in `/server/`
- Handles agent registration and authentication
- Provides REST API for proxy management
- Serves real-time web dashboard
- Manages persistent agent tokens

### Agent (Windows & Linux)
- Located in `/agent/`
- Registers with server and receives persistent token
- Monitors local system and proxy usage
- Sends statistics to server every 30 seconds
- Cross-platform compatible

## Building

### Windows
```batch
build_all.bat
```
Binaries will be placed in the `bin` directory.

### Linux
```bash
chmod +x build_all.sh
./build_all.sh
```
Binaries will be placed in the `bin` directory.

## Components Built

- `server` (or `server.exe`): Server executable for Windows
- `agent_linux`: Agent executable for Linux
- `agent_windows.exe`: Agent executable for Windows

## Running

### Server
```bash
./bin/server
```
The server will start on port 8080 by default. Access the web dashboard at `http://localhost:8080`

### Agent
```bash
./agent --server-url ws://localhost:8080/agent/ws --agent-name "MyAgent"
```

## Web Dashboard

The web dashboard provides real-time monitoring of all connected agents:
- Shows agent status (online/offline)
- Displays system statistics (CPU, memory, temperature)
- Shows program and proxy information
- Updates every 10 seconds
- Lists all connected agents with their details

## API Endpoints

- `GET /api/proxies` - Get all proxies
- `POST /api/proxies` - Add a proxy
- `DELETE /api/proxies/{id}` - Delete a proxy
- `GET /api/agents` - Get all agents
- `GET /api/agents/{id}` - Get specific agent
- `GET /ws` - WebSocket for web interface
- `GET /agent/ws` - WebSocket for agents

## Authentication Flow

1. New agents connect and send a "register" message with their name
2. Server registers the agent, assigns a unique ID and token
3. Server sends back the token to the agent
4. Agent stores the token persistently
5. On subsequent connections, agent sends "authenticate" message with token
6. Server validates token and allows connection
7. If agent reconnects after restart, it uses the stored token