package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/taasezer/TaaNOS/config"
	"github.com/taasezer/TaaNOS/internal/logger"
	"github.com/taasezer/TaaNOS/internal/pipeline"
)

// pipelineDoneMsg is sent when the pipeline finishes.
type pipelineDoneMsg struct {
	input  string
	output string
	err    error
}

// chatDoneMsg is sent when a chat response arrives.
type chatDoneMsg struct {
	input    string
	response string
	err      error
}

// state tracks what the REPL is currently doing.
type state int

const (
	stateIdle     state = iota
	stateThinking
)

// Model is the bubbletea model for the TaaNOS interactive REPL.
type Model struct {
	textInput textinput.Model
	spinner   spinner.Model
	state     state
	cfg       *config.Config
	log       *logger.Logger
	history   []historyEntry
	width     int
	height    int
	quitting  bool
	currentInput string
}

// historyEntry stores one input/output pair in the session.
type historyEntry struct {
	input      string
	output     string
	isErr      bool
	isPipeline bool   // true if output is from AI pipeline (needs rich formatting)
	time       string
}

// Styles
var (
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D4AA")).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(0, 1)

	modelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD43B")).
		Bold(true)

	promptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA")).
		Bold(true)

	dimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#51CF66"))

	thinkingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD43B"))

	inputEchoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C8DFF")).
		Bold(true)

	outputStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#C0C0C0"))

	borderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333"))

	cmdStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA"))

	helpKeyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD43B")).
		Bold(true)

	helpDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))
)

