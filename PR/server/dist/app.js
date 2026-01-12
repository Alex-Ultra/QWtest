// Configuration
const API_BASE_URL = '/api/v1';
const AUTH_TOKEN = localStorage.getItem('auth_token') || 'web-interface-token'; // Default token from config
const WS_URL = `ws://${window.location.host}/ws?token=${AUTH_TOKEN}`;

// Global variables
let agents = [];
let autoRefreshInterval = null;
let wsConnection = null;
let hasWsConnection = false;

// Chart instances
let hashrateChart = null;
let temperatureChart = null;

// DOM Elements
const elements = {
    connectionStatus: document.getElementById('connection-status'),
    refreshBtn: document.getElementById('refresh-btn'),
    autoRefreshSelect: document.getElementById('auto-refresh-select'),
    onlineCount: document.getElementById('online-count'),
    offlineCount: document.getElementById('offline-count'),
    totalHashrate: document.getElementById('total-hashrate'),
    avgTemp: document.getElementById('avg-temp'),
    commandSelect: document.getElementById('command-select'),
    executeCommandBtn: document.getElementById('execute-command-btn'),
    agentGrid: document.getElementById('agent-grid'),
    agentTable: document.getElementById('agent-table'),
    agentTableBody: document.getElementById('agent-table-body'),
    gridViewBtn: document.getElementById('grid-view-btn'),
    tableViewBtn: document.getElementById('table-view-btn'),
    searchInput: document.getElementById('search-input'),
    proxyStatus: document.getElementById('proxy-status'),
    startProxyBtn: document.getElementById('start-proxy-btn'),
    stopProxyBtn: document.getElementById('stop-proxy-btn'),
    restartProxyBtn: document.getElementById('restart-proxy-btn'),
    viewLogsBtn: document.getElementById('view-logs-btn'),
    agentModal: document.getElementById('agent-modal'),
    closeModal: document.querySelector('.close'),
    agentDetails: document.getElementById('agent-details')
};

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    initializeApp();
});

function initializeApp() {
    setupEventListeners();
    loadInitialData();
    setupWebSocket();
    setupCharts();
    
    // Start auto-refresh if configured
    const savedInterval = localStorage.getItem('autoRefreshInterval');
    if (savedInterval) {
        elements.autoRefreshSelect.value = savedInterval;
        startAutoRefresh(parseInt(savedInterval));
    }
}

function setupEventListeners() {
    // Refresh button
    elements.refreshBtn.addEventListener('click', loadInitialData);
    
    // Auto-refresh selector
    elements.autoRefreshSelect.addEventListener('change', (e) => {
        const interval = parseInt(e.target.value);
        localStorage.setItem('autoRefreshInterval', interval);
        startAutoRefresh(interval);
    });
    
    // Command panel
    elements.commandSelect.addEventListener('change', updateCommandButtonState);
    elements.executeCommandBtn.addEventListener('click', executeBulkCommand);
    
    // View toggle buttons
    elements.gridViewBtn.addEventListener('click', () => switchView('grid'));
    elements.tableViewBtn.addEventListener('click', () => switchView('table'));
    
    // Search input
    elements.searchInput.addEventListener('input', filterAgents);
    
    // Proxy controls
    elements.startProxyBtn.addEventListener('click', () => sendProxyCommand('start'));
    elements.stopProxyBtn.addEventListener('click', () => sendProxyCommand('stop'));
    elements.restartProxyBtn.addEventListener('click', () => sendProxyCommand('restart'));
    elements.viewLogsBtn.addEventListener('click', viewProxyLogs);
    
    // Modal close
    elements.closeModal.addEventListener('click', () => {
        elements.agentModal.style.display = 'none';
    });
    
    // Close modal when clicking outside
    window.addEventListener('click', (e) => {
        if (e.target === elements.agentModal) {
            elements.agentModal.style.display = 'none';
        }
    });
}

function startAutoRefresh(seconds) {
    if (autoRefreshInterval) {
        clearInterval(autoRefreshInterval);
        autoRefreshInterval = null;
    }
    
    if (seconds > 0) {
        autoRefreshInterval = setInterval(loadInitialData, seconds * 1000);
    }
}

async function loadInitialData() {
    try {
        await Promise.all([
            loadAgents(),
            loadProxyStatus()
        ]);
        
        updateDashboardOverview();
        renderAgentList();
        updateCharts();
    } catch (error) {
        console.error('Error loading initial data:', error);
        showNotification(`Error loading data: ${error.message}`, 'error');
    }
}

