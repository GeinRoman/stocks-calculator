package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"stocks_calculator/internal/server/app"
	"stocks_calculator/internal/server/handler"
	"stocks_calculator/internal/server/moex"
	"stocks_calculator/internal/server/repo"
	"stocks_calculator/pkg/tokenmanager"
	"time"

	_ "github.com/lib/pq"
)

const constr = "postgresql://postgres:1234@localhost:5432/stocks_calculator?sslmode=disable"
const secretKey = "dev_secret_key_do_not_use_in_production"

func Run() error {
	db, err := sql.Open("postgres", constr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	moex := moex.New(100)
	tokenManager := tokenmanager.New([]byte(secretKey), time.Minute*10)
	repo := repo.New(db)
	app := app.New(repo, tokenManager, moex)
	handler := handler.New(app, tokenManager)
	// TODO: load config and pipe it to the func below
	fmt.Println("Starting server")
	err = http.ListenAndServe(":3333", handler)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Print("Server closed")
		return nil
	}

	return fmt.Errorf("Server stopped with an error (%w)", err)
}
