package pipeline

import (
	"fmt"
	"time"

	"github.com/taasezer/TaaNOS/config"
	appctx "github.com/taasezer/TaaNOS/internal/context"
	"github.com/taasezer/TaaNOS/internal/executor"
	"github.com/taasezer/TaaNOS/internal/intent"
	"github.com/taasezer/TaaNOS/internal/interaction"
	"github.com/taasezer/TaaNOS/internal/logger"
	osutil "github.com/taasezer/TaaNOS/internal/os"
	"github.com/taasezer/TaaNOS/internal/planner"
	"github.com/taasezer/TaaNOS/internal/validator"
)

// ExecutionMode defines how TaaNOS presents and executes plans.
type ExecutionMode string

const (
	ModeExplain ExecutionMode = "explain"
	ModeGuided  ExecutionMode = "guided"
	ModeAuto    ExecutionMode = "auto"
)

// RawInput is the parsed CLI input entering the pipeline.
type RawInput struct {
	RawText       string
	ExecutionMode ExecutionMode
	Verbose       bool
	DryRun        bool
	Force         bool
	Timestamp     time.Time
}

// Pipeline orchestrates the sequential execution of all stages.
type Pipeline struct {
	config    *config.Config
	logger    *logger.Logger
	extractor *intent.Extractor
	platform  osutil.Platform
}

