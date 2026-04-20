package intent

import (
	"fmt"
	"runtime"
)

// SystemPrompt — ultra-compact for speed. Large models don't need examples.
const SystemPrompt = `TaaNOS on %s. Return ONLY JSON:
{"intent":"<description>","category":"<package_management|service_management|file_operation|network|system_info|unknown>","action":"<install|remove|start|stop|restart|enable|disable|create|delete|list|show|update|configure>","parameters":{"target":"<name>","options":[],"scope":"system"},"confidence":<0-1>,"suggested_commands":["<real %s commands>"]}
%s
Chat/greetings: category "unknown", confidence 0.1, suggested_commands [].`

// BuildSystemPrompt generates OS-specific system prompt.
func BuildSystemPrompt() string {
	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf(SystemPrompt, platform, "PowerShell",
			"Use winget for packages, PowerShell cmdlets for everything. No sudo/apt.")
	default:
		return fmt.Sprintf(SystemPrompt, platform, "shell",
			"Use apt/yum/dnf for packages, systemctl for services. Use sudo when needed.")
	}
}

// UserPromptTemplate wraps the user input.
const UserPromptTemplate = `%s`
