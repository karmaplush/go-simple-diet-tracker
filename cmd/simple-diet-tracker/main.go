package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karmaplush/simple-diet-tracker/internal/app"
	"github.com/karmaplush/simple-diet-tracker/internal/config"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()
	log := SetupLogger(cfg.Env)

	log.Info("Config parsed, logger initialized")

	application := app.New(log, cfg)

	log.Info("Simple diet tracker app initialized")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)

	go func() {
		log.Info("Starting server...")
		if err := application.TrackerApp.HttpServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			errCh <- err
		}
		slog.Info("Server is shutting down...")
	}()

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.TrackerApp.HttpServer.Shutdown(ctx); err != nil {
		slog.Error("Error during server shutdown", slog.String("err", err.Error()))
	}

	slog.Info("Server stopped gracefully")
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	}

	return log
}
