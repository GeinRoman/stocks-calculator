package main

import (
	"log"
	"os"
	"stocks_calculator/internal/cli/app"
	"stocks_calculator/internal/cli/commands"

	"golang.org/x/term"
)

func main() {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		log.Fatal("This is a terminal-only utility. Please run it from an interactive terminal.")
	}
	if err := app.ReadConfig(); err != nil {
		log.Fatal(err.Error())
	}
	commands.Execute()
}
