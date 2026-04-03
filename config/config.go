package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	configDirName  = ".taanos"
	configFileName = "config.yaml"
)

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Version: 1,
		Ollama: OllamaConfig{
			Endpoint:   "http://localhost:11434",
			Model:      "llama3.2",
			Timeout:    30 * time.Second,
			MaxRetries: 2,
		},
		Execution: ExecutionConfig{
			DefaultMode:              "guided",
			RequireApprovalAboveRisk: 5,
			MaxStepTimeout:           600 * time.Second,
		},
		Logging: LoggingConfig{
			Level:        "info",
			Directory:    filepath.Join(homeDir(), configDirName, "logs"),
			MaxLogFiles:  100,
			MaxLogSizeMB: 50,
		},
		Recovery: RecoveryConfig{
			Enabled:      true,
			MaxRetries:   3,
			RetryDelay:   5 * time.Second,
			AutoRollback: false,
		},
		Safety: SafetyConfig{
			BlockedActions:          []string{},
			RequireRootConfirmation: true,
			MaxRiskScore:            7,
		},
	}
}

// Load reads the config from ~/.taanos/config.yaml.
// If the file doesn't exist, returns defaults and creates the file.
func Load() (*Config, error) {
	cfgPath := ConfigPath()

	cfg := Default()

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config file
			if writeErr := Save(cfg); writeErr != nil {
				return cfg, fmt.Errorf("failed to create default config: %w", writeErr)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	cfgPath := ConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	header := []byte("# TaaNOS Configuration\n# Auto-generated — edit with care\n\n")
	data = append(header, data...)

	return os.WriteFile(cfgPath, data, 0644)
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(homeDir(), configDirName, configFileName)
}

// DataDir returns the TaaNOS data directory.
func DataDir() string {
	return filepath.Join(homeDir(), configDirName)
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}
