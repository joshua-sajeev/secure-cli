package auth

import (
	"database/sql"
	"fmt"
	"time"
)

// UserDetails contains information about a user
type UserDetails struct {
	ID             int64
	Username       string
	CreatedAt      time.Time
	LastLogin      *time.Time
	FailedAttempts int
	LockedOut      bool
}

// parseTimestamp parses various SQLite timestamp formats
func parseTimestamp(ts string) (time.Time, error) {
	if ts == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp %q: %w", ts, err)
	}

	return t, nil
}

// GetUserDetails retrieves detailed information about a user by ID
func GetUserDetails(db *sql.DB, userID int64) (*UserDetails, error) {
	var username string
	var createdAtStr string
	var lastLoginStr sql.NullString
	var failedAttempts int
	var lockoutUntilStr sql.NullString

	err := db.QueryRow(`
		SELECT 
			id, 
			username, 
			created_at, 
			last_login,
			failed_attempts,
			lockout_until
		FROM users 
		WHERE id = ?
	`, userID).Scan(
		&userID,
		&username,
		&createdAtStr,
		&lastLoginStr,
		&failedAttempts,
		&lockoutUntilStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query user details: %w", err)
	}

	details := &UserDetails{
		ID:             userID,
		Username:       username,
		FailedAttempts: failedAttempts,
	}

	// Parse created_at timestamp
	createdAt, err := parseTimestamp(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	details.CreatedAt = createdAt

	// Parse last_login if available
	if lastLoginStr.Valid && lastLoginStr.String != "" {
		lastLogin, err := parseTimestamp(lastLoginStr.String)
		if err == nil {
			details.LastLogin = &lastLogin
		}
	}

	// Check if user is locked out
	if lockoutUntilStr.Valid && lockoutUntilStr.String != "" {
		lockoutUntil, err := parseTimestamp(lockoutUntilStr.String)
		if err == nil && lockoutUntil.After(time.Now()) {
			details.LockedOut = true
		}
	}

	return details, nil
}
