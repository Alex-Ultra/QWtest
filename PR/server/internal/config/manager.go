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
		if err := cm.createDefaultServerConfig(); err != nil {
			return fmt.Errorf("failed to create default server config: %v", err)
		}
		fmt.Printf("Created default server config at: %s\n", cm.configPath)
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

// createDefaultServerConfig creates a default server configuration file
func (cm *ConfigManager) createDefaultServerConfig() error {
	defaultConfig := &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Tokens: map[string]string{
			"abc123":              "agent-1",
			"def456":              "agent-2",
			"xyz789":              "agent-3",
			"web-interface-token": "web-interface",
		},
		Paths: PathsConfig{
			MonitorBinDir:   "./assets/bins/monitor/",
			ProxyBinPath:    "./assets/bins/proxy/xmrig-proxy",
			ProxyConfigPath: "./configs/proxy-config.json",
			WebDistDir:      "./dist/",
		},
		MetricsInterval: 30,
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %v", err)
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// ensurePaths ensures all paths in the configuration exist
func (cm *ConfigManager) ensurePaths(paths *PathsConfig) error {
	// Ensure monitor bin directory exists
	if err := os.MkdirAll(paths.MonitorBinDir, 0755); err != nil {
		return fmt.Errorf("failed to create monitor bin directory: %v", err)
	}

	// Ensure proxy config exists
	if err := cm.ensureProxyConfig(paths.ProxyConfigPath); err != nil {
		return fmt.Errorf("failed to ensure proxy config: %v", err)
	}

	// Ensure proxy bin directory exists
	proxyBinDir := filepath.Dir(paths.ProxyBinPath)
	if err := os.MkdirAll(proxyBinDir, 0755); err != nil {
		return fmt.Errorf("failed to create proxy bin directory: %v", err)
	}

	// Ensure web dist directory exists
	if err := os.MkdirAll(paths.WebDistDir, 0755); err != nil {
		return fmt.Errorf("failed to create web dist directory: %v", err)
	}

	return nil
}

// ensureProxyConfig ensures the proxy config file exists
func (cm *ConfigManager) ensureProxyConfig(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %v", err)
		}

		// Create default proxy config
		defaultProxyConfig := `{
  "pools": [
    {
      "url": "pool.supportxmr.com:3333",
      "user": "YOUR_XMR_ADDRESS",
      "pass": "x",
      "keepalive": true,
      "tls": false
    }
  ],
  "upstream": {
    "url": "localhost:3333",
    "user": "YOUR_XMR_ADDRESS",
    "pass": "x"
  },
  "bind": [
    "127.0.0.1:3333"
  ],
  "access-log": {
    "file": "access.log",
    "max-files": 3
  },
  "donate-level": 1
}`
		return os.WriteFile(configPath, []byte(defaultProxyConfig), 0644)
	} else if err != nil {
		return fmt.Errorf("error checking proxy config file: %v", err)
	}
	return nil
}

// ValidateConfig validates the loaded configuration
func (cm *ConfigManager) ValidateConfig(config *Config) error {
	var errors []string

	// Validate required fields
	if strings.TrimSpace(config.Server.Port) == "" {
		errors = append(errors, "server.port is required")
	}
	if strings.TrimSpace(config.Server.Host) == "" {
		errors = append(errors, "server.host is required")
	}
	if config.MetricsInterval <= 0 {
		errors = append(errors, "metrics_interval must be greater than 0")
	}
	if config.Tokens == nil || len(config.Tokens) == 0 {
		errors = append(errors, "at least one token is required")
	}

	// Validate paths exist
	if _, err := os.Stat(config.Paths.MonitorBinDir); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("monitor bin directory does not exist: %s", config.Paths.MonitorBinDir))
	}
	if _, err := os.Stat(config.Paths.ProxyConfigPath); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("proxy config does not exist: %s", config.Paths.ProxyConfigPath))
	}
	if _, err := os.Stat(config.Paths.WebDistDir); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("web dist directory does not exist: %s", config.Paths.WebDistDir))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}