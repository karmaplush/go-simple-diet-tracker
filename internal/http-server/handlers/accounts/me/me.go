package me

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/karmaplush/simple-diet-tracker/internal/domain/models"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
	"github.com/karmaplush/simple-diet-tracker/internal/services/account"
)

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=AccountProvider
type AccountProvider interface {
	GetAccountByContextJWT(ctx context.Context) (models.Account, error)
}

func New(
	log *slog.Logger,
	accountProvider AccountProvider,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.accounts.me.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		acc, err := accountProvider.GetAccountByContextJWT(r.Context())

		if err != nil {

			if errors.Is(err, account.ErrInvalidJWT) {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.ErrorMessage("incorrect credentials"))
				return
			}

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorMessage("unexpected error"))
			return
		}

		render.JSON(w, r, acc)
	}
}
