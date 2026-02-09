package app

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/term"
)

const (
	clearLine     string = "\r\033[K"
	clearPrevLine string = "\033[A\033[2K"
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

func askNumber(min int, max int, message string) (int, error) {
	fmt.Printf("%s. Min: %d, Max: %d\n", message, min, max)
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			return 0, fmt.Errorf("Fail to read user input. %s", err)
		}

		num, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || num < min || num > max {
			fmt.Print(clearPrevLine)
			continue
		}

		fmt.Print(clearPrevLine)
		fmt.Print(clearPrevLine)
		return num, nil
	}

}

func askWeights(names []string) ([]int, error) {
	l := len(names)
	weights := make([]int, l)

	left := 100
	for i := 0; i < l-1; i++ {
		if left == 0 {
			weights[i] = 0
			continue
		}
		num, err := askNumber(
			0,
			left,
			fmt.Sprintf("Write weight (proportion) for %q", names[i]),
		)
		if err != nil {
			return nil, err
		}
		weights[i] = num
		left -= num
	}

	fmt.Printf("Residual weight for %q is %d.\n", names[l-1], left)
	weights[l-1] = left

	return weights, nil
}
