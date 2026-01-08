# Proxy Control System

This is a comprehensive proxy control system with persistent agent authentication and real-time monitoring dashboard.

## Features

- **Persistent Agent Authentication**: Agents register with the server and receive tokens that persist across restarts
- **Real-time Monitoring Dashboard**: Web interface showing connected agents with live updates
- **Automatic Status Management**: Agents marked as offline after 3 minutes without stats
- **Regular Stats Collection**: Agents send system statistics every 30 seconds
- **Web Interface Updates**: Dashboard refreshes every 10 seconds

## Architecture

### Server Components

- **Authentication Manager**: Handles agent registration and token-based authentication
- **WebSocket Endpoints**:
  - `/ws` - For web interface clients
  - `/agent/ws` - For agent connections
- **API Endpoints**:
  - `/api/agents` - Get list of all agents
  - `/api/agents/{id}` - Get specific agent
  - `/api/proxies` - Manage proxy configurations

### Agent Components

- **Automatic Registration**: Registers with server on first run
- **Token Persistence**: Saves authentication token to config file
- **Regular Stats Reporting**: Sends system and program stats every 30 seconds
- **Auto-Reconnection**: Automatically reconnects if connection drops

## How It Works

### Agent Authentication Process

1. Agent connects to `/agent/ws` endpoint
2. If agent has a saved token, tries to authenticate
3. If authentication fails or no token exists, registers as new agent
4. Server generates unique ID and token for the agent
5. Token is saved locally and used for future connections

### Data Flow

1. Agents collect system stats (CPU, memory, temperature) and program stats
2. Stats are sent to server every 30 seconds via WebSocket
3. Server updates agent status and broadcasts to web clients
4. Web interface receives updates in real-time through WebSocket
5. Server marks agents as offline after 3 minutes without stats

### Status Management

- Agents are considered **online** if they've sent stats within the last 3 minutes
- Agents are considered **offline** if no stats received for more than 3 minutes
- Web interface updates every 10 seconds to reflect current status

## Running the System

### Prerequisites

- Go 1.16 or higher

### Server Setup

```bash
cd proxy_control_system/server
go mod tidy
go run main.go
```

The server will start on port 8080 by default.

### Agent Setup

```bash
cd proxy_control_system/agent
go mod tidy
go run main.go
```

The agent will connect to the server and register itself.

### Web Interface

Open your browser and navigate to `http://localhost:8080` to access the dashboard.

## Configuration

The agent uses a `config.json` file to store its authentication token and settings. The file is automatically created after the first successful registration.

Example `config.json`:
```json
{
  "server_url": "ws://localhost:8080",
  "token": "your-auth-token-here",
  "name": "agent_hostname",
  "program": "firefox",
  "proxy": "proxy_1"
}
```

## Security

- Each agent gets a unique authentication token
- Tokens are generated server-side using cryptographically secure random generation
- No sensitive information is stored in plaintext
- Communication happens over WebSocket (can be upgraded to WSS for HTTPS)

## Customization

You can modify the agent to monitor specific programs or collect additional metrics by updating the `collectSystemStats()` and `collectProgramStats()` functions.