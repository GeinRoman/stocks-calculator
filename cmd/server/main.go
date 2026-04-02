package main

import (
	"log"
	"stocks_calculator/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
