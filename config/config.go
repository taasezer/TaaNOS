package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/taasezer/TaaNOS/internal/ollama"
	"gopkg.in/yaml.v3"
)

const (
	configDirName  = ".taanos"
	configFileName = "config.yaml"
)

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Version:        1,
		Theme:          "taaos",
		PackageManager: "auto",
		Ollama: OllamaConfig{
			Endpoint:   "http://127.0.0.1:11434",
			Model:      "auto",
			Timeout:    120 * time.Second,
			MaxRetries: 1,
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

// AutoDetectModel verifies if the configured Ollama model is available.
// It ensures Ollama is running if installed, then checks active models.
// If the model is "auto", "llama3.2" (legacy mock default), empty, or not found locally:
// it detects and sets the first available local model from Ollama.
// If Ollama is offline or has no models, it sets "offline (0ms)".
func AutoDetectModel(cfg *Config) {
	ep := ollama.NormalizeEndpoint(cfg.Ollama.Endpoint)
	_ = ollama.EnsureRunning(ep)

	client := &http.Client{Timeout: 800 * time.Millisecond}
	resp, err := client.Get(ep + "/api/tags")
	if err == nil {
		defer resp.Body.Close()
		var data struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		if json.NewDecoder(resp.Body).Decode(&data) == nil && len(data.Models) > 0 {
			// Check if current model is in data.Models
			found := false
			for _, m := range data.Models {
				if m.Name == cfg.Ollama.Model {
					found = true
					break
				}
			}
			if !found || cfg.Ollama.Model == "auto" || cfg.Ollama.Model == "llama3.2" || cfg.Ollama.Model == "" {
				cfg.Ollama.Model = data.Models[0].Name
				_ = Save(cfg)
			}
			return
		}
	}

	// Ollama unreachable or no models
	if cfg.Ollama.Model == "auto" || cfg.Ollama.Model == "llama3.2" || cfg.Ollama.Model == "" {
		cfg.Ollama.Model = "offline (0ms)"
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
			AutoDetectModel(cfg)
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

	AutoDetectModel(cfg)

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
