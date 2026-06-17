package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/taasezer/TaaNOS/config"
)

// SessionRecord represents one saved REPL session.
type SessionRecord struct {
	ID           string              `json:"id"`
	StartedAt    string              `json:"started_at"`
	Conversation []ConversationEntry `json:"conversation"`
	History      []savedHistoryEntry `json:"history"`
}

// savedHistoryEntry is a serializable version of historyEntry.
type savedHistoryEntry struct {
	Input      string `json:"input"`
	Output     string `json:"output"`
	IsPipeline bool   `json:"is_pipeline"`
	IsErr      bool   `json:"is_err"`
	Time       string `json:"time"`
}

// sessionsFile returns the path to the sessions JSON file.
func sessionsFile() string {
	return filepath.Join(config.DataDir(), "sessions.json")
}

// LoadSessions reads all saved sessions from disk.
func LoadSessions() ([]SessionRecord, error) {
	data, err := os.ReadFile(sessionsFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sessions []SessionRecord
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

// SaveSession appends the current session to disk.
func SaveSession(conversation []ConversationEntry, history []historyEntry) error {
	sessions, _ := LoadSessions()

	// Keep max 20 sessions
	if len(sessions) >= 20 {
		sessions = sessions[len(sessions)-19:]
	}

	// Convert history entries
	saved := make([]savedHistoryEntry, len(history))
	for i, h := range history {
		saved[i] = savedHistoryEntry{
			Input:      h.input,
			Output:     h.output,
			IsPipeline: h.isPipeline,
			IsErr:      h.isErr,
			Time:       h.time,
		}
	}

	session := SessionRecord{
		ID:           time.Now().Format("20060102-150405"),
		StartedAt:    time.Now().Format("2006-01-02 15:04:05"),
		Conversation: conversation,
		History:      saved,
	}

	sessions = append(sessions, session)

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(sessionsFile()), 0755)
	return os.WriteFile(sessionsFile(), data, 0644)
}

// LoadLastSession returns the most recent session for display.
func LoadLastSession() *SessionRecord {
	sessions, err := LoadSessions()
	if err != nil || len(sessions) == 0 {
		return nil
	}
	return &sessions[len(sessions)-1]
}

// LoadSessionByID finds a session by its ID.
func LoadSessionByID(id string) *SessionRecord {
	sessions, err := LoadSessions()
	if err != nil {
		return nil
	}
	for i := range sessions {
		if sessions[i].ID == id {
			return &sessions[i]
		}
	}
	return nil
}