// New creates a new Pipeline with the given configuration.
func New(cfg *config.Config, log *logger.Logger) *Pipeline {
	ext := intent.NewExtractor(intent.ExtractorConfig{
		Endpoint:   cfg.Ollama.Endpoint,
		Model:      cfg.Ollama.Model,
		Timeout:    cfg.Ollama.Timeout,
		MaxRetries: cfg.Ollama.MaxRetries,
	})

	// Detect platform (may fail on unsupported OS — handled gracefully)
	platform, err := osutil.Detect()
	if err != nil {
		log.Warn("pipeline", "Platform detection failed (non-Linux OS?)", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return &Pipeline{
		config:    cfg,
		logger:    log,
		extractor: ext,
		platform:  platform,
	}
}

// Run executes the full pipeline for the given input.
func (p *Pipeline) Run(input RawInput) error {
	p.logger.Info(string(StageParsing), "Pipeline started", map[string]interface{}{
		"raw_text":       input.RawText,
		"execution_mode": string(input.ExecutionMode),
		"dry_run":        input.DryRun,
	})

	// ── Stage 1: Input Parsing (already done by CLI) ──
	if input.Verbose {
		fmt.Printf("  ⚙ Stage: Input Parsing\n")
		fmt.Printf("    Input: %q\n", input.RawText)
		fmt.Printf("    Mode:  %s\n\n", input.ExecutionMode)
	}

	// ── Stage 2: Intent Extraction (Ollama) ──
	p.logger.Info(string(StageIntent), "Extracting intent via Ollama", nil)

	if input.Verbose {
		fmt.Printf("  ⚙ Stage: Intent Extraction\n")
		fmt.Printf("    Model: %s\n", p.config.Ollama.Model)
		fmt.Printf("    Endpoint: %s\n", p.config.Ollama.Endpoint)
	}

	fmt.Printf("\n🧠 Analyzing intent...\n")

	intentResult, err := p.extractor.Extract(input.RawText)
	if err != nil {
		p.logger.Error(string(StageIntent), "Intent extraction failed",
			map[string]interface{}{"error": err.Error()})
		return NewPipelineError(ErrOllamaConn, string(StageIntent), err.Error(), err)
	}

	p.logger.Info(string(StageIntent), "Intent extracted successfully", map[string]interface{}{
		"intent":     intentResult.Intent,
		"category":   string(intentResult.Category),
		"action":     string(intentResult.Action),
		"target":     intentResult.Parameters.Target,
		"confidence": intentResult.Confidence,
		"time_ms":    intentResult.ExtractionTimeMs,
	})

	// Check confidence threshold
	if intentResult.Confidence < 0.5 {
		msg := fmt.Sprintf("Low confidence (%.0f%%) — intent is ambiguous. Please rephrase your request.",
			intentResult.Confidence*100)
		fmt.Printf("\n⚠️  %s\n", msg)
		fmt.Printf("   Detected: %s\n", intentResult.Intent)
		return NewPipelineError(ErrIntentUnknown, string(StageIntent), msg, nil)
	}

	// Check for unknown category
	if intentResult.Category == intent.CategoryUnknown {
		msg := "Input does not appear to be a system administration task."
		fmt.Printf("\n❌ %s\n", msg)
		fmt.Printf("   Detected: %s\n", intentResult.Intent)
		return NewPipelineError(ErrIntentUnknown, string(StageIntent), msg, nil)
	}

	// Display extracted intent
	fmt.Printf("\n✅ Intent Extracted\n")
	fmt.Printf("   Description: %s\n", intentResult.Intent)
	fmt.Printf("   Category:    %s\n", intentResult.Category)
	fmt.Printf("   Action:      %s\n", intentResult.Action)
	fmt.Printf("   Target:      %s\n", intentResult.Parameters.Target)
	fmt.Printf("   Confidence:  %.0f%%\n", intentResult.Confidence*100)
	fmt.Printf("   Time:        %dms\n", intentResult.ExtractionTimeMs)

	if input.Verbose {
		fmt.Printf("   Raw LLM:     %s\n", intentResult.RawLLMResponse)
	}

	// ── AI BOUNDARY — NO AI PAST THIS POINT ──
	fmt.Printf("\n   ────────────────────────────────────\n")
	fmt.Printf("   ⛔ AI BOUNDARY — deterministic pipeline from here\n")
	fmt.Printf("   ────────────────────────────────────\n")

	// ── Stage 3: Context Analysis ──
	p.logger.Info(string(StageContext), "Analyzing system context", nil)
	fmt.Printf("\n🔍 Analyzing system context...\n")

	if p.platform == nil {
		msg := "Platform not available — context analysis requires a supported OS (Linux)"
		p.logger.Error(string(StageContext), msg, nil)
		return NewPipelineError(ErrContextOS, string(StageContext), msg, nil)
	}

	analyzer := appctx.NewAnalyzer(p.platform)
	sysCtx, err := analyzer.Analyze(
		string(intentResult.Category),
		string(intentResult.Action),
		intentResult.Parameters.Target,
	)
	if err != nil {
		p.logger.Error(string(StageContext), "Context analysis failed",
			map[string]interface{}{"error": err.Error()})
		return NewPipelineError(ErrContextOS, string(StageContext), err.Error(), err)
	}

	// Validate critical context
	if !sysCtx.PackageManager.Available && intentResult.Category == intent.CategoryPackageManagement {
		msg := fmt.Sprintf("Package manager '%s' is not available on this system", sysCtx.PackageManager.Name)
		p.logger.Error(string(StageContext), msg, nil)
		return NewPipelineError(ErrContextPkg, string(StageContext), msg, nil)
	}

	// Display context
	fmt.Printf("\n✅ System Context\n")
	fmt.Printf("   OS:          %s %s %s\n", sysCtx.OS.Distro, sysCtx.OS.Version, sysCtx.OS.Arch)
	fmt.Printf("   Kernel:      %s\n", sysCtx.OS.Kernel)
	fmt.Printf("   Pkg Manager: %s (available: %v)\n", sysCtx.PackageManager.Name, sysCtx.PackageManager.Available)
	fmt.Printf("   User:        %s (uid=%d, root=%v, sudo=%v)\n",
		sysCtx.User.Name, sysCtx.User.UID, sysCtx.User.IsRoot, sysCtx.User.SudoAvailable)
	fmt.Printf("   Disk Free:   %d MB\n", sysCtx.Resources.DiskFreeMB)
	fmt.Printf("   Memory Free: %d MB\n", sysCtx.Resources.MemoryFreeMB)

	// Display target state if probed
	if sysCtx.TargetState.Installed != nil {
		fmt.Printf("   Target:      installed=%v\n", *sysCtx.TargetState.Installed)
	}
	if sysCtx.TargetState.Running != nil {
		fmt.Printf("   Target:      running=%v\n", *sysCtx.TargetState.Running)
	}
	if sysCtx.TargetState.Enabled != nil {
		fmt.Printf("   Target:      enabled=%v\n", *sysCtx.TargetState.Enabled)
	}

	p.logger.Info(string(StageContext), "Context analysis complete", map[string]interface{}{
		"os":          sysCtx.OS.Distro,
		"pkg_manager": sysCtx.PackageManager.Name,
		"user":        sysCtx.User.Name,
		"disk_free":   sysCtx.Resources.DiskFreeMB,
		"memory_free": sysCtx.Resources.MemoryFreeMB,
	})

	// ── Stage 4: Planning ──
	p.logger.Info(string(StagePlanning), "Building execution plan", nil)
	fmt.Printf("\n📐 Building execution plan...\n")

	pln := planner.NewPlanner()
	execPlan, err := pln.BuildPlan(intentResult, sysCtx)
	if err != nil {
		p.logger.Error(string(StagePlanning), "Plan building failed",
			map[string]interface{}{"error": err.Error()})
		return NewPipelineError(ErrPlanUnsupported, string(StagePlanning), err.Error(), err)
	}

	p.logger.SetPlanID(execPlan.PlanID)

	// Display the plan
	fmt.Printf("\n✅ Execution Plan\n")
	fmt.Printf("   Plan ID:   %s\n", execPlan.PlanID[:8])
	fmt.Printf("   Summary:   %s\n", execPlan.IntentSummary)
	fmt.Printf("   Steps:     %d\n", len(execPlan.Steps))
	fmt.Printf("   Risk:      %s\n", execPlan.RiskLevel)
	fmt.Printf("   Est. Time: %ds\n", execPlan.EstimatedDurationSeconds)

	for _, step := range execPlan.Steps {
		rootTag := ""
		if step.RequiresRoot {
			rootTag = " [sudo]"
		}
		canFailTag := ""
		if step.CanFail {
			canFailTag = " (can fail)"
		}
		fmt.Printf("\n   Step %d: %s%s%s\n", step.ID, step.Description, rootTag, canFailTag)
		fmt.Printf("     → %s\n", step.CommandTemplate)
		fmt.Printf("     Timeout: %s\n", step.Timeout)
		if step.RollbackAction != "" {
			fmt.Printf("     Rollback: %s\n", step.RollbackAction)
		}
	}

	p.logger.Info(string(StagePlanning), "Execution plan built", map[string]interface{}{
		"plan_id":    execPlan.PlanID,
		"steps":      len(execPlan.Steps),
		"risk_level": string(execPlan.RiskLevel),
	})

	// ── Stage 5: Validation ──
	p.logger.Info(string(StageValidation), "Running safety checks", nil)
	fmt.Printf("\n🛡️  Running safety checks...\n")

	val := validator.NewValidator(p.config)
	valReport := val.Validate(execPlan, sysCtx)

	// Display validation report
	if valReport.IsValid {
		fmt.Printf("\n✅ Validation Passed\n")
	} else {
		fmt.Printf("\n❌ Validation FAILED\n")
	}
	fmt.Printf("   Risk Score: %d/%d\n", valReport.RiskScore, valReport.MaxRiskScore)

	for _, check := range valReport.Checks {
		statusIcon := "✅"
		if !check.Passed {
			statusIcon = "❌"
		}
		fmt.Printf("   %s %s: %s\n", statusIcon, check.Check, check.Details)
	}

	for _, warning := range valReport.Warnings {
		fmt.Printf("   ⚠️  %s\n", warning)
	}

	p.logger.Info(string(StageValidation), "Validation complete", map[string]interface{}{
		"is_valid":   valReport.IsValid,
		"risk_score": valReport.RiskScore,
		"blocked":    valReport.Blocked,
	})

	// Block execution if validation failed (unless --force)
	if valReport.Blocked && !input.Force {
		msg := fmt.Sprintf("Plan blocked: %s", valReport.BlockReason)
		fmt.Printf("\n⛔ %s\n", msg)
		return NewPipelineError(ErrValidRisk, string(StageValidation), msg, nil)
	}

	if valReport.Blocked && input.Force {
		fmt.Printf("\n⚠️  Validation warnings overridden with --force\n")
	}

	// ── Stage 6: Interaction ──
	p.logger.Info(string(StageInteraction), "Presenting plan to user", nil)

	handler := interaction.NewHandler()
	decision := handler.PresentPlan(
		string(input.ExecutionMode),
		execPlan,
		valReport,
		input.DryRun,
	)

	p.logger.Info(string(StageInteraction), "User decision received", map[string]interface{}{
		"approved":      decision.Approved,
		"mode":          decision.ExecutionMode,
		"skipped_steps": decision.SkippedSteps,
	})

	// Explain mode or user rejected — stop here
	if !decision.Approved {
		if input.ExecutionMode == ModeExplain {
			p.logger.Info(string(StageInteraction), "Explain mode — no execution", nil)
			return nil
		}
		return NewPipelineError(ErrUserAbort, string(StageInteraction), "Execution cancelled by user", nil)
	}

	// ── Stage 7: Execution ──
	p.logger.Info(string(StageExecution), "Starting execution", nil)
	fmt.Printf("\n▶️  Executing...\n\n")

	eng := executor.NewEngine(input.DryRun)
	execResult := eng.Execute(execPlan, decision, string(input.ExecutionMode))

	// Display result summary
	fmt.Printf("\n════════════════════════════════════════\n")
	switch execResult.Status {
	case executor.ExecSuccess:
		fmt.Printf("✅ All steps completed successfully.\n")
	case executor.ExecPartialFailure:
		fmt.Printf("⚠️  Partial completion: %d/%d steps succeeded.\n",
			execResult.StepsCompleted, execResult.StepsTotal)
	case executor.ExecFailure:
		fmt.Printf("❌ Execution failed at step %d.\n", len(execResult.StepResults))
	case executor.ExecAborted:
		fmt.Printf("⛔ Execution aborted by user.\n")
	}
	fmt.Printf("  Total time: %s\n", execResult.TotalDuration.Round(time.Millisecond))
	fmt.Printf("════════════════════════════════════════\n")

	p.logger.Info(string(StageExecution), "Execution complete", map[string]interface{}{
		"status":          string(execResult.Status),
		"steps_completed": execResult.StepsCompleted,
		"steps_total":     execResult.StepsTotal,
		"duration_ms":     execResult.TotalDuration.Milliseconds(),
	})

	if execResult.Status == executor.ExecFailure {
		return NewPipelineError(ErrExecFail, string(StageExecution),
			fmt.Sprintf("Execution failed: %d/%d steps completed",
				execResult.StepsCompleted, execResult.StepsTotal), nil)
	}

	p.logger.Info(string(StageLogging), "Pipeline completed successfully", nil)
	return nil
}

