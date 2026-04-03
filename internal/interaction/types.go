package interaction

// UserDecision represents the user's response to the execution plan.
type UserDecision struct {
	Approved      bool     `json:"approved"`
	SkippedSteps  []int    `json:"skipped_steps"`
	ExecutionMode string   `json:"execution_mode"`
}
