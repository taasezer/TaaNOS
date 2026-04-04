package recovery

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/taasezer/TaaNOS/internal/executor"
	"github.com/taasezer/TaaNOS/internal/planner"
)

// Engine handles rollback of failed plan executions.
type Engine struct{}

// NewEngine creates a recovery engine.
func NewEngine() *Engine {
	return &Engine{}
}

// RollbackResult holds the outcome of a rollback attempt.
type RollbackResult struct {
	StepID    int
	Action    string
	Command   string
	Success   bool
	Error     string
	Duration  time.Duration
}

// Rollback attempts to reverse completed steps after a failure.
// Only rolls back steps that have a defined rollback action.
func (e *Engine) Rollback(
	plan *planner.ExecutionPlan,
	execResult *executor.ExecutionResult,
) []RollbackResult {

	var results []RollbackResult

	// Walk steps in reverse order — only rollback completed steps
	for i := len(execResult.StepResults) - 1; i >= 0; i-- {
		sr := execResult.StepResults[i]
		if sr.Status != "success" {
			continue // Only rollback successful steps
		}

		// Find the original step
		if sr.StepID <= 0 || sr.StepID > len(plan.Steps) {
			continue
		}
		step := plan.Steps[sr.StepID-1]

		if step.RollbackAction == "" {
			continue // No rollback defined
		}

		// Build rollback command
		rollbackDef, err := planner.Lookup(
			getCategory(step.Action),
			step.RollbackAction,
			"", // will try generic first
		)
		if err != nil {
			results = append(results, RollbackResult{
				StepID:  step.ID,
				Action:  step.RollbackAction,
				Success: false,
				Error:   fmt.Sprintf("rollback action not found: %v", err),
			})
			continue
		}

		// Render the rollback command
		rollbackCmd := planner.RenderTemplate(rollbackDef.CommandTemplate, step.Parameters)

		fmt.Printf("  ↩ Rolling back step %d: %s\n", step.ID, rollbackCmd)

		// Execute rollback
		result := e.executeRollback(step.ID, step.RollbackAction, rollbackCmd, step.RequiresRoot)
		results = append(results, result)
	}

	return results
}

// executeRollback runs a single rollback command.
func (e *Engine) executeRollback(stepID int, action, cmdStr string, requiresRoot bool) RollbackResult {
	result := RollbackResult{
		StepID:  stepID,
		Action:  action,
		Command: cmdStr,
	}

	var cmd *exec.Cmd
	if requiresRoot {
		cmd = exec.Command("sudo", "bash", "-c", cmdStr)
	} else {
		cmd = exec.Command("bash", "-c", cmdStr)
	}

	start := time.Now()
	_, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		fmt.Printf("    ✗ Rollback failed: %s\n", err.Error())
	} else {
		result.Success = true
		fmt.Printf("    ✓ Rolled back (%s)\n", result.Duration.Round(time.Millisecond))
	}

	return result
}

// getCategory extracts the category from an action string like "package_management_install".
func getCategory(action string) string {
	// Actions are formatted as "category_action"
	// Find the last underscore and split
	for i := len(action) - 1; i >= 0; i-- {
		if action[i] == '_' {
			return action[:i]
		}
	}
	return action
}
