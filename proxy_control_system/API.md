# API Documentation

## Server Endpoints

### WebSocket Connection
- **Endpoint**: `/ws`
- **Method**: WebSocket upgrade
- **Description**: Main communication channel for agents to connect and send data

### Proxies Management
- **Endpoint**: `/api/proxies`
- **Methods**:
  - `GET` - Retrieve all proxy configurations
  - `POST` - Create a new proxy configuration
- **GET Response**:
```json
[
  {
    "ID": "proxy_1",
    "Address": "proxy.example.com",
    "Port": "8080",
    "Username": "user1",
    "Password": "pass1",
    "Status": "active",
    "Agents": ["agent_1", "agent_2"]
  }
]
```

- **POST Request**:
```json
{
  "ID": "proxy_1",
  "Address": "proxy.example.com",
  "Port": "8080",
  "Username": "user1",
  "Password": "pass1"
}
```

- **POST Response**:
```json
{
  "ID": "proxy_1",
  "Address": "proxy.example.com",
  "Port": "8080",
  "Username": "user1",
  "Password": "pass1",
  "Status": "active",
  "Agents": []
}
```

### Agents Status
- **Endpoint**: `/api/agents`
- **Method**: `GET`
- **Description**: Retrieve status of all connected agents
- **Response**:
```json
[
  {
    "ID": "agent_1",
    "Hostname": "workstation-1",
    "OS": "windows",
    "LastSeen": "2023-01-01T12:00:00Z",
    "CPU": 25.5,
    "Memory": 60.2,
    "Temp": 45.0,
    "ProxyID": "proxy_1"
  }
]
```

## Agent Data Format

Agents send the following data structure to the server via WebSocket:

```json
{
  "ID": "agent_1",
  "Hostname": "workstation-1",
  "OS": "windows",
  "Stats": {
    "CPU": 25.5,
    "Memory": 60.2,
    "Temp": 45.0
  },
  "ProxyID": "proxy_1"
}
```

## Agent Configuration

Agents are configured using a `config.json` file with the following structure:

```json
{
  "server_url": "ws://localhost:8080/ws",
  "agent_id": "agent_1",
  "program": "firefox",
  "proxy_id": "proxy_1"
}
```

## Error Handling

- Standard HTTP status codes are used
- 200 OK for successful requests
- 400 Bad Request for invalid input
- 404 Not Found for missing resources
- 405 Method Not Allowed for unsupported methods
- 500 Internal Server Error for server-side issues