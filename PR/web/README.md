# Monitor Dashboard Web Interface

## Overview
This web interface provides a comprehensive dashboard for managing your mining farm. It allows you to monitor agent status, control mining operations, view metrics, and send commands to agents.

## Features

### Dashboard Overview
- Real-time display of online/offline agents count
- Total hashrate calculation
- Average temperature monitoring
- Quick status overview

### Agent Management
- Grid and table views of agents
- Detailed agent information
- Search and filter capabilities
- Status indicators with color coding

### Command Panel
- Bulk operations on all or selected agents
- Support for all agent commands:
  - Start Mining
  - Stop Mining
  - Restart Monitor
  - Update Config
  - Reboot PC

### Monitoring
- Real-time charts for hashrate and temperature
- Detailed metrics for each agent
- Auto-refresh with configurable intervals
- WebSocket-based real-time updates

### Proxy Control
- Start/Stop/Restart proxy functionality
- Proxy status monitoring
- Log viewing capabilities

## Installation and Setup

### Prerequisites
- Go (for the server application)
- A running server instance with the API endpoints

### Configuration
The web interface is designed to work with the existing server configuration. The default authentication token used is `web-interface-token` which should be defined in your server's `config.yaml` file.

### Integration
1. The web interface files should be placed in the `./dist/` directory relative to your server executable
2. The server is configured to serve static files from the `web_dist_dir` path specified in `config.yaml`
3. Make sure the `web_dist_dir` in your config points to where you place the web files

## How to Use

### 1. Starting the Server
Make sure your server is running and accessible. The web interface will connect to the same server instance.

### 2. Accessing the Dashboard
Open your browser and navigate to `http://<server-ip>:<port>` where your server is running.

### 3. Authentication
The web interface uses the default token `web-interface-token` defined in the server config. No additional authentication is required for the web interface itself.

### 4. Navigation
- **Dashboard**: Shows an overview of your farm status
- **Agent List**: View and manage individual agents
- **Charts**: Visual representation of metrics
- **Proxy Control**: Manage the proxy service

### 5. Managing Agents
- **Grid View**: Click on an agent card to see details
- **Table View**: Select multiple agents using checkboxes
- **Individual Actions**: Use the buttons on each agent card/table row
- **Bulk Actions**: Use the command panel at the top

### 6. Commands
- Select "Apply to all" or "Select specific agents"
- Choose a command from the dropdown
- Click "Execute" to send the command

### 7. Real-time Updates
The dashboard automatically refreshes data at intervals you can configure (default: 10 seconds). WebSocket connections provide real-time metrics updates.

## API Endpoints Used

The web interface communicates with the server using these API endpoints:

- `GET /api/v1/agents` - Get all agents data
- `GET /api/v1/agent/{id}` - Get specific agent data
- `POST /api/v1/agent/{id}/command` - Send command to agent
- `GET /api/v1/commands/queue` - Get command queue
- `GET /api/v1/proxy/status` - Get proxy status
- `POST /api/v1/proxy/start` - Start proxy
- `POST /api/v1/proxy/stop` - Stop proxy
- `GET /api/v1/download/monitor/{platform}` - Download monitor binary
- `WS /ws` - WebSocket connection for real-time updates

## Customization

### Changing the Authentication Token
If you want to use a different token, update the `AUTH_TOKEN` constant in `app.js` and ensure the same token exists in your server's `config.yaml`.

### UI Customization
- Modify `styles.css` to change the appearance
- Update `app.js` to modify functionality
- Edit `index.html` to change the layout

## Troubleshooting

### Connection Issues
- Ensure the server is running and accessible
- Check that the authentication token is correct
- Verify that CORS settings allow connections from your browser

### Data Not Updating
- Check server logs for API errors
- Verify WebSocket connections are working
- Ensure the auto-refresh interval is set correctly

### Commands Not Working
- Confirm that agents are online and connected
- Check server logs for command processing errors
- Verify that the agent token configuration is correct

## Security Notes

- The web interface communicates with the server API using tokens
- Ensure your server is properly secured with authentication
- Consider using HTTPS in production environments
- The default web interface token should be changed in production

## Support

For issues with the web interface, check:
- Browser console for JavaScript errors
- Server logs for API errors
- Network tab for failed requests