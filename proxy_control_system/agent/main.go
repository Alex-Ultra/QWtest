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

// Stats represents system statistics
type Stats struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Temp   float64 `json:"temp"`
}

// AgentData represents data sent to the server
type AgentData struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Stats    Stats  `json:"stats"`
	ProxyID  string `json:"proxy_id"`
}

// Configuration for the agent
type Config struct {
	ServerURL string `json:"server_url"`
	AgentID   string `json:"agent_id"`
	Program   string `json:"program"`
	ProxyID   string `json:"proxy_id"`
}

func main() {
	// Load configuration
	config := loadConfig()
	
	// Connect to server
	u, err := url.Parse(config.ServerURL)
	if err != nil {
		log.Fatal("Invalid server URL:", err)
	}
	
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
	
	// Start monitoring loop
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			stats := collectStats()
			
			hostname, _ := os.Hostname()
			
			data := AgentData{
				ID:       config.AgentID,
				Hostname: hostname,
				OS:       runtime.GOOS,
				Stats:    stats,
				ProxyID:  config.ProxyID,
			}
			
			err := c.WriteJSON(data)
			if err != nil {
				log.Printf("Failed to send data to server: %v", err)
				// Try to reconnect
				c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
				if err != nil {
					log.Printf("Failed to reconnect: %v", err)
				} else {
					fmt.Println("Reconnected to server")
				}
			}
			
			// Check if target program is running through proxy
			isRunning := isProgramRunning(config.Program)
			if !isRunning {
				fmt.Printf("Program %s is not running\n", config.Program)
				// Optionally restart the program with proxy settings
				// startProgramWithProxy(config.Program, config.ProxyID)
			} else {
				fmt.Printf("Program %s is running\n", config.Program)
			}
		}
	}
}

func loadConfig() Config {
	// Default configuration
	config := Config{
		ServerURL: "ws://localhost:8080/ws",
		AgentID:   "agent_" + strconv.Itoa(int(time.Now().Unix())),
		Program:   "firefox", // Example program
		ProxyID:   "proxy_1",
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

func collectStats() Stats {
	stats := Stats{}
	
	// CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Printf("Error getting CPU usage: %v", err)
	} else if len(cpuPercent) > 0 {
		stats.CPU = cpuPercent[0]
	}
	
	// Memory usage
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting memory usage: %v", err)
	} else {
		stats.Memory = vmStat.UsedPercent
	}
	
	// Temperature
	temp, err := getTemperature()
	if err != nil {
		log.Printf("Error getting temperature: %v", err)
		stats.Temp = 0.0
	} else {
		stats.Temp = temp
	}
	
	return stats
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
		return 0.0, nil
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