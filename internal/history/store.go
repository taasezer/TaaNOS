package history

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides SQLite-backed plan history for TaaNOS.
type Store struct {
	db *sql.DB
}

// PlanRecord represents a stored plan execution record.
type PlanRecord struct {
	ID              string
	PlanID          string
	Intent          string
	Category        string
	Action          string
	Target          string
	Status          string // success, partial_failure, failure, aborted, explain
	RiskLevel       string
	StepsTotal      int
	StepsCompleted  int
	DurationMs      int64
	StepsJSON       string // JSON-encoded step results
	OS              string
	PkgManager      string
	User            string
	CreatedAt       time.Time
}

// NewStore creates or opens a SQLite history database.
func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "history.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open history database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return store, nil
}

// migrate creates tables if they don't exist.
func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS plan_history (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		plan_id         TEXT NOT NULL UNIQUE,
		intent          TEXT NOT NULL,
		category        TEXT NOT NULL,
		action          TEXT NOT NULL,
		target          TEXT DEFAULT '',
		status          TEXT NOT NULL,
		risk_level      TEXT DEFAULT 'low',
		steps_total     INTEGER DEFAULT 0,
		steps_completed INTEGER DEFAULT 0,
		duration_ms     INTEGER DEFAULT 0,
		steps_json      TEXT DEFAULT '[]',
		os_info         TEXT DEFAULT '',
		pkg_manager     TEXT DEFAULT '',
		user_name       TEXT DEFAULT '',
		created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_plan_history_status ON plan_history(status);
	CREATE INDEX IF NOT EXISTS idx_plan_history_category ON plan_history(category);
	CREATE INDEX IF NOT EXISTS idx_plan_history_created ON plan_history(created_at);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Save stores a plan execution record.
func (s *Store) Save(record *PlanRecord) error {
	_, err := s.db.Exec(`
		INSERT INTO plan_history (
			plan_id, intent, category, action, target,
			status, risk_level, steps_total, steps_completed,
			duration_ms, steps_json, os_info, pkg_manager, user_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.PlanID, record.Intent, record.Category, record.Action, record.Target,
		record.Status, record.RiskLevel, record.StepsTotal, record.StepsCompleted,
		record.DurationMs, record.StepsJSON, record.OS, record.PkgManager, record.User,
	)
	if err != nil {
		return fmt.Errorf("failed to save plan record: %w", err)
	}
	return nil
}

// GetRecent returns the N most recent plan records.
func (s *Store) GetRecent(limit int) ([]PlanRecord, error) {
	rows, err := s.db.Query(`
		SELECT plan_id, intent, category, action, target,
			   status, risk_level, steps_total, steps_completed,
			   duration_ms, os_info, pkg_manager, user_name, created_at
		FROM plan_history
		ORDER BY created_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var records []PlanRecord
	for rows.Next() {
		var r PlanRecord
		err := rows.Scan(
			&r.PlanID, &r.Intent, &r.Category, &r.Action, &r.Target,
			&r.Status, &r.RiskLevel, &r.StepsTotal, &r.StepsCompleted,
			&r.DurationMs, &r.OS, &r.PkgManager, &r.User, &r.CreatedAt,
		)
		if err != nil {
			continue
		}
		records = append(records, r)
	}

	return records, nil
}

// GetByStatus returns records filtered by status.
func (s *Store) GetByStatus(status string, limit int) ([]PlanRecord, error) {
	rows, err := s.db.Query(`
		SELECT plan_id, intent, category, action, target,
			   status, risk_level, steps_total, steps_completed,
			   duration_ms, os_info, pkg_manager, user_name, created_at
		FROM plan_history
		WHERE status = ?
		ORDER BY created_at DESC
		LIMIT ?`, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var records []PlanRecord
	for rows.Next() {
		var r PlanRecord
		err := rows.Scan(
			&r.PlanID, &r.Intent, &r.Category, &r.Action, &r.Target,
			&r.Status, &r.RiskLevel, &r.StepsTotal, &r.StepsCompleted,
			&r.DurationMs, &r.OS, &r.PkgManager, &r.User, &r.CreatedAt,
		)
		if err != nil {
			continue
		}
		records = append(records, r)
	}

	return records, nil
}

// GetByPlanID retrieves a specific plan record by its ID.
func (s *Store) GetByPlanID(planID string) (*PlanRecord, error) {
	var r PlanRecord
	err := s.db.QueryRow(`
		SELECT plan_id, intent, category, action, target,
			   status, risk_level, steps_total, steps_completed,
			   duration_ms, steps_json, os_info, pkg_manager, user_name, created_at
		FROM plan_history
		WHERE plan_id = ?`, planID).Scan(
		&r.PlanID, &r.Intent, &r.Category, &r.Action, &r.Target,
		&r.Status, &r.RiskLevel, &r.StepsTotal, &r.StepsCompleted,
		&r.DurationMs, &r.StepsJSON, &r.OS, &r.PkgManager, &r.User, &r.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plan %s not found", planID)
		}
		return nil, fmt.Errorf("failed to query plan: %w", err)
	}
	return &r, nil
}

// Count returns the total number of records.
func (s *Store) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM plan_history").Scan(&count)
	return count, err
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// MarshalStepResults converts step results to JSON for storage.
func MarshalStepResults(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return "[]"
	}
	return string(b)
}