// New creates a new REPL model.
func New(cfg *config.Config, log *logger.Logger) Model {
	ti := textinput.New()
	ti.Placeholder = "Ask TaaNOS anything..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60
	ti.PromptStyle = promptStyle
	ti.Prompt = "❯ "

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = thinkingStyle

	return Model{
		textInput: ti,
		spinner:   s,
		state:     stateIdle,
		cfg:       cfg,
		log:       log,
		history:   []historyEntry{},
		width:     80,
		height:    24,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = min(msg.Width-6, 120)
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlD:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			if m.state != stateIdle {
				return m, nil
			}

			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				return m, nil
			}

			m.textInput.Reset()

			// Handle REPL commands
			lower := strings.ToLower(input)
			switch {
			case lower == "exit" || lower == "quit" || lower == "q":
				m.quitting = true
				return m, tea.Quit

			case lower == "clear" || lower == "cls":
				m.history = []historyEntry{}
				return m, nil

			case lower == "help" || lower == "?":
				m.history = append(m.history, historyEntry{
					input:  input,
					output: m.helpText(),
					time:   time.Now().Format("15:04:05"),
				})
				return m, nil

			case lower == "status":
				m.history = append(m.history, historyEntry{
					input:  input,
					output: m.statusText(),
					time:   time.Now().Format("15:04:05"),
				})
				return m, nil

			case lower == "model":
				m.history = append(m.history, historyEntry{
					input:  input,
					output: fmt.Sprintf("Current model: %s", m.cfg.Ollama.Model),
					time:   time.Now().Format("15:04:05"),
				})
				return m, nil

			case strings.HasPrefix(lower, "model "):
				newModel := strings.TrimSpace(input[6:])
				m.cfg.Ollama.Model = newModel
				if err := config.Save(m.cfg); err != nil {
					m.history = append(m.history, historyEntry{
						input: input, output: "Failed to save: " + err.Error(),
						isErr: true, time: time.Now().Format("15:04:05"),
					})
				} else {
					m.history = append(m.history, historyEntry{
						input: input, output: fmt.Sprintf("✅ Model changed to: %s", newModel),
						time: time.Now().Format("15:04:05"),
					})
				}
				return m, nil

			case lower == "mode":
				m.history = append(m.history, historyEntry{
					input: input,
					output: fmt.Sprintf("Current mode: %s\nAvailable: explain, guided, auto", m.cfg.Execution.DefaultMode),
					time: time.Now().Format("15:04:05"),
				})
				return m, nil

			case strings.HasPrefix(lower, "mode "):
				newMode := strings.TrimSpace(lower[5:])
				switch newMode {
				case "explain", "guided", "auto":
					m.cfg.Execution.DefaultMode = newMode
					m.history = append(m.history, historyEntry{
						input: input, output: fmt.Sprintf("✅ Mode changed to: %s", newMode),
						time: time.Now().Format("15:04:05"),
					})
				default:
					m.history = append(m.history, historyEntry{
						input: input, output: "Unknown mode. Use: explain, guided, auto",
						isErr: true, time: time.Now().Format("15:04:05"),
					})
				}
				return m, nil
			}

			// Route: system query → pipeline, otherwise → chat
			m.state = stateThinking
			m.currentInput = input
			m.textInput.Blur()

			if isSystemQuery(input) {
				return m, tea.Batch(
					m.spinner.Tick,
					m.runPipeline(input),
				)
			}
			return m, tea.Batch(
				m.spinner.Tick,
				m.runChat(input),
			)
		}

	case pipelineDoneMsg:
		m.state = stateIdle
		m.textInput.Focus()

		entry := historyEntry{
			input:      msg.input,
			isPipeline: true,
			time:       time.Now().Format("15:04:05"),
		}
		if msg.err != nil {
			entry.output = msg.output
			entry.isErr = true
		} else {
			entry.output = msg.output
		}
		m.history = append(m.history, entry)
		return m, textinput.Blink

	case chatDoneMsg:
		m.state = stateIdle
		m.textInput.Focus()

		entry := historyEntry{
			input: msg.input,
			time:  time.Now().Format("15:04:05"),
		}
		if msg.err != nil {
			entry.output = msg.err.Error()
			entry.isErr = true
		} else {
			entry.output = msg.response
		}
		m.history = append(m.history, entry)
		return m, textinput.Blink

	case spinner.TickMsg:
		if m.state == stateThinking {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	if m.state == stateIdle {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "\n  " + dimStyle.Render("👋 TaaNOS session ended.") + "\n\n"
	}

	var b strings.Builder

	// Header
	b.WriteString("\n")
	header := headerStyle.Render(" TaaNOS v0.1.0 ")
	model := modelStyle.Render(fmt.Sprintf(" [%s] ", m.cfg.Ollama.Model))
	b.WriteString("  " + header + model + "\n")
	b.WriteString("  " + borderStyle.Render(strings.Repeat("─", min(m.width-4, 76))) + "\n")

	// Calculate how many history entries we can show
	maxHistory := m.height - 8 // leave room for header, input, footer
	startIdx := 0
	if len(m.history) > maxHistory && maxHistory > 0 {
		startIdx = len(m.history) - maxHistory
	}

	// History
	for _, entry := range m.history[startIdx:] {
		// Input echo
		b.WriteString("\n  " + inputEchoStyle.Render("❯ "+entry.input))
		b.WriteString("  " + dimStyle.Render(entry.time) + "\n")

		// Output — pipeline gets rich formatting, commands get plain
		if entry.isPipeline && !entry.isErr {
			formatted := FormatPipelineOutput(entry.output, m.width)
			b.WriteString(formatted)
		} else {
			lines := strings.Split(entry.output, "\n")
			for _, line := range lines {
				if entry.isErr {
					b.WriteString("  " + errorStyle.Render("  ✗ "+line) + "\n")
				} else {
					b.WriteString("  " + outputStyle.Render("  "+line) + "\n")
				}
			}
		}
	}

	// Current state
	if m.state == stateThinking {
		b.WriteString("\n  " + inputEchoStyle.Render("❯ "+m.currentInput) + "\n")
		b.WriteString("  " + thinkingStyle.Render(fmt.Sprintf("  %s %s thinking...",
			m.spinner.View(), m.cfg.Ollama.Model)) + "\n")
	} else {
		b.WriteString("\n  " + m.textInput.View() + "\n")
		b.WriteString("  " + dimStyle.Render("Type 'help' for commands, 'exit' to quit") + "\n")
	}

	return b.String()
}

// runPipeline runs the pipeline and captures its stdout output.
func (m *Model) runPipeline(input string) tea.Cmd {
	cfg := m.cfg
	log := m.log

	return func() tea.Msg {
		// Capture stdout — pipeline uses fmt.Printf
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			return pipelineDoneMsg{input: input, output: err.Error(), err: err}
		}
		os.Stdout = w

		// Parse flags from input
		mode := pipeline.ExecutionMode(cfg.Execution.DefaultMode)
		verbose := false
		dryRun := false
		force := false
		var textParts []string

		words := strings.Fields(input)
		for i := 0; i < len(words); i++ {
			switch words[i] {
			case "-m", "--mode":
				if i+1 < len(words) {
					i++
					switch words[i] {
					case "explain":
						mode = pipeline.ModeExplain
					case "guided":
						mode = pipeline.ModeGuided
					case "auto":
						mode = pipeline.ModeAuto
					}
				}
			case "-v", "--verbose":
				verbose = true
			case "-n", "--dry-run":
				dryRun = true
			case "-f", "--force":
				force = true
			default:
				textParts = append(textParts, words[i])
			}
		}

		rawText := strings.Join(textParts, " ")
		if rawText == "" {
			return pipelineDoneMsg{input: input, output: "No input text provided", err: fmt.Errorf("empty")}
		}

		pInput := pipeline.RawInput{
			RawText:       rawText,
			ExecutionMode: mode,
			Verbose:       verbose,
			DryRun:        dryRun,
			Force:         force,
			Timestamp:     time.Now(),
		}

		p := pipeline.New(cfg, log)
		pipeErr := p.Run(pInput)

		// Restore stdout and read captured output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()

		captured := strings.TrimSpace(buf.String())

		if pipeErr != nil {
			errMsg := pipeErr.Error()
			if captured != "" {
				return pipelineDoneMsg{input: input, output: captured, err: nil}
			}
			return pipelineDoneMsg{input: input, output: errMsg, err: pipeErr}
		}

		if captured == "" {
			captured = "✓ Pipeline completed"
		}

		return pipelineDoneMsg{input: input, output: captured}
	}
}

