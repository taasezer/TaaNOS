package intent

import (
	"fmt"
	"runtime"
)

// SystemPrompt — production prompt with OS-specific examples.
// Small models (tinyllama 1.1B) require examples to produce valid JSON with suggested_commands.
const SystemPrompt = `You are TaaNOS running on %s. Return ONLY a JSON object.

%s

{"intent":"...","category":"package_management|service_management|file_operation|network|system_info|unknown","action":"install|remove|start|stop|restart|enable|disable|create|delete|list|show|update|configure","parameters":{"target":"...","options":[],"scope":"system"},"confidence":0.0-1.0,"suggested_commands":["..."]}

%s

Casual chat or greetings: category "unknown", confidence 0.1, suggested_commands [].`

// BuildSystemPrompt generates OS-specific system prompt with real examples.
func BuildSystemPrompt() string {
	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf(SystemPrompt,
			platform,
			"Use PowerShell and Windows commands only. Never suggest sudo or apt.",
			`User: "install nginx"
{"intent":"Install nginx","category":"package_management","action":"install","parameters":{"target":"nginx","options":[],"scope":"system"},"confidence":0.95,"suggested_commands":["winget install nginx"]}
User: "check my go version"
{"intent":"Check Go version","category":"system_info","action":"show","parameters":{"target":"go","options":[],"scope":"system"},"confidence":0.9,"suggested_commands":["go version"]}
User: "show disk space"
{"intent":"Show disk usage","category":"system_info","action":"show","parameters":{"target":"disk","options":[],"scope":"system"},"confidence":0.9,"suggested_commands":["Get-PSDrive -PSProvider FileSystem"]}
User: "check docker version"
{"intent":"Check Docker version","category":"system_info","action":"show","parameters":{"target":"docker","options":[],"scope":"system"},"confidence":0.9,"suggested_commands":["docker --version","docker info"]}
User: "restart apache service"
{"intent":"Restart Apache","category":"service_management","action":"restart","parameters":{"target":"apache","options":[],"scope":"system"},"confidence":0.9,"suggested_commands":["Restart-Service Apache2.4"]}`)
	default:
		return fmt.Sprintf(SystemPrompt,
			platform,
			"Use Linux shell commands. Prefix privileged commands with sudo.",
			`User: "install nginx"
{"intent":"Install nginx","category":"package_management","action":"install","parameters":{"target":"nginx","options":[],"scope":"system"},"confidence":0.95,"suggested_commands":["sudo apt-get install -y nginx"]}
User: "check my go version"
{"intent":"Check Go version","category":"system_info","action":"show","parameters":{"target":"go","options":[],"scope":"system"},"confidence":0.9,"suggested_commands":["go version"]}
User: "show disk space"
{"intent":"Show disk usage","category":"system_info","action":"show","parameters":{"target":"disk","options":[],"scope":"system"},"confidence":0.9,"suggested_commands":["df -h"]}
User: "check docker version"
{"intent":"Check Docker version","category":"system_info","action":"show","parameters":{"target":"docker","options":[],"scope":"system"},"confidence":0.9,"suggested_commands":["docker --version","docker info"]}
User: "restart apache service"
{"intent":"Restart Apache","category":"service_management","action":"restart","parameters":{"target":"apache2","options":[],"scope":"system"},"confidence":0.9,"suggested_commands":["sudo systemctl restart apache2"]}`)
	}
}

// UserPromptTemplate wraps the user input.
const UserPromptTemplate = `User: "%s"`
