package ollama

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// NormalizeEndpoint ensures localhost is converted to 127.0.0.1 for deterministic IPv4 connection.
func NormalizeEndpoint(endpoint string) string {
	if strings.Contains(endpoint, "localhost") {
		return strings.Replace(endpoint, "localhost", "127.0.0.1", 1)
	}
	return endpoint
}

// IsRunning checks whether the Ollama server is currently listening.
func IsRunning(endpoint string) bool {
	ep := NormalizeEndpoint(endpoint)
	client := &http.Client{Timeout: 400 * time.Millisecond}
	resp, err := client.Get(ep + "/api/tags")
	if err == nil {
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}
	return false
}

// EnsureRunning checks if Ollama is running. If not, it attempts to launch
// 'ollama serve' in the background as a detached daemon, polling until ready.
func EnsureRunning(endpoint string) error {
	if IsRunning(endpoint) {
		return nil
	}

	// Locate ollama executable
	ollamaBin, err := exec.LookPath("ollama")
	if err != nil || ollamaBin == "" {
		if runtime.GOOS == "windows" {
			userHome, _ := os.UserHomeDir()
			candidates := []string{
				filepath.Join(userHome, "AppData", "Local", "Programs", "Ollama", "ollama.exe"),
				`C:\Program Files\Ollama\ollama.exe`,
			}
			for _, c := range candidates {
				if _, statErr := os.Stat(c); statErr == nil {
					ollamaBin = c
					break
				}
			}
		} else {
			candidates := []string{
				"/usr/local/bin/ollama",
				"/usr/bin/ollama",
			}
			for _, c := range candidates {
				if _, statErr := os.Stat(c); statErr == nil {
					ollamaBin = c
					break
				}
			}
		}
	}

	if ollamaBin == "" {
		return fmt.Errorf("ollama executable not found on system")
	}

	// Spawn 'ollama serve' detached in background
	cmd := exec.Command(ollamaBin, "serve")
	setDetached(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to auto-start ollama serve: %w", err)
	}

	// Poll up to 3.5 seconds for server to start listening
	ep := NormalizeEndpoint(endpoint)
	deadline := time.Now().Add(3500 * time.Millisecond)
	for time.Now().Before(deadline) {
		time.Sleep(150 * time.Millisecond)
		checkClient := &http.Client{Timeout: 250 * time.Millisecond}
		resp, checkErr := checkClient.Get(ep + "/api/tags")
		if checkErr == nil {
			resp.Body.Close()
			return nil // Ollama is ready!
		}
	}

	return nil
}
