package registration

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

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=RegistrationProvider
type RegistrationProvider interface {
	Registration(ctx context.Context, email string, password string) (err error)
}

type Request struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func New(
	log *slog.Logger,
	registration RegistrationProvider,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.accounts.registration.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		render.Status(r, http.StatusBadRequest)

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", slog.String("err", err.Error()))
			render.JSON(w, r, response.ErrorMessage("invalid request"))
			return
		}

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", slog.String("err", err.Error()))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		if err := registration.Registration(r.Context(), req.Email, req.Password); err != nil {
			if errors.Is(err, grpc.ErrUserExists) {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.ErrorMessage("user is already exists"))
				return
			}

			if errors.Is(err, grpc.ErrInvalidArgument) {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.ErrorMessage("invalid credentials"))
				return
			}

			if errors.Is(err, grpc.ErrGRPCUnexpected) {
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response.ErrorMessage("internal error"))
				return
			}

			log.Error("internal error", slog.String("err", err.Error()))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorMessage("internal error"))
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, nil)

	}
}
