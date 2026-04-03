package config

import "time"

// Config is the root configuration structure for TaaNOS.
type Config struct {
	Version   int              `yaml:"version"`
	Ollama    OllamaConfig     `yaml:"ollama"`
	Execution ExecutionConfig  `yaml:"execution"`
	Logging   LoggingConfig    `yaml:"logging"`
	Recovery  RecoveryConfig   `yaml:"recovery"`
	Safety    SafetyConfig     `yaml:"safety"`
}

// OllamaConfig holds Ollama connection settings.
type OllamaConfig struct {
	Endpoint      string        `yaml:"endpoint"`
	Model         string        `yaml:"model"`
	Timeout       time.Duration `yaml:"timeout_seconds"`
	MaxRetries    int           `yaml:"max_retries"`
}

// ExecutionConfig holds execution behavior settings.
type ExecutionConfig struct {
	DefaultMode              string `yaml:"default_mode"`
	RequireApprovalAboveRisk int    `yaml:"require_approval_above_risk"`
	MaxStepTimeout           time.Duration `yaml:"max_step_timeout_seconds"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level         string `yaml:"level"`
	Directory     string `yaml:"directory"`
	MaxLogFiles   int    `yaml:"max_log_files"`
	MaxLogSizeMB  int    `yaml:"max_log_size_mb"`
}

// RecoveryConfig holds recovery behavior settings.
type RecoveryConfig struct {
	Enabled           bool          `yaml:"enabled"`
	MaxRetries        int           `yaml:"max_retries"`
	RetryDelay        time.Duration `yaml:"retry_delay_seconds"`
	AutoRollback      bool          `yaml:"auto_rollback"`
}

// SafetyConfig holds safety constraints.
type SafetyConfig struct {
	BlockedActions           []string `yaml:"blocked_actions"`
	RequireRootConfirmation  bool     `yaml:"require_root_confirmation"`
	MaxRiskScore             int      `yaml:"max_risk_score"`
}
