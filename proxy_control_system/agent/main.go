package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/gorilla/websocket"
)

// SystemStats represents system statistics
type SystemStats struct {
	Hostname    string    `json:"hostname"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	Temperature float64   `json:"temperature"`
	Timestamp   time.Time `json:"timestamp"`
}

// ProgramStats represents program statistics
type ProgramStats struct {
	ProgramName string            `json:"program_name"`
	Proxy       string            `json:"proxy"`
	ProcessInfo map[string]string `json:"process_info"`
	Timestamp   time.Time         `json:"timestamp"`
}

// Configuration for the agent
type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token"`
	Name      string `json:"name"`
	Program   string `json:"program"`
	Proxy     string `json:"proxy"`
}

func main() {
	// Load configuration
	config := loadConfig()
	
	// Connect to server
	u, err := url.Parse(config.ServerURL)
	if err != nil {
		log.Fatal("Invalid server URL:", err)
	}
	
	// Change the path to agent WebSocket endpoint
	u.Path = "/agent/ws"
	
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}
	defer c.Close()
	
	fmt.Println("Connected to server")
	
	// Authenticate or register
	var token string
	if config.Token != "" {
		// Try to authenticate with existing token
		authMsg := map[string]interface{}{
			"type":  "authenticate",
			"token": config.Token,
		}
		err = c.WriteJSON(authMsg)
		if err != nil {
			log.Printf("Failed to send authentication message: %v", err)
		}
		
		// Wait for response
		var response map[string]interface{}
		err = c.ReadJSON(&response)
		if err != nil {
			log.Printf("Failed to read authentication response: %v", err)
		} else if response["type"] == "error" {
			log.Printf("Authentication failed: %v", response["message"])
			// Fall back to registration
			token = registerAgent(c, config)
		} else {
			token = config.Token
			fmt.Println("Authenticated with existing token")
		}
	} else {
		// Register new agent
		token = registerAgent(c, config)
	}
	
	// Update config with new token
	config.Token = token
	saveConfig(config)
	
	// Start monitoring loop - send stats every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			systemStats := collectSystemStats()
			programStats := collectProgramStats(config)
			
			statsMsg := map[string]interface{}{
				"type":          "stats",
				"system_stats":  systemStats,
				"program_stats": programStats,
			}
			
			err := c.WriteJSON(statsMsg)
			if err != nil {
				log.Printf("Failed to send stats to server: %v", err)
				// Try to reconnect
				c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
				if err != nil {
					log.Printf("Failed to reconnect: %v", err)
					// Wait a bit before trying again
					time.Sleep(5 * time.Second)
					continue
				} else {
					fmt.Println("Reconnected to server")
					// Re-authenticate
					authMsg := map[string]interface{}{
						"type":  "authenticate",
						"token": config.Token,
					}
					err = c.WriteJSON(authMsg)
					if err != nil {
						log.Printf("Failed to send authentication message: %v", err)
						continue
					}
					
					// Wait for response
					var response map[string]interface{}
					err = c.ReadJSON(&response)
					if err != nil {
						log.Printf("Failed to read authentication response: %v", err)
						continue
					} else if response["type"] == "error" {
						log.Printf("Re-authentication failed: %v", response["message"])
						continue
					}
				}
			}
			
			// Check if target program is running through proxy
			isRunning := isProgramRunning(config.Program)
			if !isRunning {
				fmt.Printf("Program %s is not running\n", config.Program)
				// Optionally restart the program with proxy settings
				// startProgramWithProxy(config.Program, config.Proxy)
			} else {
				fmt.Printf("Program %s is running\n", config.Program)
			}
		}
	}
}

func registerAgent(c *websocket.Conn, config Config) string {
	// Register new agent
	regMsg := map[string]interface{}{
		"type": "register",
		"name": config.Name,
	}
	err := c.WriteJSON(regMsg)
	if err != nil {
		log.Printf("Failed to send registration message: %v", err)
		return ""
	}
	
	// Wait for registration response
	var response map[string]interface{}
	err = c.ReadJSON(&response)
	if err != nil {
		log.Printf("Failed to read registration response: %v", err)
		return ""
	}
	
	if response["type"] != "registered" {
		log.Printf("Registration failed: %v", response)
		return ""
	}
	
	token, ok := response["token"].(string)
	if !ok {
		log.Printf("Invalid registration response: token not found")
		return ""
	}
	
	agentID, ok := response["agent_id"].(string)
	if !ok {
		log.Printf("Invalid registration response: agent_id not found")
		return ""
	}
	
	fmt.Printf("Registered new agent: %s with token: %s\n", agentID, token)
	
	return token
}

