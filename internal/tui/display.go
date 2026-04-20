package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Display styles for rich pipeline output
var (
	// Cards
	cardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Padding(0, 1).
		MarginTop(1)

	intentCardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D4AA")).
		Padding(0, 1)

	cmdCardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFD43B")).
		Padding(0, 1)

	// Labels
	labelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Width(14)

	valueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E0E0E0"))

	intentTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA")).
		Bold(true)

	cmdTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD43B")).
		Bold(true)

	cmdItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	stageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C8DFF"))

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF9F1C"))

	// Confidence colors
	confHighStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#51CF66")).
		Bold(true)

	confMedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD43B")).
		Bold(true)

	confLowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)

	// Execution
	execSuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#51CF66"))

	execFailStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))

	modeExplainStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C8DFF")).
		Italic(true)

	separatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333"))
)

// FormatPipelineOutput parses raw pipeline output and renders it with rich styling.
func FormatPipelineOutput(raw string, width int) string {
	if raw == "" {
		return ""
	}

	cardWidth := min(width-6, 72)

	lines := strings.Split(raw, "\n")
	var result strings.Builder

	var intentLines []string
	var cmdLines []string
	var stageLines []string
	var warningLines []string
	var execLines []string

	inIntent := false
	inCmds := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		switch {
		case strings.Contains(trimmed, "Analyzing intent"):
			stageLines = append(stageLines, "⠿ Analyzing intent...")

		case strings.Contains(trimmed, "Intent Extracted"):
			inIntent = true
			inCmds = false

		case strings.Contains(trimmed, "Suggested Commands"):
			inIntent = false
			inCmds = true

		case strings.Contains(trimmed, "Analyzing system context"):
			inIntent = false
			inCmds = false
			stageLines = append(stageLines, "⠿ System context analyzed")

		case strings.Contains(trimmed, "Explain mode"):
			stageLines = append(stageLines, "")
			execLines = append(execLines, modeExplainStyle.Render("📖 Explain mode — no commands executed"))

		case strings.Contains(trimmed, "Execute these commands"):
			inCmds = false // already captured

		case strings.Contains(trimmed, "Executing..."):
			execLines = append(execLines, stageStyle.Render("⚙️  Executing..."))

		case strings.Contains(trimmed, "✅ Done") || strings.Contains(trimmed, "All commands completed"):
			execLines = append(execLines, execSuccessStyle.Render("✓ "+trimmed))

		case strings.Contains(trimmed, "❌ Failed"):
			execLines = append(execLines, execFailStyle.Render("✗ "+trimmed))

		case strings.Contains(trimmed, "Aborted"):
			execLines = append(execLines, warningStyle.Render("⛔ "+trimmed))

		case strings.Contains(trimmed, "⚠"):
			warningLines = append(warningLines, trimmed)

		case strings.Contains(trimmed, "────"):
			// separator, skip

		case inIntent:
			intentLines = append(intentLines, trimmed)

		case inCmds:
			cmdLines = append(cmdLines, trimmed)

		default:
			if strings.HasPrefix(trimmed, "[") && strings.Contains(trimmed, "/") {
				execLines = append(execLines, stageStyle.Render("  "+trimmed))
			}
		}
	}

	// Render intent card
	if len(intentLines) > 0 {
		var card strings.Builder
		card.WriteString(intentTitleStyle.Render("✦ Intent Extracted") + "\n")

		for _, line := range intentLines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				label := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Style confidence by value
				if label == "Confidence" {
					value = styleConfidence(value)
				}

				card.WriteString(labelStyle.Render(label) + " " + valueStyle.Render(value) + "\n")
			} else {
				card.WriteString(valueStyle.Render(line) + "\n")
			}
		}

		styled := intentCardStyle.Width(cardWidth).Render(card.String())
		result.WriteString(styled + "\n")
	}

	// Render suggested commands card
	if len(cmdLines) > 0 {
		var card strings.Builder
		card.WriteString(cmdTitleStyle.Render("💡 Suggested Commands") + "\n")

		for _, line := range cmdLines {
			// Parse "1. command" format
			cleaned := strings.TrimSpace(line)
			re := regexp.MustCompile(`^\d+\.\s*(.+)$`)
			if matches := re.FindStringSubmatch(cleaned); len(matches) > 1 {
				card.WriteString("  " + cmdItemStyle.Render("▸ "+matches[1]) + "\n")
			} else if cleaned != "" {
				card.WriteString("  " + cmdItemStyle.Render("▸ "+cleaned) + "\n")
			}
		}

		styled := cmdCardStyle.Width(cardWidth).Render(card.String())
		result.WriteString(styled + "\n")
	}

	// Render warnings
	for _, w := range warningLines {
		result.WriteString(warningStyle.Render("  "+w) + "\n")
	}

	// Render stage info
	for _, s := range stageLines {
		if s == "" {
			continue
		}
		result.WriteString(stageStyle.Render("  "+s) + "\n")
	}

	// Render execution results
	for _, e := range execLines {
		result.WriteString("  " + e + "\n")
	}

	return result.String()
}

// styleConfidence returns a colored confidence value.
func styleConfidence(value string) string {
	// Extract number from string like "90%" or "45%"
	value = strings.TrimSpace(value)
	var num int
	fmt.Sscanf(value, "%d", &num)

	switch {
	case num >= 70:
		return confHighStyle.Render(value)
	case num >= 40:
		return confMedStyle.Render(value)
	default:
		return confLowStyle.Render(value)
	}
}
