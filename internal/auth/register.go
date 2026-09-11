// Package auth handles user registration, password hashing, and authentication logic.
package auth

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrInvalidUsername    = errors.New("username cannot be empty")
	ErrInvalidPassword    = errors.New("password cannot be empty")
	ErrPasswordTooLong    = errors.New("password exceeds 72 bytes")
	ErrUserExists         = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

// CheckUsernameExists checks if a username is already registered in the database
func CheckUsernameExists(db *sql.DB, username string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check username: %w", err)
	}
	return exists, nil
}

// Register creates a new user account with username and password
func Register(db *sql.DB, username, password string) error {
	if username == "" {
		return ErrInvalidUsername
	}

	if password == "" {
		return ErrInvalidPassword
	}

	if len([]byte(password)) > 72 {
		return ErrPasswordTooLong
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	const query = `
		INSERT INTO users (username, password_hash)
		VALUES (?, ?)
	`

	_, err = db.Exec(query, username, passwordHash)
	if err == nil {
		return nil
	}
	if sqliteErr, ok := errors.AsType[*sqlite.Error](err); ok {
		if sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return ErrUserExists
		}
	}

	return fmt.Errorf("create user: %w", err)
}
