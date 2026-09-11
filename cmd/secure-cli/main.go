package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ergochat/readline"

	"github.com/joshu-sajeev/secure-cli/internal/cli"
	"github.com/joshu-sajeev/secure-cli/internal/db"
)

func main() {
	fmt.Println("Secure CLI starting up...")

	// Initialize the SQLite database and run migrations.
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/app/data/secure-cli.db"
	}

	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Critical error: %v", err)
	}

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

	var currentUserID int64

	for {
		if currentUserID > 0 {
			rl.SetPrompt("secure-cli [user]> ")
		} else {
			rl.SetPrompt("secure-cli [guest]> ")
		}

		input, err := rl.ReadLine()
		if err != nil {
			_, _ = fmt.Fprintln(rl, "\nGoodbye!")
			return
		}

		input = strings.TrimSpace(input)

		switch input {
		case "":
			continue

		case "register":
			cli.Register(database, rl)

		case "login":
			userID, err := cli.Login(database, rl)
			if err == nil {
				currentUserID = userID
				fmt.Fprintln(rl, "")
			}

		case "whoami":
			if currentUserID == 0 {
				fmt.Fprintln(rl, "You are not logged in. Use 'login' to authenticate.")
				continue
			}
			cli.WhoAmI(database, rl, currentUserID)

		case "logout":
			if currentUserID == 0 {
				fmt.Fprintln(rl, "You are not logged in.")
				continue
			}
			currentUserID = 0
			fmt.Fprintln(rl, "Logged out successfully.")

		case "help":
			showHelp(rl, currentUserID > 0)

		case "exit":
			_, _ = fmt.Fprintln(rl, "Goodbye!")
			return

		default:
			_, _ = fmt.Fprintf(rl, "Unknown command: %s\n", input)
		}
	}
}

// showHelp displays help message based on login state
func showHelp(rl *readline.Instance, isLoggedIn bool) {
	if !isLoggedIn {
		fmt.Fprintln(rl, "")
		fmt.Fprintln(rl, "═════════════════════════════════════════")
		fmt.Fprintln(rl, "  Guest Commands")
		fmt.Fprintln(rl, "═════════════════════════════════════════")
		fmt.Fprintln(rl, "  register       Create a new account")
		fmt.Fprintln(rl, "  login          Authenticate with username and password")
		fmt.Fprintln(rl, "  help           Show this help message")
		fmt.Fprintln(rl, "  exit           Quit the application")
		fmt.Fprintln(rl, "═════════════════════════════════════════")
		fmt.Fprintln(rl, "")
	} else {
		fmt.Fprintln(rl, "")
		fmt.Fprintln(rl, "═════════════════════════════════════════")
		fmt.Fprintln(rl, "  User Commands")
		fmt.Fprintln(rl, "═════════════════════════════════════════")
		fmt.Fprintln(rl, "  whoami         Display your account details")
		fmt.Fprintln(rl, "  logout         End your current session")
		fmt.Fprintln(rl, "  help           Show this help message")
		fmt.Fprintln(rl, "  exit           Quit the application")
		fmt.Fprintln(rl, "═════════════════════════════════════════")
		fmt.Fprintln(rl, "")
	}
}