// helpText returns the REPL help text.
func (m *Model) helpText() string {
	var b strings.Builder
	b.WriteString("TaaNOS Interactive REPL\n")
	b.WriteString(strings.Repeat("─", 50) + "\n")
	b.WriteString("\nCommands:\n")
	cmds := []struct{ key, desc string }{
		{"<any text>", "Ask AI to analyze and suggest commands"},
		{"help, ?", "Show this help menu"},
		{"status", "Show TaaNOS system status"},
		{"model", "Show current AI model"},
		{"model <name>", "Change AI model"},
		{"mode", "Show current execution mode"},
		{"mode <mode>", "Set mode: explain, guided, auto"},
		{"clear, cls", "Clear screen"},
		{"exit, quit, q", "Exit TaaNOS"},
		{"Ctrl+D", "Exit TaaNOS"},
	}
	for _, e := range cmds {
		b.WriteString(fmt.Sprintf("  %-18s %s\n", e.key, e.desc))
	}
	b.WriteString("\nFlags (use with any query):\n")
	flags := []struct{ key, desc string }{
		{"-m explain", "Show plan without executing"},
		{"-m guided", "Ask for confirmation before each step"},
		{"-m auto", "Execute automatically after one confirmation"},
		{"-v, --verbose", "Show detailed pipeline output"},
		{"-n, --dry-run", "Full pipeline, skip execution"},
		{"-f, --force", "Bypass non-critical warnings"},
	}
	for _, e := range flags {
		b.WriteString(fmt.Sprintf("  %-18s %s\n", e.key, e.desc))
	}
	b.WriteString("\nExamples:\n")
	b.WriteString("  install nginx\n")
	b.WriteString("  -m explain check disk space\n")
	b.WriteString("  -m guided upgrade all packages\n")
	return b.String()
}

// statusText returns status info.
func (m *Model) statusText() string {
	return fmt.Sprintf("Model:    %s\nEndpoint: %s\nMode:     %s\nVersion:  0.1.0-dev",
		m.cfg.Ollama.Model, m.cfg.Ollama.Endpoint, m.cfg.Execution.DefaultMode)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isSystemQuery checks if the input looks like a system administration request.
func isSystemQuery(input string) bool {
	lower := strings.ToLower(input)

	// If it has flags, it's definitely a system query
	if strings.Contains(lower, "-m ") || strings.Contains(lower, "--mode") ||
		strings.Contains(lower, "-v") || strings.Contains(lower, "-n") || strings.Contains(lower, "-f") {
		return true
	}

	systemKeywords := []string{
		"install", "remove", "uninstall", "upgrade", "update", "delete",
		"start", "stop", "restart", "enable", "disable",
		"check", "show", "list", "find", "search", "version",
		"create", "mkdir", "touch", "copy", "move", "rename",
		"port", "network", "firewall", "ping", "dns",
		"disk", "memory", "cpu", "ram", "process",
		"service", "package", "docker", "nginx", "apache",
		"systemctl", "apt", "winget", "brew", "pip", "npm",
		"permission", "chmod", "chown", "sudo",
		"log", "config", "configure", "set",
		"kill", "reboot", "shutdown",
		"git", "ssh", "curl", "wget",
	}

	for _, kw := range systemKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	return false
}

// runChat sends a conversational message to Ollama.
func (m *Model) runChat(input string) tea.Cmd {
	cfg := m.cfg

	return func() tea.Msg {
		response, err := Chat(
			cfg.Ollama.Endpoint,
			cfg.Ollama.Model,
			input,
			cfg.Ollama.Timeout,
		)
		if err != nil {
			return chatDoneMsg{input: input, err: err}
		}
		return chatDoneMsg{input: input, response: response}
	}
}
