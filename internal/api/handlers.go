package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"monitor-server/internal/auth"
	"monitor-server/internal/config"
	"monitor-server/internal/models"
	"monitor-server/internal/proxy"
	"monitor-server/internal/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for development
		// In production, you should restrict this
		return true
	},
}

type Handlers struct {
	hub      *ws.Hub
	cfg      *config.Config
	agents   *sync.Map
	commands *sync.Map
	proxyCtrl *proxy.Controller
}

func InitHandlers(router *mux.Router, hub *ws.Hub, cfg *config.Config) {
	handlers := &Handlers{
		hub:      hub,
		cfg:      cfg,
		agents:   &sync.Map{},
		commands: &sync.Map{},
		proxyCtrl: proxy.NewController(cfg.Paths.ProxyBinPath, cfg.Paths.ProxyConfigPath),
	}

	// Apply auth middleware to protected routes
	protected := router.PathPrefix("/api/v1").Subrouter()
	protected.Use(auth.AuthMiddleware(cfg))

	// Agent routes
	protected.HandleFunc("/agents", handlers.GetAgents).Methods("GET")
	protected.HandleFunc("/agent/{id}", handlers.GetAgent).Methods("GET")
	protected.HandleFunc("/agent/{id}/command", handlers.SendCommand).Methods("POST")

	// Command queue route
	protected.HandleFunc("/commands/queue", handlers.GetCommandQueue).Methods("GET")

	// Download routes (no auth required)
	router.HandleFunc("/api/v1/download/monitor/{platform}", handlers.DownloadMonitor).Methods("GET")

	// Proxy routes
	protected.HandleFunc("/proxy/start", handlers.StartProxy).Methods("POST")
	protected.HandleFunc("/proxy/stop", handlers.StopProxy).Methods("POST")
	protected.HandleFunc("/proxy/status", handlers.GetProxyStatus).Methods("GET")

	// WebSocket endpoint
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.WebSocketHandler(w, r)
	})

	// Serve static files
	serveStaticFiles(cfg, router)
}

func (h *Handlers) GetAgents(w http.ResponseWriter, r *http.Request) {
	agents := make([]*models.Agent, 0)
	h.agents.Range(func(key, value interface{}) bool {
		agent := value.(*models.Agent)
		agents = append(agents, agent)
		return true
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func (h *Handlers) GetAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if agent, ok := h.agents.Load(id); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agent)
	} else {
		http.Error(w, "Agent not found", http.StatusNotFound)
	}
}

func (h *Handlers) SendCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["id"]

	var req models.CommandRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Error parsing request JSON", http.StatusBadRequest)
		return
	}

	command := &models.Command{
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		AgentID: agentID,
		Action:  req.Action,
		Payload: req.Payload,
		Created: time.Now().Unix(),
		Sent:    false,
	}

	h.commands.Store(command.ID, command)

	// Check if the agent is connected and send the command immediately
	client, ok := h.hub.GetClient(agentID)
	if ok && client.Conn != nil {
		// Send command via WebSocket
		msgBytes, _ := json.Marshal(command)
		h.hub.SendToClient(agentID, msgBytes)
		
		// Mark as sent
		command.Sent = true
		h.commands.Store(command.ID, command)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "command queued"})
}

func (h *Handlers) GetCommandQueue(w http.ResponseWriter, r *http.Request) {
	commands := make([]*models.Command, 0)
	h.commands.Range(func(key, value interface{}) bool {
		command := value.(*models.Command)
		commands = append(commands, command)
		return true
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commands)
}

func (h *Handlers) DownloadMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	platform := vars["platform"]

	// Determine the filename based on platform
	var filename string
	switch platform {
	case "windows":
		filename = "monitor.exe"
	case "linux":
		filename = "monitor"
	default:
		http.Error(w, "Unsupported platform", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.cfg.Paths.MonitorBinDir, filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Binary not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	http.ServeFile(w, r, filePath)
}

func (h *Handlers) StartProxy(w http.ResponseWriter, r *http.Request) {
	err := h.proxyCtrl.Start()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting proxy: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "proxy started"})
}

func (h *Handlers) StopProxy(w http.ResponseWriter, r *http.Request) {
	err := h.proxyCtrl.Stop()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error stopping proxy: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "proxy stopped"})
}

func (h *Handlers) GetProxyStatus(w http.ResponseWriter, r *http.Request) {
	status := h.proxyCtrl.Status()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *Handlers) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Extract token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		log.Println("No token provided in WebSocket connection")
		return
	}

	// Find agent ID associated with the token
	var agentID string
	var found bool
	for tkn, id := range h.cfg.Tokens {
		if tkn == token {
			agentID = id
			found = true
			break
		}
	}

	if !found {
		log.Printf("Invalid token in WebSocket connection: %s", token)
		return
	}

	// Create client
	client := &ws.Client{
		ID:  agentID,
		Hub: h.hub,
		Conn: conn,
		Send: make(chan []byte, 256),
		Agent: &models.Agent{
			ID: agentID,
			Token: token,
		},
	}

	client.Hub.Register(client)

	// Handle incoming messages from the client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			client.Hub.Unregister(client)
			break
		}

		// Handle different types of messages
		var msgMap map[string]interface{}
		err = json.Unmarshal(message, &msgMap)
		if err != nil {
			log.Printf("Error unmarshaling WebSocket message: %v", err)
			continue
		}

		// Check if this is metrics data
		if _, hasHashrate := msgMap["hashrate"]; hasHashrate {
			// This is metrics data
			var metrics models.Metrics
			metricsData, _ := json.Marshal(msgMap)
			json.Unmarshal(metricsData, &metrics)

			// Update agent metrics
			if agent, ok := h.agents.Load(agentID); ok {
				a := agent.(*models.Agent)
				a.Hashrate = metrics.Hashrate
				a.CPUUsage = metrics.CPUUsage
				a.RAMUsage = metrics.RAMUsage
				a.TempCPU = metrics.TempCPU
				a.Status = metrics.Status
				a.LastSeen = time.Now().Unix()
			} else {
				// Agent not registered yet, create a new one
				newAgent := &models.Agent{
					ID:       agentID,
					Token:    token,
					IP:       getClientIP(r),
					Platform: getStringValue(msgMap, "platform", "unknown"),
					LastSeen: time.Now().Unix(),
					Hashrate: metrics.Hashrate,
					CPUUsage: metrics.CPUUsage,
					RAMUsage: metrics.RAMUsage,
					TempCPU:  metrics.TempCPU,
					Status:   metrics.Status,
					Version:  getStringValue(msgMap, "version", "1.0.0"),
				}
				h.agents.Store(agentID, newAgent)
			}
		}
	}
}



func serveStaticFiles(cfg *config.Config, router *mux.Router) {
	staticDir := cfg.Paths.WebDistDir
	fs := http.FileServer(http.Dir(staticDir))
	
	// Serve static files under /static prefix
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	
	// Catch-all handler for SPA routing
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't serve index.html for API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		
		// Serve index.html for all other routes (SPA)
		http.ServeFile(w, r, staticDir+"/index.html")
	})
}

func getClientIP(r *http.Request) string {
	// Get the real IP if behind a proxy
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return strings.Split(ip, ",")[0]
	}
	
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	
	host := r.RemoteAddr
	colonIndex := strings.LastIndex(host, ":")
	if colonIndex != -1 {
		return host[:colonIndex]
	}
	
	return host
}

func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}