package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Tokens         map[string]string `yaml:"tokens"`
	Paths          Paths             `yaml:"paths"`
	MetricsInterval int              `yaml:"metrics_interval"`
}

type Paths struct {
	MonitorBinDir    string `yaml:"monitor_bin_dir"`
	ProxyBinPath     string `yaml:"proxy_bin_path"`
	ProxyConfigPath  string `yaml:"proxy_config_path"`
	WebDistDir       string `yaml:"web_dist_dir"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}