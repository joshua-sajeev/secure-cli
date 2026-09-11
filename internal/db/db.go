// Package db manages SQLite connections and schema initialization.
package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// InitDB initializes the SQLite database connection using the provided
// Data Source Name (DSN), verifies connectivity via Ping, and sets up
// the required database schema automatically.
func InitDB(dataSourceName string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	if err := createSchema(database); err != nil {
		return nil, fmt.Errorf("failed to create database schema: %w", err)
	}

	log.Println("SQLite database initialized and schema applied successfully.")
	return database, nil
}

// createSchema runs initial DDL statements to create the users table
// supporting authentication, lockout tracking, and session timestamps.
func createSchema(db *sql.DB) error {
	schemaQuery := `
	PRAGMA foreign_keys = ON;

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		failed_attempts INTEGER DEFAULT 0,
		lockout_until DATETIME DEFAULT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_login DATETIME DEFAULT NULL,
		totp_secret TEXT DEFAULT NULL,
		totp_enabled INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_activity DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`

	_, err := db.Exec(schemaQuery)
	if err != nil {
		return fmt.Errorf("executing schema DDL failed: %w", err)
	}

	_, _ = db.Exec("ALTER TABLE users ADD COLUMN totp_secret TEXT DEFAULT NULL")
	_, _ = db.Exec("ALTER TABLE users ADD COLUMN totp_enabled INTEGER DEFAULT 0")

	return nil
}
