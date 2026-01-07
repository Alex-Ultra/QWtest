# Proxy Control System

This system consists of a server and agents that monitor programs using proxies, along with system statistics collection.

## Components

### Server
- Runs on Windows
- Manages proxy configurations
- Receives data from agents
- Provides API endpoints to manage proxies and view agent status

### Agent
- Runs on both Windows and Linux
- Monitors programs that should be using proxies
- Collects system statistics (CPU, memory, temperature)
- Reports data to the server

## Architecture

The system uses WebSocket connections for real-time communication between agents and the server.

## Building

### For Windows
```bash
# Build server
build_server.bat

# Build agent
build_agent.bat
```

### For Linux
```bash
# Build server
./build_server.sh

# Build agent
./build_agent.sh
```

## Usage

### Running the Server
```bash
# On Windows
bin/server.exe

# On Linux
./bin/server
```

The server will start on port 8080 and provide the following endpoints:
- `/ws` - WebSocket endpoint for agent connections
- `/api/proxies` - REST API for proxy management (GET/POST)
- `/api/agents` - REST API to view connected agents (GET)

### Running the Agent
```bash
# On Windows
bin/agent.exe

# On Linux
./bin/agent
```

The agent will connect to the server and start reporting statistics.

## Configuration

The agent can be configured using `config.json`:

```json
{
  "server_url": "ws://localhost:8080/ws",
  "agent_id": "agent_1",
  "program": "firefox",
  "proxy_id": "proxy_1"
}
```

## Features

- Real-time monitoring of programs using proxies
- System statistics collection (CPU, memory, temperature)
- Cross-platform support (Windows and Linux agents)
- REST API for proxy management
- WebSocket-based communication