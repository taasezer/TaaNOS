package planner

import (
	"fmt"

	"github.com/google/uuid"
	appctx "github.com/taasezer/TaaNOS/internal/context"
	"github.com/taasezer/TaaNOS/internal/intent"
)

// Planner converts an IntentResult + SystemContext into a deterministic ExecutionPlan.
// NO AI. NO randomness. Same input + same context = same plan, always.
type Planner struct{}

// NewPlanner creates a new Planner.
func NewPlanner() *Planner {
	return &Planner{}
}

// BuildPlan creates an ExecutionPlan from the intent and system context.
func (p *Planner) BuildPlan(intentResult *intent.IntentResult, sysCtx *appctx.SystemContext) (*ExecutionPlan, error) {
	category := string(intentResult.Category)
	action := string(intentResult.Action)
	target := intentResult.Parameters.Target
	pkgManager := sysCtx.PackageManager.Name

	// Lookup the primary action in the registry
	primaryDef, err := Lookup(category, action, pkgManager)
	if err != nil {
		return nil, fmt.Errorf("unsupported operation: %w", err)
	}

	plan := &ExecutionPlan{
		PlanID:        uuid.New().String(),
		IntentSummary: intentResult.Intent,
		Steps:         []ExecutionStep{},
		RiskLevel:     RiskLow,
	}

	stepID := 1
	totalDuration := 0

	// Add pre-steps (e.g., apt update before apt install)
	for _, preAction := range primaryDef.PreSteps {
		var preDef *ActionDef
		var preErr error

		switch preAction {
		case "update_index":
			preDef, preErr = LookupUpdateIndex(pkgManager)
		default:
			preDef, preErr = Lookup(category, preAction, pkgManager)
		}

		if preErr != nil {
			// Pre-step not found — skip but don't fail
			continue
		}

		params := map[string]string{"target": target}
		rendered := RenderTemplate(preDef.CommandTemplate, params)

		plan.Steps = append(plan.Steps, ExecutionStep{
			ID:              stepID,
			Action:          fmt.Sprintf("%s_%s", category, preAction),
			Description:     preDef.Description,
			CommandTemplate: rendered,
			Parameters:      params,
			RequiresRoot:    preDef.RequiresRoot,
			CanFail:         preDef.CanFail,
			Timeout:         preDef.Timeout,
		})
		stepID++
		totalDuration += int(preDef.Timeout.Seconds())
	}

	// Add the primary action
	params := map[string]string{"target": target}
	for i, opt := range intentResult.Parameters.Options {
		params[fmt.Sprintf("option_%d", i)] = opt
	}
	rendered := RenderTemplate(primaryDef.CommandTemplate, params)

	plan.Steps = append(plan.Steps, ExecutionStep{
		ID:              stepID,
		Action:          fmt.Sprintf("%s_%s", category, action),
		Description:     primaryDef.Description,
		CommandTemplate: rendered,
		Parameters:      params,
		RequiresRoot:    primaryDef.RequiresRoot,
		CanFail:         primaryDef.CanFail,
		Timeout:         primaryDef.Timeout,
		RollbackAction:  primaryDef.RollbackAction,
	})
	stepID++
	totalDuration += int(primaryDef.Timeout.Seconds())

	// Add post-install steps for service-related packages
	if category == "package_management" && action == "install" {
		// After installing, enable the service if it's a daemon package
		enableDef, err := Lookup("service_management", "enable", "")
		if err == nil {
			plan.Steps = append(plan.Steps, ExecutionStep{
				ID:              stepID,
				Action:          "service_management_enable",
				Description:     fmt.Sprintf("Enable %s to start on boot", target),
				CommandTemplate: RenderTemplate(enableDef.CommandTemplate, params),
				Parameters:      params,
				RequiresRoot:    enableDef.RequiresRoot,
				CanFail:         true, // Not all packages have services
				Timeout:         enableDef.Timeout,
				RollbackAction:  enableDef.RollbackAction,
			})
			stepID++
			totalDuration += int(enableDef.Timeout.Seconds())
		}
	}

	plan.EstimatedDurationSeconds = totalDuration
	plan.RiskLevel = p.assessRisk(plan, sysCtx)

	return plan, nil
}

// assessRisk calculates the risk level based on plan contents and context.
func (p *Planner) assessRisk(plan *ExecutionPlan, sysCtx *appctx.SystemContext) RiskLevel {
	score := 0

	for _, step := range plan.Steps {
		// Root operations are inherently riskier
		if step.RequiresRoot {
			score += 2
		}

		// Non-recoverable steps (no rollback)
		if step.RollbackAction == "" && !step.CanFail {
			score++
		}

		// Certain actions are inherently risky
		switch {
		case contains(step.Action, "remove"):
			score += 2
		case contains(step.Action, "delete"):
			score += 3
		case contains(step.Action, "upgrade"):
			score += 2
		}
	}

	// Context-based risk adjustments
	if sysCtx.User.IsRoot {
		score++ // Running as root is riskier
	}
	if sysCtx.Resources.DiskFreeMB < 1000 {
		score++ // Low disk space
	}

	switch {
	case score <= 3:
		return RiskLow
	case score <= 6:
		return RiskMedium
	case score <= 8:
		return RiskHigh
	default:
		return RiskCritical
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
