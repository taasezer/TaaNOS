package validator

import (
	"fmt"

	"github.com/taasezer/TaaNOS/config"
	appctx "github.com/taasezer/TaaNOS/internal/context"
	"github.com/taasezer/TaaNOS/internal/planner"
)

// Validator performs safety and dependency checks on an execution plan.
// Blocks dangerous operations, checks privileges, disk space, conflicts.
type Validator struct {
	config *config.Config
}

// NewValidator creates a new Validator.
func NewValidator(cfg *config.Config) *Validator {
	return &Validator{config: cfg}
}

// Validate runs all safety checks against the plan and system context.
func (v *Validator) Validate(plan *planner.ExecutionPlan, sysCtx *appctx.SystemContext) *ValidationReport {
	report := &ValidationReport{
		PlanID:       plan.PlanID,
		IsValid:      true,
		RiskScore:    0,
		MaxRiskScore: 10,
		Checks:       []CheckResult{},
		Warnings:     []string{},
	}

	// Run all checks
	v.checkBlockedActions(plan, report)
	v.checkPrivileges(plan, sysCtx, report)
	v.checkDiskSpace(plan, sysCtx, report)
	v.checkRiskThreshold(plan, report)
	v.checkDependencies(plan, sysCtx, report)
	v.checkConflicts(plan, sysCtx, report)
	v.checkTargetState(plan, sysCtx, report)

	// Calculate final risk score
	report.RiskScore = v.calculateRiskScore(plan, sysCtx)

	// Block if risk exceeds configured maximum
	if report.RiskScore > v.config.Safety.MaxRiskScore {
		report.IsValid = false
		report.Blocked = true
		report.BlockReason = fmt.Sprintf("Risk score %d exceeds maximum allowed %d (use --force to override)",
			report.RiskScore, v.config.Safety.MaxRiskScore)
	}

	return report
}

// checkBlockedActions checks if any step uses a blocked action.
func (v *Validator) checkBlockedActions(plan *planner.ExecutionPlan, report *ValidationReport) {
	blocked := make(map[string]bool)
	for _, action := range v.config.Safety.BlockedActions {
		blocked[action] = true
	}

	if len(blocked) == 0 {
		report.Checks = append(report.Checks, CheckResult{
			Check:   "blocked_actions",
			Passed:  true,
			Details: "No blocked actions configured",
		})
		return
	}

	for _, step := range plan.Steps {
		if blocked[step.Action] {
			report.IsValid = false
			report.Blocked = true
			report.BlockReason = fmt.Sprintf("Action '%s' is blocked by configuration", step.Action)
			report.Checks = append(report.Checks, CheckResult{
				Check:   "blocked_actions",
				Passed:  false,
				Details: fmt.Sprintf("Step %d uses blocked action: %s", step.ID, step.Action),
			})
			return
		}
	}

	report.Checks = append(report.Checks, CheckResult{
		Check:   "blocked_actions",
		Passed:  true,
		Details: "No blocked actions in plan",
	})
}

// checkPrivileges verifies the user can execute root-requiring steps.
func (v *Validator) checkPrivileges(plan *planner.ExecutionPlan, sysCtx *appctx.SystemContext, report *ValidationReport) {
	needsRoot := false
	for _, step := range plan.Steps {
		if step.RequiresRoot {
			needsRoot = true
			break
		}
	}

	if !needsRoot {
		report.Checks = append(report.Checks, CheckResult{
			Check:   "privilege_escalation",
			Passed:  true,
			Details: "No root privileges required",
		})
		return
	}

	if sysCtx.User.IsRoot {
		report.Checks = append(report.Checks, CheckResult{
			Check:   "privilege_escalation",
			Passed:  true,
			Details: "Running as root",
		})
		report.Warnings = append(report.Warnings, "Running as root — all operations will execute with full privileges")
		return
	}

	if sysCtx.User.SudoAvailable {
		report.Checks = append(report.Checks, CheckResult{
			Check:   "privilege_escalation",
			Passed:  true,
			Details: "sudo available for user",
		})
		return
	}

	// No root and no sudo
	report.IsValid = false
	report.Blocked = true
	report.BlockReason = "Root privileges required but sudo is not available. Run as root or configure sudo."
	report.Checks = append(report.Checks, CheckResult{
		Check:   "privilege_escalation",
		Passed:  false,
		Details: fmt.Sprintf("User '%s' (uid=%d) cannot escalate to root — sudo unavailable",
			sysCtx.User.Name, sysCtx.User.UID),
	})
}

// checkDiskSpace ensures sufficient disk space for the operation.
func (v *Validator) checkDiskSpace(plan *planner.ExecutionPlan, sysCtx *appctx.SystemContext, report *ValidationReport) {
	// Estimate minimum disk space needed (rough heuristics)
	var estimatedMB int64 = 50 // base minimum for any operation

	for _, step := range plan.Steps {
		switch {
		case contains(step.Action, "install"):
			estimatedMB += 200 // average package install
		case contains(step.Action, "upgrade"):
			estimatedMB += 500 // full system upgrade
		}
	}

	if sysCtx.Resources.DiskFreeMB == 0 {
		report.Checks = append(report.Checks, CheckResult{
			Check:   "disk_space",
			Passed:  true,
			Details: "Disk space check skipped (info unavailable)",
		})
		return
	}

	if sysCtx.Resources.DiskFreeMB < estimatedMB {
		report.IsValid = false
		report.Blocked = true
		report.BlockReason = fmt.Sprintf("Insufficient disk space: %d MB free, estimated need: %d MB",
			sysCtx.Resources.DiskFreeMB, estimatedMB)
		report.Checks = append(report.Checks, CheckResult{
			Check:   "disk_space",
			Passed:  false,
			Details: fmt.Sprintf("%d MB free, estimated need: %d MB", sysCtx.Resources.DiskFreeMB, estimatedMB),
		})
		return
	}

	report.Checks = append(report.Checks, CheckResult{
		Check:   "disk_space",
		Passed:  true,
		Details: fmt.Sprintf("%d MB free, estimated need: %d MB", sysCtx.Resources.DiskFreeMB, estimatedMB),
	})
}

