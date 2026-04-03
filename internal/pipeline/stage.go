package pipeline

// Stage represents the current pipeline stage for tracking.
type Stage string

const (
	StageParsing     Stage = "parsing"
	StageIntent      Stage = "intent_extraction"
	StageContext     Stage = "context_analysis"
	StagePlanning    Stage = "planning"
	StageValidation  Stage = "validation"
	StageInteraction Stage = "interaction"
	StageExecution   Stage = "execution"
	StageLogging     Stage = "logging"
	StageRecovery    Stage = "recovery"
)

// Status represents the pipeline execution status.
type Status string

const (
	StatusIdle       Status = "idle"
	StatusRunning    Status = "running"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusAborted    Status = "aborted"
	StatusBlocked    Status = "blocked"
)

// StageResult holds the outcome of a single pipeline stage.
type StageResult struct {
	Stage   Stage
	Status  Status
	Data    interface{}
	Error   *PipelineError
}
