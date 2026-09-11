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
	ErrUserAlreadyExists  = errors.New("username already exists")
	ErrWeakPassword       = errors.New("password must be at least 8 characters long")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrLockedOut         = errors.New("account is locked out due to too many failed attempts")
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
func Register(db *sql.DB, username, password string) (int64, error) {
	if username == "" {
		return 0, ErrInvalidUsername
	}

	if password == "" {
		return 0, ErrInvalidPassword
	}

	if len(password) < 8 {
		return 0, ErrWeakPassword
	}

	if len([]byte(password)) > 72 {
		return 0, ErrPasswordTooLong
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}

	const query = `
		INSERT INTO users (username, password_hash)
		VALUES (?, ?)
	`

	result, err := db.Exec(query, username, passwordHash)
	if err != nil {
		if sqliteErr, ok := errors.AsType[*sqlite.Error](err); ok {
			if sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				return 0, ErrUserAlreadyExists
			}
		}
		return 0, fmt.Errorf("create user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get user id: %w", err)
	}

	return userID, nil
}
