package intent

import (
	"bytes"
	"context"
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
			// Do NOT set Client.Timeout here.
			// We use per-request context timeouts instead,
			// because model cold-start can exceed the inference timeout.
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

		// Don't retry on connection/timeout errors — they won't resolve
		if isConnectionError(err) {
			return nil, fmt.Errorf("Ollama connection failed: %w\n\nMake sure Ollama is running:\n  ollama serve\n\nTips:\n  - Increase timeout: taanos config set ollama.timeout 120s\n  - Use a smaller model: taanos model tinyllama\n\nEndpoint: %s",
				err, e.config.Endpoint)
		}
	}

	return nil, fmt.Errorf("intent extraction failed after %d attempts: %w",
		e.config.MaxRetries+1, lastErr)
}

// CheckConnection verifies that Ollama is reachable.
func (e *Extractor) CheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", e.config.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot reach Ollama at %s: %w", e.config.Endpoint, err)
	}
	resp, err := e.client.Do(req)
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
		System: BuildSystemPrompt(),
		Stream: false,
		Format: "json",
		Options: map[string]interface{}{
			"temperature": 0.0,  // Fully deterministic output
			"top_p":       0.9,
			"num_predict": 256,  // Intent JSON never exceeds ~200 tokens
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", e.config.Endpoint)

	// Use context-based timeout for the inference request.
	// This allows the model to load (cold start can take 60s+).
	ctx, cancel := context.WithTimeout(context.Background(), e.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
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
		return nil, fmt.Errorf("attempt %d: %w (model_response: %q)", attempt+1, err, ollamaResp.Response)
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
		"context deadline exceeded",
		"Client.Timeout",
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
