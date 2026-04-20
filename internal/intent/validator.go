package intent

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Validate checks that an IntentResult conforms to the allowed schema.
// Returns a cleaned IntentResult and any validation error.
func Validate(result *IntentResult) error {
	// 1. Category must be in the allowed set
	if !ValidCategories[result.Category] {
		return fmt.Errorf("invalid category: %q (allowed: %s)",
			result.Category, allowedCategories())
	}

	// 2. Action must be in the allowed set
	if !ValidActions[result.Action] {
		return fmt.Errorf("invalid action: %q (allowed: %s)",
			result.Action, allowedActions())
	}

	// 3. Confidence must be between 0.0 and 1.0
	if result.Confidence < 0.0 || result.Confidence > 1.0 {
		return fmt.Errorf("confidence out of range: %f (must be 0.0–1.0)",
			result.Confidence)
	}

	// 4. Intent description must not be empty
	if strings.TrimSpace(result.Intent) == "" {
		return fmt.Errorf("intent description is empty")
	}

	return nil
}

// ParseAndValidate parses raw JSON bytes into an IntentResult and validates it.
func ParseAndValidate(data []byte) (*IntentResult, error) {
	// Strip markdown code fences if the LLM wraps them
	cleaned := cleanJSON(data)

	var result IntentResult
	if err := json.Unmarshal(cleaned, &result); err != nil {
		// Fallback: if model returned plain text instead of JSON,
		// try to build a best-effort result from the raw text
		fallback := tryPlainTextFallback(string(data))
		if fallback != nil {
			return fallback, nil
		}
		return nil, fmt.Errorf("JSON parse error: %w (raw: %.200s)", err, string(data))
	}

	// Normalize category and action before validation (helps small models)
	result.Category = normalizeCategory(result.Category)
	result.Action = normalizeAction(result.Action)

	// Smart fallback: if category is "unknown" but action+target are valid,
	// infer the category from the action. Small models often get the action
	// right but leave category empty.
	if result.Category == "unknown" && result.Parameters.Target != "" {
		result.Category = inferCategoryFromAction(result.Action)
	}

	// Also try to infer from intent text keywords
	if result.Category == "unknown" && strings.TrimSpace(result.Intent) != "" {
		result.Category = inferCategoryFromText(result.Intent)
	}

	// If confidence is 0 (model didn't set it), give a default
	if result.Confidence == 0 && result.Category != "unknown" {
		result.Confidence = 0.5
	}

	// If intent is empty, generate one
	if strings.TrimSpace(result.Intent) == "" {
		result.Intent = fmt.Sprintf("%s %s", result.Action, result.Parameters.Target)
	}

	if err := Validate(&result); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return &result, nil
}

// tryPlainTextFallback attempts to extract intent from plain text
// when the model fails to produce JSON. Returns nil if it can't.
func tryPlainTextFallback(raw string) *IntentResult {
	lower := strings.ToLower(raw)

	// Check if response contains any recognizable action keywords
	actions := map[string]Action{
		"install": "install", "remove": "remove", "delete": "delete",
		"start": "start", "stop": "stop", "restart": "restart",
		"update": "update", "upgrade": "update", "show": "show",
		"list": "list", "create": "create", "enable": "enable",
		"disable": "disable", "configure": "configure",
	}

	for keyword, action := range actions {
		if strings.Contains(lower, string(keyword)) {
			// Try to find a target word after the action
			parts := strings.Fields(lower)
			target := ""
			for i, p := range parts {
				if p == keyword && i+1 < len(parts) {
					target = parts[i+1]
					break
				}
			}

			cat := normalizeCategory(Category(""))
			if action == "install" || action == "remove" || action == "update" {
				cat = "package_management"
			} else if action == "start" || action == "stop" || action == "restart" || action == "enable" || action == "disable" {
				cat = "service_management"
			}

			return &IntentResult{
				Intent:     fmt.Sprintf("%s %s", action, target),
				Category:   cat,
				Action:     action,
				Parameters: Parameters{Target: target, Options: FlexibleStrings{}, Scope: "system"},
				Confidence: 0.4, // Low confidence for fallback
			}
		}
	}

	return nil
}

// inferCategoryFromAction maps an action to the most likely category.
func inferCategoryFromAction(act Action) Category {
	switch act {
	case "install", "remove", "update":
		return "package_management"
	case "start", "stop", "restart", "enable", "disable":
		return "service_management"
	case "create", "delete":
		return "file_operation"
	case "list", "show":
		return "system_info"
	case "configure":
		return "system_info"
	default:
		return "unknown"
	}
}

// inferCategoryFromText tries to determine category from intent description text.
func inferCategoryFromText(text string) Category {
	lower := strings.ToLower(text)

	pkgKeywords := []string{"install", "package", "apt", "yum", "dnf", "pacman", "uninstall", "upgrade", "winget", "choco", "brew", "pip", "npm"}
	svcKeywords := []string{"service", "restart", "systemctl", "daemon", "start", "stop", "enable", "disable"}
	fileKeywords := []string{"file", "directory", "folder", "delete", "create", "remove file", "path", "move", "copy", "rename", "permission", "chmod"}
	netKeywords := []string{"port", "network", "firewall", "ip", "dns", "socket", "connection", "ping", "curl", "wget", "http", "ssh"}
	infoKeywords := []string{"disk", "memory", "cpu", "uptime", "info", "status", "usage", "space", "version", "check", "show", "list", "docker", "kernel", "process", "ram", "temperature", "battery", "os", "system"}

	for _, kw := range pkgKeywords {
		if strings.Contains(lower, kw) {
			return "package_management"
		}
	}
	for _, kw := range svcKeywords {
		if strings.Contains(lower, kw) {
			return "service_management"
		}
	}
	for _, kw := range fileKeywords {
		if strings.Contains(lower, kw) {
			return "file_operation"
		}
	}
	for _, kw := range netKeywords {
		if strings.Contains(lower, kw) {
			return "network"
		}
	}
	for _, kw := range infoKeywords {
		if strings.Contains(lower, kw) {
			return "system_info"
		}
	}

	return "unknown"
}

