package ollama

import (
	"testing"
)

func TestEnsureRunning(t *testing.T) {
	err := EnsureRunning("http://127.0.0.1:11434")
	t.Logf("EnsureRunning result: %v", err)
	t.Logf("IsRunning result: %v", IsRunning("http://127.0.0.1:11434"))
	if !IsRunning("http://127.0.0.1:11434") {
		t.Fatalf("Ollama should be running!")
	}
}
