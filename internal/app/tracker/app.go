package trackerapp

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/karmaplush/simple-diet-tracker/internal/config"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/login"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/me"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/registration"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/records/create"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/records/delete"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/records/list"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/middlewares/logger"
	"github.com/karmaplush/simple-diet-tracker/internal/services/account"
	"github.com/karmaplush/simple-diet-tracker/internal/services/auth"
	"github.com/karmaplush/simple-diet-tracker/internal/services/record"
)

type App struct {
	Log        *slog.Logger
	HttpServer *http.Server
}

func New(
	log *slog.Logger,
	cfg *config.Config,
	authService *auth.Auth,
	accountService *account.Account,
	recordService *record.Record,
) *App {

	tokenAuth := jwtauth.New("HS256", []byte(cfg.AppSecret), nil)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/accounts/login", login.New(log, authService))
	router.Post("/accounts/registration", registration.New(log, authService))

	// Protected routes
	router.Group(func(router chi.Router) {
		router.Use(jwtauth.Verifier(tokenAuth))
		router.Use(jwtauth.Authenticator)

		router.Get("/accounts/me", me.New(log, accountService))

		router.Get("/records", list.New(log, recordService))
		router.Post("/records", create.New(log, recordService))
		router.Delete("/records/{recordId}", delete.New(log, recordService))
	})

	return &App{
		Log: log,
		HttpServer: &http.Server{
			Addr:         cfg.HttpServer.Address,
			Handler:      router,
			ReadTimeout:  cfg.HttpServer.Timeout,
			WriteTimeout: cfg.HttpServer.Timeout,
			IdleTimeout:  cfg.HttpServer.IdleTimeout,
		},
	}
}
