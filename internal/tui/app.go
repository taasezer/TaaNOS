package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
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

// warmupDoneMsg signals that model warmup completed.
type warmupDoneMsg struct{}

// state tracks what the REPL is currently doing.
type state int

const (
	stateIdle state = iota
	stateThinking
	stateConfirm // waiting for y/n to execute commands
)

// Model is the bubbletea model for the TaaNOS interactive REPL.
type Model struct {
	textInput    textinput.Model
	spinner      spinner.Model
	state        state
	cfg          *config.Config
	log          *logger.Logger
	history      []historyEntry
	conversation []ConversationEntry // session memory for AI context
	pendingCmds  []string            // commands waiting for y/n approval
	pendingInput string              // original input for pending commands
	scrollOffset int                 // 0 = bottom (latest), positive = scrolled up
	showWelcome  bool                // show welcome banner until first input
	width        int
	height       int
	quitting     bool
	currentInput string
}

// execDoneMsg is sent when command execution finishes.
type execDoneMsg struct {
	input  string
	output string
	err    error
}

// historyEntry stores one input/output pair in the session.
type historyEntry struct {
	input      string
	output     string
	isErr      bool
	isPipeline bool // true if output is from AI pipeline (needs rich formatting)
	time       string
}

// Styles — Claude Code inspired theme with penguin mascot
var (
	// Brand colors
	brandColor  = lipgloss.Color("#00D4AA")
	accentColor = lipgloss.Color("#7C8DFF")
	warnColor   = lipgloss.Color("#FFD43B")
	errColor    = lipgloss.Color("#FF6B6B")
	okColor     = lipgloss.Color("#51CF66")
	dimColor    = lipgloss.Color("#555555")
	textColor   = lipgloss.Color("#C0C0C0")
	bgDark      = lipgloss.Color("#1a1a2e")



	modelStyle = lipgloss.NewStyle().
			Foreground(warnColor).
			Bold(true)

	promptStyle = lipgloss.NewStyle().
			Foreground(brandColor).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(errColor)

	successStyle = lipgloss.NewStyle().
			Foreground(okColor)

	thinkingStyle = lipgloss.NewStyle().
			Foreground(warnColor)

	inputEchoStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	outputStyle = lipgloss.NewStyle().
			Foreground(textColor)



	cmdStyle = lipgloss.NewStyle().
			Foreground(brandColor)



	footerStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E0E0")).
			Bold(true)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333"))
)

