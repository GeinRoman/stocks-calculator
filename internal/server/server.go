package server

import (
	"errors"
	"fmt"
	"net/http"
	"stocks_calculator/internal/server/handler"
)

// @title           Stocks Calculator API
// @version         1.0
// @host            localhost:3333
// @BasePath        /
func Run() error {
	handler := handler.New()
	// TODO: load config and pipe it to the func below
	fmt.Println("Starting server")
	err := http.ListenAndServe(":3333", handler)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Print("Server closed")
		return nil
	}

	return fmt.Errorf("Server stopped with an error (%w)", err)
}
