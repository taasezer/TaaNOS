package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
You can have conversations, answer questions, write code, explain concepts, and help with anything.
Be natural and conversational. Give short answers for simple questions, detailed answers when needed. Write full code when asked.
IMPORTANT: Always respond in the SAME LANGUAGE the user writes in. If they write in Turkish, respond in Turkish. If they write in Japanese, respond in Japanese. Match their language exactly.
If the user asks about system tasks, remind them they can use commands like "install nginx" or "check disk space".
You have memory of this conversation session. Use it to give contextual answers.`

// ConversationEntry represents one exchange in conversation history.
type ConversationEntry struct {
	Role    string // "user" or "assistant"
	Content string
}

// BuildConversationPrompt creates a prompt with conversation history.
func BuildConversationPrompt(history []ConversationEntry, currentInput string) string {
	if len(history) == 0 {
		return currentInput
	}

	var b strings.Builder

	// Include last 6 exchanges (3 pairs) — compact for speed
	start := 0
	if len(history) > 6 {
		start = len(history) - 6
	}

	b.WriteString("Context:\n")
	for _, entry := range history[start:] {
		content := entry.Content
		// Truncate long responses to save tokens
		if len(content) > 80 {
			content = content[:80] + "..."
		}
		if entry.Role == "user" {
			b.WriteString("User: " + content + "\n")
		} else {
			b.WriteString("AI: " + content + "\n")
		}
	}
	b.WriteString("\nUser: " + currentInput)
	return b.String()
}

// Chat sends a conversational message to Ollama with session memory.
func Chat(endpoint, model, userInput string, timeout time.Duration) (string, error) {
	return ChatWithHistory(endpoint, model, userInput, nil, timeout)
}

// ChatWithHistory sends a message with conversation context.
func ChatWithHistory(endpoint, model, userInput string, history []ConversationEntry, timeout time.Duration) (string, error) {
	prompt := BuildConversationPrompt(history, userInput)

	reqBody := chatRequest{
		Model:  model,
		Prompt: prompt,
		System: chatSystemPrompt,
		Stream: false,
		Options: map[string]interface{}{
			"num_predict": 2048,
			"temperature": 0.7,
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
