package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	monitor_agent "pr-agent/internal/config"
	"pr-agent/internal/monitor"
	"pr-agent/internal/updater"
	"pr-agent/internal/ws"
)

func main() {
	// Load configuration
	cfg, err := monitor_agent.LoadConfig("configs/agent-config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create updater
	updater := updater.NewBinaryUpdater(cfg.ServerURL, cfg.Token)
	
	// Check if monitor binary exists, download if needed
	err = updater.CheckAndDownloadMonitor(cfg.Paths.MonitorBin)
	if err != nil {
		log.Fatalf("Failed to download monitor binary: %v", err)
	}

	// Create monitor controller
	monitorCtrl := monitor.NewController(
		cfg.Paths.MonitorBin,
		cfg.Paths.MonitorConfig,
		cfg.Paths.MonitorLogFile,
	)

	// Create WebSocket client
	wsClient := ws.NewClient(cfg.ServerURL, cfg.Token)

	// Set up command handler
	wsClient.OnCommand = func(cmd map[string]interface{}) {
		action, ok := cmd["action"].(string)
		if !ok {
			log.Printf("Invalid command format: missing action")
			return
		}

		switch action {
		case "start_monitor":
			err := monitorCtrl.Start()
			if err != nil {
				log.Printf("Failed to start monitor: %v", err)
			} else {
				log.Printf("Monitor started successfully")
			}
		case "stop_monitor":
			err := monitorCtrl.Stop()
			if err != nil {
				log.Printf("Failed to stop monitor: %v", err)
			} else {
				log.Printf("Monitor stopped successfully")
			}
		case "restart_monitor":
			err := monitorCtrl.Restart()
			if err != nil {
				log.Printf("Failed to restart monitor: %v", err)
			} else {
				log.Printf("Monitor restarted successfully")
			}
		case "update_config":
			payload, ok := cmd["payload"].(map[string]interface{})
			if !ok {
				log.Printf("Invalid config update payload")
				return
			}
			
			// Handle config update
			configPath, ok := payload["config_path"].(string)
			if ok {
				err := monitorCtrl.UpdateConfig(configPath)
				if err != nil {
					log.Printf("Failed to update config: %v", err)
				} else {
					log.Printf("Config updated successfully")
				}
			}
		default:
			log.Printf("Unknown command: %s", action)
		}
	}

	// Set up connection handlers
	wsClient.OnConnect = func() {
		log.Printf("Connected to server")
	}
	
	wsClient.OnDisconnect = func() {
		log.Printf("Disconnected from server")
	}

	// Connect to server
	err = wsClient.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	// Start metrics reporting loop
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.MetricsInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if !wsClient.Connected {
				continue
			}

			// In a real implementation, these values would come from actual monitoring
			hashrate := getMockHashrate() // Placeholder function
			cpuUsage := getMockCPUUsage() // Placeholder function
			ramUsage := getMockRAMUsage() // Placeholder function
			tempCPU := getMockTempCPU()   // Placeholder function
			status := getMockStatus()     // Placeholder function

			err := wsClient.SendMetrics(hashrate, cpuUsage, ramUsage, tempCPU, status)
			if err != nil {
				log.Printf("Failed to send metrics: %v", err)
			}
		}
	}()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	log.Printf("Agent started, connected to %s", cfg.ServerURL)
	
	// Wait for signal to shut down
	<-sigChan
	log.Printf("Shutting down agent...")
	
	wsClient.Disconnect()
	
	log.Printf("Agent stopped")
}

// Mock functions for demonstration purposes
// In a real implementation, these would collect actual system metrics
func getMockHashrate() float64 {
	return 420.5
}

func getMockCPUUsage() float64 {
	return 75.3
}

func getMockRAMUsage() float64 {
	return 62.1
}

func getMockTempCPU() int {
	return 65
}

func getMockStatus() string {
	return "mining" // Could be "mining", "stopped", "paused", etc.
}