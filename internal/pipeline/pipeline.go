package pipeline

import (
	"fmt"
	"time"

	"github.com/taasezer/TaaNOS/config"
	appctx "github.com/taasezer/TaaNOS/internal/context"
	"github.com/taasezer/TaaNOS/internal/intent"
	"github.com/taasezer/TaaNOS/internal/logger"
	osutil "github.com/taasezer/TaaNOS/internal/os"
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

	// ── Stage 4–9: Not yet implemented ──
	stages := []struct {
		name Stage
		desc string
	}{
		{StagePlanning, "Plan Building"},
		{StageValidation, "Safety Validation"},
		{StageInteraction, "User Interaction"},
		{StageExecution, "Execution"},
		{StageLogging, "Logging"},
		{StageRecovery, "Recovery"},
	}

	for _, s := range stages {
		p.logger.Debug(string(s.name), fmt.Sprintf("Stage stub: %s (not yet implemented)", s.desc), nil)
	}

	fmt.Printf("\n📋 Remaining pipeline stages (Phase 5–8) not yet implemented.\n")
	fmt.Printf("   Intent + Context stages completed successfully.\n")

	p.logger.Info(string(StageLogging), "Pipeline completed (intent + context)", nil)
	return nil
}

