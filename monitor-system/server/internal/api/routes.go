package api

import (
	"net/http"
	
	"github.com/gorilla/mux"
	"monitor-server/internal/config"
	"monitor-server/internal/ws"
)

func SetupRoutes(router *mux.Router, hub *ws.Hub, cfg *config.Config) {
	// Защищенные маршруты для веб-интерфейса
	protected := router.PathPrefix("/api/v1").Subrouter()
	protected.HandleFunc("/agents", GetAgentsHandler).Methods("GET")
	protected.HandleFunc("/agent/{id}", GetAgentHandler).Methods("GET")
	protected.HandleFunc("/agent/{id}/command", SendCommandHandler).Methods("POST")
	protected.HandleFunc("/commands/queue", GetCommandsQueueHandler).Methods("GET")
	protected.HandleFunc("/proxy/start", StartProxyHandler).Methods("POST")
	protected.HandleFunc("/proxy/stop", StopProxyHandler).Methods("POST")
	protected.HandleFunc("/proxy/status", GetProxyStatusHandler).Methods("GET")

	// Маршрут для скачивания Monitor (доступен для агентов)
	router.HandleFunc("/api/v1/download/monitor/{platform}", func(w http.ResponseWriter, r *http.Request) {
		DownloadMonitorHandler(cfg, w, r)
	}).Methods("GET")
}