package tui

import (
	"runtime"
	"strings"

	"github.com/taasezer/TaaNOS/internal/intent"
)

// localMatch holds a pre-built intent result for common commands.
type localMatch struct {
	Category string
	Action   string
	Commands []string
}

// tryLocalMatch attempts to resolve the input without calling Ollama.
// Returns nil if no match — caller should fall back to AI pipeline.
func tryLocalMatch(input string) *intent.IntentResult {
	lower := strings.ToLower(strings.TrimSpace(input))
	words := strings.Fields(lower)
	if len(words) < 2 {
		return nil
	}

	verb := words[0]
	target := strings.Join(words[1:], " ")

	var m *localMatch

	switch verb {
	case "install":
		m = &localMatch{
			Category: "package_management",
			Action:   "install",
			Commands: installCmd(target),
		}
	case "remove", "uninstall":
		m = &localMatch{
			Category: "package_management",
			Action:   "remove",
			Commands: removeCmd(target),
		}
	case "update", "upgrade":
		m = &localMatch{
			Category: "package_management",
			Action:   "update",
			Commands: updateCmd(target),
		}
	case "restart":
		m = &localMatch{
			Category: "service_management",
			Action:   "restart",
			Commands: serviceCmd("restart", target),
		}
	case "start":
		m = &localMatch{
			Category: "service_management",
			Action:   "start",
			Commands: serviceCmd("start", target),
		}
	case "stop":
		m = &localMatch{
			Category: "service_management",
			Action:   "stop",
			Commands: serviceCmd("stop", target),
		}
	default:
		return nil
	}

	if m == nil {
		return nil
	}

	cmds := make(intent.FlexibleStrings, len(m.Commands))
	copy(cmds, m.Commands)

	return &intent.IntentResult{
		Intent:            strings.ToUpper(verb[:1]) + verb[1:] + " " + target,
		Category:          intent.Category(m.Category),
		Action:            intent.Action(m.Action),
		Parameters:        intent.Parameters{Target: target, Scope: "system"},
		Confidence:        1.0,
		SuggestedCommands: cmds,
		ExtractionTimeMs:  0,
	}
}

func installCmd(target string) []string {
	if runtime.GOOS == "windows" {
		return []string{"winget install " + target}
	}
	return []string{"sudo apt-get install -y " + target}
}

func removeCmd(target string) []string {
	if runtime.GOOS == "windows" {
		return []string{"winget uninstall " + target}
	}
	return []string{"sudo apt-get remove -y " + target}
}

func updateCmd(target string) []string {
	if runtime.GOOS == "windows" {
		if target == "all" || target == "system" || target == "everything" {
			return []string{"winget upgrade --all"}
		}
		return []string{"winget upgrade " + target}
	}
	if target == "all" || target == "system" || target == "everything" {
		return []string{"sudo apt-get update && sudo apt-get upgrade -y"}
	}
	return []string{"sudo apt-get install --only-upgrade " + target}
}

func serviceCmd(action, target string) []string {
	if runtime.GOOS == "windows" {
		switch action {
		case "restart":
			return []string{"Restart-Service " + target}
		case "start":
			return []string{"Start-Service " + target}
		case "stop":
			return []string{"Stop-Service " + target}
		}
	}
	return []string{"sudo systemctl " + action + " " + target}
}
