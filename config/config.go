package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Backends BackendsConfig `yaml:"backends"`
	Strategy string         `yaml:"strategy"` // "roundrobin" or "leastconnection"
	Circuit  CircuitConfig  `yaml:"circuit"`
	Health   HealthConfig   `yaml:"health"`
	Metrics  MetricsConfig  `yaml:"metrics"`
}

type ServerConfig struct {
	Port    int    `yaml:"port"`
	Address string `yaml:"address"`
}

type BackendsConfig struct {
	URLs []string `yaml:"urls"`
}

type CircuitConfig struct {
	FailureThreshold int           `yaml:"failure_threshold"`
	Timeout          time.Duration `yaml:"timeout"`
}

type HealthConfig struct {
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    8090,
			Address: "0.0.0.0",
		},
		Backends: BackendsConfig{
			URLs: []string{
				"http://localhost:8081",
				"http://localhost:8082",
				"http://localhost:8083",
			},
		},
		Strategy: "roundrobin",
		Circuit: CircuitConfig{
			FailureThreshold: 3,
			Timeout:          10 * time.Second,
		},
		Health: HealthConfig{
			Interval: 5 * time.Second,
			Timeout:  2 * time.Second,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Port:    8090,
			Path:    "/metrics",
		},
	}
}

// LoadConfig loads configuration from a YAML file
// If filePath is empty, it returns the default config
func LoadConfig(filePath string) (*Config, error) {
	cfg := DefaultConfig()

	// If no file path provided, return default
	if filePath == "" {
		return cfg, nil
	}

	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return default config with a warning
			fmt.Printf("Config file %s not found, using defaults\n", filePath)
			return cfg, nil
		}
		return nil, err
	}

	// Read and parse the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// LoadConfigWithEnv loads config and applies environment variable overrides
func LoadConfigWithEnv(filePath string) (*Config, error) {
	cfg, err := LoadConfig(filePath)
	if err != nil {
		return nil, err
	}

	// Override with environment variables if set
	if port := os.Getenv("LB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}
	if addr := os.Getenv("LB_ADDRESS"); addr != "" {
		cfg.Server.Address = addr
	}
	if strategy := os.Getenv("LB_STRATEGY"); strategy != "" {
		cfg.Strategy = strategy
	}
	if interval := os.Getenv("HEALTH_CHECK_INTERVAL"); interval != "" {
		d, _ := time.ParseDuration(interval)
		cfg.Health.Interval = d
	}
	if threshold := os.Getenv("CIRCUIT_FAILURE_THRESHOLD"); threshold != "" {
		fmt.Sscanf(threshold, "%d", &cfg.Circuit.FailureThreshold)
	}

	return cfg, nil
}

// WriteDefaultConfig writes the default configuration to a file
func WriteDefaultConfig(filePath string) error {
	cfg := DefaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Default config written to %s\n", filePath)
	return nil
}