// capybara builds a colored capybara logo
func capybara() string {
	colors := map[rune]lipgloss.TerminalColor{
		'D': lipgloss.Color("#333333"), // Dark outline
		'B': lipgloss.Color("#8B5A2B"), // Brown body
		'L': lipgloss.Color("#D68A59"), // Light brown/snout
		'_': lipgloss.NoColor{},        // Transparent
	}

	raw := `
⬜⬜⬜⬜⬛⬛⬜⬜⬜⬜⬜⬜⬜⬜
⬜⬜⬜⬛🟫🟫⬛⬛⬛⬛⬛⬜⬜⬜
⬜⬜⬜⬛🟧🟫⬛🟫🟫🟫🟧⬛⬜⬜
⬜⬜⬛🟫🟫🟫🟫⬛🟫🟫🟧⬛⬜⬜
⬜⬛🟫🟫🟫🟫🟫🟫🟫🟫🟧⬛⬜⬜
⬛🟫🟫🟧🟫🟫🟫🟫🟫🟫⬛⬜⬜⬜
⬛🟫🟫🟫🟫🟫🟫🟫🟫🟫⬛⬜⬜⬜
⬛🟫🟫🟧🟫🟫🟫🟫🟫🟫⬛⬜⬜⬜
⬛🟫🟫🟧🟫🟧🟫🟫🟫⬛⬜⬜⬜⬜
⬜⬛🟫🟫⬛⬛🟫🟧🟫⬛⬜⬜⬜⬜`

	raw = strings.Trim(raw, "\n\r")
	lines := strings.Split(raw, "\n")
	for i := range lines {
		lines[i] = strings.ReplaceAll(lines[i], "⬜", "_")
		lines[i] = strings.ReplaceAll(lines[i], "⬛", "D")
		lines[i] = strings.ReplaceAll(lines[i], "🟫", "B")
		lines[i] = strings.ReplaceAll(lines[i], "🟧", "L")
	}

	var b strings.Builder
	for i := 0; i < len(lines); i += 2 {
		topRow := lines[i]
		botRow := ""
		if i+1 < len(lines) {
			botRow = lines[i+1]
		}
		for j := 0; j < len(topRow); j++ {
			topP := rune(topRow[j])
			botP := rune('_')
			if j < len(botRow) {
				botP = rune(botRow[j])
			}

			if topP == '_' && botP == '_' {
				b.WriteString(" ")
				continue
			}

			style := lipgloss.NewStyle()
			char := "▀"

			if topP != '_' {
				style = style.Foreground(colors[topP])
			} else {
				char = "▄"
			}
			if botP != '_' {
				if char == "▀" {
					style = style.Background(colors[botP])
				} else {
					style = style.Foreground(colors[botP])
				}
			}

			b.WriteString(style.Render(char))
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

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

	// Load previous session summary
	var prevHistory []historyEntry
	var prevConversation []ConversationEntry
	if last := LoadLastSession(); last != nil {
		// Show a session separator
		prevHistory = append(prevHistory, historyEntry{
			input:  "📂 Previous session (" + last.StartedAt + ")",
			output: fmt.Sprintf("%d messages", len(last.History)),
			time:   last.ID,
		})
		// Load last 5 entries as summary
		start := 0
		if len(last.History) > 5 {
			start = len(last.History) - 5
		}
		for _, h := range last.History[start:] {
			prevHistory = append(prevHistory, historyEntry{
				input: h.Input, output: h.Output,
				isPipeline: h.IsPipeline, isErr: h.IsErr, time: h.Time,
			})
		}
		// Load conversation memory from last session
		prevConversation = last.Conversation
	}

	return Model{
		textInput:    ti,
		spinner:      s,
		state:        stateIdle,
		cfg:          cfg,
		log:          log,
		history:      prevHistory,
		conversation: prevConversation,
		showWelcome:  true,
		width:        80,
		height:       24,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.warmupModel())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = min(msg.Width-6, 120)
		return m, nil

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			m.scrollOffset += 3
			return m, nil
		case tea.MouseWheelDown:
			m.scrollOffset -= 3
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
			return m, nil
		}

	case tea.KeyMsg:
		// Handle confirm state — y/n for command approval
		if m.state == stateConfirm {
			switch msg.String() {
			case "y", "Y":
				m.state = stateThinking
				cmds := make([]string, len(m.pendingCmds))
				copy(cmds, m.pendingCmds)
				input := m.pendingInput
				m.pendingCmds = nil
				m.pendingInput = ""
				return m, tea.Batch(m.spinner.Tick, m.executeCommands(input, cmds))
			case "n", "N", "esc":
				m.state = stateIdle
				m.textInput.Focus()
				m.history = append(m.history, historyEntry{
					input: "execute?", output: "⛔ Cancelled by user",
					time: time.Now().Format("15:04:05"),
				})
				m.pendingCmds = nil
				m.pendingInput = ""
				return m, textinput.Blink
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyEsc:
			if m.state == stateThinking {
				m.state = stateIdle
				m.textInput.Focus()
				m.history = append(m.history, historyEntry{
					input: m.currentInput, output: "⛔ Cancelled",
					time: time.Now().Format("15:04:05"),
				})
				m.currentInput = ""
				return m, textinput.Blink
			}

		case tea.KeyCtrlD:
			SaveSession(m.conversation, m.history)
			m.quitting = true
			return m, tea.Quit

		case tea.KeyPgUp:
			m.scrollOffset += 5
			return m, nil
		case tea.KeyPgDown:
			m.scrollOffset -= 5
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
			return m, nil
		case tea.KeyUp:
			if m.state == stateIdle && m.textInput.Value() == "" {
				m.scrollOffset++
				return m, nil
			}
		case tea.KeyDown:
			if m.state == stateIdle && m.textInput.Value() == "" {
				m.scrollOffset--
				if m.scrollOffset < 0 {
					m.scrollOffset = 0
				}
				return m, nil
			}

		case tea.KeyEnter:
			m.scrollOffset = 0    // reset scroll on new input
			m.showWelcome = false // hide welcome banner after first interaction
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
				SaveSession(m.conversation, m.history)
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

			case lower == "history":
				m.history = append(m.history, historyEntry{
					input:  input,
					output: m.sessionHistoryText(),
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
					input:  input,
					output: fmt.Sprintf("Current mode: %s\nAvailable: explain, guided, auto", m.cfg.Execution.DefaultMode),
					time:   time.Now().Format("15:04:05"),
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

			// Route: local match > pipeline > chat
			m.state = stateThinking
			m.currentInput = input
			m.textInput.Blur()

			// Try instant local match first (0ms)
			if result := tryLocalMatch(input); result != nil {
				var out strings.Builder
				out.WriteString("\n✅ Intent Extracted\n")
				out.WriteString(fmt.Sprintf("   Description: %s\n", result.Intent))
				out.WriteString(fmt.Sprintf("   Category:    %s\n", result.Category))
				out.WriteString(fmt.Sprintf("   Action:      %s\n", result.Action))
				out.WriteString(fmt.Sprintf("   Target:      %s\n", result.Parameters.Target))
				out.WriteString("   Confidence:  100%\n")
				out.WriteString("   Time:        0ms (local)\n")
				if len(result.SuggestedCommands) > 0 {
					out.WriteString("\n💡 Suggested Commands:\n")
					for i, cmd := range result.SuggestedCommands {
						out.WriteString(fmt.Sprintf("   %d. %s\n", i+1, cmd))
					}
				}
				m.history = append(m.history, historyEntry{
					input: input, output: out.String(),
					isPipeline: true, time: time.Now().Format("15:04:05"),
				})
				m.conversation = append(m.conversation,
					ConversationEntry{Role: "user", Content: input},
					ConversationEntry{Role: "assistant", Content: fmt.Sprintf("Suggested: %s", strings.Join([]string(result.SuggestedCommands), ", "))},
				)
				// Enter confirm state for approval
				m.pendingCmds = []string(result.SuggestedCommands)
				m.pendingInput = input
				m.state = stateConfirm
				return m, nil
			}

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
		// Record in conversation memory
		m.conversation = append(m.conversation,
			ConversationEntry{Role: "user", Content: msg.input},
			ConversationEntry{Role: "assistant", Content: "[System command analysis completed]"},
		)
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
		// Record in conversation memory
		m.conversation = append(m.conversation,
			ConversationEntry{Role: "user", Content: msg.input},
			ConversationEntry{Role: "assistant", Content: msg.response},
		)
		return m, textinput.Blink

	case execDoneMsg:
		m.state = stateIdle
		m.textInput.Focus()

		entry := historyEntry{
			input: "⚙️ executing",
			time:  time.Now().Format("15:04:05"),
		}
		if msg.err != nil {
			entry.output = msg.output
			entry.isErr = true
		} else {
			entry.output = msg.output
		}
		m.history = append(m.history, entry)
		m.conversation = append(m.conversation,
			ConversationEntry{Role: "assistant", Content: "Executed: " + msg.output},
		)
		return m, textinput.Blink

	case warmupDoneMsg:
		// Model is now loaded and warm — no visible change
		return m, nil

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
		return "\n  " + dimStyle.Render("👋 TaaNOS session ended. See you!") + "\n\n"
	}

	var b strings.Builder
	w := min(m.width, 100)

	// ── Top header bar (Claude Code Style) ──
	if m.showWelcome {
		cwd, _ := os.Getwd()
		logo := capybara()

		info := fmt.Sprintf("%s\n%s\n%s",
			versionStyle.Render("TaaNOS v1.5.0"),
			modelStyle.Render(m.cfg.Ollama.Model),
			pathStyle.Render(cwd))

		header := lipgloss.JoinHorizontal(lipgloss.Top, logo, "   ", info)

		b.WriteString("\n" + lipgloss.NewStyle().MarginLeft(2).Render(header) + "\n\n")
	}

	// ── Separator ──
	b.WriteString("  " + separatorStyle.Render(strings.Repeat("─", w-4)) + "\n")

	// ── Chat history ──
	var historyLines []string
	for _, entry := range m.history {
		historyLines = append(historyLines,
			"  "+inputEchoStyle.Render("❯ "+entry.input)+"  "+dimStyle.Render(entry.time))

		if entry.isPipeline && !entry.isErr {
			formatted := FormatPipelineOutput(entry.output, m.width)
			for _, line := range strings.Split(formatted, "\n") {
				if line != "" {
					historyLines = append(historyLines, line)
				}
			}
		} else {
			w := m.width - 8
			if w < 10 {
				w = 10
			}
			wrapped := lipgloss.NewStyle().Width(w).Render(entry.output)
			lines := strings.Split(wrapped, "\n")
			for _, line := range lines {
				if entry.isErr {
					historyLines = append(historyLines, "  "+errorStyle.Render("  ✗ "+line))
				} else {
					historyLines = append(historyLines, "  "+outputStyle.Render("  "+line))
				}
			}
		}
		historyLines = append(historyLines, "")
	}

	// ── Scroll window ──
	welcomeOffset := 0
	if m.showWelcome {
		welcomeOffset = 14
	}
	maxVisible := m.height - 8 - welcomeOffset
	if maxVisible < 5 {
		maxVisible = 5
	}
	totalLines := len(historyLines)

	offset := m.scrollOffset
	maxScroll := totalLines - maxVisible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if offset > maxScroll {
		offset = maxScroll
	}

	endIdx := totalLines - offset
	if endIdx > totalLines {
		endIdx = totalLines
	}
	startIdx := endIdx - maxVisible
	if startIdx < 0 {
		startIdx = 0
	}

	if endIdx > startIdx {
		for _, line := range historyLines[startIdx:endIdx] {
			b.WriteString(line + "\n")
		}
	}

	if m.scrollOffset > 0 {
		b.WriteString("  " + dimStyle.Render(fmt.Sprintf("  ↑ %d more lines (PgUp/PgDn)", m.scrollOffset)) + "\n")
	}

	// ── Bottom separator ──
	b.WriteString("  " + separatorStyle.Render(strings.Repeat("─", w-4)) + "\n")

	// ── Input area ──
	switch m.state {
	case stateThinking:
		b.WriteString("  " + inputEchoStyle.Render("❯ "+m.currentInput) + "\n")
		b.WriteString("  " + thinkingStyle.Render(fmt.Sprintf("  %s %s thinking...",
			m.spinner.View(), m.cfg.Ollama.Model)) + "\n")
		b.WriteString("  " + dimStyle.Render("  Press ESC to cancel") + "\n")
	case stateConfirm:
		b.WriteString("  " + thinkingStyle.Render("🚀 Execute these commands?") + "\n")
		for i, cmd := range m.pendingCmds {
			b.WriteString("  " + cmdStyle.Render(fmt.Sprintf("   %d. %s", i+1, cmd)) + "\n")
		}
		b.WriteString("  " + successStyle.Render("[y]") + " execute  " + errorStyle.Render("[n]") + " cancel\n")
	default:
		b.WriteString("  " + m.textInput.View() + "\n")
	}

	// ── Footer bar ──
	footer := footerStyle.Render("  ? for shortcuts · ↵ for agents")
	b.WriteString(footer + "\n")

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
	b.WriteString("TaaNOS v0.2.0 — Interactive REPL\n")
	b.WriteString(strings.Repeat("─", 50) + "\n")
	b.WriteString("\nCommands:\n")
	cmds := []struct{ key, desc string }{
		{"<any text>", "Ask AI or execute system commands"},
		{"help, ?", "Show this help menu"},
		{"status", "Show TaaNOS system status"},
		{"history", "Show past chat sessions"},
		{"model", "Show current AI model"},
		{"model <name>", "Change AI model"},
		{"mode", "Show current execution mode"},
		{"mode <mode>", "Set mode: explain, guided, auto"},
		{"clear, cls", "Clear screen"},
		{"exit, quit, q", "Exit TaaNOS (saves session)"},
		{"Ctrl+D", "Exit TaaNOS (saves session)"},
	}
	for _, e := range cmds {
		b.WriteString(fmt.Sprintf("  %-18s %s\n", e.key, e.desc))
	}
	b.WriteString("\nKeyboard:\n")
	keys := []struct{ key, desc string }{
		{"ESC", "Cancel current request"},
		{"PgUp / PgDn", "Scroll history"},
		{"Mouse wheel", "Scroll history"},
	}
	for _, e := range keys {
		b.WriteString(fmt.Sprintf("  %-18s %s\n", e.key, e.desc))
	}
	b.WriteString("\nExamples:\n")
	b.WriteString("  install nginx\n")
	b.WriteString("  -m explain check disk space\n")
	b.WriteString("  merhaba, nasılsın?\n")
	b.WriteString("\nSession memory: Your chats are saved and auto-loaded on next start.\n")
	return b.String()
}

// statusText returns status info.
func (m *Model) statusText() string {
	return fmt.Sprintf("Model:    %s\nEndpoint: %s\nMode:     %s\nVersion:  0.2.0",
		m.cfg.Ollama.Model, m.cfg.Ollama.Endpoint, m.cfg.Execution.DefaultMode)
}

// sessionHistoryText shows past chat sessions.
func (m *Model) sessionHistoryText() string {
	sessions, err := LoadSessions()
	if err != nil || len(sessions) == 0 {
		return "No past sessions found."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Chat Sessions (%d):\n", len(sessions)))
	b.WriteString(strings.Repeat("─", 50) + "\n")

	// Show last 10 sessions
	start := 0
	if len(sessions) > 10 {
		start = len(sessions) - 10
	}
	for _, s := range sessions[start:] {
		msgCount := len(s.History)
		lastMsg := ""
		if msgCount > 0 {
			lastMsg = s.History[msgCount-1].Input
			if len(lastMsg) > 30 {
				lastMsg = lastMsg[:27] + "..."
			}
		}
		b.WriteString(fmt.Sprintf("  %s  %d msgs  %s\n", s.StartedAt, msgCount, lastMsg))
	}
	b.WriteString("\nLast session is auto-loaded on start.\n")
	b.WriteString("Use 'taanos history <id>' in terminal for details.")
	return b.String()
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

// runChat sends a conversational message to Ollama with session memory.
func (m *Model) runChat(input string) tea.Cmd {
	cfg := m.cfg
	history := make([]ConversationEntry, len(m.conversation))
	copy(history, m.conversation)

	return func() tea.Msg {
		response, err := ChatWithHistory(
			cfg.Ollama.Endpoint,
			cfg.Ollama.Model,
			input,
			history,
			cfg.Ollama.Timeout,
		)
		if err != nil {
			return chatDoneMsg{input: input, err: err}
		}
		return chatDoneMsg{input: input, response: response}
	}
}

// warmupModel sends a tiny request on REPL start to pre-load the model.
func (m *Model) warmupModel() tea.Cmd {
	cfg := m.cfg
	return func() tea.Msg {
		// Send a minimal request to load the model into memory
		Chat(cfg.Ollama.Endpoint, cfg.Ollama.Model, "hi", 30*time.Second)
		return warmupDoneMsg{}
	}
}

// executeCommands runs approved commands and returns output.
func (m *Model) executeCommands(input string, cmds []string) tea.Cmd {
	return func() tea.Msg {
		var output strings.Builder
		allOk := true

		for i, cmd := range cmds {
			output.WriteString(fmt.Sprintf("[%d/%d] %s\n", i+1, len(cmds), cmd))

			var execCmd *exec.Cmd
			if runtime.GOOS == "windows" {
				execCmd = exec.Command("powershell", "-Command", cmd)
			} else {
				execCmd = exec.Command("sh", "-c", cmd)
			}

			out, err := execCmd.CombinedOutput()
			if err != nil {
				output.WriteString(fmt.Sprintf("  ❌ Error: %v\n", err))
				if len(out) > 0 {
					output.WriteString(fmt.Sprintf("  %s\n", strings.TrimSpace(string(out))))
				}
				allOk = false
			} else {
				if len(out) > 0 {
					output.WriteString(fmt.Sprintf("  %s\n", strings.TrimSpace(string(out))))
				}
				output.WriteString("  ✅ Done\n")
			}
		}

		if allOk {
			output.WriteString("\n✅ All commands completed successfully")
		}

		return execDoneMsg{input: input, output: output.String()}
	}
}
