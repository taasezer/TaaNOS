package tui

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for the UI.
type Theme struct {
	Name        string
	BrandColor  lipgloss.Color
	AccentColor lipgloss.Color
	WarnColor   lipgloss.Color
	ErrColor    lipgloss.Color
	OkColor     lipgloss.Color
	DimColor    lipgloss.Color
	TextColor   lipgloss.Color
	BgColor     lipgloss.Color
}

// Predefined Themes
var (
	ThemeDefault = Theme{
		Name:        "Default",
		BrandColor:  lipgloss.Color("#00D4AA"),
		AccentColor: lipgloss.Color("#7C8DFF"),
		WarnColor:   lipgloss.Color("#FFD43B"),
		ErrColor:    lipgloss.Color("#FF6B6B"),
		OkColor:     lipgloss.Color("#51CF66"),
		DimColor:    lipgloss.Color("#555555"),
		TextColor:   lipgloss.Color("#C0C0C0"),
		BgColor:     lipgloss.Color("#1a1a2e"),
	}

	ThemeOcean = Theme{
		Name:        "Ocean",
		BrandColor:  lipgloss.Color("#00B4D8"),
		AccentColor: lipgloss.Color("#90E0EF"),
		WarnColor:   lipgloss.Color("#F4A261"),
		ErrColor:    lipgloss.Color("#E76F51"),
		OkColor:     lipgloss.Color("#2A9D8F"),
		DimColor:    lipgloss.Color("#4A4E69"),
		TextColor:   lipgloss.Color("#CAF0F8"),
		BgColor:     lipgloss.Color("#03045E"),
	}

	ThemeDracula = Theme{
		Name:        "Dracula",
		BrandColor:  lipgloss.Color("#FF79C6"),
		AccentColor: lipgloss.Color("#BD93F9"),
		WarnColor:   lipgloss.Color("#F1FA8C"),
		ErrColor:    lipgloss.Color("#FF5555"),
		OkColor:     lipgloss.Color("#50FA7B"),
		DimColor:    lipgloss.Color("#6272A4"),
		TextColor:   lipgloss.Color("#F8F8F2"),
		BgColor:     lipgloss.Color("#282A36"),
	}

	ThemeMatrix = Theme{
		Name:        "Matrix",
		BrandColor:  lipgloss.Color("#00FF41"),
		AccentColor: lipgloss.Color("#008F11"),
		WarnColor:   lipgloss.Color("#FFFF00"),
		ErrColor:    lipgloss.Color("#FF0000"),
		OkColor:     lipgloss.Color("#00FF41"),
		DimColor:    lipgloss.Color("#003B00"),
		TextColor:   lipgloss.Color("#00FF41"),
		BgColor:     lipgloss.Color("#0D0208"),
	}
)

var Themes = map[string]Theme{
	"default": ThemeDefault,
	"ocean":   ThemeOcean,
	"dracula": ThemeDracula,
	"matrix":  ThemeMatrix,
}

var CurrentTheme = ThemeDefault

// UpdateStyles updates the global lipgloss styles based on the current theme.
func UpdateStyles(t Theme) {
	CurrentTheme = t
	brandColor = t.BrandColor
	accentColor = t.AccentColor
	warnColor = t.WarnColor
	errColor = t.ErrColor
	okColor = t.OkColor
	dimColor = t.DimColor
	textColor = t.TextColor

	// Rebuild styles
	modelStyle = lipgloss.NewStyle().Foreground(warnColor).Bold(true)
	promptStyle = lipgloss.NewStyle().Foreground(brandColor).Bold(true)
	dimStyle = lipgloss.NewStyle().Foreground(dimColor)
	errorStyle = lipgloss.NewStyle().Foreground(errColor)
	successStyle = lipgloss.NewStyle().Foreground(okColor)
	thinkingStyle = lipgloss.NewStyle().Foreground(warnColor)
	inputEchoStyle = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	outputStyle = lipgloss.NewStyle().Foreground(textColor)
	cmdStyle = lipgloss.NewStyle().Foreground(brandColor)
	footerStyle = lipgloss.NewStyle().Foreground(dimColor)
}
