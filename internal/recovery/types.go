package recovery

// Strategy represents the recovery approach for a failed step.
type Strategy string

const (
	StrategyRetry    Strategy = "retry"
	StrategyRollback Strategy = "rollback"
	StrategySkip     Strategy = "skip"
	StrategyAbort    Strategy = "abort"
)

// RollbackStep describes a single rollback operation.
type RollbackStep struct {
	Action     string            `json:"action"`
	Parameters map[string]string `json:"parameters"`
}

// RecoveryAction describes how to handle a pipeline failure.
type RecoveryAction struct {
	PlanID        string         `json:"plan_id"`
	FailedStepID  int            `json:"failed_step_id"`
	Strategy      Strategy       `json:"strategy"`
	RetryCount    int            `json:"retry_count"`
	MaxRetries    int            `json:"max_retries"`
	RollbackSteps []RollbackStep `json:"rollback_steps"`
}
