package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Server represents the main server structure
type Server struct {
	agents    map[string]*Agent
	proxyList map[string]*Proxy
	mutex     sync.RWMutex
	upgrader  websocket.Upgrader
}

// Agent represents a connected agent
type Agent struct {
	ID       string
	Hostname string
	OS       string
	Conn     *websocket.Conn
	LastSeen time.Time
	CPU      float64
	Memory   float64
	Temp     float64
	ProxyID  string
}

// Proxy represents a proxy configuration
type Proxy struct {
	ID       string
	Address  string
	Port     string
	Username string
	Password string
	Status   string
	Agents   []string
}

// Stats represents system statistics
type Stats struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Temp   float64 `json:"temp"`
}

// AgentData represents data sent by agents
type AgentData struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Stats    Stats  `json:"stats"`
	ProxyID  string `json:"proxy_id"`
}

var server = &Server{
	agents:    make(map[string]*Agent),
	proxyList: make(map[string]*Proxy),
	upgrader: websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin for simplicity
		},
	},
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api/proxies", handleProxies)
	http.HandleFunc("/api/agents", handleAgents)
	
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	for {
		var data AgentData
		err := conn.ReadJSON(&data)
		if err != nil {
			log.Printf("Failed to read message: %v", err)
			server.removeAgent(data.ID)
			break
		}

		server.handleAgentData(data, conn)
	}
}

func (s *Server) handleAgentData(data AgentData, conn *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if agent already exists
	if agent, exists := s.agents[data.ID]; exists {
		// Update existing agent
		agent.Hostname = data.Hostname
		agent.OS = data.OS
		agent.Conn = conn
		agent.LastSeen = time.Now()
		agent.CPU = data.Stats.CPU
		agent.Memory = data.Stats.Memory
		agent.Temp = data.Stats.Temp
		agent.ProxyID = data.ProxyID
	} else {
		// Create new agent
		s.agents[data.ID] = &Agent{
			ID:       data.ID,
			Hostname: data.Hostname,
			OS:       data.OS,
			Conn:     conn,
			LastSeen: time.Now(),
			CPU:      data.Stats.CPU,
			Memory:   data.Stats.Memory,
			Temp:     data.Stats.Temp,
			ProxyID:  data.ProxyID,
		}
	}

	fmt.Printf("Agent %s connected from %s (OS: %s)\n", data.ID, data.Hostname, data.OS)
}

func (s *Server) removeAgent(agentID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if agent, exists := s.agents[agentID]; exists {
		if proxy, proxyExists := s.proxyList[agent.ProxyID]; proxyExists {
			// Remove agent from proxy's agent list
			newAgents := []string{}
			for _, id := range proxy.Agents {
				if id != agentID {
					newAgents = append(newAgents, id)
				}
			}
			proxy.Agents = newAgents
		}
		delete(s.agents, agentID)
	}
}

func handleProxies(w http.ResponseWriter, r *http.Request) {
	server.mutex.RLock()
	defer server.mutex.RUnlock()

	if r.Method == "GET" {
		proxies := make([]Proxy, 0, len(server.proxyList))
		for _, proxy := range server.proxyList {
			proxies = append(proxies, *proxy)
		}
		json.NewEncoder(w).Encode(proxies)
		return
	}

	if r.Method == "POST" {
		var proxy Proxy
		if err := json.NewDecoder(r.Body).Decode(&proxy); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		proxy.Status = "active"
		server.proxyList[proxy.ID] = &proxy
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(proxy)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleAgents(w http.ResponseWriter, r *http.Request) {
	server.mutex.RLock()
	defer server.mutex.RUnlock()

	agents := make([]Agent, 0, len(server.agents))
	for _, agent := range server.agents {
		agents = append(agents, *agent)
	}
	json.NewEncoder(w).Encode(agents)
}