package login

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/karmaplush/simple-diet-tracker/internal/clients/auth/grpc"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
)

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=LoginProvider
type LoginProvider interface {
	Login(ctx context.Context, email string, password string) (token string, err error)
}

type Request struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	Token string `json:"token"`
}

func New(
	log *slog.Logger,
	loginProvider LoginProvider,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.accounts.login.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			render.Status(r, http.StatusBadRequest)
			log.Error("failed to decode request body", slog.String("err", err.Error()))
			render.JSON(w, r, response.ErrorMessage("invalid request"))
			return
		}

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Info("failed to validate request body", slog.String("err", err.Error()))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		token, err := loginProvider.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, grpc.ErrInvalidArgument) || errors.Is(err, grpc.ErrUserNotFound) {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.ErrorMessage("invalid credentials"))
				return
			}
			if errors.Is(err, grpc.ErrGRPCUnexpected) {
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response.ErrorMessage("internal error"))
				return
			}

			render.Status(r, http.StatusInternalServerError)
			log.Error("internal error", slog.String("err", err.Error()))
			render.JSON(w, r, response.ErrorMessage("internal error"))
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			Token: token,
		})

	}
}