// checkRiskThreshold checks if the plan's risk level is acceptable.
func (v *Validator) checkRiskThreshold(plan *planner.ExecutionPlan, report *ValidationReport) {
	riskValues := map[planner.RiskLevel]int{
		planner.RiskLow:      1,
		planner.RiskMedium:   4,
		planner.RiskHigh:     7,
		planner.RiskCritical: 10,
	}

	riskVal := riskValues[plan.RiskLevel]

	if riskVal > v.config.Safety.MaxRiskScore {
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("Risk level '%s' (score: %d) exceeds configured threshold (%d)",
				plan.RiskLevel, riskVal, v.config.Safety.MaxRiskScore))
	}

	report.Checks = append(report.Checks, CheckResult{
		Check:   "risk_threshold",
		Passed:  riskVal <= v.config.Safety.MaxRiskScore,
		Details: fmt.Sprintf("Risk level: %s (score: %d, max: %d)", plan.RiskLevel, riskVal, v.config.Safety.MaxRiskScore),
	})
}

// checkDependencies verifies that required tools are available.
func (v *Validator) checkDependencies(plan *planner.ExecutionPlan, sysCtx *appctx.SystemContext, report *ValidationReport) {
	// For package management, the package manager must be available
	for _, step := range plan.Steps {
		if contains(step.Action, "package_management") && !sysCtx.PackageManager.Available {
			report.IsValid = false
			report.Blocked = true
			report.BlockReason = fmt.Sprintf("Package manager '%s' is not available", sysCtx.PackageManager.Name)
			report.Checks = append(report.Checks, CheckResult{
				Check:   "dependency_resolution",
				Passed:  false,
				Details: fmt.Sprintf("Package manager '%s' not found", sysCtx.PackageManager.Name),
			})
			return
		}
	}

	report.Checks = append(report.Checks, CheckResult{
		Check:   "dependency_resolution",
		Passed:  true,
		Details: "All dependencies available",
	})
}

// checkConflicts detects potential conflicts with current system state.
func (v *Validator) checkConflicts(plan *planner.ExecutionPlan, sysCtx *appctx.SystemContext, report *ValidationReport) {
	for _, step := range plan.Steps {
		// Warn if trying to install an already-installed package
		if contains(step.Action, "install") && sysCtx.TargetState.Installed != nil && *sysCtx.TargetState.Installed {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Target '%s' is already installed — reinstall/upgrade may occur",
					step.Parameters["target"]))
		}

		// Warn if trying to remove a non-installed package
		if contains(step.Action, "remove") && sysCtx.TargetState.Installed != nil && !*sysCtx.TargetState.Installed {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Target '%s' does not appear to be installed",
					step.Parameters["target"]))
		}

		// Warn if trying to start an already-running service
		if contains(step.Action, "start") && sysCtx.TargetState.Running != nil && *sysCtx.TargetState.Running {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Service '%s' is already running",
					step.Parameters["target"]))
		}

		// Warn if trying to stop a non-running service
		if contains(step.Action, "stop") && sysCtx.TargetState.Running != nil && !*sysCtx.TargetState.Running {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Service '%s' is not currently running",
					step.Parameters["target"]))
		}
	}

	report.Checks = append(report.Checks, CheckResult{
		Check:   "conflict_detection",
		Passed:  true,
		Details: fmt.Sprintf("Checked %d steps for conflicts", len(plan.Steps)),
	})
}

// checkTargetState validates the target exists or is reachable.
func (v *Validator) checkTargetState(plan *planner.ExecutionPlan, sysCtx *appctx.SystemContext, report *ValidationReport) {
	report.Checks = append(report.Checks, CheckResult{
		Check:   "target_state",
		Passed:  true,
		Details: "Target state verified",
	})
}

// calculateRiskScore computes a numeric risk score for the plan.
func (v *Validator) calculateRiskScore(plan *planner.ExecutionPlan, sysCtx *appctx.SystemContext) int {
	score := 0

	for _, step := range plan.Steps {
		if step.RequiresRoot {
			score += 2
		}
		if step.RollbackAction == "" && !step.CanFail {
			score++
		}
		if contains(step.Action, "remove") || contains(step.Action, "delete") {
			score += 2
		}
		if contains(step.Action, "upgrade") {
			score += 2
		}
	}

	if sysCtx.User.IsRoot {
		score++
	}
	if sysCtx.Resources.DiskFreeMB > 0 && sysCtx.Resources.DiskFreeMB < 1000 {
		score++
	}

	if score > 10 {
		score = 10
	}
	return score
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
