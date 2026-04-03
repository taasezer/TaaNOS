package intent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ollamaRequest is the request body for Ollama's /api/generate endpoint.
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// ollamaResponse is the response body from Ollama's /api/generate endpoint.
type ollamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
	TotalDuration  int64 `json:"total_duration"`
	LoadDuration   int64 `json:"load_duration"`
	PromptEvalCount int  `json:"prompt_eval_count"`
	EvalCount       int  `json:"eval_count"`
	EvalDuration    int64 `json:"eval_duration"`
}

// ExtractorConfig holds configuration for the intent extractor.
type ExtractorConfig struct {
	Endpoint   string
	Model      string
	Timeout    time.Duration
	MaxRetries int
}

// Extractor connects to Ollama and extracts structured intent from natural language.
type Extractor struct {
	config ExtractorConfig
	client *http.Client
}

// NewExtractor creates an Extractor with the given config.
func NewExtractor(cfg ExtractorConfig) *Extractor {
	return &Extractor{
		config: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Extract sends the user's input to Ollama and returns a validated IntentResult.
// Retries up to MaxRetries on parse/validation failures.
func (e *Extractor) Extract(userInput string) (*IntentResult, error) {
	var lastErr error

	for attempt := 0; attempt <= e.config.MaxRetries; attempt++ {
		result, err := e.doExtract(userInput, attempt)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// Don't retry on connection errors — they won't resolve
		if isConnectionError(err) {
			return nil, fmt.Errorf("Ollama connection failed: %w\n\nMake sure Ollama is running:\n  ollama serve\n\nEndpoint: %s",
				err, e.config.Endpoint)
		}
	}

	return nil, fmt.Errorf("intent extraction failed after %d attempts: %w",
		e.config.MaxRetries+1, lastErr)
}

// CheckConnection verifies that Ollama is reachable.
func (e *Extractor) CheckConnection() error {
	resp, err := e.client.Get(e.config.Endpoint)
	if err != nil {
		return fmt.Errorf("cannot reach Ollama at %s: %w", e.config.Endpoint, err)
	}
	defer resp.Body.Close()
	return nil
}

// doExtract performs a single extraction attempt.
func (e *Extractor) doExtract(userInput string, attempt int) (*IntentResult, error) {
	start := time.Now()

	prompt := fmt.Sprintf(UserPromptTemplate, userInput)

	reqBody := ollamaRequest{
		Model:  e.config.Model,
		Prompt: prompt,
		System: SystemPrompt,
		Stream: false,
		Format: "json",
		Options: map[string]interface{}{
			"temperature": 0.1, // Low temperature for deterministic output
			"top_p":       0.9,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", e.config.Endpoint)
	resp, err := e.client.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("Ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Ollama response: %w", err)
	}

	// Parse the Ollama response wrapper
	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response envelope: %w", err)
	}

	// Parse and validate the LLM's actual response (the JSON intent)
	result, err := ParseAndValidate([]byte(ollamaResp.Response))
	if err != nil {
		return nil, fmt.Errorf("attempt %d: %w", attempt+1, err)
	}

	// Attach metadata
	result.RawLLMResponse = ollamaResp.Response
	result.ExtractionTimeMs = time.Since(start).Milliseconds()

	return result, nil
}

// isConnectionError checks if the error is a network/connection issue.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	connectionIndicators := []string{
		"connection refused",
		"no such host",
		"network is unreachable",
		"cannot reach",
		"dial tcp",
		"i/o timeout",
	}
	for _, indicator := range connectionIndicators {
		if contains(errStr, indicator) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
