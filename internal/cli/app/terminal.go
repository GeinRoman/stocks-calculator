package app

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
	"unicode"
)

const (
	clearLine     = "\r\033[K"
	clearPrevLine = "\033[A"
)

func confirmation(message string) (bool, error) {
	fmt.Printf("%s Y[es], N[o]?\n", message)
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			return false, err
		}

		text := scanner.Text()
		if len(text) < 1 {
			continue
		}

		char := unicode.ToLower(rune(text[0]))
		switch char {
		case 'y':
			return true, nil
		case 'n':
			return false, nil
		default:
			fmt.Print(clearPrevLine)
		}
	}
}

func askPassword() (string, error) {
	fmt.Print("Please, enter password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Print(clearLine)
	return string(password), err
}
