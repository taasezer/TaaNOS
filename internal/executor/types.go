package executor

import "time"

// StepResult holds the outcome of executing a single plan step.
type StepResult struct {
	StepID     int           `json:"step_id"`
	Status     string        `json:"status"`
	ExitCode   int           `json:"exit_code"`
	Stdout     string        `json:"stdout"`
	Stderr     string        `json:"stderr"`
	Duration   time.Duration `json:"duration_ms"`
	Timestamp  time.Time     `json:"timestamp"`
}

// ExecutionStatus represents the overall execution outcome.
type ExecutionStatus string

const (
	ExecSuccess        ExecutionStatus = "success"
	ExecPartialFailure ExecutionStatus = "partial_failure"
	ExecFailure        ExecutionStatus = "failure"
	ExecAborted        ExecutionStatus = "aborted"
)

// ExecutionResult is the output of the Executor stage.
type ExecutionResult struct {
	PlanID         string          `json:"plan_id"`
	Status         ExecutionStatus `json:"status"`
	StepsCompleted int             `json:"steps_completed"`
	StepsTotal     int             `json:"steps_total"`
	StepResults    []StepResult    `json:"step_results"`
	TotalDuration  time.Duration   `json:"total_duration_ms"`
}
