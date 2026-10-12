package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/naphat1243/go-service-kit/internal/config"
	"github.com/naphat1243/go-service-kit/internal/handler"
	apphttp "github.com/naphat1243/go-service-kit/internal/http"
	"github.com/naphat1243/go-service-kit/internal/repository"
	"github.com/naphat1243/go-service-kit/internal/service"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	users := service.NewUserService(repository.NewUserRepository(pool))

	e := apphttp.NewServer(cfg, log,
		handler.NewHealthHandler(pool),
		handler.NewUserHandler(users),
	)

	return apphttp.Start(ctx, e, cfg, log)
}
