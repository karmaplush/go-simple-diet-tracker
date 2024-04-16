package create

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
)

type RecordCreator interface {
	CreateRecordForCurrentUser(
		ctx context.Context,
		dateRecord time.Time,
		value int,
	) error
}

type Request struct {
	Value      int       `json:"value"      validate:"required,gte=1"`
	DateRecord time.Time `json:"dateRecord" validate:"required"`
}

func New(
	log *slog.Logger,
	recordCreator RecordCreator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.records.create.New"

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
			log.Info("invalid request", slog.String("err", err.Error()))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		if err := recordCreator.CreateRecordForCurrentUser(r.Context(), req.DateRecord, req.Value); err != nil {
			render.JSON(w, r, response.ErrorMessage("unexpected error"))
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, nil)
	}
}
