package app

import (
	"context"
	"log/slog"
	"os"

	trackerapp "github.com/karmaplush/simple-diet-tracker/internal/app/tracker"
	grpcauthclient "github.com/karmaplush/simple-diet-tracker/internal/clients/auth/grpc"
	"github.com/karmaplush/simple-diet-tracker/internal/config"
	"github.com/karmaplush/simple-diet-tracker/internal/services/account"
	"github.com/karmaplush/simple-diet-tracker/internal/services/auth"
	"github.com/karmaplush/simple-diet-tracker/internal/services/record"
	"github.com/karmaplush/simple-diet-tracker/internal/storage/sqlite"
)

type App struct {
	TrackerApp *trackerapp.App
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {

	sqliteStorage, err := sqlite.New(cfg.StoragePath)

	if err != nil {
		log.Error(
			"error was occured due storage initalization",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	log.Info("storage initialized")

	grpcAuthClient, err := grpcauthclient.New(context.Background(), log, cfg)
	if err != nil {
		log.Error("failed to init grpc auth client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("gRPC Auth client initialized")

	authService := auth.New(log, grpcAuthClient, grpcAuthClient, sqliteStorage, sqliteStorage)
	accountService := account.New(log, sqliteStorage, sqliteStorage)
	recordService := record.New(log, sqliteStorage, sqliteStorage, sqliteStorage, accountService)

	trackerApp := trackerapp.New(
		log,
		cfg,
		authService,
		accountService,
		recordService,
	)

	return &App{
		TrackerApp: trackerApp,
	}

}
