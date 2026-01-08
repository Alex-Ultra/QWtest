package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Session struct {
	Username  string
	CreatedAt time.Time
}

var (
	proxies = make(map[string]string)
	authManager = NewAuthManager()
	broadcast = make(chan Agent, 10) // Channel to broadcast agent updates
	clients = make(map[*websocket.Conn]bool) // Connected web clients
	agentConnections = make(map[string]*websocket.Conn) // Agent connections
	sessions = make(map[string]*Session) // Store active sessions
	sessionMutex = sync.RWMutex{} // Mutex for session map
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	router := mux.NewRouter()
	
	// API routes
	router.HandleFunc("/api/login", loginHandler).Methods("POST")
	router.HandleFunc("/api/logout", logoutHandler).Methods("POST")
	router.HandleFunc("/api/proxies", getProxies).Methods("GET")
	router.HandleFunc("/api/proxies", addProxy).Methods("POST")
	router.HandleFunc("/api/proxies/{id}", deleteProxy).Methods("DELETE")
	router.HandleFunc("/api/agents", getAgents).Methods("GET")
	router.HandleFunc("/api/agents/{id}", getAgent).Methods("GET")
	
	// WebSocket endpoints
	router.HandleFunc("/ws", wsEndpoint) // For web interface
	router.HandleFunc("/agent/ws", agentWsEndpoint) // For agents
	
	// Static file serving for web interface
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	
	// Start the broadcast handler
	go handleBroadcasts()
	
	// Start the agent status updater
	go updateAgentStatuses()
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// Generate a random session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Validate session token
func validateSession(token string) bool {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()
	
	session, exists := sessions[token]
	if !exists {
		return false
	}
	
	// Check if session is expired (expire after 24 hours)
	if time.Since(session.CreatedAt) > 24*time.Hour {
		delete(sessions, token)
		return false
	}
	
	return true
}

// Create a new session
func createSession(username string) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}
	
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	
	sessions[token] = &Session{
		Username:  username,
		CreatedAt: time.Now(),
	}
	
	return token, nil
}

// Remove a session
func removeSession(token string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	
	delete(sessions, token)
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Check credentials (for now using a default user)
	if credentials.Username == "admin" && credentials.Password == "admin123" {
		token, err := createSession(credentials.Username)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": token,
			"status": "success",
		})
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

// Logout handler
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		http.Error(w, "Authorization header missing or invalid", http.StatusUnauthorized)
		return
	}
	
	token := authHeader[7:]
	removeSession(token)
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

// WebSocket endpoint for web interface
func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// Check for session token in Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		http.Error(w, "Authorization header missing or invalid", http.StatusUnauthorized)
		return
	}
	
	token := authHeader[7:]
	if !validateSession(token) {
		http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Add client to the clients map
	clients[conn] = true
	defer func() {
		delete(clients, conn)
	}()

	// Send initial agent list
	sendCurrentAgents(conn)

	// Keep connection alive
	for {
		// Read message to detect disconnection
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			break
		}
		
		// Handle authentication message
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}
		
		if msg["type"] == "auth" {
			clientToken, ok := msg["token"].(string)
			if !ok || clientToken != token {
				log.Printf("Invalid auth token from client")
				conn.Close()
				break
			}
			
			// Token validated successfully
			continue
		}
	}
}

// WebSocket endpoint for agents
func agentWsEndpoint(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Agent WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Read the initial registration message
	var regMsg struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Token string `json:"token"`
	}
	
	err = conn.ReadJSON(&regMsg)
	if err != nil {
		log.Printf("Failed to read registration message: %v", err)
		return
	}

	var agent *Agent
	var authenticated bool

	if regMsg.Type == "register" {
		// New agent registration
		ip := getClientIP(r)
		agent, err = authManager.RegisterAgent(regMsg.Name, ip)
		if err != nil {
			log.Printf("Failed to register agent: %v", err)
			return
		}
		
		// Send back the token for future authentication
		response := map[string]interface{}{
			"type": "registered",
			"token": agent.Token,
			"agent_id": agent.ID,
		}
		conn.WriteJSON(response)
		
		// Store agent connection
		agentConnections[agent.ID] = conn
		authenticated = true
	} else if regMsg.Type == "authenticate" && regMsg.Token != "" {
		// Existing agent authentication
		agent, authenticated = authManager.Authenticate(regMsg.Token)
		if !authenticated {
			log.Printf("Agent authentication failed with token: %s", regMsg.Token)
			return
		}
		
		// Store agent connection
		agentConnections[agent.ID] = conn
	}

	if !authenticated || agent == nil {
		return
	}

	// Handle messages from authenticated agent
	for {
		var msg struct {
			Type string `json:"type"`
			SystemStats SystemStats `json:"system_stats"`
			ProgramStats ProgramStats `json:"program_stats"`
		}
		
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Failed to read message from agent %s: %v", agent.ID, err)
			// Remove agent connection
			delete(agentConnections, agent.ID)
			break
		}

		if msg.Type == "stats" {
			// Update agent stats
			err = authManager.UpdateAgentStats(agent.ID, msg.SystemStats, msg.ProgramStats)
			if err != nil {
				log.Printf("Failed to update agent stats: %v", err)
				continue
			}
			
			// Broadcast updated agent info to web clients
			updatedAgent, _ := authManager.GetAgentByID(agent.ID)
			if updatedAgent != nil {
				broadcast <- *updatedAgent
			}
		}
	}
}

func getProxies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proxies)
}

func addProxy(w http.ResponseWriter, r *http.Request) {
	var proxy struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&proxy); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	proxies[proxy.ID] = proxy.Value
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(proxy)
}

func deleteProxy(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	delete(proxies, id)
	w.WriteHeader(http.StatusOK)
}

func getAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	agents := authManager.GetAgents()
	json.NewEncoder(w).Encode(agents)
}

func getAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	agent, exists := authManager.GetAgentByID(id)
	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// Broadcast agent updates to all connected web clients
func handleBroadcasts() {
	for {
		agent := <-broadcast
		// Convert to a format suitable for JSON serialization
		agentJSON, err := json.Marshal(agent)
		if err != nil {
			log.Printf("Error marshaling agent: %v", err)
			continue
		}
		
		// Send to all connected web clients
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, agentJSON)
			if err != nil {
				log.Printf("Error sending to client: %v", err)
				delete(clients, client)
				client.Close()
			}
		}
	}
}

// Update agent statuses periodically
func updateAgentStatuses() {
	for {
		time.Sleep(30 * time.Second) // Check every 30 seconds
		
		// Get all agents and update their statuses
		agents := authManager.GetAgents()
		for _, agent := range agents {
			// The status is already updated in GetAgents() method
			// Now broadcast the updated status to web clients
			broadcast <- *agent
		}
	}
}

// Send current agents list to a new client
func sendCurrentAgents(conn *websocket.Conn) {
	agents := authManager.GetAgents()
	
	for _, agent := range agents {
		agentJSON, err := json.Marshal(agent)
		if err != nil {
			log.Printf("Error marshaling agent: %v", err)
			continue
		}
		
		err = conn.WriteMessage(websocket.TextMessage, agentJSON)
		if err != nil {
			log.Printf("Error sending agent to client: %v", err)
			break
		}
	}
}

// Get client IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	
	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	// Use remote address
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}