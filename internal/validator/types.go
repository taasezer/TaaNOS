package validator

// CheckResult represents the outcome of a single validation check.
type CheckResult struct {
	Check   string `json:"check"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}

// ValidationReport is the output of the Validator stage.
type ValidationReport struct {
	PlanID       string        `json:"plan_id"`
	IsValid      bool          `json:"is_valid"`
	RiskScore    int           `json:"risk_score"`
	MaxRiskScore int           `json:"max_risk_score"`
	Checks       []CheckResult `json:"checks"`
	Blocked      bool          `json:"blocked"`
	BlockReason  string        `json:"block_reason,omitempty"`
	Warnings     []string      `json:"warnings"`
}
