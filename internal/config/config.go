package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Mate       MateConfig       `yaml:"mate"`
	Navigator  NavigatorConfig  `yaml:"navigator"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Hardware   HardwareConfig   `yaml:"hardware"`
	Logging    LoggingConfig    `yaml:"logging"`
}

// MateConfig contains mate identification
type MateConfig struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// NavigatorConfig contains Fleet Navigator connection settings
type NavigatorConfig struct {
	URL                   string        `yaml:"url"`
	ReconnectInterval     time.Duration `yaml:"reconnect_interval"`
	MaxReconnectAttempts  int           `yaml:"max_reconnect_attempts"`
}

// MonitoringConfig contains monitoring settings
type MonitoringConfig struct {
	Interval time.Duration     `yaml:"interval"`
	Enabled  MonitoringEnabled `yaml:"enabled"`
}

// MonitoringEnabled defines which monitors are active
type MonitoringEnabled struct {
	CPU         bool `yaml:"cpu"`
	Memory      bool `yaml:"memory"`
	Disk        bool `yaml:"disk"`
	Temperature bool `yaml:"temperature"`
	Network     bool `yaml:"network"`
	GPU         bool `yaml:"gpu"`
	Processes   bool `yaml:"processes"`
}

// HardwareConfig contains hardware-specific settings
type HardwareConfig struct {
	CPU         CPUConfig         `yaml:"cpu"`
	Memory      MemoryConfig      `yaml:"memory"`
	Disk        DiskConfig        `yaml:"disk"`
	Temperature TemperatureConfig `yaml:"temperature"`
	Network     NetworkConfig     `yaml:"network"`
	GPU         GPUConfig         `yaml:"gpu"`
}

// CPUConfig contains CPU monitoring settings
type CPUConfig struct {
	CollectPerCore bool `yaml:"collect_per_core"`
}

// MemoryConfig contains memory monitoring settings
type MemoryConfig struct {
	IncludeSwap bool `yaml:"include_swap"`
}

// DiskConfig contains disk monitoring settings
type DiskConfig struct {
	MountPoints    []string `yaml:"mount_points"`
	AlertThreshold int      `yaml:"alert_threshold"`
}

// TemperatureConfig contains temperature monitoring settings
type TemperatureConfig struct {
	Sensors        []string `yaml:"sensors"`
	AlertThreshold float64  `yaml:"alert_threshold"`
}

// NetworkConfig contains network monitoring settings
type NetworkConfig struct {
	Interfaces     []string `yaml:"interfaces"`
	CollectTraffic bool     `yaml:"collect_traffic"`
}

// GPUConfig contains GPU monitoring settings
type GPUConfig struct {
	NvidiaOnly bool `yaml:"nvidia_only"` // Currently only NVIDIA is supported
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// Load reads and parses the configuration file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Mate.ID == "" {
		return fmt.Errorf("mate.id is required")
	}
	if c.Navigator.URL == "" {
		return fmt.Errorf("navigator.url is required")
	}
	if c.Monitoring.Interval <= 0 {
		return fmt.Errorf("monitoring.interval must be positive")
	}
	return nil
}
