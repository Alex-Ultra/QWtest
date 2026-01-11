package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"monitor-server/internal/api"
	"monitor-server/internal/config"
	"monitor-server/internal/ws"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Create router
	router := mux.NewRouter()

	// Initialize API handlers
	api.InitHandlers(router, hub, cfg)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}