package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Secure CLI starting up...")
	fmt.Println("Type 'help' or 'exit' to quit.")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("secure-cli [guest]> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "help" {
			fmt.Println("Available commands: register, login, help, exit")
			continue
		}

		fmt.Printf("Unknown command: %s\n", input)
	}
}
