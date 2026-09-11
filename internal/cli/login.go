package cli

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ergochat/readline"

	"github.com/joshu-sajeev/secure-cli/internal/auth"
)

// Login handles user authentication with username and password.
func Login(db *sql.DB, rl *readline.Instance) (int64, error) {
	fmt.Fprintln(rl, "Login")

	rl.DisableHistory()
	defer rl.EnableHistory()

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

	return userID, nil
}

// LoginWithSession handles user authentication and registers a session.
func LoginWithSession(
	db *sql.DB,
	rl *readline.Instance,
	sessionConfig auth.SessionConfig,
) (int64, int64, error) {
	fmt.Fprintln(rl, "Login")

	rl.DisableHistory()
	defer rl.EnableHistory()

	rl.SetPrompt("Username: ")

	username, err := rl.ReadLine()
	if err != nil {
		return 0, 0, err
	}

	username = strings.TrimSpace(username)

	if username == "" {
		fmt.Fprintln(rl, "Username cannot be empty.")
		return 0, 0, fmt.Errorf("empty username")
	}

	rl.SetPrompt("")
	fmt.Fprint(rl, "Password: ")

	password, err := readPassword()
	if err != nil {
		fmt.Fprintf(rl, "\nFailed to read password: %v\n", err)
		return 0, 0, fmt.Errorf("failed to read password")
	}

	if len(password) == 0 {
		fmt.Fprintln(rl, "\nPassword cannot be empty.")
		return 0, 0, fmt.Errorf("empty password")
	}

	rl.SetPrompt("secure-cli [guest]> ")

	userID, err := auth.Login(db, username, string(password))
	if err != nil {
		fmt.Fprintf(rl, "\nLogin failed: %v\n", err)
		return 0, 0, err
	}

	userDetails, err := auth.GetUserDetails(db, userID)
	if err != nil {
		fmt.Fprintf(rl, "\nFailed to retrieve user details: %v\n", err)
		return 0, 0, fmt.Errorf("get user details: %w", err)
	}

	if userDetails.MFAEnabled {
		valid, err := VerifyTOTPDuringLogin(
			rl,
			userDetails.TOTPSecret,
		)

		if err != nil || !valid {
			return 0, 0, fmt.Errorf("2FA verification failed")
		}
	}

	sessionID, err := auth.CreateSession(
		db,
		userID,
		sessionConfig,
	)
	if err != nil {
		fmt.Fprintf(rl, "\nSession creation failed: %v\n", err)
		return 0, 0, err
	}

	fmt.Fprintln(rl, "\nLogin successful.")

	displayUserInfo(db, rl, userID)

	return userID, sessionID, nil
}

// displayUserInfo shows a formatted summary of user details.
func displayUserInfo(
	db *sql.DB,
	rl *readline.Instance,
	userID int64,
) {
	userDetails, err := auth.GetUserDetails(db, userID)
	if err != nil {
		fmt.Fprintf(
			rl,
			"Warning: Could not retrieve user details: %v\n",
			err,
		)
		return
	}

	fmt.Fprintln(
		rl,
		"═════════════════════════════════════════",
	)
	fmt.Fprintln(rl, "  User Details")
	fmt.Fprintln(
		rl,
		"═════════════════════════════════════════",
	)

	fmt.Fprintf(
		rl,
		"  Username:  %s\n",
		userDetails.Username,
	)

	fmt.Fprintf(
		rl,
		"  User ID:   %d\n",
		userDetails.ID,
	)

	fmt.Fprintf(
		rl,
		"  Created:   %s\n",
		userDetails.CreatedAt.Format("2006-01-02 15:04:05"),
	)

	if userDetails.LastLogin != nil {
		fmt.Fprintf(
			rl,
			"  Last Login: %s\n",
			userDetails.LastLogin.Format("2006-01-02 15:04:05"),
		)
	} else {
		fmt.Fprintln(
			rl,
			"  Last Login: Never",
		)
	}

	if userDetails.FailedAttempts > 0 {
		fmt.Fprintf(
			rl,
			"  Failed Attempts: %d\n",
			userDetails.FailedAttempts,
		)
	}

	fmt.Fprintln(
		rl,
		"═════════════════════════════════════════",
	)
	fmt.Fprintln(rl, "")
}
