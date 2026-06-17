package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/taasezer/TaaNOS/config"
	"github.com/taasezer/TaaNOS/internal/logger"
	"github.com/taasezer/TaaNOS/internal/pipeline"
	"github.com/taasezer/TaaNOS/internal/setup"
	"github.com/taasezer/TaaNOS/internal/tui"
)

const version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		cmdREPL()
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
  history       Show past chat sessions (history <id> for detail)
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

func cmdREPL() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Logging.Directory, logger.Level(cfg.Logging.Level))
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: logger error: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	m := tui.New(cfg, log)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "taanos: TUI error: %v\n", err)
		os.Exit(1)
	}
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
	sessions, err := tui.LoadSessions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: history error: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No chat sessions yet. Start one with: taanos")
		return
	}

	// If a session ID is provided, show that session's detail
	if len(os.Args) > 2 {
		sessionID := os.Args[2]
		for _, s := range sessions {
			if s.ID == sessionID {
				fmt.Printf("\n╔══════════════════════════════════════════════════════════╗\n")
				fmt.Printf("║  Session: %-46s ║\n", s.StartedAt)
				fmt.Printf("╠══════════════════════════════════════════════════════════╣\n")
				for _, h := range s.History {
					icon := "💬"
					if h.IsPipeline {
						icon = "⚙️ "
					}
					if h.IsErr {
						icon = "❌"
					}
					fmt.Printf("║ %s %-52s ║\n", icon, truncate(h.Input, 50))
				}
				fmt.Printf("╠══════════════════════════════════════════════════════════╣\n")
				fmt.Printf("║  Resume: taanos  (session memory auto-loaded)           ║\n")
				fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")
				return
			}
		}
		fmt.Fprintf(os.Stderr, "taanos: session '%s' not found\n", sessionID)
		os.Exit(1)
	}

	// List all sessions
	fmt.Printf("\n╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              TaaNOS Chat Sessions (%d)                  ║\n", len(sessions))
	fmt.Printf("╠══════════════════════════════════════════════════════════╣\n")

	for i := len(sessions) - 1; i >= 0; i-- {
		s := sessions[i]
		msgCount := len(s.History)
		lastMsg := ""
		if msgCount > 0 {
			lastMsg = truncate(s.History[msgCount-1].Input, 30)
		}
		fmt.Printf("║  %-14s  %s  %2d msgs  %-18s ║\n",
			s.ID, s.StartedAt[:16], msgCount, lastMsg)
	}

	fmt.Printf("╠══════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Detail:  taanos history <session_id>                   ║\n")
	fmt.Printf("║  Resume:  taanos  (last session auto-loaded)            ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
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
