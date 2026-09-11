package cli

import (
	"database/sql"
	"fmt"

	"github.com/ergochat/readline"

	"github.com/joshu-sajeev/secure-cli/internal/auth"
)

// WhoAmI displays the current user's account details.
func WhoAmI(db *sql.DB, rl *readline.Instance, userID int64) {
	if userID == 0 {
		fmt.Fprintln(rl, "You are not logged in.")
		return
	}

	details, err := auth.GetUserDetails(db, userID)
	if err != nil {
		fmt.Fprintf(rl, "Error retrieving user details: %v\n", err)
		return
	}

	fmt.Fprintln(rl)
	fmt.Fprintln(rl, "┌─────────────────────────────────────────────┐")
	fmt.Fprintln(rl, "│              User Account Details           │")
	fmt.Fprintln(rl, "├─────────────────────────────────────────────┤")

	fmt.Fprintf(rl, "│ %-17s %-25s │\n", "Username:", details.Username)
	fmt.Fprintf(rl, "│ %-17s %-25d │\n", "User ID:", details.ID)
	fmt.Fprintf(
		rl,
		"│ %-17s %-25s │\n",
		"Registered:",
		details.CreatedAt.Format("2006-01-02 15:04:05"),
	)

	if details.LastLogin != nil {
		fmt.Fprintf(
			rl,
			"│ %-17s %-25s │\n",
			"Last Login:",
			details.LastLogin.Format("2006-01-02 15:04:05"),
		)
	} else {
		fmt.Fprintf(rl, "│ %-17s %-25s │\n", "Last Login:", "Never")
	}

	fmt.Fprintln(rl, "└─────────────────────────────────────────────┘")
	fmt.Fprintln(rl)
}
