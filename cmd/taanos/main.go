package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/taasezer/TaaNOS/config"
	"github.com/taasezer/TaaNOS/internal/history"
	"github.com/taasezer/TaaNOS/internal/logger"
	"github.com/taasezer/TaaNOS/internal/pipeline"
	"github.com/taasezer/TaaNOS/internal/setup"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "version", "--version", "-V":
		printVersion()

	case "help", "--help", "-h":
		printUsage()

	case "status":
		cmdStatus()

	case "config":
		cmdConfig()

	case "history":
		cmdHistory()

	case "init":
		cmdInit()

	case "model":
		cmdModel()

	default:
		// Everything else is treated as natural language input
		cmdRun(os.Args[1:])
	}
}

func printVersion() {
	fmt.Printf("TaaNOS v%s\n", version)
	fmt.Println("A deterministic, pipeline-based, local AI-powered CLI system.")
	fmt.Println("https://github.com/taasezer/TaaNOS")
}

func printUsage() {
	fmt.Println(`
╔══════════════════════════════════════════════════════════╗
║                      TaaNOS CLI                         ║
║   Deterministic AI-Powered System Operations Engine     ║
╚══════════════════════════════════════════════════════════╝

USAGE:
  taanos <natural language input>     Execute an operation
  taanos [command]                    Run a subcommand

COMMANDS:
  version       Show version information
  status        Show TaaNOS system status
  config        Show current configuration
  history       Show execution history
  init          First-time setup wizard (Ollama + model detection)
  model         View or change the current AI model

MODE FLAGS:
  -m, --mode    Execution mode: explain | guided | auto  (default: guided)
  -v, --verbose Show detailed output for each pipeline stage
  -n, --dry-run Full pipeline run, but skip actual command execution
  -f, --force   Bypass non-critical validation warnings
  -l, --log-level  Set log verbosity: debug | info | warn | error

EXAMPLES:
  taanos install nginx
  taanos -m explain upgrade all packages
  taanos --dry-run remove docker
  taanos status

SECURITY:
  AI is used ONLY for intent extraction. All commands are deterministic
  and mapped from a hardcoded action registry. No AI-generated commands.`)
	fmt.Println()
}

func cmdRun(args []string) {
	// Parse flags from args
	mode := pipeline.ModeGuided
	verbose := false
	dryRun := false
	force := false
	var textParts []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m", "--mode":
			if i+1 < len(args) {
				i++
				switch args[i] {
				case "explain":
					mode = pipeline.ModeExplain
				case "guided":
					mode = pipeline.ModeGuided
				case "auto":
					mode = pipeline.ModeAuto
				default:
					fmt.Fprintf(os.Stderr, "taanos: unknown mode '%s' (use: explain, guided, auto)\n", args[i])
					os.Exit(1)
				}
			}
		case "-v", "--verbose":
			verbose = true
		case "-n", "--dry-run":
			dryRun = true
		case "-f", "--force":
			force = true
		default:
			textParts = append(textParts, args[i])
		}
	}

	rawText := strings.Join(textParts, " ")
	if rawText == "" {
		fmt.Fprintln(os.Stderr, "taanos: no input provided")
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging.Directory, logger.Level(cfg.Logging.Level))
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: logger error: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	// Build input
	input := pipeline.RawInput{
		RawText:       rawText,
		ExecutionMode: mode,
		Verbose:       verbose,
		DryRun:        dryRun,
		Force:         force,
		Timestamp:     time.Now(),
	}

	// Run pipeline
	p := pipeline.New(cfg, log)
	if err := p.Run(input); err != nil {
		fmt.Fprintf(os.Stderr, "\ntaanos: pipeline error: %v\n", err)
		os.Exit(1)
	}
}

func cmdStatus() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("╔══════════════════════════════════╗")
	fmt.Println("║         TaaNOS Status            ║")
	fmt.Println("╠══════════════════════════════════╣")
	fmt.Printf("║ Version:    %-20s ║\n", version)
	fmt.Printf("║ Mode:       %-20s ║\n", cfg.Execution.DefaultMode)
	fmt.Printf("║ Ollama:     %-20s ║\n", cfg.Ollama.Endpoint)
	fmt.Printf("║ Model:      %-20s ║\n", cfg.Ollama.Model)
	fmt.Printf("║ Log Level:  %-20s ║\n", cfg.Logging.Level)
	fmt.Printf("║ Config:     %-20s ║\n", config.ConfigPath())
	fmt.Println("╚══════════════════════════════════╝")
}

func cmdConfig() {
	if len(os.Args) > 2 && os.Args[2] == "set" {
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "taanos: usage: taanos config set <key> <value>")
			os.Exit(1)
		}
		fmt.Printf("taanos: config set not yet implemented (Phase 2 skeleton)\n")
		fmt.Printf("  key:   %s\n", os.Args[3])
		fmt.Printf("  value: %s\n", os.Args[4])
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config file: %s\n\n", config.ConfigPath())
	fmt.Printf("ollama:\n")
	fmt.Printf("  endpoint:  %s\n", cfg.Ollama.Endpoint)
	fmt.Printf("  model:     %s\n", cfg.Ollama.Model)
	fmt.Printf("  timeout:   %s\n", cfg.Ollama.Timeout)
	fmt.Printf("  retries:   %d\n", cfg.Ollama.MaxRetries)
	fmt.Printf("\nexecution:\n")
	fmt.Printf("  mode:      %s\n", cfg.Execution.DefaultMode)
	fmt.Printf("  risk_gate: %d\n", cfg.Execution.RequireApprovalAboveRisk)
	fmt.Printf("\nlogging:\n")
	fmt.Printf("  level:     %s\n", cfg.Logging.Level)
	fmt.Printf("  directory: %s\n", cfg.Logging.Directory)
	fmt.Printf("\nsafety:\n")
	fmt.Printf("  max_risk:  %d\n", cfg.Safety.MaxRiskScore)
	fmt.Printf("  root_confirm: %v\n", cfg.Safety.RequireRootConfirmation)
}

func cmdHistory() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	store, err := history.NewStore(cfg.Logging.Directory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: history error: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	count, _ := store.Count()
	if count == 0 {
		fmt.Println("No execution history yet.")
		return
	}

	records, err := store.GetRecent(20)
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: history query error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                  TaaNOS History (%d records)             ║\n", count)
	fmt.Printf("╠══════════════════════════════════════════════════════════╣\n")

	for _, r := range records {
		statusIcon := "✅"
		switch r.Status {
		case "failure":
			statusIcon = "❌"
		case "partial_failure":
			statusIcon = "⚠️ "
		case "explain":
			statusIcon = "📖"
		case "aborted":
			statusIcon = "⛔"
		}

		fmt.Printf("║ %s %-50s ║\n", statusIcon,
			fmt.Sprintf("[%s] %s → %s/%s %s (%dms, risk:%s)",
				r.PlanID[:8], r.CreatedAt.Format("2006-01-02 15:04"),
				r.Category, r.Action, r.Target, r.DurationMs, r.RiskLevel))
	}

	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")
}

func cmdInit() {
	wiz := setup.NewWizard()
	if err := wiz.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\ntaanos: setup error: %v\n", err)
		os.Exit(1)
	}
}

func cmdModel() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Printf("Current model: %s\n", cfg.Ollama.Model)
		fmt.Println("Usage: taanos model <model_name>")
		return
	}

	newModel := os.Args[2]
	cfg.Ollama.Model = newModel

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "taanos: failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Model changed to: %s\n", newModel)
}
