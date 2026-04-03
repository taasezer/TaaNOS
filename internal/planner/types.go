package planner

import "time"

// ExecutionStep represents a single step in an execution plan.
type ExecutionStep struct {
	ID              int               `json:"id"`
	Action          string            `json:"action"`
	Description     string            `json:"description"`
	CommandTemplate string            `json:"command_template"`
	Parameters      map[string]string `json:"parameters"`
	RequiresRoot    bool              `json:"requires_root"`
	CanFail         bool              `json:"can_fail"`
	Timeout         time.Duration     `json:"timeout_seconds"`
	RollbackAction  string            `json:"rollback_action,omitempty"`
}

// RiskLevel represents the risk classification of a plan.
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// ExecutionPlan is the deterministic plan built by the Planner.
type ExecutionPlan struct {
	PlanID                  string          `json:"plan_id"`
	IntentSummary           string          `json:"intent_summary"`
	Steps                   []ExecutionStep `json:"steps"`
	EstimatedDurationSeconds int            `json:"estimated_duration_seconds"`
	RiskLevel               RiskLevel       `json:"risk_level"`
}
