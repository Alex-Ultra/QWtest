package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"pr-server/internal/api"
	"pr-server/internal/config"
	"pr-server/internal/ws"
)

func main() {
	// Determine config path relative to executable location
	configPath := filepath.Join("configs", "config.yaml")
	
	// Also check for absolute path or working directory
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		// Try alternative paths for cross-platform compatibility
		// Check if we're in a subdirectory, try relative to root
		altConfigPath := filepath.Join("..", "configs", "config.yaml")
		cfg, err = config.LoadConfig(altConfigPath)
		if err != nil {
			// Check if config file exists in current directory
			if _, statErr := os.Stat("config.yaml"); statErr == nil {
				cfg, err = config.LoadConfig("config.yaml")
				if err != nil {
					log.Fatalf("Failed to load config from any location: original path '%s': %v, alt path '%s': %v, local 'config.yaml': %v", 
						configPath, err, altConfigPath, err, err)
				}
				log.Printf("Loaded config from local directory")
			} else {
				log.Fatalf("Failed to load config from any location: original path '%s': %v, alt path '%s': %v, local 'config.yaml' not found: %v", 
					configPath, err, altConfigPath, err, statErr)
			}
		} else {
			log.Printf("Loaded config from alternative path: %s", altConfigPath)
		}
	} else {
		log.Printf("Loaded config from: %s", configPath)
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