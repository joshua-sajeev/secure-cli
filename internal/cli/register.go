// Package cli handles user interaction and command processing for the application.
package cli

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ergochat/readline"
	"github.com/joshu-sajeev/secure-cli/internal/auth"
)

// Register handles new user registration with username and password validation.
func Register(db *sql.DB, rl *readline.Instance) {
	fmt.Fprintln(rl, "Register")

	rl.DisableHistory()
	defer rl.EnableHistory()

	// Username input.
	rl.SetPrompt("Username: ")

	username, err := rl.ReadLine()
	if err != nil {
		return
	}

	username = strings.TrimSpace(username)

	if username == "" {
		fmt.Fprintln(rl, "Username cannot be empty.")
		return
	}

	// Check if username already exists before asking for password.
	exists, err := auth.CheckUsernameExists(db, username)
	if err != nil {
		fmt.Fprintf(rl, "Error checking username: %v\n", err)
		return
	}

	if exists {
		fmt.Fprintln(rl, "Username already exists.")
		return
	}

	// Password input.
	rl.SetPrompt("")
	fmt.Fprint(rl, "Password: ")

	password, err := readPassword()
	if err != nil {
		fmt.Fprintf(rl, "\nFailed to read password: %v\n", err)
		return
	}

	fmt.Fprintln(rl)

	// Bcrypt limit: 72 bytes.
	if len(password) > 72 {
		fmt.Fprintln(rl, "Password exceeds 72 bytes.")
		return
	}

	if len(password) == 0 {
		fmt.Fprintln(rl, "Password cannot be empty.")
		return
	}

	rl.SetPrompt("secure-cli [guest]> ")

	_, err = auth.Register(db, username, string(password))
	if err != nil {
		fmt.Fprintf(rl, "Registration failed: %v\n", err)
		return
	}

	fmt.Fprintln(rl, "Registration successful.")
}
