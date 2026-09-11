package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ergochat/readline"

	"github.com/joshu-sajeev/secure-cli/internal/auth"
	"github.com/joshu-sajeev/secure-cli/internal/cli"
	"github.com/joshu-sajeev/secure-cli/internal/db"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║         Secure CLI Login System v1.1                      ║")
	fmt.Println("║    With Two-Factor Authentication Support                 ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/app/data/secure-cli.db"
	}

	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Critical error: %v", err)
	}

	_ = auth.CleanupExpiredSessions(database)

	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error while closing database: %v", err)
		}
	}()

	completer := readline.NewPrefixCompleter(
		readline.PcItem("register"),
		readline.PcItem("login"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
		readline.PcItem("whoami"),
		readline.PcItem("logout"),
		readline.PcItem("enable-2fa"),
		readline.PcItem("disable-2fa"),
		readline.PcItem("2fa-status"),
	)

	config := &readline.Config{
		Prompt:       "secure-cli [guest]> ",
		HistoryFile:  "~/.secure-cli_history",
		AutoComplete: completer,
	}

	rl, err := readline.NewFromConfig(config)
	if err != nil {
		log.Fatalf("Failed to initialize CLI: %v", err)
	}
	defer func() {
		if err := rl.Close(); err != nil {
			log.Fatalf("Failed to close readline instance: %v", err)
		}
	}()

	_, _ = fmt.Fprintln(rl, "Type 'help' for available commands or 'exit' to quit.")
	fmt.Fprintln(rl, "")

	var currentUserID int64
	var currentUsername string
	var currentSessionID int64
	sessionConfig := auth.DefaultSessionConfig()

	for {
		if currentSessionID > 0 {
			valid, checkErr := auth.IsSessionValid(database, currentSessionID, sessionConfig)
			if !valid || checkErr != nil {
				fmt.Fprintln(rl, "\n⏱️  Session expired. Please login again.")
				currentUserID = 0
				currentUsername = ""
				currentSessionID = 0
			} else {
				_ = auth.UpdateSessionActivity(database, currentSessionID)
			}
		}

		if currentUserID > 0 {
			rl.SetPrompt(fmt.Sprintf("secure-cli [%s]> ", currentUsername))
		} else {
			rl.SetPrompt("secure-cli [guest]> ")
		}

		input, err := rl.ReadLine()
		if err != nil {
			if currentSessionID > 0 {
				_ = auth.DeleteSession(database, currentSessionID)
			}
			fmt.Fprintln(rl, "")
			fmt.Fprintln(rl, "👋 Goodbye!")
			return
		}

		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		switch input {
		case "register":
			cli.Register(database, rl)

		case "login":
			userID, sessionID, err := cli.LoginWithSession(database, rl, sessionConfig)
			if err == nil {
				currentUserID = userID
				currentSessionID = sessionID

				user, _ := auth.GetUserDetails(database, userID)
				if user != nil {
					currentUsername = user.Username
				}
				fmt.Fprintln(rl, "")
			}

		case "whoami":
			if currentUserID == 0 {
				fmt.Fprintln(rl, "❌ You are not logged in. Use 'login' to authenticate.")
				continue
			}
			cli.WhoAmI(database, rl, currentUserID)

		case "logout":
			if currentUserID == 0 {
				fmt.Fprintln(rl, "❌ You are not logged in.")
				continue
			}
			if currentSessionID > 0 {
				_ = auth.DeleteSession(database, currentSessionID)
			}
			currentUserID = 0
			currentUsername = ""
			currentSessionID = 0
			fmt.Fprintln(rl, "✅ Logged out successfully.")

		case "enable-2fa":
			if currentUserID == 0 {
				fmt.Fprintln(rl, "❌ You must be logged in to enable 2FA.")
				continue
			}
			_ = cli.Enable2FA(database, rl, currentUserID, currentUsername)

		case "disable-2fa":
			if currentUserID == 0 {
				fmt.Fprintln(rl, "❌ You must be logged in to disable 2FA.")
				continue
			}
			_ = cli.Disable2FA(database, rl, currentUserID, currentUsername)

		case "2fa-status":
			if currentUserID == 0 {
				fmt.Fprintln(rl, "❌ You must be logged in to check 2FA status.")
				continue
			}
			_ = cli.Status2FA(database, rl, currentUserID)

		case "help":
			showHelp(rl, currentUserID > 0)

		case "exit":
			if currentSessionID > 0 {
				_ = auth.DeleteSession(database, currentSessionID)
			}
			fmt.Fprintln(rl, "")
			fmt.Fprintln(rl, "👋 Goodbye!")
			return

		default:
			fmt.Fprintf(rl, "❌ Unknown command: '%s' (type 'help' for available commands)\n", input)
		}
	}
}

// showHelp displays context-aware help message
func showHelp(rl *readline.Instance, isLoggedIn bool) {
	if !isLoggedIn {
		fmt.Fprintln(rl, "")
		fmt.Fprintln(rl, "╔════════════════════════════════════════════════════════════╗")
		fmt.Fprintln(rl, "║                     GUEST COMMANDS                         ║")
		fmt.Fprintln(rl, "╠════════════════════════════════════════════════════════════╣")
		fmt.Fprintln(rl, "║                                                            ║")
		fmt.Fprintln(rl, "║  register                Create a new account              ║")
		fmt.Fprintln(rl, "║  login                   Authenticate with credentials     ║")
		fmt.Fprintln(rl, "║  help                    Show this help message            ║")
		fmt.Fprintln(rl, "║  exit                    Quit the application              ║")
		fmt.Fprintln(rl, "║                                                            ║")
		fmt.Fprintln(rl, "╚════════════════════════════════════════════════════════════╝")
		fmt.Fprintln(rl, "")
	} else {
		fmt.Fprintln(rl, "")
		fmt.Fprintln(rl, "╔════════════════════════════════════════════════════════════╗")
		fmt.Fprintln(rl, "║                    AUTHENTICATED COMMANDS                  ║")
		fmt.Fprintln(rl, "╠════════════════════════════════════════════════════════════╣")
		fmt.Fprintln(rl, "║                                                            ║")
		fmt.Fprintln(rl, "║  whoami                  Display your account details      ║")
		fmt.Fprintln(rl, "║  logout                  End your current session          ║")
		fmt.Fprintln(rl, "║                                                            ║")
		fmt.Fprintln(rl, "║  TWO-FACTOR AUTHENTICATION:                                ║")
		fmt.Fprintln(rl, "║  enable-2fa              Enable TOTP-based 2FA             ║")
		fmt.Fprintln(rl, "║  disable-2fa             Disable 2FA                       ║")
		fmt.Fprintln(rl, "║  2fa-status              Show current 2FA status           ║")
		fmt.Fprintln(rl, "║                                                            ║")
		fmt.Fprintln(rl, "║  help                    Show this help message            ║")
		fmt.Fprintln(rl, "║  exit                    Quit the application              ║")
		fmt.Fprintln(rl, "║                                                            ║")
		fmt.Fprintln(rl, "╚════════════════════════════════════════════════════════════╝")
		fmt.Fprintln(rl, "")
	}
}
