package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	
	"github.com/gorilla/mux"
	"monitor-server/internal/auth"
	"monitor-server/internal/config"
	"monitor-server/internal/models"
)

var (
	agents = make(map[string]*models.Agent)
	commandsQueue = make([]*models.Command, 0)
	agentsMutex = &sync.Mutex{}
)

func GetAgentsHandler(w http.ResponseWriter, r *http.Request) {
	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !auth.ValidateToken(authToken) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	agentsList := make([]models.Agent, 0, len(agents))
	for _, agent := range agents {
		// Обновляем статус на основе времени последнего подключения
		if time.Now().Unix()-agent.LastSeen > 60 { // Если не было активности более 60 секунд
			agent.Status = "offline"
		}
		agentsList = append(agentsList, *agent)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agentsList)
}

func GetAgentHandler(w http.ResponseWriter, r *http.Request) {
	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !auth.ValidateToken(authToken) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	agentID := vars["id"]

	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	agent, exists := agents[agentID]
	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	// Обновляем статус на основе времени последнего подключения
	if time.Now().Unix()-agent.LastSeen > 60 { // Если не было активности более 60 секунд
		agent.Status = "offline"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

func SendCommandHandler(w http.ResponseWriter, r *http.Request) {
	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !auth.ValidateToken(authToken) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	agentID := vars["id"]

	var cmd models.Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.ID = generateID()
	cmd.AgentID = agentID
	cmd.Created = time.Now().Unix()
	cmd.Sent = false

	// Добавляем команду в очередь
	commandsQueue = append(commandsQueue, &cmd)

	// TODO: Отправить команду через WebSocket агенту

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Command queued"})
}

func GetCommandsQueueHandler(w http.ResponseWriter, r *http.Request) {
	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !auth.ValidateToken(authToken) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commandsQueue)
}

func DownloadMonitorHandler(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	platform := mux.Vars(r)["platform"]
	
	// Проверяем токен агента
	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !auth.ValidateTokenForDownload(authToken, platform) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Путь к бинарнику в зависимости от платформы
	binPath := ""
	switch platform {
	case "windows":
		binPath = cfg.Paths.MonitorBinDir + "monitor.exe"
	case "linux":
		binPath = cfg.Paths.MonitorBinDir + "monitor"
	default:
		http.Error(w, "Unsupported platform", http.StatusBadRequest)
		return
	}

	// Отправляем файл
	http.ServeFile(w, r, binPath)
}

func StartProxyHandler(w http.ResponseWriter, r *http.Request) {
	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !auth.ValidateToken(authToken) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Реализовать запуск Proxy
	log.Println("Starting proxy...")
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Proxy started"})
}

func StopProxyHandler(w http.ResponseWriter, r *http.Request) {
	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !auth.ValidateToken(authToken) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Реализовать остановку Proxy
	log.Println("Stopping proxy...")
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Proxy stopped"})
}

func GetProxyStatusHandler(w http.ResponseWriter, r *http.Request) {
	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !auth.ValidateToken(authToken) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Реализовать получение статуса Proxy
	status := "stopped" // или "running"
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

// Вспомогательная функция для генерации ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}