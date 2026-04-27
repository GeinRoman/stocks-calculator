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

func Run() error {
	fmt.Println("Reading config")
	config, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Println("Setup database")
	db, err := setupDb(fmt.Sprintf(
		"postgresql://%s:%s@%s?sslmode=%s",
		config.DbParams.Username,
		config.DbParams.Password,
		config.DbParams.Address,
		config.DbParams.SslMode,
	))
	if err != nil {
		return err
	}
	defer db.Close()

	moex := moex.New(config.MaxConnections)
	tokenManager := tokenmanager.New(
		[]byte(config.JwtKey),
		time.Minute*time.Duration(config.JwtExpirationTimeMin),
	)
	repo := repo.New(db)
	app := app.New(repo, tokenManager, moex)
	handler := handler.New(app, tokenManager)

	fmt.Println("Starting server")
	fmt.Println("Server running ...")
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Port), handler)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Print("Server closed")
		return nil
	}

	return fmt.Errorf("Server stopped with an error (%w)", err)
}

func setupDb(constr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", constr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
