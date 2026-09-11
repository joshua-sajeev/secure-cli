package cli

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ergochat/readline"

	"github.com/joshu-sajeev/secure-cli/internal/auth"
)

// Login handles user authentication with username and password
func Login(db *sql.DB, rl *readline.Instance) (int64, error) {
	fmt.Fprintln(rl, "Login")

	rl.DisableHistory()
	defer rl.EnableHistory()

	// Username input
	rl.SetPrompt("Username: ")
	username, err := rl.ReadLine()
	if err != nil {
		return 0, err
	}

	username = strings.TrimSpace(username)

	if username == "" {
		fmt.Fprintln(rl, "Username cannot be empty.")
		return 0, fmt.Errorf("empty username")
	}

	rl.SetPrompt("")
	fmt.Fprint(rl, "Password: ")

	password, err := readPassword()
	if err != nil {
		fmt.Fprintf(rl, "\nFailed to read password: %v\n", err)
		return 0, fmt.Errorf("failed to read password")
	}
	if len(password) == 0 {
		fmt.Fprintln(rl, "\nPassword cannot be empty.")
		return 0, fmt.Errorf("empty password")
	}

	rl.SetPrompt("secure-cli [guest]> ")

	userID, err := auth.Login(db, username, string(password))
	if err != nil {
		fmt.Fprintf(rl, "\nLogin failed: %v\n", err)
		return 0, err
	}

	fmt.Fprintln(rl, "\nLogin successful.")
	return userID, nil
}
