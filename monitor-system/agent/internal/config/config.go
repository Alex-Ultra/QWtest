package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	ServerURL       string `yaml:"server_url"`
	Token           string `yaml:"token"`

	Paths           Paths  `yaml:"paths"`
	MetricsInterval int    `yaml:"metrics_interval"`
}

type Paths struct {
	MonitorBin      string `yaml:"monitor_bin"`
	MonitorConfig   string `yaml:"monitor_config"`
	MonitorLogFile  string `yaml:"monitor_log_file"`
}

func LoadAgentConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AgentConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}