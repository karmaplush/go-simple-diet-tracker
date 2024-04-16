package main

import (
	"log/slog"
	"os"

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

	log.Info("config parsed, logger initialized")

	application := app.New(log, cfg)

	log.Info("Simple diet tracker app initialized")

	if err := application.TrackerApp.HttpServer.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
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