func loadConfig() Config {
	// Default configuration
	config := Config{
		ServerURL: "ws://localhost:8080",
		Name:      "agent_" + getHostname(),
		Program:   "firefox", // Example program
		Proxy:     "proxy_1",
	}
	
	// Try to load from config file
	if _, err := os.Stat("config.json"); err == nil {
		file, err := os.ReadFile("config.json")
		if err == nil {
			json.Unmarshal(file, &config)
		}
	}
	
	return config
}

func saveConfig(config Config) {
	file, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal config: %v", err)
		return
	}
	
	err = os.WriteFile("config.json", file, 0644)
	if err != nil {
		log.Printf("Failed to save config: %v", err)
	}
}

func collectSystemStats() SystemStats {
	stats := SystemStats{}
	
	hostname, _ := os.Hostname()
	stats.Hostname = hostname
	stats.Timestamp = time.Now()
	
	// CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Printf("Error getting CPU usage: %v", err)
	} else if len(cpuPercent) > 0 {
		stats.CPUUsage = cpuPercent[0]
	}
	
	// Memory usage
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting memory usage: %v", err)
	} else {
		stats.MemoryUsage = vmStat.UsedPercent
	}
	
	// Temperature
	temp, err := getTemperature()
	if err != nil {
		log.Printf("Error getting temperature: %v", err)
		stats.Temperature = 0.0
	} else {
		stats.Temperature = temp
	}
	
	return stats
}

func collectProgramStats(config Config) ProgramStats {
	programStats := ProgramStats{
		ProgramName: config.Program,
		Proxy:       config.Proxy,
		Timestamp:   time.Now(),
		ProcessInfo: make(map[string]string),
	}
	
	// Add any relevant process information
	isRunning := isProgramRunning(config.Program)
	programStats.ProcessInfo["running"] = fmt.Sprintf("%t", isRunning)
	
	return programStats
}

func getTemperature() (float64, error) {
	var temp float64
	var err error
	
	switch runtime.GOOS {
	case "windows":
		temp, err = getWindowsTemperature()
	case "linux":
		temp, err = getLinuxTemperature()
	default:
		// For other OS, return 0 as default
		return 40.0, nil
	}
	
	return temp, err
}

func getWindowsTemperature() (float64, error) {
	// On Windows, we can use wmic to get temperature
	// Note: This may require admin privileges and only works on some systems
	cmd := exec.Command("wmic", "path", "Win32_TemperatureProbe", "get", "CurrentReading")
	output, err := cmd.Output()
	if err != nil {
		// If wmic fails, try other methods or return an average temperature
		return 40.0, nil // Return a default temperature
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "CurrentReading" {
			temp, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
			if err == nil {
				return temp, nil
			}
		}
	}
	
	return 40.0, nil // Default if no temperature found
}

func getLinuxTemperature() (float64, error) {
	// On Linux, we can read from thermal zones
	// First, try to get from /sys/class/thermal
	cmd := exec.Command("cat", "/sys/class/thermal/thermal_zone*/temp")
	output, err := cmd.Output()
	if err != nil {
		// If the command fails, try to install and use sensors
		return 40.0, nil // Return default temperature
	}
	
	lines := strings.Split(string(output), "\n")
	var totalTemp float64
	var count int
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			temp, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
			if err == nil {
				totalTemp += temp / 1000.0 // Convert from millidegrees to degrees
				count++
			}
		}
	}
	
	if count > 0 {
		return totalTemp / float64(count), nil
	}
	
	return 40.0, nil // Default if no temperature found
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func isProgramRunning(programName string) bool {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("tasklist")
	case "linux":
		cmd = exec.Command("ps", "aux")
	default:
		return false
	}
	
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return strings.Contains(strings.ToLower(string(output)), strings.ToLower(programName))
}

func startProgramWithProxy(programName string, proxyID string) {
	// This is a placeholder for starting a program with proxy settings
	// Implementation would depend on the specific program and proxy type
	fmt.Printf("Starting %s with proxy %s\n", programName, proxyID)
}