async function loadAgents() {
    try {
        const response = await fetch(`${API_BASE_URL}/agents`, {
            headers: {
                'Authorization': `Bearer ${AUTH_TOKEN}`
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        agents = data;
        
        // Update connection status
        updateConnectionStatus(true);
        
        return data;
    } catch (error) {
        console.error('Error loading agents:', error);
        updateConnectionStatus(false);
        throw error;
    }
}

async function loadProxyStatus() {
    try {
        const response = await fetch(`${API_BASE_URL}/proxy/status`, {
            headers: {
                'Authorization': `Bearer ${AUTH_TOKEN}`
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const status = await response.json();
        elements.proxyStatus.textContent = status.running ? 'Running' : 'Stopped';
        elements.proxyStatus.style.color = status.running ? '#2ecc71' : '#e74c3c';
    } catch (error) {
        console.error('Error loading proxy status:', error);
        elements.proxyStatus.textContent = 'Error';
        elements.proxyStatus.style.color = '#e74c3c';
    }
}

function updateConnectionStatus(connected) {
    hasWsConnection = connected;
    elements.connectionStatus.className = connected ? 
        'status-indicator status-online' : 
        'status-indicator status-offline';
    elements.connectionStatus.textContent = connected ? 'Connected' : 'Disconnected';
}

function updateDashboardOverview() {
    const onlineAgents = agents.filter(agent => agent.status !== 'offline');
    const offlineAgents = agents.filter(agent => agent.status === 'offline');
    const totalHashrate = agents.reduce((sum, agent) => sum + (agent.hashrate || 0), 0);
    const avgTemp = agents.length > 0 ? 
        agents.reduce((sum, agent) => sum + (agent.temp_cpu || 0), 0) / agents.length : 0;
    
    elements.onlineCount.textContent = onlineAgents.length;
    elements.offlineCount.textContent = offlineAgents.length;
    elements.totalHashrate.textContent = `${totalHashrate.toFixed(2)} H/s`;
    elements.avgTemp.textContent = `${avgTemp.toFixed(1)}°C`;
}

function renderAgentList() {
    const viewMode = localStorage.getItem('viewMode') || 'grid';
    
    if (viewMode === 'grid') {
        renderAgentGrid();
    } else {
        renderAgentTable();
    }
}

function renderAgentGrid() {
    elements.agentGrid.innerHTML = '';
    
    agents.forEach(agent => {
        const card = createAgentCard(agent);
        elements.agentGrid.appendChild(card);
    });
}

function renderAgentTable() {
    elements.agentTableBody.innerHTML = '';
    
    agents.forEach(agent => {
        const row = createAgentTableRow(agent);
        elements.agentTableBody.appendChild(row);
    });
}

function createAgentCard(agent) {
    const card = document.createElement('div');
    card.className = 'agent-card';
    
    const statusClass = getStatusClass(agent.status);
    const lastSeen = new Date(agent.last_seen * 1000).toLocaleString();
    
    card.innerHTML = `
        <div class="agent-card-header">
            <div class="agent-id">${agent.id}</div>
            <div class="agent-status ${statusClass}">${agent.status}</div>
        </div>
        <div class="agent-platform">${agent.platform}</div>
        <div class="agent-metrics">
            <div class="metric">
                <div class="metric-label">Hashrate</div>
                <div class="metric-value">${(agent.hashrate || 0).toFixed(2)} H/s</div>
            </div>
            <div class="metric">
                <div class="metric-label">CPU Temp</div>
                <div class="metric-value">${agent.temp_cpu || 0}°C</div>
            </div>
            <div class="metric">
                <div class="metric-label">CPU Load</div>
                <div class="metric-value">${(agent.cpu_usage || 0).toFixed(1)}%</div>
            </div>
            <div class="metric">
                <div class="metric-label">RAM Usage</div>
                <div class="metric-value">${(agent.ram_usage || 0).toFixed(1)}%</div>
            </div>
            <div class="metric">
                <div class="metric-label">Version</div>
                <div class="metric-value">${agent.version}</div>
            </div>
            <div class="metric">
                <div class="metric-label">Last Seen</div>
                <div class="metric-value">${lastSeen}</div>
            </div>
        </div>
        <div class="agent-actions">
            <button class="btn btn-success" onclick="sendAgentCommand('${agent.id}', 'start_monitor')">Start</button>
            <button class="btn btn-warning" onclick="sendAgentCommand('${agent.id}', 'stop_monitor')">Stop</button>
            <button class="btn btn-info" onclick="sendAgentCommand('${agent.id}', 'restart_monitor')">Restart</button>
            <button class="btn btn-secondary" onclick="sendAgentCommand('${agent.id}', 'update_config')">Config</button>
            <button class="btn btn-danger" onclick="sendAgentCommand('${agent.id}', 'reboot')">Reboot</button>
            <button class="btn btn-primary" onclick="showAgentDetails('${agent.id}')">Details</button>
        </div>
    `;
    
    return card;
}

function createAgentTableRow(agent) {
    const row = document.createElement('tr');
    const statusClass = getStatusClass(agent.status);
    const lastSeen = new Date(agent.last_seen * 1000).toLocaleString();
    
    row.innerHTML = `
        <td><input type="checkbox" class="agent-checkbox" data-agent-id="${agent.id}"></td>
        <td>${agent.id}</td>
        <td><span class="agent-status ${statusClass}">${agent.status}</span></td>
        <td>${agent.platform}</td>
        <td>${(agent.hashrate || 0).toFixed(2)} H/s</td>
        <td>${agent.temp_cpu || 0}°C</td>
        <td>${(agent.cpu_usage || 0).toFixed(1)}%</td>
        <td>${(agent.ram_usage || 0).toFixed(1)}%</td>
        <td>${agent.version}</td>
        <td>${lastSeen}</td>
        <td>
            <button class="btn btn-success" onclick="sendAgentCommand('${agent.id}', 'start_monitor')">Start</button>
            <button class="btn btn-warning" onclick="sendAgentCommand('${agent.id}', 'stop_monitor')">Stop</button>
            <button class="btn btn-info" onclick="sendAgentCommand('${agent.id}', 'restart_monitor')">Restart</button>
        </td>
    `;
    
    return row;
}

function getStatusClass(status) {
    switch (status) {
        case 'mining':
            return 'status-mining';
        case 'stopped':
            return 'status-stopped';
        case 'offline':
            return 'status-offline';
        default:
            return 'status-offline';
    }
}

function switchView(mode) {
    localStorage.setItem('viewMode', mode);
    
    if (mode === 'grid') {
        elements.agentGrid.style.display = 'grid';
        elements.agentTable.style.display = 'none';
        elements.gridViewBtn.classList.add('active');
        elements.tableViewBtn.classList.remove('active');
        renderAgentGrid();
    } else {
        elements.agentGrid.style.display = 'none';
        elements.agentTable.style.display = 'table';
        elements.tableViewBtn.classList.add('active');
        elements.gridViewBtn.classList.remove('active');
        renderAgentTable();
    }
}

function filterAgents() {
    const searchTerm = elements.searchInput.value.toLowerCase();
    
    const filteredAgents = agents.filter(agent => 
        agent.id.toLowerCase().includes(searchTerm) ||
        agent.platform.toLowerCase().includes(searchTerm) ||
        agent.status.toLowerCase().includes(searchTerm) ||
        agent.version.toLowerCase().includes(searchTerm)
    );
    
    // Update the displayed list based on current view mode
    const viewMode = localStorage.getItem('viewMode') || 'grid';
    
    if (viewMode === 'grid') {
        elements.agentGrid.innerHTML = '';
        filteredAgents.forEach(agent => {
            const card = createAgentCard(agent);
            elements.agentGrid.appendChild(card);
        });
    } else {
        elements.agentTableBody.innerHTML = '';
        filteredAgents.forEach(agent => {
            const row = createAgentTableRow(agent);
            elements.agentTableBody.appendChild(row);
        });
    }
}

function updateCommandButtonState() {
    const selectedCommand = elements.commandSelect.value;
    elements.executeCommandBtn.disabled = !selectedCommand;
}

async function executeBulkCommand() {
    const command = elements.commandSelect.value;
    if (!command) return;
    
    const selectionMode = document.querySelector('input[name="selection-mode"]:checked').value;
    let targetAgents = [];
    
    if (selectionMode === 'all') {
        targetAgents = agents.map(agent => agent.id);
    } else {
        // Get selected agents from checkboxes
        const checkboxes = document.querySelectorAll('.agent-checkbox:checked');
        targetAgents = Array.from(checkboxes).map(cb => cb.dataset.agentId);
    }
    
    if (targetAgents.length === 0) {
        showNotification('No agents selected', 'warning');
        return;
    }
    
    // Confirm action
    const confirmed = confirm(`Are you sure you want to execute "${command}" on ${targetAgents.length} agent(s)?`);
    if (!confirmed) return;
    
    // Execute command for each agent
    const promises = targetAgents.map(agentId => 
        sendAgentCommandInternal(agentId, command)
    );
    
    try {
        await Promise.all(promises);
        showNotification(`Command "${command}" sent to ${targetAgents.length} agent(s)`, 'success');
        
        // Refresh data after a delay
        setTimeout(loadInitialData, 1000);
    } catch (error) {
        showNotification(`Error sending command: ${error.message}`, 'error');
    }
}

async function sendAgentCommand(agentId, command) {
    try {
        await sendAgentCommandInternal(agentId, command);
        showNotification(`Command "${command}" sent to agent ${agentId}`, 'success');
        
        // Refresh data after a delay
        setTimeout(loadInitialData, 1000);
    } catch (error) {
        showNotification(`Error sending command to agent ${agentId}: ${error.message}`, 'error');
    }
}

async function sendAgentCommandInternal(agentId, command) {
    const response = await fetch(`${API_BASE_URL}/agent/${agentId}/command`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${AUTH_TOKEN}`
        },
        body: JSON.stringify({
            action: command
        })
    });
    
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const result = await response.json();
    return result;
}

async function sendProxyCommand(action) {
    let url, successMsg;
    
    switch (action) {
        case 'start':
            url = `${API_BASE_URL}/proxy/start`;
            successMsg = 'Proxy started successfully';
            break;
        case 'stop':
            url = `${API_BASE_URL}/proxy/stop`;
            successMsg = 'Proxy stopped successfully';
            break;
        case 'restart':
            // For restart, we'll stop then start
            try {
                await sendProxyCommand('stop');
                setTimeout(() => sendProxyCommand('start'), 1000);
                return;
            } catch (error) {
                showNotification(`Error restarting proxy: ${error.message}`, 'error');
                return;
            }
        default:
            return;
    }
    
    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${AUTH_TOKEN}`
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const result = await response.json();
        showNotification(successMsg, 'success');
        
        // Refresh proxy status
        setTimeout(loadProxyStatus, 500);
    } catch (error) {
        showNotification(`Error ${action}ing proxy: ${error.message}`, 'error');
    }
}

function viewProxyLogs() {
    alert('Proxy logs feature would be implemented here. In a real implementation, this would fetch and display proxy logs.');
}

function showAgentDetails(agentId) {
    const agent = agents.find(a => a.id === agentId);
    if (!agent) return;
    
    const lastSeen = new Date(agent.last_seen * 1000).toLocaleString();
    
    elements.agentDetails.innerHTML = `
        <h3>Agent: ${agent.id}</h3>
        <div class="agent-detail-row">
            <strong>ID:</strong> ${agent.id}
        </div>
        <div class="agent-detail-row">
            <strong>IP:</strong> ${agent.ip}
        </div>
        <div class="agent-detail-row">
            <strong>Platform:</strong> ${agent.platform}
        </div>
        <div class="agent-detail-row">
            <strong>Status:</strong> <span class="agent-status ${getStatusClass(agent.status)}">${agent.status}</span>
        </div>
        <div class="agent-detail-row">
            <strong>Version:</strong> ${agent.version}
        </div>
        <div class="agent-detail-row">
            <strong>Hashrate:</strong> ${(agent.hashrate || 0).toFixed(2)} H/s
        </div>
        <div class="agent-detail-row">
            <strong>CPU Temperature:</strong> ${agent.temp_cpu || 0}°C
        </div>
        <div class="agent-detail-row">
            <strong>CPU Usage:</strong> ${(agent.cpu_usage || 0).toFixed(1)}%
        </div>
        <div class="agent-detail-row">
            <strong>RAM Usage:</strong> ${(agent.ram_usage || 0).toFixed(1)}%
        </div>
        <div class="agent-detail-row">
            <strong>Last Seen:</strong> ${lastSeen}
        </div>
        <div class="agent-actions" style="margin-top: 20px;">
            <button class="btn btn-success" onclick="sendAgentCommand('${agent.id}', 'start_monitor')">Start Mining</button>
            <button class="btn btn-warning" onclick="sendAgentCommand('${agent.id}', 'stop_monitor')">Stop Mining</button>
            <button class="btn btn-info" onclick="sendAgentCommand('${agent.id}', 'restart_monitor')">Restart Monitor</button>
            <button class="btn btn-secondary" onclick="sendAgentCommand('${agent.id}', 'update_config')">Update Config</button>
            <button class="btn btn-danger" onclick="sendAgentCommand('${agent.id}', 'reboot')">Reboot PC</button>
        </div>
    `;
    
    elements.agentModal.style.display = 'block';
}

function setupWebSocket() {
    try {
        wsConnection = new WebSocket(WS_URL);
        
        wsConnection.onopen = function(event) {
            console.log('WebSocket connected');
            updateConnectionStatus(true);
        };
        
        wsConnection.onmessage = function(event) {
            try {
                const message = JSON.parse(event.data);
                // Handle real-time updates
                if (message.type === 'metrics_update') {
                    // Update agent metrics in real-time
                    updateAgentMetrics(message.agentId, message.metrics);
                }
            } catch (e) {
                console.error('Error parsing WebSocket message:', e);
            }
        };
        
        wsConnection.onclose = function(event) {
            console.log('WebSocket disconnected');
            updateConnectionStatus(false);
            // Attempt to reconnect after a delay
            setTimeout(setupWebSocket, 5000);
        };
        
        wsConnection.onerror = function(error) {
            console.error('WebSocket error:', error);
            updateConnectionStatus(false);
        };
    } catch (error) {
        console.error('Error establishing WebSocket connection:', error);
        updateConnectionStatus(false);
    }
}

function updateAgentMetrics(agentId, metrics) {
    // Find and update the agent in our local array
    const agentIndex = agents.findIndex(a => a.id === agentId);
    if (agentIndex !== -1) {
        agents[agentIndex] = { ...agents[agentIndex], ...metrics };
        
        // Update dashboard overview
        updateDashboardOverview();
        
        // Re-render the agent list if needed
        renderAgentList();
        
        // Update charts
        updateCharts();
    }
}

function setupCharts() {
    const hashrateCtx = document.getElementById('hashrate-chart').getContext('2d');
    const tempCtx = document.getElementById('temperature-chart').getContext('2d');
    
    hashrateChart = new Chart(hashrateCtx, {
        type: 'bar',
        data: {
            labels: [],
            datasets: [{
                label: 'Hashrate (H/s)',
                data: [],
                backgroundColor: 'rgba(52, 152, 219, 0.6)',
                borderColor: 'rgba(52, 152, 219, 1)',
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
    
    temperatureChart = new Chart(tempCtx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Temperature (°C)',
                data: [],
                fill: false,
                borderColor: 'rgb(231, 76, 60)',
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
}

function updateCharts() {
    if (!hashrateChart || !temperatureChart) return;
    
    // Prepare data for hashrate chart
    const onlineAgents = agents.filter(agent => agent.status !== 'offline');
    const agentIds = onlineAgents.map(agent => agent.id.substring(0, 8) + '...');
    const hashrates = onlineAgents.map(agent => agent.hashrate || 0);
    
    hashrateChart.data.labels = agentIds;
    hashrateChart.data.datasets[0].data = hashrates;
    hashrateChart.update();
    
    // Prepare data for temperature chart
    const temps = onlineAgents.map(agent => agent.temp_cpu || 0);
    
    temperatureChart.data.labels = agentIds;
    temperatureChart.data.datasets[0].data = temps;
    temperatureChart.update();
}

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.innerHTML = `
        <i class="fas fa-${getNotificationIcon(type)}"></i>
        <span>${message}</span>
    `;
    
    document.getElementById('notifications').appendChild(notification);
    
    // Remove notification after 5 seconds
    setTimeout(() => {
        notification.remove();
    }, 5000);
}

function getNotificationIcon(type) {
    switch (type) {
        case 'success': return 'check-circle';
        case 'error': return 'exclamation-circle';
        case 'warning': return 'exclamation-triangle';
        default: return 'info-circle';
    }
}

// Expose functions to global scope for inline event handlers
window.sendAgentCommand = sendAgentCommand;
window.showAgentDetails = showAgentDetails;