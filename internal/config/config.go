package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server           ServerConfig            `yaml:"server"`
	Tokens           map[string]string       `yaml:"tokens"`
	Paths            PathsConfig             `yaml:"paths"`
	MetricsInterval  int                     `yaml:"metrics_interval"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type PathsConfig struct {
	MonitorBinDir     string `yaml:"monitor_bin_dir"`
	ProxyBinPath      string `yaml:"proxy_bin_path"`
	ProxyConfigPath   string `yaml:"proxy_config_path"`
	WebDistDir        string `yaml:"web_dist_dir"`
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