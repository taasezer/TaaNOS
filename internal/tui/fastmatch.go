package tui

import (
	"github.com/taasezer/TaaNOS/internal/intent"
)

// tryLocalMatch attempts to resolve the input without calling Ollama.
// Returns nil if no match — caller should fall back to AI pipeline.
func tryLocalMatch(input string) *intent.IntentResult {
	return intent.TryLocalMatch(input)
}
