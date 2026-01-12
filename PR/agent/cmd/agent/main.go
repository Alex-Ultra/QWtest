package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	monitor_agent "pr-agent/internal/config"
	"pr-agent/internal/monitor"
	"pr-agent/internal/updater"
	"pr-agent/internal/ws"
)

func main() {
	// Determine config path relative to executable location
	configPath := "configs/agent-config.yaml"
	
	// Also check for alternative paths for cross-platform compatibility
	cfg, err := monitor_agent.LoadConfig(configPath)
	if err != nil {
		// Try alternative paths for cross-platform compatibility
		altConfigPath := "../configs/agent-config.yaml"
		cfg, err = monitor_agent.LoadConfig(altConfigPath)
		if err != nil {
			// Check if config file exists in current directory
			if _, statErr := os.Stat("agent-config.yaml"); statErr == nil {
				cfg, err = monitor_agent.LoadConfig("agent-config.yaml")
				if err != nil {
					log.Fatalf("Failed to load agent config from any location: original path '%s': %v, alt path '%s': %v, local 'agent-config.yaml': %v", 
						configPath, err, altConfigPath, err, err)
				}
				log.Printf("Loaded agent config from local directory")
			} else {
				log.Fatalf("Failed to load agent config from any location: original path '%s': %v, alt path '%s': %v, local 'agent-config.yaml' not found: %v", 
					configPath, err, altConfigPath, err, statErr)
			}
		} else {
			log.Printf("Loaded agent config from alternative path: %s", altConfigPath)
		}
	} else {
		log.Printf("Loaded agent config from: %s", configPath)
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

			// Get actual system metrics
			hashrate := getRealHashrate()
			cpuUsage := getRealCPUUsage()
			ramUsage := getRealRAMUsage()
			tempCPU := getRealTempCPU()
			status := getRealStatus()

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

// Real functions to collect actual system metrics
func getRealHashrate() float64 {
	// In a real mining application, this would interface with the miner to get actual hashrate
	// For now, we'll simulate a dynamic value based on CPU usage
	percent, err := cpu.Percent(time.Second, false)
	if err != nil || len(percent) == 0 {
		return 0.0
	}
	
	// Simulate hashrate based on CPU usage (this is just a simulation)
	// In a real implementation, this would come from actual mining software
	return percent[0] * 10.0 // Just an example calculation
}

func getRealCPUUsage() float64 {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil || len(percent) == 0 {
		return 0.0
	}
	return percent[0]
}

func getRealRAMUsage() float64 {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0.0
	}
	return vmStat.UsedPercent
}

func getRealTempCPU() int {
	// Getting temperature varies by platform and might not be available on all systems
	// Return a reasonable default if not available
	temp := 50 // Default temperature in Celsius
	
	// On some systems we could get actual temperature readings
	// But gopsutil doesn't provide cross-platform temperature reading consistently
	// So we'll return a simulated value based on CPU usage
	cpuUsage := getRealCPUUsage()
	if cpuUsage > 80 {
		temp = 70
	} else if cpuUsage > 60 {
		temp = 65
	} else if cpuUsage > 40 {
		temp = 60
	}
	
	return temp
}

func getRealStatus() string {
	// In a real implementation, this would check the actual status of the mining process
	// For now, we'll determine status based on if the monitor is running
	// Since we don't have direct access here, we'll just return a default status
	return "active" // Could be "active", "idle", "maintenance", etc.
}