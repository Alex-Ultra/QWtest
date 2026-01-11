package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerURL       string       `yaml:"server_url"`
	Token           string       `yaml:"token"`
	Paths           PathsConfig  `yaml:"paths"`
	MetricsInterval int          `yaml:"metrics_interval"`
}

type PathsConfig struct {
	MonitorBin      string `yaml:"monitor_bin"`
	MonitorConfig   string `yaml:"monitor_config"`
	MonitorLogFile  string `yaml:"monitor_log_file"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}