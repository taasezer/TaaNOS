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

	if err := Validate(&result); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return &result, nil
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
