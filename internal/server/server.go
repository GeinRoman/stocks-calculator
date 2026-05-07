package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	log.Println("Reading config...")
	config, err := loadConfig()
	if err != nil {
		return err
	}

	log.Println("Connecting to database...")
	db, err := setupDb(fmt.Sprintf(
		"postgresql://%s:%s@db:5432/stocks_calculator?sslmode=disable",
		config.DBUsername,
		config.DBPassword,
	))
	if err != nil {
		return err
	}
	defer db.Close()
	log.Println("Database connected")

	moex := moex.New(config.MaxConnections)
	tokenManager := tokenmanager.New(
		[]byte(config.JwtKey),
		time.Minute*time.Duration(config.JwtExpirationTimeMin),
	)
	repo := repo.New(db)
	app := app.New(repo, tokenManager, moex)
	server := handler.NewHttpServer(app, tokenManager, config.Port)

	log.Printf("Server listening on :%d\n", config.Port)
	err = runServer(server)
	log.Print("\nExecution stopped")

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
