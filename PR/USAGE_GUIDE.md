# Monitor Dashboard System Usage Guide

## Overview
This document provides comprehensive instructions on how to use the Monitor Dashboard system, which includes a server application, agent applications, and a web interface for centralized management of your mining farm.

## Project Structure
```
PR/
├── agent/                 # Agent application source code
│   ├── cmd/agent/        # Agent main entry point
│   ├── configs/          # Agent configuration files
│   ├── internal/         # Internal modules
│   └── go.mod/go.sum     # Go module dependencies
├── server/               # Server application source code
│   ├── cmd/server/       # Server main entry point
│   ├── configs/          # Server configuration files
│   ├── internal/         # Internal modules
│   └── go.mod/go.sum     # Go module dependencies
├── web/                  # Web interface source files
├── dist/                 # Web interface distribution files
├── build_web.sh          # Build script
├── monitor-dashboard-project.zip  # Complete project archive
└── web-interface.zip     # Web interface only archive
```

## Getting Started

### 1. Prerequisites
- Go 1.19+ installed on both server and agent machines
- Git (optional, for cloning the repository)

### 2. Installation

#### Option A: Using the ZIP Archive
1. Extract `monitor-dashboard-project.zip` to your desired location
2. Navigate to the extracted directory

#### Option B: Building from Source
```bash
# Clone the repository (if available)
cd PR

# Build the server
cd server
go build ./cmd/server

# Build the agent
cd ../agent
go build ./cmd/agent
```

### 3. Configuration

#### Server Configuration
Edit `server/configs/config.yaml`:

```yaml
server:
  port: "8080"              # Port for the server API and web interface
  host: "0.0.0.0"           # Host to bind to

tokens:
  "abc123": "agent-1"       # Token for agent-1
  "def456": "agent-2"       # Token for agent-2
  "xyz789": "agent-3"       # Token for agent-3
  "web-interface-token": "web-interface"  # Token for web interface

paths:
  monitor_bin_dir: "./assets/bins/monitor/"  # Directory for monitor binaries
  proxy_bin_path: "./assets/bins/proxy/xmrig-proxy"  # Path to proxy binary
  proxy_config_path: "./configs/proxy-config.json"   # Path to proxy config
  web_dist_dir: "./dist/"   # Directory for web interface files

metrics_interval: 30        # Interval for metrics collection (seconds)
```

#### Agent Configuration
Edit `agent/configs/agent-config.yaml`:

```yaml
server_url: "ws://your-server-ip:8080/ws"  # Server WebSocket URL
token: "your-agent-token"                   # Agent-specific token
metrics_interval: 30                        # Metrics reporting interval
```

### 4. Setting Up the Web Interface

The web interface is already built and located in the `dist/` directory. The server is configured to serve these files automatically from the path specified in `web_dist_dir` in the server config.

Make sure the `dist/` directory contains:
- `index.html`
- `styles.css`
- `app.js`

### 5. Running the System

#### Step 1: Start the Server
```bash
cd server
./server  # On Linux/Mac
# OR
server.exe  # On Windows
```

The server will start on the configured port (default: 8080).

#### Step 2: Configure and Run Agents
1. For each agent machine:
   - Copy the agent binary and config to the machine
   - Update the agent config with the correct server URL and token
   - Run the agent

```bash
cd agent
./agent  # On Linux/Mac
# OR
agent.exe  # On Windows
```

#### Step 3: Access the Web Interface
Open your browser and navigate to `http://your-server-ip:8080`

## Web Interface Features

### Dashboard Overview
- **Online Agents**: Count of currently connected agents
- **Offline Agents**: Count of disconnected agents
- **Total Hashrate**: Sum of all agents' hash rates
- **Average Temperature**: Average CPU temperature across all agents

### Agent Management
- **Grid View**: Visual cards showing each agent's status and metrics
- **Table View**: Tabular format for easier comparison
- **Search**: Filter agents by ID, platform, status, or version
- **Detailed View**: Click on any agent to see detailed information

### Command Panel
- **Bulk Operations**: Send commands to all agents or selected ones
- **Supported Commands**:
  - Start Mining
  - Stop Mining
  - Restart Monitor
  - Update Config
  - Reboot PC

### Real-time Monitoring
- **Charts**: Live updating graphs for hashrate and temperature
- **Auto-refresh**: Configurable intervals (5s, 10s, 30s, 60s, or off)
- **WebSocket Updates**: Real-time metrics without page refresh

### Proxy Control
- **Start/Stop/Restart**: Control the proxy service
- **Status Monitoring**: Real-time proxy status
- **Log Viewing**: Access to proxy logs

## API Endpoints

The system exposes the following REST API endpoints:

- `GET /api/v1/agents` - Get all agents data
- `GET /api/v1/agent/{id}` - Get specific agent data
- `POST /api/v1/agent/{id}/command` - Send command to agent
- `GET /api/v1/commands/queue` - Get command queue
- `GET /api/v1/proxy/status` - Get proxy status
- `POST /api/v1/proxy/start` - Start proxy
- `POST /api/v1/proxy/stop` - Stop proxy
- `GET /api/v1/download/monitor/{platform}` - Download monitor binary
- `WS /ws` - WebSocket connection for real-time updates

## Security

- All API endpoints (except downloads) require authentication via Bearer tokens
- Web interface uses the special `web-interface-token` for access
- Agent connections are authenticated using individual tokens
- Change default tokens in production environments

## Troubleshooting

### Common Issues

1. **Cannot access web interface**
   - Check if the server is running
   - Verify the server port is accessible
   - Ensure the `web_dist_dir` path is correct in config

2. **Agents not connecting**
   - Verify agent tokens match server config
   - Check network connectivity to server
   - Ensure WebSocket connections aren't blocked

3. **Commands not executing**
   - Confirm agents are online and connected
   - Check server logs for errors
   - Verify agent permissions for requested actions

4. **Charts not updating**
   - Check WebSocket connection status
   - Verify metrics_interval settings
   - Ensure agents are sending metrics

### Logs
Server logs will show connection attempts, command executions, and errors. Check the server console/output for troubleshooting information.

## Production Deployment

For production use:
1. Use HTTPS instead of HTTP
2. Secure tokens and rotate them regularly
3. Set up proper logging and monitoring
4. Configure firewalls appropriately
5. Consider load balancing for multiple server instances
6. Implement backup strategies for configurations

## Support

For technical support:
- Check server logs for error messages
- Verify network connectivity between components
- Review configuration files for correctness
- Consult the web interface console for client-side errors