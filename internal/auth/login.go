package auth

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Login authenticates a user with username and password
func Login(db *sql.DB, username, password string) (int64, error) {
	if username == "" {
		return 0, ErrInvalidUsername
	}

	if password == "" {
		return 0, ErrInvalidPassword
	}

	var userID int64
	var passwordHash string

	err := db.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = ?",
		username,
	).Scan(&userID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrInvalidCredentials
		}
		return 0, fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return 0, ErrInvalidCredentials
	}

	_, err = db.Exec("UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?", userID)
	if err != nil {
		return 0, fmt.Errorf("update last login: %w", err)
	}

	return userID, nil
}
