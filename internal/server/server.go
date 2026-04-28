package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
	server := handler.NewHttpServer(app, tokenManager, config.Port)

	fmt.Println("Starting server")
	fmt.Println("Server running ...")
	err = runServer(server)
	fmt.Print("\nExecution stopped")

	return err
}

func runServer(server *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			return fmt.Errorf("shutdown failed: %v", err)
		}
		return err

	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
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
