package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Logger provides structured JSONL logging for TaaNOS.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	level    Level
	planID   string
	minLevel int
}

var levelPriority = map[Level]int{
	LevelDebug: 0,
	LevelInfo:  1,
	LevelWarn:  2,
	LevelError: 3,
	LevelFatal: 4,
}

// New creates a new Logger that writes JSONL to the specified directory.
func New(logDir string, level Level) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := filepath.Join(logDir, fmt.Sprintf("taanos_%s.jsonl", time.Now().Format("2006-01-02")))
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		file:     f,
		level:    level,
		minLevel: levelPriority[level],
	}, nil
}

// SetPlanID sets the current plan ID for correlation.
func (l *Logger) SetPlanID(planID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.planID = planID
}

// Log writes a structured log entry.
func (l *Logger) Log(level Level, stage, message string, data map[string]interface{}) {
	if levelPriority[level] < l.minLevel {
		return
	}

	entry := LogEntry{
		LogID:         uuid.New().String(),
		Timestamp:     time.Now().UTC(),
		Level:         level,
		PipelineStage: stage,
		PlanID:        l.planID,
		Message:       message,
		Data:          data,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: failed to marshal log entry: %v\n", err)
		return
	}

	l.file.Write(jsonBytes)
	l.file.Write([]byte("\n"))
}

// Debug logs at debug level.
func (l *Logger) Debug(stage, message string, data map[string]interface{}) {
	l.Log(LevelDebug, stage, message, data)
}

// Info logs at info level.
func (l *Logger) Info(stage, message string, data map[string]interface{}) {
	l.Log(LevelInfo, stage, message, data)
}

// Warn logs at warn level.
func (l *Logger) Warn(stage, message string, data map[string]interface{}) {
	l.Log(LevelWarn, stage, message, data)
}

// Error logs at error level.
func (l *Logger) Error(stage, message string, data map[string]interface{}) {
	l.Log(LevelError, stage, message, data)
}

// Fatal logs at fatal level.
func (l *Logger) Fatal(stage, message string, data map[string]interface{}) {
	l.Log(LevelFatal, stage, message, data)
}

// Close flushes and closes the log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
