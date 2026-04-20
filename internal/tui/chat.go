package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// chatRequest is a simple Ollama generate request for conversational mode.
type chatRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	System  string                 `json:"system"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type chatResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

const chatSystemPrompt = `You are TaaNOS, a friendly and knowledgeable AI system assistant.
You can have casual conversations, answer questions, and help with anything.
Keep responses concise (2-4 sentences). Be natural, helpful, and conversational.
IMPORTANT: Always respond in the SAME LANGUAGE the user writes in. If they write in Turkish, respond in Turkish. If they write in Japanese, respond in Japanese. Match their language exactly.
If the user asks about system tasks, remind them they can use commands like "install nginx" or "check disk space".`

// Chat sends a conversational message to Ollama (no JSON mode, no intent extraction).
func Chat(endpoint, model, userInput string, timeout time.Duration) (string, error) {
	reqBody := chatRequest{
		Model:  model,
		Prompt: userInput,
		System: chatSystemPrompt,
		Stream: false,
		Options: map[string]interface{}{
			"num_predict":  200,
			"temperature":  0.7,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(endpoint+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("Ollama connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama HTTP %d: %s", resp.StatusCode, string(b))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", err
	}

	return chatResp.Response, nil
}
