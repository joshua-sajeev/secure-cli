package auth

import (
	"database/sql"
	"fmt"
	"time"
)

// SessionConfig defines session parameters
type SessionConfig struct {
	SessionDuration   time.Duration
	InactivityTimeout time.Duration
}

// DefaultSessionConfig returns default session configuration
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		SessionDuration:   24 * time.Hour,
		InactivityTimeout: 15 * time.Minute,
	}
}

// Session represents an active user session
type Session struct {
	ID           int64
	UserID       int64
	CreatedAt    time.Time
	LastActivity time.Time
	ExpiresAt    time.Time
}

// CreateSession creates a new session for a user
func CreateSession(db *sql.DB, userID int64, config SessionConfig) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("invalid user id: %d", userID)
	}

	expiresAt := time.Now().Add(config.SessionDuration)

	result, err := db.Exec(
		`INSERT INTO sessions (user_id, expires_at, last_activity) 
		 VALUES (?, ?, CURRENT_TIMESTAMP)`,
		userID,
		expiresAt.Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("create session: %w", err)
	}

	sessionID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get session id: %w", err)
	}

	return sessionID, nil
}

// GetSession retrieves session details by session ID
func GetSession(db *sql.DB, sessionID int64) (*Session, error) {
	var userID int64
	var createdAtStr, lastActivityStr, expiresAtStr string

	err := db.QueryRow(
		`SELECT id, user_id, created_at, last_activity, expires_at 
		 FROM sessions WHERE id = ?`,
		sessionID,
	).Scan(&sessionID, &userID, &createdAtStr, &lastActivityStr, &expiresAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("query session: %w", err)
	}

	createdAt, err := parseTimestamp(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	lastActivity, err := parseTimestamp(lastActivityStr)
	if err != nil {
		return nil, fmt.Errorf("parse last_activity: %w", err)
	}

	expiresAt, err := parseTimestamp(expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse expires_at: %w", err)
	}

	return &Session{
		ID:           sessionID,
		UserID:       userID,
		CreatedAt:    createdAt,
		LastActivity: lastActivity,
		ExpiresAt:    expiresAt,
	}, nil
}

// IsSessionValid checks if a session is still valid (not expired, not inactive)
func IsSessionValid(db *sql.DB, sessionID int64, config SessionConfig) (bool, error) {
	session, err := GetSession(db, sessionID)
	if err != nil {
		return false, err
	}

	now := time.Now()

	if now.After(session.ExpiresAt) {
		_ = DeleteSession(db, sessionID)
		return false, fmt.Errorf("session expired")
	}

	inactivityLimit := session.LastActivity.Add(config.InactivityTimeout)
	if now.After(inactivityLimit) {
		_ = DeleteSession(db, sessionID)
		return false, fmt.Errorf("session inactive timeout")
	}

	return true, nil
}

// UpdateSessionActivity updates the last activity timestamp for a session
func UpdateSessionActivity(db *sql.DB, sessionID int64) error {
	_, err := db.Exec(
		`UPDATE sessions SET last_activity = CURRENT_TIMESTAMP WHERE id = ?`,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("update session activity: %w", err)
	}
	return nil
}

// DeleteSession removes a session from the database
func DeleteSession(db *sql.DB, sessionID int64) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteUserSessions removes all sessions for a user
func DeleteUserSessions(db *sql.DB, userID int64) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete user sessions: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes all expired sessions from the database
func CleanupExpiredSessions(db *sql.DB) error {
	_, err := db.Exec(
		`DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`,
	)
	if err != nil {
		return fmt.Errorf("cleanup expired sessions: %w", err)
	}
	return nil
}
