package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

type Agent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	IP          string    `json:"ip"`
	LastSeen    time.Time `json:"last_seen"`
	Status      string    `json:"status"` // online, offline
	Token       string    `json:"token"`
	SystemStats SystemStats `json:"system_stats"`
	ProgramStats ProgramStats `json:"program_stats"`
}

type SystemStats struct {
	Hostname   string    `json:"hostname"`
	CPUUsage   float64   `json:"cpu_usage"`
	MemoryUsage float64  `json:"memory_usage"`
	Temperature float64  `json:"temperature"`
	Timestamp  time.Time `json:"timestamp"`
}

type ProgramStats struct {
	ProgramName string            `json:"program_name"`
	Proxy       string            `json:"proxy"`
	ProcessInfo map[string]string `json:"process_info"`
	Timestamp   time.Time         `json:"timestamp"`
}

type AuthManager struct {
	agents map[string]*Agent
	mutex  sync.RWMutex
}

func NewAuthManager() *AuthManager {
	return &AuthManager{
		agents: make(map[string]*Agent),
	}
}

func (am *AuthManager) RegisterAgent(agentName, ip string) (*Agent, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Generate unique agent ID
	agentID := generateID()
	
	// Generate authentication token
	token := generateToken()

	agent := &Agent{
		ID:       agentID,
		Name:     agentName,
		IP:       ip,
		Token:    token,
		Status:   "online",
		LastSeen: time.Now(),
	}

	am.agents[agentID] = agent
	log.Printf("Registered new agent: %s (%s)", agentName, agentID)
	
	return agent, nil
}

func (am *AuthManager) Authenticate(token string) (*Agent, bool) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	for _, agent := range am.agents {
		if agent.Token == token {
			// Update last seen time
			agent.LastSeen = time.Now()
			agent.Status = "online"
			return agent, true
		}
	}
	return nil, false
}

func (am *AuthManager) UpdateAgentStats(agentID string, stats SystemStats, programStats ProgramStats) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	agent, exists := am.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	agent.SystemStats = stats
	agent.ProgramStats = programStats
	agent.LastSeen = time.Now()
	agent.Status = "online"

	return nil
}

func (am *AuthManager) GetAgents() []*Agent {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	// Update agent statuses based on last seen time
	var agents []*Agent
	for _, agent := range am.agents {
		if time.Since(agent.LastSeen) > 3*time.Minute {
			agent.Status = "offline"
		}
		// Create a copy to avoid race conditions
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}

	return agents
}

func (am *AuthManager) GetAgentByID(agentID string) (*Agent, bool) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	agent, exists := am.agents[agentID]
	if !exists {
		return nil, false
	}

	// Update status based on last seen time
	if time.Since(agent.LastSeen) > 3*time.Minute {
		agent.Status = "offline"
	}

	// Return a copy to avoid race conditions
	agentCopy := *agent
	return &agentCopy, true
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}