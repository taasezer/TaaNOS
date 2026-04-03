package logger

import "time"

// Level represents log severity.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelFatal Level = "fatal"
)

// LogEntry is a single structured log record.
type LogEntry struct {
	LogID         string                 `json:"log_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Level         Level                  `json:"level"`
	PipelineStage string                 `json:"pipeline_stage"`
	PlanID        string                 `json:"plan_id,omitempty"`
	Message       string                 `json:"message"`
	Data          map[string]interface{} `json:"data,omitempty"`
}
