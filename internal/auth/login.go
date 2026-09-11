package auth

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Login verifies the username and password.
func Login(db *sql.DB, username, password string) (int64, error) {
	if username == "" {
		return 0, ErrInvalidUsername
	}

	if password == "" {
		return 0, ErrInvalidPassword
	}

	var userID int64
	var passwordHash string
	var failedAttempts int
	var lockoutUntilStr sql.NullString

	err := db.QueryRow(
		"SELECT id, password_hash, failed_attempts, lockout_until FROM users WHERE username = ?",
		username,
	).Scan(
		&userID,
		&passwordHash,
		&failedAttempts,
		&lockoutUntilStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrInvalidCredentials
		}

		return 0, fmt.Errorf("query user: %w", err)
	}

	if lockoutUntilStr.Valid && lockoutUntilStr.String != "" {
		lockoutUntil, err := parseTimestamp(lockoutUntilStr.String)

		if err == nil && lockoutUntil.After(time.Now()) {
			return 0, ErrLockedOut
		}
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	); err != nil {

		failedAttempts++

		var lockoutUntil sql.NullString

		if failedAttempts >= 5 {
			lockoutTime := time.Now().
				Add(15 * time.Minute).
				Format(time.RFC3339)

			lockoutUntil = sql.NullString{
				String: lockoutTime,
				Valid:  true,
			}
		}

		_, updateErr := db.Exec(
			"UPDATE users SET failed_attempts = ?, lockout_until = ? WHERE id = ?",
			failedAttempts,
			lockoutUntil,
			userID,
		)

		if updateErr != nil {
			return 0, fmt.Errorf(
				"update failed attempts: %w",
				updateErr,
			)
		}

		return 0, ErrInvalidCredentials
	}

	_, err = db.Exec(
		"UPDATE users SET failed_attempts = 0, lockout_until = NULL WHERE id = ?",
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"reset failed attempts: %w",
			err,
		)
	}

	return userID, nil
}

// UpdateLastLogin records the successful completion of authentication.
func UpdateLastLogin(db *sql.DB, userID int64) error {
	_, err := db.Exec(
		"UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?",
		userID,
	)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	return nil
}
