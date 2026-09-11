package cli

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func readPassword() ([]byte, error) {
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer term.Restore(fd, oldState)

	var password []byte
	buf := make([]byte, 1)

	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return nil, err
		}

		switch buf[0] {
		case '\r', '\n':
			fmt.Fprintln(os.Stdout)
			return password, nil

		case 127, 8: // Backspace
			if len(password) > 0 {
				password = password[:len(password)-1]
				fmt.Fprint(os.Stdout, "\b \b")
			}

		case 3: // Ctrl+C
			fmt.Fprintln(os.Stdout)
			return nil, fmt.Errorf("password input cancelled")

		default:
			password = append(password, buf[0])
			fmt.Fprint(os.Stdout, "*")
		}
	}
}
