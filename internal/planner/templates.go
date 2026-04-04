package planner

import (
	"strings"
)

// RenderTemplate replaces {target} and {options} placeholders in a command template.
func RenderTemplate(template string, params map[string]string) string {
	result := template
	for key, value := range params {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