// normalizeCategory maps partial/fuzzy category strings to valid ones.
func normalizeCategory(cat Category) Category {
	s := strings.ToLower(strings.TrimSpace(string(cat)))
	
	// Direct match
	if ValidCategories[Category(s)] {
		return Category(s)
	}

	// Comprehensive fuzzy matching — handles typos, abbreviations, synonyms
	categoryMap := map[string]Category{
		// Package management
		"package":             "package_management",
		"pkg":                 "package_management",
		"packages":            "package_management",
		"package_managemeent": "package_management",
		"package_mgmt":        "package_management",
		"packagemanagement":   "package_management",
		"software":            "package_management",
		"install":             "package_management",
		"apt":                 "package_management",
		"winget":              "package_management",
		"brew":                "package_management",
		"choco":               "package_management",
		"pip":                 "package_management",
		"npm":                 "package_management",
		// Service management
		"service":             "service_management",
		"services":            "service_management",
		"service_mgmt":        "service_management",
		"servicemanagement":   "service_management",
		"daemon":              "service_management",
		"systemctl":           "service_management",
		"systemd":             "service_management",
		// File operations
		"file":                "file_operation",
		"files":               "file_operation",
		"filesystem":          "file_operation",
		"file_ops":            "file_operation",
		"fileoperation":       "file_operation",
		"directory":           "file_operation",
		"folder":              "file_operation",
		// Network
		"net":                 "network",
		"networking":          "network",
		"firewall":            "network",
		"port":                "network",
		"dns":                 "network",
		"connection":          "network",
		// System info
		"info":                "system_info",
		"system":              "system_info",
		"sysinfo":             "system_info",
		"system_information":  "system_info",
		"check":               "system_info",
		"version":             "system_info",
		"status":              "system_info",
		"monitor":             "system_info",
		"diagnostic":          "system_info",
		"health":              "system_info",
	}

	for partial, full := range categoryMap {
		if s == partial || strings.Contains(s, partial) {
			return full
		}
	}

	// Last resort: if it's not empty and not recognized, treat as system_info
	// rather than rejecting — let the model's intent guide the pipeline
	if s != "" && s != "unknown" {
		return "system_info"
	}

	return "unknown"
}

// normalizeAction maps partial/fuzzy action strings to valid ones.
func normalizeAction(act Action) Action {
	s := strings.ToLower(strings.TrimSpace(string(act)))

	// Direct match
	if ValidActions[Action(s)] {
		return Action(s)
	}

	// Comprehensive fuzzy matching
	actionMap := map[string]Action{
		// Install
		"add": "install", "setup": "install", "deploy": "install", "download": "install", "get": "install", "fetch": "install", "pull": "install",
		// Remove
		"uninstall": "remove", "purge": "remove", "erase": "remove", "drop": "remove", "clean": "remove",
		// Start
		"begin": "start", "launch": "start", "run": "start", "open": "start", "init": "start", "boot": "start",
		// Stop
		"halt": "stop", "kill": "stop", "terminate": "stop", "end": "stop", "close": "stop", "shutdown": "stop",
		// Restart
		"reboot": "restart", "reload": "restart", "reset": "restart", "refresh": "restart",
		// Enable/Disable
		"activate": "enable", "deactivate": "disable", "turnon": "enable", "turnoff": "disable",
		// Create/Delete
		"make": "create", "mkdir": "create", "touch": "create", "new": "create", "generate": "create",
		"rm": "delete", "rmdir": "delete", "destroy": "delete", "wipe": "delete",
		// List/Show
		"ls": "list", "dir": "list", "find": "list", "search": "list", "enumerate": "list",
		"display": "show", "view": "show", "print": "show", "cat": "show", "read": "show", "info": "show",
		"status": "show", "check": "show", "verify": "show", "inspect": "show", "describe": "show", "examine": "show",
		"choose": "show", "select": "show", "query": "show", "lookup": "show", "monitor": "show",
		// Update
		"upgrade": "update", "patch": "update", "sync": "update",
		// Configure
		"set": "configure", "edit": "configure", "modify": "configure", "change": "configure", "alter": "configure", "adjust": "configure", "tune": "configure",
	}

	for partial, full := range actionMap {
		if s == partial || strings.Contains(s, partial) {
			return full
		}
	}

	// Default: show (safe, read-only)
	return "show"
}

// cleanJSON strips markdown code fences and leading/trailing whitespace.
func cleanJSON(data []byte) []byte {
	s := strings.TrimSpace(string(data))

	// Remove ```json ... ``` wrapper
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}

	// Find first { and last } to extract JSON object
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		s = s[start : end+1]
	}

	return []byte(s)
}

func allowedCategories() string {
	cats := make([]string, 0, len(ValidCategories))
	for c := range ValidCategories {
		cats = append(cats, string(c))
	}
	return strings.Join(cats, ", ")
}

func allowedActions() string {
	acts := make([]string, 0, len(ValidActions))
	for a := range ValidActions {
		acts = append(acts, string(a))
	}
	return strings.Join(acts, ", ")
}
