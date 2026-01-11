package agents

import (
	"sync"
	"time"
	"monitor-server/internal/models"
)

type Manager struct {
	agents map[string]*models.Agent
	mutex  sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		agents: make(map[string]*models.Agent),
	}
}

func (am *Manager) RegisterAgent(agent *models.Agent) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	agent.LastSeen = time.Now().Unix()
	am.agents[agent.ID] = agent
}

func (am *Manager) UpdateAgentMetrics(agentID string, metrics *models.Metrics) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	if agent, exists := am.agents[agentID]; exists {
		agent.Hashrate = metrics.Hashrate
		agent.CPUUsage = metrics.CPUUsage
		agent.RAMUsage = metrics.RAMUsage
		agent.TempCPU = metrics.TempCPU
		agent.Status = metrics.Status
		agent.LastSeen = time.Now().Unix()
	}
}

func (am *Manager) GetAgent(agentID string) (*models.Agent, bool) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	agent, exists := am.agents[agentID]
	return agent, exists
}

func (am *Manager) GetAllAgents() []*models.Agent {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	agents := make([]*models.Agent, 0, len(am.agents))
	for _, agent := range am.agents {
		// Create a copy to avoid race conditions
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}
	return agents
}

func (am *Manager) RemoveAgent(agentID string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	delete(am.agents, agentID)
}