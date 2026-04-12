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
		return nil, fmt.Errorf("JSON parse error: %w (raw: %s)", err, string(data))
	}

	// Normalize category and action before validation (helps small models)
	result.Category = normalizeCategory(result.Category)
	result.Action = normalizeAction(result.Action)

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

// normalizeCategory maps partial/fuzzy category strings to valid ones.
func normalizeCategory(cat Category) Category {
	s := strings.ToLower(strings.TrimSpace(string(cat)))
	
	// Direct match
	if ValidCategories[Category(s)] {
		return Category(s)
	}

	// Fuzzy matching
	categoryMap := map[string]Category{
		"package":    "package_management",
		"pkg":        "package_management",
		"packages":   "package_management",
		"service":    "service_management",
		"services":   "service_management",
		"file":       "file_operation",
		"files":      "file_operation",
		"filesystem": "file_operation",
		"net":        "network",
		"networking": "network",
		"info":       "system_info",
		"system":     "system_info",
		"sysinfo":    "system_info",
	}

	for partial, full := range categoryMap {
		if strings.Contains(s, partial) {
			return full
		}
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

	// Fuzzy matching
	actionMap := map[string]Action{
		"add":       "install",
		"setup":     "install",
		"uninstall": "remove",
		"purge":     "remove",
		"begin":     "start",
		"launch":    "start",
		"halt":      "stop",
		"kill":      "stop",
		"reboot":    "restart",
		"reload":    "restart",
		"activate":  "enable",
		"deactivate":"disable",
		"make":      "create",
		"mkdir":     "create",
		"rm":        "delete",
		"erase":     "delete",
		"ls":        "list",
		"dir":       "list",
		"display":   "show",
		"view":      "show",
		"get":       "show",
		"status":    "show",
		"upgrade":   "update",
		"patch":     "update",
		"set":       "configure",
		"edit":      "configure",
		"modify":    "configure",
	}

	for partial, full := range actionMap {
		if s == partial || strings.Contains(s, partial) {
			return full
		}
	}

	// Default fallback
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
