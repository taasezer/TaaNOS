package intent

import (
	"fmt"
	"runtime"
)

// SystemPrompt is the hardcoded prompt sent to Ollama for intent extraction.
// This is the ONLY place where AI interacts with TaaNOS.
//
// The model extracts intent AND suggests shell commands for the user to review.
// TaaNOS will display these suggestions and ask for approval before executing.
//
// Design: Few-shot examples ensure even tiny models (tinyllama 1.1B) follow the format.
const SystemPrompt = `You are TaaNOS, an AI system administration assistant. Analyze the user's input and return a JSON object with your analysis and suggested commands.

Return ONLY valid JSON. No explanation. No markdown. No code blocks.

Schema:
{"intent":"<what user wants>","category":"<package_management|service_management|file_operation|network|system_info|unknown>","action":"<install|remove|start|stop|restart|enable|disable|create|delete|list|show|update|configure>","parameters":{"target":"<target>","options":[],"scope":"system"},"confidence":<0.0-1.0>,"suggested_commands":["<shell command 1>","<shell command 2>"]}

IMPORTANT: suggested_commands MUST match the user's operating system.
%s

Examples:

Input: "install nginx"
{"intent":"Install the nginx web server","category":"package_management","action":"install","parameters":{"target":"nginx","options":[],"scope":"system"},"confidence":0.95,"suggested_commands":[%s]}

Input: "restart apache"
{"intent":"Restart the Apache service","category":"service_management","action":"restart","parameters":{"target":"apache2","options":[],"scope":"system"},"confidence":0.9,"suggested_commands":[%s]}

Input: "show disk usage"
{"intent":"Display disk space usage","category":"system_info","action":"show","parameters":{"target":"disk","options":[],"scope":"system"},"confidence":0.85,"suggested_commands":[%s]}

Input: "hello how are you"
{"intent":"Not a system task","category":"unknown","action":"show","parameters":{"target":"","options":[],"scope":"system"},"confidence":0.1,"suggested_commands":[]}

Rules:
- suggested_commands MUST be valid commands for the current OS
- For dangerous operations, include sudo (Linux) or Run as Admin note (Windows)
- If input is not a system task, category is "unknown" and suggested_commands is []
- target is the package name, service name, file path, or resource name`

// BuildSystemPrompt generates the system prompt with OS-specific examples.
func BuildSystemPrompt() string {
	osInfo := fmt.Sprintf("The user is running: %s/%s", runtime.GOOS, runtime.GOARCH)

	var installEx, restartEx, diskEx string

	switch runtime.GOOS {
	case "windows":
		osInfo += "\nUse Windows commands: winget, powershell, choco, net, sc, Get-Service, etc."
		installEx = `"winget install nginx"`
		restartEx = `"Restart-Service Apache2.4"`
		diskEx = `"Get-PSDrive -PSProvider FileSystem"`
	default: // linux, darwin
		osInfo += "\nUse Linux commands: apt, yum, dnf, pacman, systemctl, etc."
		installEx = `"sudo apt-get update","sudo apt-get install -y nginx"`
		restartEx = `"sudo systemctl restart apache2"`
		diskEx = `"df -h"`
	}

	return fmt.Sprintf(SystemPrompt, osInfo, installEx, restartEx, diskEx)
}

// UserPromptTemplate wraps the user's natural language input.
const UserPromptTemplate = `Input: "%s"`
