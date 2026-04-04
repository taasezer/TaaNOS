package interaction

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/taasezer/TaaNOS/internal/planner"
	"github.com/taasezer/TaaNOS/internal/validator"
)

// Handler manages user interaction for plan approval.
type Handler struct {
	reader *bufio.Reader
}

// NewHandler creates a new interaction handler.
func NewHandler() *Handler {
	return &Handler{
		reader: bufio.NewReader(os.Stdin),
	}
}

// PresentPlan shows the execution plan and validation report, then collects user decision.
// Behavior depends on execution mode:
//   - explain: show plan, never execute
//   - guided: show each step, ask at each step
//   - auto: show plan, ask once
func (h *Handler) PresentPlan(
	mode string,
	plan *planner.ExecutionPlan,
	valReport *validator.ValidationReport,
	dryRun bool,
) *UserDecision {

	switch mode {
	case "explain":
		return h.explainMode(plan)
	case "auto":
		return h.autoMode(plan, dryRun)
	default: // "guided"
		return h.guidedMode(plan, dryRun)
	}
}

// explainMode shows the full plan without executing.
func (h *Handler) explainMode(plan *planner.ExecutionPlan) *UserDecision {
	fmt.Printf("\n════════════════════════════════════════\n")
	fmt.Printf("  TaaNOS — Explain Mode (no execution)\n")
	fmt.Printf("════════════════════════════════════════\n")

	h.displayPlanSummary(plan)

	for _, step := range plan.Steps {
		h.displayStep(step, len(plan.Steps))
	}

	fmt.Printf("\n  [explain mode — no commands were executed]\n")

	return &UserDecision{
		Approved:      false,
		ExecutionMode: "explain",
	}
}

// guidedMode shows each step and asks for approval.
func (h *Handler) guidedMode(plan *planner.ExecutionPlan, dryRun bool) *UserDecision {
	fmt.Printf("\n════════════════════════════════════════\n")
	fmt.Printf("  TaaNOS — Guided Execution\n")
	fmt.Printf("════════════════════════════════════════\n")

	h.displayPlanSummary(plan)

	if dryRun {
		fmt.Printf("  [DRY RUN — commands will be shown but not executed]\n\n")
	}

	decision := &UserDecision{
		Approved:      true,
		ExecutionMode: "guided",
		SkippedSteps:  []int{},
	}

	for _, step := range plan.Steps {
		h.displayStep(step, len(plan.Steps))

		if dryRun {
			fmt.Printf("    [dry-run: skipped]\n")
			continue
		}

		response := h.promptUser(fmt.Sprintf("  Execute step %d? [Y/n/skip/abort]: ", step.ID))

		switch strings.ToLower(strings.TrimSpace(response)) {
		case "n", "no", "abort", "a":
			fmt.Printf("    ⛔ Aborted by user\n")
			decision.Approved = false
			return decision
		case "skip", "s":
			fmt.Printf("    ⏭ Step skipped\n")
			decision.SkippedSteps = append(decision.SkippedSteps, step.ID)
		case "", "y", "yes":
			// Step approved — will be executed by the executor
		}
	}

	return decision
}

// autoMode shows the full plan and asks once for approval.
func (h *Handler) autoMode(plan *planner.ExecutionPlan, dryRun bool) *UserDecision {
	fmt.Printf("\n════════════════════════════════════════\n")
	fmt.Printf("  TaaNOS — Auto Execution\n")
	fmt.Printf("════════════════════════════════════════\n")

	h.displayPlanSummary(plan)

	for _, step := range plan.Steps {
		h.displayStep(step, len(plan.Steps))
	}

	if dryRun {
		fmt.Printf("\n  [DRY RUN — no commands will be executed]\n")
		return &UserDecision{
			Approved:      false,
			ExecutionMode: "auto",
		}
	}

	response := h.promptUser("\n  Proceed with execution? [Y/n]: ")

	approved := true
	switch strings.ToLower(strings.TrimSpace(response)) {
	case "n", "no":
		approved = false
		fmt.Printf("  ⛔ Execution cancelled by user\n")
	}

	return &UserDecision{
		Approved:      approved,
		ExecutionMode: "auto",
	}
}

// displayPlanSummary shows the plan header.
func (h *Handler) displayPlanSummary(plan *planner.ExecutionPlan) {
	fmt.Printf("\n  Intent:  %s\n", plan.IntentSummary)
	fmt.Printf("  Risk:    %s\n", plan.RiskLevel)
	fmt.Printf("  Steps:   %d\n", len(plan.Steps))
	fmt.Printf("  Est:     %ds\n\n", plan.EstimatedDurationSeconds)
}

// displayStep shows a single step.
func (h *Handler) displayStep(step planner.ExecutionStep, total int) {
	rootTag := ""
	if step.RequiresRoot {
		rootTag = " [sudo]"
	}
	fmt.Printf("  Step %d/%d: %s%s\n", step.ID, total, step.Description, rootTag)
	if step.RequiresRoot {
		fmt.Printf("    → sudo %s\n", step.CommandTemplate)
	} else {
		fmt.Printf("    → %s\n", step.CommandTemplate)
	}
}

// promptUser displays a prompt and reads user input.
func (h *Handler) promptUser(prompt string) string {
	fmt.Print(prompt)
	input, _ := h.reader.ReadString('\n')
	return strings.TrimSpace(input)
}
